package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	configmock "github.com/goravel/framework/contracts/config/mocks"
	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	"github.com/stretchr/testify/assert"
)

type resourceController struct{}

func (c resourceController) Index(ctx httpcontract.Context) {
	action := ctx.Value("action")
	ctx.Response().Json(http.StatusOK, httpcontract.Json{
		"action": action,
	})
}

func (c resourceController) Show(ctx httpcontract.Context) {
	action := ctx.Value("action")
	id := ctx.Request().Input("id")
	ctx.Response().Json(http.StatusOK, httpcontract.Json{
		"action": action,
		"id":     id,
	})
}

func (c resourceController) Store(ctx httpcontract.Context) {
	action := ctx.Value("action")
	ctx.Response().Json(http.StatusOK, httpcontract.Json{
		"action": action,
	})
}

func (c resourceController) Update(ctx httpcontract.Context) {
	action := ctx.Value("action")
	id := ctx.Request().Input("id")
	ctx.Response().Json(http.StatusOK, httpcontract.Json{
		"action": action,
		"id":     id,
	})
}

func (c resourceController) Destroy(ctx httpcontract.Context) {
	action := ctx.Value("action")
	id := ctx.Request().Input("id")
	ctx.Response().Json(http.StatusOK, httpcontract.Json{
		"action": action,
		"id":     id,
	})
}

