package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	contractshttp "github.com/goravel/framework/contracts/http"
	contractsroute "github.com/goravel/framework/contracts/route"
	configmocks "github.com/goravel/framework/mocks/config"
	"github.com/stretchr/testify/suite"
)

type WithoutMiddlewareSuite struct {
	suite.Suite
}

func TestWithoutMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(WithoutMiddlewareSuite))
}

// TestRouterWithoutMiddleware tests that Router-level WithoutMiddleware
// excludes middleware for all routes registered within the group.
func (s *WithoutMiddlewareSuite) TestRouterLevelExclusion() {
	mockConfig := setupWithoutMiddlewareConfig(s)
	route := setupRoute(s, mockConfig)
	routes = make(map[string]map[string]contractshttp.Info)

	// Define a shared middleware instance so the function pointer matches
	authMiddleware := func(ctx contractshttp.Context) {
		ctx.WithValue("auth", "applied")
		ctx.Request().Next()
	}
	logMiddleware := func(ctx contractshttp.Context) {
		ctx.WithValue("log", "applied")
		ctx.Request().Next()
	}

	// Group with both middlewares, then exclude auth
	route.Middleware(authMiddleware, logMiddleware).WithoutMiddleware(authMiddleware).Group(func(router contractsroute.Router) {
		router.Get("/excluded", func(ctx contractshttp.Context) contractshttp.Response {
			return ctx.Response().Json(http.StatusOK, contractshttp.Json{
				"auth": ctx.Value("auth"),
				"log":  ctx.Value("log"),
			})
		})
	})

	// Group WITHOUT excluding - both middlewares apply
	route.Middleware(authMiddleware, logMiddleware).Group(func(router contractsroute.Router) {
		router.Get("/included", func(ctx contractshttp.Context) contractshttp.Response {
			return ctx.Response().Json(http.StatusOK, contractshttp.Json{
				"auth": ctx.Value("auth"),
				"log":  ctx.Value("log"),
			})
		})
	})

	// Excluded route: auth should be null (middleware not executed), log present
	s.assertRoute(route, "GET", "/excluded", http.StatusOK, `{"auth":null,"log":"applied"}`)

	// Included route: both present
	s.assertRoute(route, "GET", "/included", http.StatusOK, `{"auth":"applied","log":"applied"}`)
}

// TestRouterWithoutMiddlewareMultiple tests excluding multiple middlewares
func (s *WithoutMiddlewareSuite) TestRouterLevelExclusionMultiple() {
	mockConfig := setupWithoutMiddlewareConfig(s)
	route := setupRoute(s, mockConfig)
	routes = make(map[string]map[string]contractshttp.Info)

	m1 := func(ctx contractshttp.Context) { ctx.WithValue("m1", "yes"); ctx.Request().Next() }
	m2 := func(ctx contractshttp.Context) { ctx.WithValue("m2", "yes"); ctx.Request().Next() }
	m3 := func(ctx contractshttp.Context) { ctx.WithValue("m3", "yes"); ctx.Request().Next() }

	// Exclude m1 and m3, keep m2
	route.Middleware(m1, m2, m3).WithoutMiddleware(m1, m3).Get("/multi", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Json(http.StatusOK, contractshttp.Json{
			"m1": ctx.Value("m1"),
			"m2": ctx.Value("m2"),
			"m3": ctx.Value("m3"),
		})
	})

	// Only m2 should be applied (m1 and m3 excluded, shown as null)
	s.assertRoute(route, "GET", "/multi", http.StatusOK, `{"m1":null,"m2":"yes","m3":null}`)
}

