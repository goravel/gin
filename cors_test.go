package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	contractshttp "github.com/goravel/framework/contracts/http"
	configmocks "github.com/goravel/framework/mocks/config"
	"github.com/stretchr/testify/assert"
)

func TestCors(t *testing.T) {
	var (
		mockConfig *configmocks.Config
		resp       *httptest.ResponseRecorder
	)
	beforeEach := func() {
		mockConfig = &configmocks.Config{}
	}

	tests := []struct {
		name   string
		method string
		setup  func()
		assert func()
	}{
		{
			name: "allow all paths",
			setup: func() {
				mockConfig.On("GetBool", "app.debug").Return(true).Once()
				mockConfig.On("Get", "cors.paths").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.allowed_methods").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.allowed_origins").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.allowed_headers").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.exposed_headers").Return([]string{"*"}).Once()
				mockConfig.On("GetInt", "cors.max_age").Return(0).Once()
				mockConfig.On("GetBool", "cors.supports_credentials").Return(false).Once()
				ConfigFacade = mockConfig
			},
			assert: func() {
				assert.Equal(t, http.StatusNoContent, resp.Code)
				assert.Equal(t, "POST", resp.Header().Get("Access-Control-Allow-Methods"))
				assert.Equal(t, "*", resp.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "", resp.Header().Get("Access-Control-Allow-Headers"))
				assert.Equal(t, "", resp.Header().Get("Access-Control-Expose-Headers"))
			},
		},
		{
			name: "not allow path",
			setup: func() {
				mockConfig.On("GetBool", "app.debug").Return(true).Once()
				mockConfig.On("Get", "cors.paths").Return([]string{"api"}).Once()
				mockConfig.On("GetString", "http.tls.host").Return("").Once()
				mockConfig.On("GetString", "http.tls.port").Return("").Once()
				mockConfig.On("GetString", "http.tls.ssl.cert").Return("").Once()
				mockConfig.On("GetString", "http.tls.ssl.key").Return("").Once()
				ConfigFacade = mockConfig
			},
			assert: func() {
				assert.Equal(t, http.StatusNotFound, resp.Code)
				assert.Equal(t, "", resp.Header().Get("Access-Control-Allow-Methods"))
				assert.Equal(t, "", resp.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "", resp.Header().Get("Access-Control-Allow-Headers"))
				assert.Equal(t, "", resp.Header().Get("Access-Control-Expose-Headers"))
			},
		},
		{
			name: "allow path with *",
			setup: func() {
				mockConfig.On("GetBool", "app.debug").Return(true).Once()
				mockConfig.On("Get", "cors.paths").Return([]string{"any/*"}).Once()
				mockConfig.On("Get", "cors.allowed_methods").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.allowed_origins").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.allowed_headers").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.exposed_headers").Return([]string{"*"}).Once()
				mockConfig.On("GetInt", "cors.max_age").Return(0).Once()
				mockConfig.On("GetBool", "cors.supports_credentials").Return(false).Once()
				ConfigFacade = mockConfig
			},
			assert: func() {
				assert.Equal(t, http.StatusNoContent, resp.Code)
				assert.Equal(t, "POST", resp.Header().Get("Access-Control-Allow-Methods"))
				assert.Equal(t, "*", resp.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "", resp.Header().Get("Access-Control-Allow-Headers"))
				assert.Equal(t, "", resp.Header().Get("Access-Control-Expose-Headers"))
			},
		},
		{
			name: "only allow POST",
			setup: func() {
				mockConfig.On("GetBool", "app.debug").Return(true).Once()
				mockConfig.On("Get", "cors.paths").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.allowed_methods").Return([]string{"POST"}).Once()
				mockConfig.On("Get", "cors.allowed_origins").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.allowed_headers").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.exposed_headers").Return([]string{"*"}).Once()
				mockConfig.On("GetInt", "cors.max_age").Return(0).Once()
				mockConfig.On("GetBool", "cors.supports_credentials").Return(false).Once()
				ConfigFacade = mockConfig
			},
			assert: func() {
				assert.Equal(t, http.StatusNoContent, resp.Code)
				assert.Equal(t, "POST", resp.Header().Get("Access-Control-Allow-Methods"))
				assert.Equal(t, "*", resp.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "", resp.Header().Get("Access-Control-Allow-Headers"))
				assert.Equal(t, "", resp.Header().Get("Access-Control-Expose-Headers"))
			},
		},
		{
			name: "not allow POST",
			setup: func() {
				mockConfig.On("GetBool", "app.debug").Return(true).Once()
				mockConfig.On("Get", "cors.paths").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.allowed_methods").Return([]string{"GET"}).Once()
				mockConfig.On("Get", "cors.allowed_origins").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.allowed_headers").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.exposed_headers").Return([]string{"*"}).Once()
				mockConfig.On("GetInt", "cors.max_age").Return(0).Once()
				mockConfig.On("GetBool", "cors.supports_credentials").Return(false).Once()
				ConfigFacade = mockConfig
			},
			assert: func() {
				assert.Equal(t, http.StatusNoContent, resp.Code)
				assert.Equal(t, "", resp.Header().Get("Access-Control-Allow-Methods"))
				assert.Equal(t, "", resp.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "", resp.Header().Get("Access-Control-Allow-Headers"))
				assert.Equal(t, "", resp.Header().Get("Access-Control-Expose-Headers"))
			},
		},
		{
			name: "not allow origin",
			setup: func() {
				mockConfig.On("GetBool", "app.debug").Return(true).Once()
				mockConfig.On("Get", "cors.paths").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.allowed_methods").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.allowed_origins").Return([]string{"https://goravel.com"}).Once()
				mockConfig.On("Get", "cors.allowed_headers").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.exposed_headers").Return([]string{"*"}).Once()
				mockConfig.On("GetInt", "cors.max_age").Return(0).Once()
				mockConfig.On("GetBool", "cors.supports_credentials").Return(false).Once()
				ConfigFacade = mockConfig
			},
			assert: func() {
				assert.Equal(t, http.StatusNoContent, resp.Code)
				assert.Equal(t, "", resp.Header().Get("Access-Control-Allow-Methods"))
				assert.Equal(t, "", resp.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "", resp.Header().Get("Access-Control-Allow-Headers"))
				assert.Equal(t, "", resp.Header().Get("Access-Control-Expose-Headers"))
			},
		},
		{
			name: "allow specific origin",
			setup: func() {
				mockConfig.On("GetBool", "app.debug").Return(true).Once()
				mockConfig.On("Get", "cors.paths").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.allowed_methods").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.allowed_origins").Return([]string{"https://goravel.dev"}).Once()
				mockConfig.On("Get", "cors.allowed_headers").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.exposed_headers").Return([]string{"*"}).Once()
				mockConfig.On("GetInt", "cors.max_age").Return(0).Once()
				mockConfig.On("GetBool", "cors.supports_credentials").Return(false).Once()
				ConfigFacade = mockConfig
			},
			assert: func() {
				assert.Equal(t, http.StatusNoContent, resp.Code)
				assert.Equal(t, "POST", resp.Header().Get("Access-Control-Allow-Methods"))
				assert.Equal(t, "https://goravel.dev", resp.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "", resp.Header().Get("Access-Control-Allow-Headers"))
				assert.Equal(t, "", resp.Header().Get("Access-Control-Expose-Headers"))
			},
		},
		{
			name: "not allow exposed headers",
			setup: func() {
				mockConfig.On("GetBool", "app.debug").Return(true).Once()
				mockConfig.On("Get", "cors.paths").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.allowed_methods").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.allowed_origins").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.allowed_headers").Return([]string{"*"}).Once()
				mockConfig.On("Get", "cors.exposed_headers").Return([]string{"Goravel"}).Once()
				mockConfig.On("GetInt", "cors.max_age").Return(0).Once()
				mockConfig.On("GetBool", "cors.supports_credentials").Return(false).Once()
				ConfigFacade = mockConfig
			},
			assert: func() {
				assert.Equal(t, http.StatusNoContent, resp.Code)
				assert.Equal(t, "POST", resp.Header().Get("Access-Control-Allow-Methods"))
				assert.Equal(t, "*", resp.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "", resp.Header().Get("Access-Control-Allow-Headers"))
				assert.Equal(t, "", resp.Header().Get("Access-Control-Expose-Headers"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()

			g, err := NewRoute(mockConfig, nil)
			assert.Nil(t, err)
			g.GlobalMiddleware()
			g.Post("/any/{id}", func(ctx contractshttp.Context) contractshttp.Response {
				return ctx.Response().Success().Json(contractshttp.Json{
					"id": ctx.Request().Input("id"),
				})
			})

			resp = httptest.NewRecorder()
			req, err := http.NewRequest("OPTIONS", "/any/1", nil)
			assert.Nil(t, err)
			req.Header.Set("Origin", "https://goravel.dev")
			req.Header.Set("Access-Control-Request-Method", "POST")
			g.ServeHTTP(resp, req)

			test.assert()

			mockConfig.AssertExpectations(t)
		})
	}
}
