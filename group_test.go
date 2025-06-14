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

type GroupTestSuite struct {
	suite.Suite
	mockConfig *configmocks.Config
	route      *Route
}

func TestGroupTestSuite(t *testing.T) {
	suite.Run(t, new(GroupTestSuite))
}

func (s *GroupTestSuite) SetupTest() {
	routes = make(map[string]map[string]contractsroute.Info)

	s.mockConfig = configmocks.NewConfig(s.T())
	s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	s.mockConfig.EXPECT().GetInt("http.drivers.gin.body_limit", 4096).Return(4096).Once()
	ConfigFacade = s.mockConfig

	route, err := NewRoute(s.mockConfig, nil)
	s.NoError(err)

	s.route = route
}

func (s *GroupTestSuite) TestGet() {
	s.route.Get("/input/{id}", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Json(http.StatusOK, contractshttp.Json{
			"id": ctx.Request().Input("id"),
		})
	}).Name("get")

	s.assert("GET", "/input/1", http.StatusOK, "{\"id\":\"1\"}")
	s.assert("HEAD", "/input/1", http.StatusOK, "{\"id\":\"1\"}")
	s.Equal(contractsroute.Info{
		Method: MethodGet,
		Path:   "/input/{id}",
		Name:   "get",
	}, s.route.Info("get"))
}

func (s *GroupTestSuite) TestPost() {
	s.route.Post("/input/{id}", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Json(http.StatusOK, contractshttp.Json{
			"id": ctx.Request().Input("id"),
		})
	}).Name("post")

	s.assert("POST", "/input/1", http.StatusOK, "{\"id\":\"1\"}")
	s.Equal(contractsroute.Info{
		Method: MethodPost,
		Path:   "/input/{id}",
		Name:   "post",
	}, s.route.Info("post"))
}

func (s *GroupTestSuite) TestPut() {
	s.route.Put("/input/{id}", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Json(http.StatusOK, contractshttp.Json{
			"id": ctx.Request().Input("id"),
		})
	}).Name("put")

	s.assert("PUT", "/input/1", http.StatusOK, "{\"id\":\"1\"}")
	s.Equal(contractsroute.Info{
		Method: MethodPut,
		Path:   "/input/{id}",
		Name:   "put",
	}, s.route.Info("put"))
}

func (s *GroupTestSuite) TestDelete() {
	s.route.Delete("/input/{id}", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Json(http.StatusOK, contractshttp.Json{
			"id": ctx.Request().Input("id"),
		})
	}).Name("delete")

	s.assert("DELETE", "/input/1", http.StatusOK, "{\"id\":\"1\"}")
	s.Equal(contractsroute.Info{
		Method: MethodDelete,
		Path:   "/input/{id}",
		Name:   "delete",
	}, s.route.Info("delete"))
}

func (s *GroupTestSuite) TestOptions() {
	s.route.Options("/input/{id}", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Json(http.StatusOK, contractshttp.Json{
			"id": ctx.Request().Input("id"),
		})
	}).Name("options")

	s.assert("OPTIONS", "/input/1", http.StatusOK, "{\"id\":\"1\"}")
	s.Equal(contractsroute.Info{
		Method: MethodOptions,
		Path:   "/input/{id}",
		Name:   "options",
	}, s.route.Info("options"))
}

func (s *GroupTestSuite) TestPatch() {
	s.route.Patch("/input/{id}", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Json(http.StatusOK, contractshttp.Json{
			"id": ctx.Request().Input("id"),
		})
	}).Name("patch")

	s.assert("PATCH", "/input/1", http.StatusOK, "{\"id\":\"1\"}")
	s.Equal(contractsroute.Info{
		Method: MethodPatch,
		Path:   "/input/{id}",
		Name:   "patch",
	}, s.route.Info("patch"))
}

func (s *GroupTestSuite) TestAny() {
	s.route.Any("/input/{id}", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Json(http.StatusOK, contractshttp.Json{
			"id": ctx.Request().Input("id"),
		})
	}).Name("any")

	path := "/input/1"
	body := "{\"id\":\"1\"}"

	s.assert("GET", path, http.StatusOK, body)
	s.assert("POST", path, http.StatusOK, body)
	s.assert("PUT", path, http.StatusOK, body)
	s.assert("DELETE", path, http.StatusOK, body)
	s.assert("PATCH", path, http.StatusOK, body)
	s.assert("OPTIONS", path, http.StatusOK, body)

	s.Equal(contractsroute.Info{
		Method: MethodAny,
		Path:   "/input/{id}",
		Name:   "any",
	}, s.route.Info("any"))
}

