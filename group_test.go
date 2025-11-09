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
	routes = make(map[string]map[string]contractshttp.Info)

	s.mockConfig = configmocks.NewConfig(s.T())
	s.mockConfig.EXPECT().GetInt("http.drivers.gin.body_limit", 4096).Return(4096).Once()
	s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	s.mockConfig.EXPECT().Get("http.drivers.gin.template").Return(nil).Once()

	ConfigFacade = s.mockConfig

	route := &Route{
		config: s.mockConfig,
		driver: "gin",
	}
	err := route.init(nil)
	s.Require().Nil(err)

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
	s.Equal(contractshttp.Info{
		Handler: "github.com/goravel/gin.(*GroupTestSuite).TestGet.func1",
		Method:  "GET|HEAD",
		Path:    "/input/{id}",
		Name:    "get",
	}, s.route.Info("get"))
}

func (s *GroupTestSuite) TestPost() {
	s.route.Post("/input/{id}", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Json(http.StatusOK, contractshttp.Json{
			"id": ctx.Request().Input("id"),
		})
	}).Name("post")

	s.assert("POST", "/input/1", http.StatusOK, "{\"id\":\"1\"}")
	s.Equal(contractshttp.Info{
		Handler: "github.com/goravel/gin.(*GroupTestSuite).TestPost.func1",
		Method:  contractshttp.MethodPost,
		Path:    "/input/{id}",
		Name:    "post",
	}, s.route.Info("post"))
}

func (s *GroupTestSuite) TestPut() {
	s.route.Put("/input/{id}", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Json(http.StatusOK, contractshttp.Json{
			"id": ctx.Request().Input("id"),
		})
	}).Name("put")

	s.assert("PUT", "/input/1", http.StatusOK, "{\"id\":\"1\"}")
	s.Equal(contractshttp.Info{
		Handler: "github.com/goravel/gin.(*GroupTestSuite).TestPut.func1",
		Method:  contractshttp.MethodPut,
		Path:    "/input/{id}",
		Name:    "put",
	}, s.route.Info("put"))
}

func (s *GroupTestSuite) TestDelete() {
	s.route.Delete("/input/{id}", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Json(http.StatusOK, contractshttp.Json{
			"id": ctx.Request().Input("id"),
		})
	}).Name("delete")

	s.assert("DELETE", "/input/1", http.StatusOK, "{\"id\":\"1\"}")
	s.Equal(contractshttp.Info{
		Handler: "github.com/goravel/gin.(*GroupTestSuite).TestDelete.func1",
		Method:  contractshttp.MethodDelete,
		Path:    "/input/{id}",
		Name:    "delete",
	}, s.route.Info("delete"))
}

func (s *GroupTestSuite) TestOptions() {
	s.route.Options("/input/{id}", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Json(http.StatusOK, contractshttp.Json{
			"id": ctx.Request().Input("id"),
		})
	}).Name("options")

	s.assert("OPTIONS", "/input/1", http.StatusOK, "{\"id\":\"1\"}")
	s.Equal(contractshttp.Info{
		Handler: "github.com/goravel/gin.(*GroupTestSuite).TestOptions.func1",
		Method:  contractshttp.MethodOptions,
		Path:    "/input/{id}",
		Name:    "options",
	}, s.route.Info("options"))
}

func (s *GroupTestSuite) TestPatch() {
	s.route.Patch("/input/{id}", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Json(http.StatusOK, contractshttp.Json{
			"id": ctx.Request().Input("id"),
		})
	}).Name("patch")

	s.assert("PATCH", "/input/1", http.StatusOK, "{\"id\":\"1\"}")
	s.Equal(contractshttp.Info{
		Handler: "github.com/goravel/gin.(*GroupTestSuite).TestPatch.func1",
		Method:  contractshttp.MethodPatch,
		Path:    "/input/{id}",
		Name:    "patch",
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

	s.Equal(contractshttp.Info{
		Handler: "github.com/goravel/gin.(*GroupTestSuite).TestAny.func1",
		Method:  contractshttp.MethodAny,
		Path:    "/input/{id}",
		Name:    "any",
	}, s.route.Info("any"))
}

func (s *GroupTestSuite) TestResource() {
	s.route.Resource("/resource", resourceController{}).Name("resource")

	s.assert("GET", "/resource", http.StatusOK, "{\"action\":\"Index\"}")
	s.assert("GET", "/resource/1", http.StatusOK, "{\"action\":\"Show\",\"id\":\"1\"}")
	s.assert("POST", "/resource", http.StatusOK, "{\"action\":\"Store\"}")
	s.assert("PUT", "/resource/1", http.StatusOK, "{\"action\":\"Update\",\"id\":\"1\"}")
	s.assert("PATCH", "/resource/1", http.StatusOK, "{\"action\":\"Update\",\"id\":\"1\"}")
	s.assert("DELETE", "/resource/1", http.StatusOK, "{\"action\":\"Destroy\",\"id\":\"1\"}")

	s.Equal(contractshttp.Info{
		Handler: "github.com/goravel/gin.(resourceController)",
		Method:  contractshttp.MethodResource,
		Path:    "/resource",
		Name:    "resource",
	}, s.route.Info("resource"))
}

func (s *GroupTestSuite) TestStatic() {
	s.route.Static("static", "./").Name("static")

	s.assert("GET", "/static/README.md", http.StatusOK, "")

	s.Equal(contractshttp.Info{
		Method: contractshttp.MethodStatic,
		Path:   "/static",
		Name:   "static",
	}, s.route.Info("static"))
}

func (s *GroupTestSuite) TestStaticFile() {
	s.route.StaticFile("static-file", "./README.md").Name("static-file")

	s.assert("GET", "/static-file", http.StatusOK, "")

	s.Equal(contractshttp.Info{
		Method: contractshttp.MethodStaticFile,
		Path:   "/static-file",
		Name:   "static-file",
	}, s.route.Info("static-file"))
}

func (s *GroupTestSuite) TestStaticFS() {
	s.route.StaticFS("/static-fs", http.Dir("./")).Name("static-fs")

	s.assert("GET", "/static-fs", http.StatusMovedPermanently, "")

	s.Equal(contractshttp.Info{
		Method: contractshttp.MethodStaticFS,
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
	s.mockConfig.EXPECT().GetInt("http.drivers.gin.body_limit", 4096).Return(4096).Once()
	s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	s.mockConfig.EXPECT().Get("http.drivers.gin.template").Return(nil).Once()

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
	return ctx.Response().Json(http.StatusOK, contractshttp.Json{
		"action": "Index",
	})
}

func (c resourceController) Show(ctx contractshttp.Context) contractshttp.Response {
	id := ctx.Request().Input("id")

	return ctx.Response().Json(http.StatusOK, contractshttp.Json{
		"action": "Show",
		"id":     id,
	})
}

func (c resourceController) Store(ctx contractshttp.Context) contractshttp.Response {
	return ctx.Response().Json(http.StatusOK, contractshttp.Json{
		"action": "Store",
	})
}

func (c resourceController) Update(ctx contractshttp.Context) contractshttp.Response {
	id := ctx.Request().Input("id")

	return ctx.Response().Json(http.StatusOK, contractshttp.Json{
		"action": "Update",
		"id":     id,
	})
}

func (c resourceController) Destroy(ctx contractshttp.Context) contractshttp.Response {
	id := ctx.Request().Input("id")

	return ctx.Response().Json(http.StatusOK, contractshttp.Json{
		"action": "Destroy",
		"id":     id,
	})
}
