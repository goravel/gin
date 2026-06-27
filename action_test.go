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
	routes = make(map[string]map[string]contractshttp.Info)

	action := NewAction("GET", "/test-path", "test.Action")
	assert.NotNil(t, action)
	assert.IsType(t, &Action{}, action)

	routeInfo, exists := routes["/test-path"]["GET|HEAD"]
	assert.True(t, exists)
	assert.Equal(t, "GET|HEAD", routeInfo.Method)
	assert.Equal(t, "/test-path", routeInfo.Path)
	assert.Equal(t, "test.Action", routeInfo.Handler)
	assert.Empty(t, routeInfo.Name)
}

func TestAction_Name(t *testing.T) {
	routes = make(map[string]map[string]contractshttp.Info)

	action := NewAction("GET", "/named-path", "")

	namedAction := action.Name("test-route")
	assert.NotNil(t, namedAction)
	assert.IsType(t, &Action{}, namedAction)

	routeInfo, exists := routes["/named-path"]["GET|HEAD"]
	assert.True(t, exists)
	assert.Equal(t, "GET|HEAD", routeInfo.Method)
	assert.Equal(t, "/named-path", routeInfo.Path)
	assert.Equal(t, "test-route", routeInfo.Name)

	chainedAction := action.Name("new-name").Name("final-name")
	assert.NotNil(t, chainedAction)

	routeInfo, exists = routes["/named-path"]["GET|HEAD"]
	assert.True(t, exists)
	assert.Equal(t, "final-name", routeInfo.Name)
}

type actionAuthMiddleware struct{}

func (m *actionAuthMiddleware) Signature() string { return "actionAuth" }
func (m *actionAuthMiddleware) Handle(ctx contractshttp.Context) {
	ctx.Response().Json(http.StatusOK, map[string]string{"middleware": "auth"})
	ctx.Request().Next()
}

type actionThrottleMiddleware struct{}

func (m *actionThrottleMiddleware) Signature() string { return "actionThrottle" }
func (m *actionThrottleMiddleware) Handle(ctx contractshttp.Context) {
	ctx.Response().Json(http.StatusTooManyRequests, map[string]string{"error": "throttled"})
	ctx.Request().Abort()
}

func TestAction_WithoutMiddleware(t *testing.T) {
	routes = make(map[string]map[string]contractshttp.Info)

	authMiddleware := &actionAuthMiddleware{}
	throttleMiddleware := &actionThrottleMiddleware{}

	action := NewAction("GET", "/test-path", "test.Action")

	actionWithoutAuth := action.WithoutMiddleware(authMiddleware)
	assert.NotNil(t, actionWithoutAuth)
	assert.IsType(t, &Action{}, actionWithoutAuth)

	routeInfo, exists := routes["/test-path"]["GET|HEAD"]
	assert.True(t, exists)
	assert.Len(t, routeInfo.ExcludedMiddleware, 1)

	routes = make(map[string]map[string]contractshttp.Info)
	action2 := NewAction("GET", "/test-path2", "test.Action")
	action2.WithoutMiddleware(authMiddleware, throttleMiddleware)

	routeInfo2, exists := routes["/test-path2"]["GET|HEAD"]
	assert.True(t, exists)
	assert.Len(t, routeInfo2.ExcludedMiddleware, 2)

	routes = make(map[string]map[string]contractshttp.Info)
	action3 := NewAction("GET", "/test-path3", "test.Action")
	chainedAction := action3.WithoutMiddleware(authMiddleware).Name("test-route")
	assert.NotNil(t, chainedAction)

	routeInfo3, exists := routes["/test-path3"]["GET|HEAD"]
	assert.True(t, exists)
	assert.Equal(t, "test-route", routeInfo3.Name)
	assert.Len(t, routeInfo3.ExcludedMiddleware, 1)
}

type actionIntegrationMw struct{}

func (m *actionIntegrationMw) Signature() string { return "actionIntegration" }
func (m *actionIntegrationMw) Handle(ctx contractshttp.Context) {
	ctx.Request().Next()
}

func TestAction_WithoutMiddleware_Integration(t *testing.T) {
	routes = make(map[string]map[string]contractshttp.Info)

	throttleMiddleware := &actionIntegrationMw{}

	NewAction("GET", "/protected", "handler.Protected").
		WithoutMiddleware(throttleMiddleware)

	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.GET("/protected", func(c *gin.Context) {
		routeInfo, exists := routes["/protected"]["GET|HEAD"]
		if exists && routeInfo.ExcludedMiddleware != nil {
			for _, excluded := range routeInfo.ExcludedMiddleware {
				if isSameMiddleware(excluded, throttleMiddleware) {
					c.Next()
					return
				}
			}
		}
		c.Next()
	}, func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	routeInfo, exists := routes["/protected"]["GET|HEAD"]
	assert.True(t, exists)
	assert.Len(t, routeInfo.ExcludedMiddleware, 1)
}