func (s *GroupTestSuite) TestResource() {
	s.route.setMiddlewares([]contractshttp.Middleware{func(ctx contractshttp.Context) {
		ctx.WithValue("action", ctx.Request().Origin().Method)
		ctx.Request().Next()
	}})
	s.route.Resource("/resource", resourceController{}).Name("resource")

	s.assert("GET", "/resource", http.StatusOK, "{\"action\":\"GET\"}")
	s.assert("GET", "/resource/1", http.StatusOK, "{\"action\":\"GET\",\"id\":\"1\"}")
	s.assert("POST", "/resource", http.StatusOK, "{\"action\":\"POST\"}")
	s.assert("PUT", "/resource/1", http.StatusOK, "{\"action\":\"PUT\",\"id\":\"1\"}")
	s.assert("PATCH", "/resource/1", http.StatusOK, "{\"action\":\"PATCH\",\"id\":\"1\"}")
	s.assert("DELETE", "/resource/1", http.StatusOK, "{\"action\":\"DELETE\",\"id\":\"1\"}")

	s.Equal(contractsroute.Info{
		Method: MethodResource,
		Path:   "/resource",
		Name:   "resource",
	}, s.route.Info("resource"))
}

func (s *GroupTestSuite) TestStatic() {
	s.route.Static("static", "./").Name("static")

	s.assert("GET", "/static/README.md", http.StatusOK, "")

	s.Equal(contractsroute.Info{
		Method: MethodStatic,
		Path:   "/static",
		Name:   "static",
	}, s.route.Info("static"))
}

func (s *GroupTestSuite) TestStaticFile() {
	s.route.StaticFile("static-file", "./README.md").Name("static-file")

	s.assert("GET", "/static-file", http.StatusOK, "")

	s.Equal(contractsroute.Info{
		Method: MethodStaticFile,
		Path:   "/static-file",
		Name:   "static-file",
	}, s.route.Info("static-file"))
}

func (s *GroupTestSuite) TestStaticFS() {
	s.route.StaticFS("/static-fs", http.Dir("./")).Name("static-fs")

	s.assert("GET", "/static-fs", http.StatusMovedPermanently, "")

	s.Equal(contractsroute.Info{
		Method: MethodStaticFS,
		Path:   "/static-fs",
		Name:   "static-fs",
	}, s.route.Info("static-fs"))
}

func (s *GroupTestSuite) TestAbortMiddleware() {
	s.route.Middleware(abortMiddleware()).Get("/middleware/{id}", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Success().Json(contractshttp.Json{
			"id": ctx.Request().Input("id"),
		})
	})

	s.assert("GET", "/middleware/1", http.StatusNonAuthoritativeInfo, "")
}

func (s *GroupTestSuite) TestMultipleMiddleware() {
	s.route.Middleware(contextMiddleware(), contextMiddleware1()).Get("/middlewares/{id}", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Success().Json(contractshttp.Json{
			"id":   ctx.Request().Input("id"),
			"ctx":  ctx.Value("ctx"),
			"ctx1": ctx.Value("ctx1"),
		})
	})

	s.assert("GET", "/middlewares/1", http.StatusOK, "{\"ctx\":\"Goravel\",\"ctx1\":\"Hello\",\"id\":\"1\"}")
}

func (s *GroupTestSuite) TestMultiplePrefix() {
	s.route.Prefix("prefix1").Prefix("prefix2").Get("input/{id}", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Success().Json(contractshttp.Json{
			"id": ctx.Request().Input("id"),
		})
	})

	s.assert("GET", "/prefix1/prefix2/input/1", http.StatusOK, "{\"id\":\"1\"}")
}

func (s *GroupTestSuite) TestMultiplePrefixGroupMiddleware() {
	s.route.Prefix("group1").Middleware(contextMiddleware()).Group(func(route1 contractsroute.Router) {
		route1.Prefix("group2").Middleware(contextMiddleware1()).Group(func(route2 contractsroute.Router) {
			route2.Get("/middleware/{id}", func(ctx contractshttp.Context) contractshttp.Response {
				return ctx.Response().Success().Json(contractshttp.Json{
					"id":   ctx.Request().Input("id"),
					"ctx":  ctx.Value("ctx"),
					"ctx1": ctx.Value("ctx1"),
				})
			})
		})
		route1.Middleware(contextMiddleware2()).Get("/middleware/{id}", func(ctx contractshttp.Context) contractshttp.Response {
			return ctx.Response().Success().Json(contractshttp.Json{
				"id":   ctx.Request().Input("id"),
				"ctx":  ctx.Value("ctx"),
				"ctx2": ctx.Value("ctx2"),
			})
		})
	})

	s.assert("GET", "/group1/group2/middleware/1", http.StatusOK, "{\"ctx\":\"Goravel\",\"ctx1\":\"Hello\",\"id\":\"1\"}")
	s.assert("GET", "/group1/middleware/1", http.StatusOK, "{\"ctx\":\"Goravel\",\"ctx2\":\"World\",\"id\":\"1\"}")
}

