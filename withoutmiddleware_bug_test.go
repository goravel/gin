package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/stretchr/testify/assert"
)

// NOTE: Concurrent access to routes map is not supported by the framework
// Routes should be registered during application initialization (single goroutine)
// This is a pre-existing architectural limitation, not introduced by WithoutMiddleware

// TestWithoutMiddleware_NilMiddleware tests behavior with nil middleware
func TestWithoutMiddleware_NilMiddleware(t *testing.T) {
	// Clear routes map before test
	routes = make(map[string]map[string]contractshttp.Info)

	// This should not panic
	action := NewAction("GET", "/test", "handler.Test")
	assert.NotPanics(t, func() {
		action.WithoutMiddleware(nil)
	})

	routeInfo, exists := routes["/test"]["GET|HEAD"]
	assert.True(t, exists)
	// Nil middleware should still be added to the list (Go allows nil in slices)
	assert.True(t, len(routeInfo.ExcludedMiddleware) > 0)
}

// TestWithoutMiddleware_EmptyExcludedList tests route with no exclusions
func TestWithoutMiddleware_EmptyExcludedList(t *testing.T) {
	// Clear routes map before test
	routes = make(map[string]map[string]contractshttp.Info)

	var executedMiddlewares []string
	middleware := func(ctx contractshttp.Context) {
		executedMiddlewares = append(executedMiddlewares, "middleware")
		ctx.Request().Next()
	}

	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.GET("/test", func(c *gin.Context) {
		context := NewContext(c)
		defer func() {
			contextRequestPool.Put(context.request)
			contextResponsePool.Put(context.response)
			context.request = nil
			context.response = nil
			contextPool.Put(context)
		}()

		routeInfo := context.Request().Info()
		middlewares := []contractshttp.Middleware{middleware}

		for _, mw := range middlewares {
			excluded := false
			// Check if ExcludedMiddleware is nil
			if routeInfo.ExcludedMiddleware != nil {
				for _, excludedMw := range routeInfo.ExcludedMiddleware {
					if getFunctionPointer(excludedMw) == getFunctionPointer(mw) {
						excluded = true
						break
					}
				}
			}

			if !excluded {
				mw(context)
			}
		}

		c.JSON(200, gin.H{"ok": true})
	})

	// Create action WITHOUT WithoutMiddleware - ExcludedMiddleware should be nil/empty
	NewAction("GET", "/test", "handler.Test")

	executedMiddlewares = []string{}
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Middleware should run because nothing is excluded
	assert.Equal(t, []string{"middleware"}, executedMiddlewares)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestWithoutMiddleware_SameMiddlewareTwice tests excluding same middleware multiple times
func TestWithoutMiddleware_SameMiddlewareTwice(t *testing.T) {
	// Clear routes map before test
	routes = make(map[string]map[string]contractshttp.Info)

	middleware := func(ctx contractshttp.Context) {
		ctx.Request().Next()
	}

	action := NewAction("GET", "/test", "handler.Test")
	// Call WithoutMiddleware twice with same middleware
	action.WithoutMiddleware(middleware).WithoutMiddleware(middleware)

	routeInfo, exists := routes["/test"]["GET|HEAD"]
	assert.True(t, exists)
	// Should appear twice in the list (no deduplication)
	assert.Equal(t, 2, len(routeInfo.ExcludedMiddleware))
}

// TestWithoutMiddleware_DifferentFunctionPointers tests that different functions are not mixed up
func TestWithoutMiddleware_DifferentFunctionPointers(t *testing.T) {
	// Clear routes map before test
	routes = make(map[string]map[string]contractshttp.Info)

	var executedMiddlewares []string

	middleware1 := func(ctx contractshttp.Context) {
		executedMiddlewares = append(executedMiddlewares, "m1")
		ctx.Request().Next()
	}

	middleware2 := func(ctx contractshttp.Context) {
		executedMiddlewares = append(executedMiddlewares, "m2")
		ctx.Request().Next()
	}

	// Similar but different function
	middleware3 := func(ctx contractshttp.Context) {
		executedMiddlewares = append(executedMiddlewares, "m3")
		ctx.Request().Next()
	}

	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.GET("/test", func(c *gin.Context) {
		context := NewContext(c)
		defer func() {
			contextRequestPool.Put(context.request)
			contextResponsePool.Put(context.response)
			context.request = nil
			context.response = nil
			contextPool.Put(context)
		}()

		routeInfo := context.Request().Info()
		middlewares := []contractshttp.Middleware{middleware1, middleware2, middleware3}

		for _, mw := range middlewares {
			excluded := false
			if routeInfo.ExcludedMiddleware != nil {
				for _, excludedMw := range routeInfo.ExcludedMiddleware {
					if getFunctionPointer(excludedMw) == getFunctionPointer(mw) {
						excluded = true
						break
					}
				}
			}

			if !excluded {
				mw(context)
			}
		}

		c.JSON(200, gin.H{"ok": true})
	})

	// Exclude only middleware2
	NewAction("GET", "/test", "handler.Test").
		WithoutMiddleware(middleware2)

	executedMiddlewares = []string{}
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Only m1 and m3 should run, m2 should be excluded
	assert.Equal(t, []string{"m1", "m3"}, executedMiddlewares)
}

// TestWithoutMiddleware_RouteNotFound tests behavior when route info is not found
func TestWithoutMiddleware_RouteNotFound(t *testing.T) {
	// Clear routes map before test
	routes = make(map[string]map[string]contractshttp.Info)

	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.GET("/unknown", func(c *gin.Context) {
		context := NewContext(c)
		defer func() {
			contextRequestPool.Put(context.request)
			contextResponsePool.Put(context.response)
			context.request = nil
			context.response = nil
			contextPool.Put(context)
		}()

		// This route is not in the routes map
		routeInfo := context.Request().Info()
		assert.Equal(t, contractshttp.Info{}, routeInfo, "Should return empty Info")
		assert.Nil(t, routeInfo.ExcludedMiddleware, "ExcludedMiddleware should be nil")
		assert.Empty(t, routeInfo.Path, "Path should be empty")

		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/unknown", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestWithoutMiddleware_MultipleRoutesSamePath tests different methods on same path
func TestWithoutMiddleware_MultipleRoutesSamePath(t *testing.T) {
	// Clear routes map before test
	routes = make(map[string]map[string]contractshttp.Info)

	middleware := func(ctx contractshttp.Context) {
		ctx.Request().Next()
	}

	// Create GET and POST routes with same path
	NewAction("GET", "/resource", "handler.Get").WithoutMiddleware(middleware)
	NewAction("POST", "/resource", "handler.Post")
	// POST does NOT exclude middleware

	getInfo := routes["/resource"]["GET|HEAD"]
	postInfo := routes["/resource"]["POST"]

	assert.Equal(t, 1, len(getInfo.ExcludedMiddleware), "GET should have 1 excluded middleware")
	assert.Empty(t, postInfo.ExcludedMiddleware, "POST should have no excluded middleware")
}

// TestWithoutMiddleware_ChainedCalls tests method chaining behavior
func TestWithoutMiddleware_ChainedCalls(t *testing.T) {
	// Clear routes map before test
	routes = make(map[string]map[string]contractshttp.Info)

	middleware1 := func(ctx contractshttp.Context) { ctx.Request().Next() }
	middleware2 := func(ctx contractshttp.Context) { ctx.Request().Next() }

	// Chain WithoutMiddleware calls
	action := NewAction("GET", "/test", "handler.Test")
	action.WithoutMiddleware(middleware1).
		WithoutMiddleware(middleware2).
		Name("test-route")

	routeInfo := routes["/test"]["GET|HEAD"]
	assert.Equal(t, "test-route", routeInfo.Name)
	assert.Equal(t, 2, len(routeInfo.ExcludedMiddleware))
}

// TestWithoutMiddleware_ActionReturn tests that Action returns correctly
func TestWithoutMiddleware_ActionReturn(t *testing.T) {
	// Clear routes map before test
	routes = make(map[string]map[string]contractshttp.Info)

	middleware := func(ctx contractshttp.Context) { ctx.Request().Next() }

	action := NewAction("GET", "/test", "handler.Test")

	// Should return Action interface
	result := action.WithoutMiddleware(middleware)
	assert.NotNil(t, result)

	// Should be able to chain Name
	result = result.Name("test")
	assert.NotNil(t, result)

	routeInfo := routes["/test"]["GET|HEAD"]
	assert.Equal(t, "test", routeInfo.Name)
}
