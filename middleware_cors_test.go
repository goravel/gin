package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	contractshttp "github.com/goravel/framework/contracts/http"
	configmocks "github.com/goravel/framework/mocks/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCors(t *testing.T) {
	var (
		mockConfig *configmocks.Config
		resp       *httptest.ResponseRecorder
	)
	beforeEach := func() {
		mockConfig = configmocks.NewConfig(t)
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
				mockConfig.On("GetInt", "http.drivers.gin.body_limit", 4096).Return(4096).Once()
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
				mockConfig.On("GetInt", "http.drivers.gin.body_limit", 4096).Return(4096).Once()
				mockConfig.On("Get", "cors.paths").Return([]string{"api"}).Once()
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
				mockConfig.On("GetInt", "http.drivers.gin.body_limit", 4096).Return(4096).Once()
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
				mockConfig.On("GetInt", "http.drivers.gin.body_limit", 4096).Return(4096).Once()
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
				mockConfig.On("GetInt", "http.drivers.gin.body_limit", 4096).Return(4096).Once()
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
				mockConfig.On("GetInt", "http.drivers.gin.body_limit", 4096).Return(4096).Once()
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
				mockConfig.On("GetInt", "http.drivers.gin.body_limit", 4096).Return(4096).Once()
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
				mockConfig.On("GetInt", "http.drivers.gin.body_limit", 4096).Return(4096).Once()
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

			mockConfig.EXPECT().Get("http.drivers.gin.template").Return(nil).Once()

			route := &Route{
				config: mockConfig,
				driver: "gin",
			}
			err := route.init([]contractshttp.Middleware{Cors()})
			require.Nil(t, err)

			route.Post("/any/{id}", func(ctx contractshttp.Context) contractshttp.Response {
				return ctx.Response().Success().Json(contractshttp.Json{
					"id": ctx.Request().Input("id"),
				})
			})

			resp = httptest.NewRecorder()
			req, err := http.NewRequest("OPTIONS", "/any/1", nil)
			assert.Nil(t, err)

			req.Header.Set("Origin", "https://goravel.dev")
			req.Header.Set("Access-Control-Request-Method", "POST")
			route.ServeHTTP(resp, req)

			test.assert()
		})
	}
}