func (s *GroupTestSuite) TestGlobalMiddleware() {
	s.mockConfig.On("Get", "cors.paths").Return([]string{}).Once()
	s.mockConfig.On("GetString", "http.tls.host").Return("").Once()
	s.mockConfig.On("GetString", "http.tls.port").Return("").Once()
	s.mockConfig.On("GetString", "http.tls.ssl.cert").Return("").Once()
	s.mockConfig.On("GetString", "http.tls.ssl.key").Return("").Once()
	s.mockConfig.On("GetInt", "http.request_timeout", 3).Return(1).Once()

	s.route.GlobalMiddleware(func(ctx contractshttp.Context) {
		ctx.WithValue("global", "goravel")
		ctx.Request().Next()
	})
	s.route.Get("/global-middleware", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Json(http.StatusOK, contractshttp.Json{
			"global": ctx.Value("global"),
		})
	})

	s.assert("GET", "/global-middleware", http.StatusOK, "{\"global\":\"goravel\"}")
}

func (s *GroupTestSuite) TestMiddlewareConflict() {
	s.route.Prefix("conflict").Group(func(route1 contractsroute.Router) {
		route1.Middleware(contextMiddleware()).Get("/middleware1/{id}", func(ctx contractshttp.Context) contractshttp.Response {
			return ctx.Response().Success().Json(contractshttp.Json{
				"id":   ctx.Request().Input("id"),
				"ctx":  ctx.Value("ctx"),
				"ctx2": ctx.Value("ctx2"),
			})
		})
		route1.Middleware(contextMiddleware2()).Post("/middleware2/{id}", func(ctx contractshttp.Context) contractshttp.Response {
			return ctx.Response().Success().Json(contractshttp.Json{
				"id":   ctx.Request().Input("id"),
				"ctx":  ctx.Value("ctx"),
				"ctx2": ctx.Value("ctx2"),
			})
		})
	})

	s.assert("POST", "/conflict/middleware2/1", http.StatusOK, "{\"ctx\":null,\"ctx2\":\"World\",\"id\":\"1\"}")
}

// https://github.com/goravel/goravel/issues/408
func (s *GroupTestSuite) TestIssue408() {
	s.route.Prefix("prefix/{id}").Group(func(route contractsroute.Router) {
		route.Get("", func(ctx contractshttp.Context) contractshttp.Response {
			return ctx.Response().String(200, "ok")
		})
		route.Post("test/{name}", func(ctx contractshttp.Context) contractshttp.Response {
			return ctx.Response().String(200, "ok")
		})
	})

	routes := s.route.GetRoutes()
	s.Equal(2, len(routes))
	s.Equal("GET|HEAD", routes[0].Method)
	s.Equal("/prefix/{id}", routes[0].Path)
	s.Equal("POST", routes[1].Method)
	s.Equal("/prefix/{id}/test/{name}", routes[1].Path)
}

func (s *GroupTestSuite) assert(method, url string, expectCode int, expectBody string) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest(method, url, nil)
	s.NoError(err)

	s.route.ServeHTTP(w, req)

	if expectBody != "" {
		s.Equal(expectBody, w.Body.String())
	}
	s.Equal(expectCode, w.Code)
}

func abortMiddleware() contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		ctx.Request().Abort(http.StatusNonAuthoritativeInfo)
	}
}

func contextMiddleware() contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		type customKey struct{}
		var customKeyCtx customKey
		ctx.WithValue(customKeyCtx, "context with custom key")
		ctx.WithValue("ctx", "Goravel")

		ctx.Request().Next()
	}
}

func contextMiddleware1() contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		ctx.WithValue(2.2, "two point two")
		ctx.WithValue("ctx1", "Hello")

		ctx.Request().Next()
	}
}

func contextMiddleware2() contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		ctx.WithValue("ctx2", "World")

		ctx.Request().Next()
	}
}

type resourceController struct{}

func (c resourceController) Index(ctx contractshttp.Context) contractshttp.Response {
	action := ctx.Value("action")

	return ctx.Response().Json(http.StatusOK, contractshttp.Json{
		"action": action,
	})
}

func (c resourceController) Show(ctx contractshttp.Context) contractshttp.Response {
	action := ctx.Value("action")
	id := ctx.Request().Input("id")

	return ctx.Response().Json(http.StatusOK, contractshttp.Json{
		"action": action,
		"id":     id,
	})
}

func (c resourceController) Store(ctx contractshttp.Context) contractshttp.Response {
	action := ctx.Value("action")

	return ctx.Response().Json(http.StatusOK, contractshttp.Json{
		"action": action,
	})
}

func (c resourceController) Update(ctx contractshttp.Context) contractshttp.Response {
	action := ctx.Value("action")
	id := ctx.Request().Input("id")

	return ctx.Response().Json(http.StatusOK, contractshttp.Json{
		"action": action,
		"id":     id,
	})
}

func (c resourceController) Destroy(ctx contractshttp.Context) contractshttp.Response {
	action := ctx.Value("action")
	id := ctx.Request().Input("id")

	return ctx.Response().Json(http.StatusOK, contractshttp.Json{
		"action": action,
		"id":     id,
	})
}
