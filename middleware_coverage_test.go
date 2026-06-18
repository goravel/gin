package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/stretchr/testify/assert"
)

// TestMiddlewareToGinHandler_Filtering tests the filtering logic in middlewareToGinHandler
func TestMiddlewareToGinHandler_Filtering(t *testing.T) {
	// Clear routes map before test
	routes = make(map[string]map[string]contractshttp.Info)

	var executionOrder []string

	// Create different middlewares
	authMiddleware := func(ctx contractshttp.Context) {
		executionOrder = append(executionOrder, "auth")
		ctx.Request().Next()
	}

	throttleMiddleware := func(ctx contractshttp.Context) {
		executionOrder = append(executionOrder, "throttle")
		ctx.Request().Next()
	}

	corsMiddleware := func(ctx contractshttp.Context) {
		executionOrder = append(executionOrder, "cors")
		ctx.Request().Next()
	}

	gin.SetMode(gin.TestMode)
	engine := gin.New()

	// Setup route with all middlewares through the normal gin middleware system
	excludedList := []contractshttp.Middleware{throttleMiddleware}

	engine.GET("/test", func(c *gin.Context) {
		// Simulate what happens in middlewareToGinHandler
		context := NewContext(c)
		defer func() {
			// This ensures the defer cleanup code is executed
			contextRequestPool.Put(context.request)
			contextResponsePool.Put(context.response)
			context.request = nil
			context.response = nil
			contextPool.Put(context)
		}()

		// Create route info for this test
		routes["/test"] = map[string]contractshttp.Info{
			"GET": {
				Handler:            "handler.Test",
				Method:             "GET",
				Path:               "/test",
				ExcludedMiddleware: excludedList,
			},
		}

		// Test auth middleware (not excluded)
		middlewareToGinHandler(authMiddleware)(c)

		// Test throttle middleware (excluded)
		middlewareToGinHandler(throttleMiddleware)(c)

		// Test cors middleware (not excluded)
		middlewareToGinHandler(corsMiddleware)(c)

		c.JSON(200, gin.H{"ok": true})
	})

	executionOrder = []string{}
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Verify execution order - throttle should be skipped
	// Note: The actual filtering happens inside middlewareToGinHandler
	assert.Contains(t, executionOrder, "auth")
	assert.NotContains(t, executionOrder, "throttle")
	assert.Contains(t, executionOrder, "cors")
}

// TestMiddlewareToGinHandler_EmptyExcludedList tests with empty excluded middleware list
func TestMiddlewareToGinHandler_EmptyExcludedList(t *testing.T) {
	// Clear routes map before test
	routes = make(map[string]map[string]contractshttp.Info)

	executed := false
	middleware := func(ctx contractshttp.Context) {
		executed = true
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

		// Route with empty excluded list
		routes["/test"] = map[string]contractshttp.Info{
			"GET": {
				Handler:            "handler.Test",
				Method:             "GET",
				Path:               "/test",
				ExcludedMiddleware: []contractshttp.Middleware{},
			},
		}

		middlewareToGinHandler(middleware)(c)
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.True(t, executed, "Middleware should execute when excluded list is empty")
}

// TestMiddlewareToGinHandler_ContextPool tests the context pool cleanup
func TestMiddlewareToGinHandler_ContextPool(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	var called bool
	middleware := func(ctx contractshttp.Context) {
		called = true
		ctx.Request().Next()
	}

	engine.GET("/test", func(c *gin.Context) {
		// This ensures the defer is executed and pools are used
		middlewareToGinHandler(middleware)(c)
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestGetFunctionPointer tests the helper function
func TestGetFunctionPointer(t *testing.T) {
	middleware1 := func(ctx contractshttp.Context) {
		ctx.Request().Next()
	}

	middleware2 := func(ctx contractshttp.Context) {
		ctx.Request().Next()
	}

	ptr1 := getFunctionPointer(middleware1)
	ptr2 := getFunctionPointer(middleware2)
	ptr1Again := getFunctionPointer(middleware1)

	// Same function should have same pointer
	assert.Equal(t, ptr1, ptr1Again)

	// Different functions should have different pointers
	assert.NotEqual(t, ptr1, ptr2)

	// Pointer should not be zero
	assert.NotZero(t, ptr1)
}
