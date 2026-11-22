package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	contractshttp "github.com/goravel/framework/contracts/http"
	contractsroute "github.com/goravel/framework/contracts/route"
	configmocks "github.com/goravel/framework/mocks/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type countMiddleware struct {
	count int
}

func (r *countMiddleware) Signature() string { return "count" }

func (r *countMiddleware) Handle(ctx contractshttp.Context) {
	r.count++
	ctx.Request().Next()
}

func newDomainTestRoute(t *testing.T, globalMiddleware ...contractshttp.Middleware) *Route {
	routes = make(map[string]map[string]contractshttp.Info)

	mockConfig := configmocks.NewConfig(t)
	mockConfig.EXPECT().GetInt("http.drivers.gin.body_limit", 4096).Return(4096).Once()
	mockConfig.EXPECT().GetBool("app.debug").Return(false).Once()
	mockConfig.EXPECT().Get("http.drivers.gin.template").Return(nil).Once()
	ConfigFacade = mockConfig

	route := &Route{
		config: mockConfig,
		driver: "gin",
	}
	route.init(globalMiddleware)

	return route
}

func stringHandler(body string) contractshttp.HandlerFunc {
	return func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().String(http.StatusOK, body)
	}
}

func serveHost(t *testing.T, route *Route, method, host, path string) *httptest.ResponseRecorder {
	t.Helper()

	request, err := http.NewRequest(method, "http://"+host+path, nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	route.ServeHTTP(recorder, request)

	return recorder
}

func TestDomain(t *testing.T) {
	route := newDomainTestRoute(t)

	route.Domain("admin.example.com").Group(func(router contractsroute.Router) {
		router.Get("/", stringHandler("admin home"))
		router.Prefix("users").Get("/{id}", func(ctx contractshttp.Context) contractshttp.Response {
			return ctx.Response().String(http.StatusOK, "admin user "+ctx.Request().Route("id"))
		})
	})
	route.Domain("{account}.example.com").Get("/", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().String(http.StatusOK, "account "+ctx.Request().Route("account"))
	})
	route.Get("/", stringHandler("home"))
	route.Get("/about", stringHandler("about"))

	tests := []struct {
		name       string
		method     string
		host       string
		path       string
		expectCode int
		expectBody string
	}{
		{name: "the domain's route", method: "GET", host: "admin.example.com", path: "/", expectCode: http.StatusOK, expectBody: "admin home"},
		{name: "with a port and in upper case", method: "GET", host: "Admin.Example.com:3000", path: "/", expectCode: http.StatusOK, expectBody: "admin home"},
		{name: "with a path parameter", method: "GET", host: "admin.example.com", path: "/users/1", expectCode: http.StatusOK, expectBody: "admin user 1"},
		{name: "HEAD on the domain", method: "HEAD", host: "admin.example.com", path: "/", expectCode: http.StatusOK},
		{name: "a static domain before a parameter", method: "GET", host: "admin.example.com", path: "/", expectCode: http.StatusOK, expectBody: "admin home"},
		{name: "with a domain parameter", method: "GET", host: "goravel.example.com", path: "/", expectCode: http.StatusOK, expectBody: "account goravel"},
		{name: "a parameter matches one label", method: "GET", host: "a.b.example.com", path: "/", expectCode: http.StatusOK, expectBody: "home"},
		{name: "another host uses the routes without a domain", method: "GET", host: "example.com", path: "/", expectCode: http.StatusOK, expectBody: "home"},
		{name: "a path the domain doesn't have falls back to any host", method: "GET", host: "admin.example.com", path: "/about", expectCode: http.StatusOK, expectBody: "about"},
		{name: "a domain route is not on other hosts", method: "GET", host: "example.com", path: "/users/1", expectCode: http.StatusNotFound},
		{name: "not found on the domain", method: "GET", host: "admin.example.com", path: "/missing", expectCode: http.StatusNotFound},
		{name: "a method the domain doesn't have falls back", method: "POST", host: "admin.example.com", path: "/", expectCode: http.StatusNotFound},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := serveHost(t, route, test.method, test.host, test.path)

			assert.Equal(t, test.expectCode, recorder.Code)
			if test.expectBody != "" {
				assert.Equal(t, test.expectBody, recorder.Body.String())
			}
		})
	}
}

func TestDomainRunsTheGlobalMiddlewareOnce(t *testing.T) {
	counter := &countMiddleware{}
	route := newDomainTestRoute(t, counter)

	route.Domain("admin.example.com").Get("/", stringHandler("admin home"))
	route.Get("/about", stringHandler("about"))

	assert.Equal(t, "admin home", serveHost(t, route, "GET", "admin.example.com", "/").Body.String())
	assert.Equal(t, 1, counter.count)

	// Passed to the routes of any host.
	assert.Equal(t, "about", serveHost(t, route, "GET", "admin.example.com", "/about").Body.String())
	assert.Equal(t, 2, counter.count)
}

func TestDomainWithFallback(t *testing.T) {
	route := newDomainTestRoute(t)

	route.Domain("admin.example.com").Get("/", stringHandler("admin home"))
	route.Fallback(stringHandler("fallback"))

	recorder := serveHost(t, route, "GET", "admin.example.com", "/missing")
	assert.Equal(t, "fallback", recorder.Body.String())
}

func TestDomainRoutesInfo(t *testing.T) {
	route := newDomainTestRoute(t)

	route.Domain("admin.example.com").Get("/", stringHandler("admin home")).Name("admin.home")
	route.Get("/", stringHandler("home")).Name("home")

	assert.Equal(t, contractshttp.Info{
		Domain:  "admin.example.com",
		Handler: "github.com/goravel/gin.stringHandler.func1",
		Method:  "GET|HEAD",
		Name:    "admin.home",
		Path:    "/",
	}, route.Info("admin.home"))
	assert.Equal(t, contractshttp.Info{
		Handler: "github.com/goravel/gin.stringHandler.func1",
		Method:  "GET|HEAD",
		Name:    "home",
		Path:    "/",
	}, route.Info("home"))
	assert.Len(t, route.GetRoutes(), 2)
}

func TestDomainRouterMatch(t *testing.T) {
	router := newDomainRouter("{account}.{region}.Example.com", nil)

	params, ok := router.match("goravel.eu.example.com")
	assert.True(t, ok)
	assert.Equal(t, []string{"account", "region"}, router.params)
	assert.Len(t, params, 2)
	assert.Equal(t, "goravel", params.ByName("account"))
	assert.Equal(t, "eu", params.ByName("region"))

	_, ok = router.match("eu.example.com")
	assert.False(t, ok)
	_, ok = router.match("goravel.eu.example.org")
	assert.False(t, ok)

	static := newDomainRouter("Admin.Example.com", nil)
	_, ok = static.match("admin.example.com")
	assert.True(t, ok)
	_, ok = static.match("adminXexample.com")
	assert.False(t, ok)
}