func TestGroup(t *testing.T) {
	var (
		gin        *Route
		mockConfig *configmock.Config
	)
	beforeEach := func() {
		mockConfig = &configmock.Config{}
		mockConfig.On("GetBool", "app.debug").Return(true).Once()
		mockConfig.On("Get", "cors.paths").Return([]string{}).Once()
		ConfigFacade = mockConfig

		gin = NewRoute(mockConfig)
	}
	tests := []struct {
		name       string
		setup      func(req *http.Request)
		method     string
		url        string
		expectCode int
		expectBody string
	}{
		{
			name: "Get",
			setup: func(req *http.Request) {
				gin.Get("/input/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Json(http.StatusOK, httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "GET",
			url:        "/input/1",
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name: "Post",
			setup: func(req *http.Request) {
				gin.Post("/input/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "POST",
			url:        "/input/1",
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name: "Put",
			setup: func(req *http.Request) {
				gin.Put("/input/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "PUT",
			url:        "/input/1",
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name: "Delete",
			setup: func(req *http.Request) {
				gin.Delete("/input/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "DELETE",
			url:        "/input/1",
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name: "Options",
			setup: func(req *http.Request) {
				gin.Options("/input/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "OPTIONS",
			url:        "/input/1",
			expectCode: http.StatusOK,
		},
		{
			name: "Patch",
			setup: func(req *http.Request) {
				gin.Patch("/input/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "PATCH",
			url:        "/input/1",
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name: "Any Get",
			setup: func(req *http.Request) {
				gin.Any("/any/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "GET",
			url:        "/any/1",
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name: "Any Post",
			setup: func(req *http.Request) {
				gin.Any("/any/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "POST",
			url:        "/any/1",
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name: "Any Put",
			setup: func(req *http.Request) {
				gin.Any("/any/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "PUT",
			url:        "/any/1",
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name: "Any Delete",
			setup: func(req *http.Request) {
				gin.Any("/any/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "DELETE",
			url:        "/any/1",
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name: "Any Patch",
			setup: func(req *http.Request) {
				gin.Any("/any/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "PATCH",
			url:        "/any/1",
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name: "Resource Index",
			setup: func(req *http.Request) {
				mockConfig.On("Get", "cors.paths").Return([]string{}).Times(5)
				mockConfig.On("GetString", "http.tls.host").Return("").Once()
				mockConfig.On("GetString", "http.tls.port").Return("").Once()
				mockConfig.On("GetString", "http.tls.ssl.cert").Return("").Once()
				mockConfig.On("GetString", "http.tls.ssl.key").Return("").Once()

				resource := resourceController{}
				gin.GlobalMiddleware(func(ctx httpcontract.Context) {
					ctx.WithValue("action", "index")
					ctx.Request().Next()
				})
				gin.Resource("/resource", resource)
			},
			method:     "GET",
			url:        "/resource",
			expectCode: http.StatusOK,
			expectBody: "{\"action\":\"index\"}",
		},
		{
			name: "Resource Show",
			setup: func(req *http.Request) {
				mockConfig.On("Get", "cors.paths").Return([]string{}).Times(5)
				mockConfig.On("GetString", "http.tls.host").Return("").Once()
				mockConfig.On("GetString", "http.tls.port").Return("").Once()
				mockConfig.On("GetString", "http.tls.ssl.cert").Return("").Once()
				mockConfig.On("GetString", "http.tls.ssl.key").Return("").Once()

				resource := resourceController{}
				gin.GlobalMiddleware(func(ctx httpcontract.Context) {
					ctx.WithValue("action", "show")
					ctx.Request().Next()
				})
				gin.Resource("/resource", resource)
			},
			method:     "GET",
			url:        "/resource/1",
			expectCode: http.StatusOK,
			expectBody: "{\"action\":\"show\",\"id\":\"1\"}",
		},
		{
			name: "Resource Store",
			setup: func(req *http.Request) {
				mockConfig.On("Get", "cors.paths").Return([]string{}).Times(5)
				mockConfig.On("GetString", "http.tls.host").Return("").Once()
				mockConfig.On("GetString", "http.tls.port").Return("").Once()
				mockConfig.On("GetString", "http.tls.ssl.cert").Return("").Once()
				mockConfig.On("GetString", "http.tls.ssl.key").Return("").Once()

				resource := resourceController{}
				gin.GlobalMiddleware(func(ctx httpcontract.Context) {
					ctx.WithValue("action", "store")
					ctx.Request().Next()
				})
				gin.Resource("/resource", resource)
			},
			method:     "POST",
			url:        "/resource",
			expectCode: http.StatusOK,
			expectBody: "{\"action\":\"store\"}",
		},
		{
			name: "Resource Update (PUT)",
			setup: func(req *http.Request) {
				mockConfig.On("Get", "cors.paths").Return([]string{}).Times(5)
				mockConfig.On("GetString", "http.tls.host").Return("").Once()
				mockConfig.On("GetString", "http.tls.port").Return("").Once()
				mockConfig.On("GetString", "http.tls.ssl.cert").Return("").Once()
				mockConfig.On("GetString", "http.tls.ssl.key").Return("").Once()

				resource := resourceController{}
				gin.GlobalMiddleware(func(ctx httpcontract.Context) {
					ctx.WithValue("action", "update")
					ctx.Request().Next()
				})
				gin.Resource("/resource", resource)
			},
			method:     "PUT",
			url:        "/resource/1",
			expectCode: http.StatusOK,
			expectBody: "{\"action\":\"update\",\"id\":\"1\"}",
		},
		{
			name: "Resource Update (PATCH)",
			setup: func(req *http.Request) {
				mockConfig.On("Get", "cors.paths").Return([]string{}).Times(5)
				mockConfig.On("GetString", "http.tls.host").Return("").Once()
				mockConfig.On("GetString", "http.tls.port").Return("").Once()
				mockConfig.On("GetString", "http.tls.ssl.cert").Return("").Once()
				mockConfig.On("GetString", "http.tls.ssl.key").Return("").Once()

				resource := resourceController{}
				gin.GlobalMiddleware(func(ctx httpcontract.Context) {
					ctx.WithValue("action", "update")
					ctx.Request().Next()
				})
				gin.Resource("/resource", resource)
			},
			method:     "PATCH",
			url:        "/resource/1",
			expectCode: http.StatusOK,
			expectBody: "{\"action\":\"update\",\"id\":\"1\"}",
		},
		{
			name: "Resource Destroy",
			setup: func(req *http.Request) {
				mockConfig.On("Get", "cors.paths").Return([]string{}).Times(5)
				mockConfig.On("GetString", "http.tls.host").Return("").Once()
				mockConfig.On("GetString", "http.tls.port").Return("").Once()
				mockConfig.On("GetString", "http.tls.ssl.cert").Return("").Once()
				mockConfig.On("GetString", "http.tls.ssl.key").Return("").Once()

				resource := resourceController{}
				gin.GlobalMiddleware(func(ctx httpcontract.Context) {
					ctx.WithValue("action", "destroy")
					ctx.Request().Next()
				})
				gin.Resource("/resource", resource)
			},
			method:     "DELETE",
			url:        "/resource/1",
			expectCode: http.StatusOK,
			expectBody: "{\"action\":\"destroy\",\"id\":\"1\"}",
		},
		{
			name: "Static",
			setup: func(req *http.Request) {
				gin.Static("static", "./")
			},
			method:     "GET",
			url:        "/static/README.md",
			expectCode: http.StatusOK,
		},
		{
			name: "StaticFile",
			setup: func(req *http.Request) {
				gin.StaticFile("static-file", "./README.md")
			},
			method:     "GET",
			url:        "/static-file",
			expectCode: http.StatusOK,
		},
		{
			name: "StaticFS",
			setup: func(req *http.Request) {
				gin.StaticFS("static-fs", http.Dir("./"))
			},
			method:     "GET",
			url:        "/static-fs",
			expectCode: http.StatusMovedPermanently,
		},
		{
			name: "Abort Middleware",
			setup: func(req *http.Request) {
				gin.Middleware(abortMiddleware()).Get("/middleware/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "GET",
			url:        "/middleware/1",
			expectCode: http.StatusNonAuthoritativeInfo,
		},
		{
			name: "Multiple Middleware",
			setup: func(req *http.Request) {
				gin.Middleware(contextMiddleware(), contextMiddleware1()).Get("/middlewares/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id":   ctx.Request().Input("id"),
						"ctx":  ctx.Value("ctx"),
						"ctx1": ctx.Value("ctx1"),
					})
				})
			},
			method:     "GET",
			url:        "/middlewares/1",
			expectCode: http.StatusOK,
			expectBody: "{\"ctx\":\"Goravel\",\"ctx1\":\"Hello\",\"id\":\"1\"}",
		},
		{
			name: "Multiple Prefix",
			setup: func(req *http.Request) {
				gin.Prefix("prefix1").Prefix("prefix2").Get("input/{id}", func(ctx httpcontract.Context) {
					ctx.Response().Success().Json(httpcontract.Json{
						"id": ctx.Request().Input("id"),
					})
				})
			},
			method:     "GET",
			url:        "/prefix1/prefix2/input/1",
			expectCode: http.StatusOK,
			expectBody: "{\"id\":\"1\"}",
		},
		{
			name: "Multiple Prefix Group Middleware",
			setup: func(req *http.Request) {
				mockConfig.On("Get", "cors.paths").Return([]string{}).Once()

				gin.Prefix("group1").Middleware(contextMiddleware()).Group(func(route1 route.Route) {
					route1.Prefix("group2").Middleware(contextMiddleware1()).Group(func(route2 route.Route) {
						route2.Get("/middleware/{id}", func(ctx httpcontract.Context) {
							ctx.Response().Success().Json(httpcontract.Json{
								"id":   ctx.Request().Input("id"),
								"ctx":  ctx.Value("ctx").(string),
								"ctx1": ctx.Value("ctx1").(string),
							})
						})
					})
					route1.Middleware(contextMiddleware2()).Get("/middleware/{id}", func(ctx httpcontract.Context) {
						ctx.Response().Success().Json(httpcontract.Json{
							"id":   ctx.Request().Input("id"),
							"ctx":  ctx.Value("ctx").(string),
							"ctx2": ctx.Value("ctx2").(string),
						})
					})
				})
			},
			method:     "GET",
			url:        "/group1/group2/middleware/1",
			expectCode: http.StatusOK,
			expectBody: "{\"ctx\":\"Goravel\",\"ctx1\":\"Hello\",\"id\":\"1\"}",
		},
		{
			name: "Multiple Group Middleware",
			setup: func(req *http.Request) {
				mockConfig.On("Get", "cors.paths").Return([]string{}).Once()

				gin.Prefix("group1").Middleware(contextMiddleware()).Group(func(route1 route.Route) {
					route1.Prefix("group2").Middleware(contextMiddleware1()).Group(func(route2 route.Route) {
						route2.Get("/middleware/{id}", func(ctx httpcontract.Context) {
							ctx.Response().Success().Json(httpcontract.Json{
								"id":   ctx.Request().Input("id"),
								"ctx":  ctx.Value("ctx").(string),
								"ctx1": ctx.Value("ctx1").(string),
							})
						})
					})
					route1.Middleware(contextMiddleware2()).Get("/middleware/{id}", func(ctx httpcontract.Context) {
						ctx.Response().Success().Json(httpcontract.Json{
							"id":   ctx.Request().Input("id"),
							"ctx":  ctx.Value("ctx").(string),
							"ctx2": ctx.Value("ctx2").(string),
						})
					})
				})
			},
			method:     "GET",
			url:        "/group1/middleware/1",
			expectCode: http.StatusOK,
			expectBody: "{\"ctx\":\"Goravel\",\"ctx2\":\"World\",\"id\":\"1\"}",
		},
		{
			name: "Global Middleware",
			setup: func(req *http.Request) {
				mockConfig.On("GetString", "http.tls.host").Return("").Once()
				mockConfig.On("GetString", "http.tls.port").Return("").Once()
				mockConfig.On("GetString", "http.tls.ssl.cert").Return("").Once()
				mockConfig.On("GetString", "http.tls.ssl.key").Return("").Once()

				gin.GlobalMiddleware(func(ctx httpcontract.Context) {
					ctx.WithValue("global", "goravel")
					ctx.Request().Next()
				})
				gin.Get("/global-middleware", func(ctx httpcontract.Context) {
					ctx.Response().Json(http.StatusOK, httpcontract.Json{
						"global": ctx.Value("global"),
					})
				})
			},
			method:     "GET",
			url:        "/global-middleware",
			expectCode: http.StatusOK,
			expectBody: "{\"global\":\"goravel\"}",
		},
		{
			name: "Middleware Conflict",
			setup: func(req *http.Request) {
				mockConfig.On("Get", "cors.paths").Return([]string{}).Once()

				gin.Prefix("conflict").Group(func(route1 route.Route) {
					route1.Middleware(contextMiddleware()).Get("/middleware1/{id}", func(ctx httpcontract.Context) {
						ctx.Response().Success().Json(httpcontract.Json{
							"id":   ctx.Request().Input("id"),
							"ctx":  ctx.Value("ctx"),
							"ctx2": ctx.Value("ctx2"),
						})
					})
					route1.Middleware(contextMiddleware2()).Post("/middleware2/{id}", func(ctx httpcontract.Context) {
						ctx.Response().Success().Json(httpcontract.Json{
							"id":   ctx.Request().Input("id"),
							"ctx":  ctx.Value("ctx"),
							"ctx2": ctx.Value("ctx2"),
						})
					})
				})
			},
			method:     "POST",
			url:        "/conflict/middleware2/1",
			expectCode: http.StatusOK,
			expectBody: "{\"ctx\":null,\"ctx2\":\"World\",\"id\":\"1\"}",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(test.method, test.url, nil)
			if test.setup != nil {
				test.setup(req)
			}
			gin.ServeHTTP(w, req)

			if test.expectBody != "" {
				assert.Equal(t, test.expectBody, w.Body.String(), test.name)
			}
			assert.Equal(t, test.expectCode, w.Code, test.name)
			mockConfig.AssertExpectations(t)
		})
	}
}

func TestAddCorsMiddleware(t *testing.T) {
	var (
		group       *Group
		mockConfig  *configmock.Config
		middlewares []gin.HandlerFunc
	)

	beforeEach := func() {
		mockConfig = new(configmock.Config)
		group = &Group{config: mockConfig}
	}

	tests := []struct {
		name                   string
		setup                  func()
		fullPath               string
		expectMiddlewareLength int
	}{
		{
			name: "cors.paths is empty",
			setup: func() {
				mockConfig.On("Get", "cors.paths").Return([]string{}).Once()
			},
			expectMiddlewareLength: 0,
		},
		{
			name: "cors.paths contains *",
			setup: func() {
				mockConfig.On("Get", "cors.paths").Return([]string{"api/*"}).Once()
			},
			fullPath:               "/api/v1/user",
			expectMiddlewareLength: 1,
		},
		{
			name: "cors.paths contains * and path has no /",
			setup: func() {
				mockConfig.On("Get", "cors.paths").Return([]string{"/api/*"}).Once()
			},
			fullPath:               "api/v1/user",
			expectMiddlewareLength: 1,
		},
		{
			name: "cors.paths is *",
			setup: func() {
				mockConfig.On("Get", "cors.paths").Return([]string{"*"}).Once()
			},
			fullPath:               "/api/v1/user",
			expectMiddlewareLength: 1,
		},
		{
			name: "cors.paths is a specific path",
			setup: func() {
				mockConfig.On("Get", "cors.paths").Return([]string{"api"}).Once()
			},
			fullPath:               "/api",
			expectMiddlewareLength: 1,
		},
		{
			name: "cors.paths is a specific path and path has no /",
			setup: func() {
				mockConfig.On("Get", "cors.paths").Return([]string{"/api"}).Once()
			},
			fullPath:               "api",
			expectMiddlewareLength: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()
			result := group.addCorsMiddleware(middlewares, test.fullPath)
			assert.Equal(t, test.expectMiddlewareLength, len(result))
		})
	}
}

func abortMiddleware() httpcontract.Middleware {
	return func(ctx httpcontract.Context) {
		ctx.Request().AbortWithStatus(http.StatusNonAuthoritativeInfo)
	}
}

func contextMiddleware() httpcontract.Middleware {
	return func(ctx httpcontract.Context) {
		ctx.WithValue("ctx", "Goravel")

		ctx.Request().Next()
	}
}

func contextMiddleware1() httpcontract.Middleware {
	return func(ctx httpcontract.Context) {
		ctx.WithValue("ctx1", "Hello")

		ctx.Request().Next()
	}
}

func contextMiddleware2() httpcontract.Middleware {
	return func(ctx httpcontract.Context) {
		ctx.WithValue("ctx2", "World")

		ctx.Request().Next()
	}
}