// TestRouterWithoutMiddlewareNested tests nested groups inherit exclusions
func (s *WithoutMiddlewareSuite) TestRouterLevelExclusionNested() {
	mockConfig := setupWithoutMiddlewareConfig(s)
	route := setupRoute(s, mockConfig)
	routes = make(map[string]map[string]contractshttp.Info)

	authMiddleware := func(ctx contractshttp.Context) {
		ctx.WithValue("auth", "applied")
		ctx.Request().Next()
	}
	extraMiddleware := func(ctx contractshttp.Context) {
		ctx.WithValue("extra", "applied")
		ctx.Request().Next()
	}

	// Outer group excludes auth, inner group adds extra
	route.Middleware(authMiddleware).WithoutMiddleware(authMiddleware).Group(func(outer contractsroute.Router) {
		outer.Get("/outer", func(ctx contractshttp.Context) contractshttp.Response {
			return ctx.Response().Json(http.StatusOK, contractshttp.Json{"auth": ctx.Value("auth")})
		})

		outer.Middleware(extraMiddleware).Group(func(inner contractsroute.Router) {
			inner.Get("/inner", func(ctx contractshttp.Context) contractshttp.Response {
				return ctx.Response().Json(http.StatusOK, contractshttp.Json{
					"auth":  ctx.Value("auth"),
					"extra": ctx.Value("extra"),
				})
			})
		})
	})

	// Outer route: auth excluded (null value, middleware not executed)
	s.assertRoute(route, "GET", "/outer", http.StatusOK, `{"auth":null}`)
	// Inner route: auth excluded, extra applied (exclusion inherited)
	s.assertRoute(route, "GET", "/inner", http.StatusOK, `{"auth":null,"extra":"applied"}`)
}

// TestRouterWithoutMiddlewareAbort tests that excluded aborting middleware
// no longer blocks the request.
func (s *WithoutMiddlewareSuite) TestRouterLevelExclusionAbort() {
	mockConfig := setupWithoutMiddlewareConfig(s)
	route := setupRoute(s, mockConfig)
	routes = make(map[string]map[string]contractshttp.Info)

	// A middleware that blocks the request
	blockMiddleware := func(ctx contractshttp.Context) {
		ctx.Request().Abort(http.StatusForbidden)
	}
	passMiddleware := func(ctx contractshttp.Context) {
		ctx.WithValue("pass", "yes")
		ctx.Request().Next()
	}

	// Exclude the blocking middleware
	route.Middleware(blockMiddleware, passMiddleware).WithoutMiddleware(blockMiddleware).Get("/open", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Json(http.StatusOK, contractshttp.Json{"pass": ctx.Value("pass")})
	})

	// Request should succeed because blockMiddleware is excluded
	s.assertRoute(route, "GET", "/open", http.StatusOK, `{"pass":"yes"}`)

	// Now WITHOUT exclusion - request should be blocked
	route.Middleware(blockMiddleware, passMiddleware).Get("/blocked", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Json(http.StatusOK, contractshttp.Json{"pass": ctx.Value("pass")})
	})

	s.assertRoute(route, "GET", "/blocked", http.StatusForbidden, "")
}

func (s *WithoutMiddlewareSuite) assertRoute(route *Route, method, url string, expectCode int, expectBody string) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest(method, url, nil)
	s.NoError(err)
	route.ServeHTTP(w, req)
	if expectBody != "" {
		s.Equal(expectBody, w.Body.String())
	}
	s.Equal(expectCode, w.Code)
}

func setupWithoutMiddlewareConfig(s *WithoutMiddlewareSuite) *configmocks.Config {
	mockConfig := configmocks.NewConfig(s.T())
	mockConfig.EXPECT().GetInt("http.drivers.gin.body_limit", 4096).Return(4096).Maybe()
	mockConfig.EXPECT().GetBool("app.debug").Return(true).Maybe()
	mockConfig.EXPECT().Get("http.drivers.gin.template").Return(nil).Maybe()
	ConfigFacade = mockConfig
	return mockConfig
}

func setupRoute(s *WithoutMiddlewareSuite, mockConfig *configmocks.Config) *Route {
	route := &Route{
		config: mockConfig,
		driver: "gin",
	}
	err := route.init(nil)
	s.Require().Nil(err)
	return route
}
