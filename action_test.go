package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/stretchr/testify/assert"
)

func TestNewAction(t *testing.T) {
	// Clear routes map before test
	routes = make(map[string]map[string]contractshttp.Info)

	// Test creating a new action
	action := NewAction("GET", "/test-path", "test.Action")
	assert.NotNil(t, action)
	assert.IsType(t, &Action{}, action)

	// Verify route was added to routes map
	routeInfo, exists := routes["/test-path"]["GET|HEAD"]
	assert.True(t, exists)
	assert.Equal(t, "GET|HEAD", routeInfo.Method)
	assert.Equal(t, "/test-path", routeInfo.Path)
	assert.Equal(t, "test.Action", routeInfo.Handler)
	assert.Empty(t, routeInfo.Name)
}

func TestAction_Name(t *testing.T) {
	// Clear routes map before test
	routes = make(map[string]map[string]contractshttp.Info)

	// Create a new action
	action := NewAction("GET", "/named-path", "")

	// Test setting name
	namedAction := action.Name("test-route")
	assert.NotNil(t, namedAction)
	assert.IsType(t, &Action{}, namedAction)

	// Verify route info was updated
	routeInfo, exists := routes["/named-path"]["GET|HEAD"]
	assert.True(t, exists)
	assert.Equal(t, "GET|HEAD", routeInfo.Method)
	assert.Equal(t, "/named-path", routeInfo.Path)
	assert.Equal(t, "test-route", routeInfo.Name)

	// Test method chaining
	chainedAction := action.Name("new-name").Name("final-name")
	assert.NotNil(t, chainedAction)

	// Verify final route info
	routeInfo, exists = routes["/named-path"]["GET|HEAD"]
	assert.True(t, exists)
	assert.Equal(t, "final-name", routeInfo.Name)
}

func TestAction_WithoutMiddleware(t *testing.T) {
	// Clear routes map before test
	routes = make(map[string]map[string]contractshttp.Info)

	// Create test middlewares
	authMiddleware := func(ctx contractshttp.Context) {
		ctx.Response().Json(http.StatusOK, map[string]string{"middleware": "auth"})
		ctx.Request().Next()
	}
	throttleMiddleware := func(ctx contractshttp.Context) {
		ctx.Response().Json(http.StatusTooManyRequests, map[string]string{"error": "throttled"})
		ctx.Request().Abort()
	}

	// Create a new action
	action := NewAction("GET", "/test-path", "test.Action")

	// Test excluding middleware
	actionWithoutAuth := action.WithoutMiddleware(authMiddleware)
	assert.NotNil(t, actionWithoutAuth)
	assert.IsType(t, &Action{}, actionWithoutAuth)

	// Verify route info was updated
	routeInfo, exists := routes["/test-path"]["GET|HEAD"]
	assert.True(t, exists)
	assert.Len(t, routeInfo.ExcludedMiddleware, 1)

	// Test excluding multiple middlewares at once
	routes = make(map[string]map[string]contractshttp.Info) // Clear for fresh test
	action2 := NewAction("GET", "/test-path2", "test.Action")
	action2.WithoutMiddleware(authMiddleware, throttleMiddleware)

	routeInfo2, exists := routes["/test-path2"]["GET|HEAD"]
	assert.True(t, exists)
	assert.Len(t, routeInfo2.ExcludedMiddleware, 2)

	// Test method chaining
	routes = make(map[string]map[string]contractshttp.Info) // Clear for fresh test
	action3 := NewAction("GET", "/test-path3", "test.Action")
	chainedAction := action3.WithoutMiddleware(authMiddleware).Name("test-route")
	assert.NotNil(t, chainedAction)

	// Verify final route info
	routeInfo3, exists := routes["/test-path3"]["GET|HEAD"]
	assert.True(t, exists)
	assert.Equal(t, "test-route", routeInfo3.Name)
	assert.Len(t, routeInfo3.ExcludedMiddleware, 1)
}

func TestAction_WithoutMiddleware_Integration(t *testing.T) {
	// Clear routes map before test
	routes = make(map[string]map[string]contractshttp.Info)

	// Create test middlewares
	throttleMiddleware := func(ctx contractshttp.Context) {
		ctx.Request().Next()
	}

	// Create a route with middlewares
	NewAction("GET", "/protected", "handler.Protected").
		WithoutMiddleware(throttleMiddleware)

	// Create a test gin engine
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	// Register the route with auth middleware (simulating goravel's middleware application)
	engine.GET("/protected", func(c *gin.Context) {
		// Simulate the middleware filtering logic
		routeInfo, exists := routes["/protected"]["GET|HEAD"]
		if exists && routeInfo.ExcludedMiddleware != nil {
			// Check if throttleMiddleware is excluded
			for _, excluded := range routeInfo.ExcludedMiddleware {
				// Compare by function pointer
				if getFunctionPointer(excluded) == getFunctionPointer(throttleMiddleware) {
					// Skip throttleMiddleware
					c.Next()
					return
				}
			}
		}
		// Execute throttleMiddleware
		c.Next()
	}, func(c *gin.Context) {
		// Handler
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Make a request
	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Verify the route info has the excluded middleware
	routeInfo, exists := routes["/protected"]["GET|HEAD"]
	assert.True(t, exists)
	assert.Len(t, routeInfo.ExcludedMiddleware, 1)
}
