package gin

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	contractshttp "github.com/goravel/framework/contracts/http"
	configmocks "github.com/goravel/framework/mocks/config"
	"github.com/stretchr/testify/assert"
)

func TestTls(t *testing.T) {
	var (
		mockConfig       *configmocks.Config
		responseRecorder *httptest.ResponseRecorder
	)
	beforeEach := func() {
		mockConfig = &configmocks.Config{}
	}

	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "not use tls",
			setup: func() {
				mockConfig.On("GetBool", "app.debug").Return(true).Once()
				mockConfig.On("GetInt", "http.drivers.gin.body_limit", 4096).Return(4096).Once()
				mockConfig.On("GetInt", "http.drivers.gin.header_limit", 4096).Return(4096).Once()
				mockConfig.On("Get", "cors.paths").Return([]string{}).Once()
				mockConfig.On("GetString", "http.tls.host").Return("").Once()
				mockConfig.On("GetString", "http.tls.port").Return("").Once()
				mockConfig.On("GetString", "http.tls.ssl.cert").Return("").Once()
				mockConfig.On("GetString", "http.tls.ssl.key").Return("").Once()
				ConfigFacade = mockConfig
			},
		},
		{
			name: "use tls",
			setup: func() {
				mockConfig.On("GetBool", "app.debug").Return(true).Once()
				mockConfig.On("GetInt", "http.drivers.gin.body_limit", 4096).Return(4096).Once()
				mockConfig.On("GetInt", "http.drivers.gin.header_limit", 4096).Return(4096).Once()
				mockConfig.On("Get", "cors.paths").Return([]string{}).Once()
				mockConfig.On("GetString", "http.tls.host").Return("127.0.0.1").Once()
				mockConfig.On("GetString", "http.tls.port").Return("3000").Once()
				mockConfig.On("GetString", "http.tls.ssl.cert").Return("test_ca.crt").Once()
				mockConfig.On("GetString", "http.tls.ssl.key").Return("test_ca.key").Once()
				ConfigFacade = mockConfig
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
			g.Any("/any/{id}", func(ctx contractshttp.Context) contractshttp.Response {
				return ctx.Response().Success().Json(contractshttp.Json{
					"id": ctx.Request().Input("id"),
				})
			})

			responseRecorder = httptest.NewRecorder()
			req, err := http.NewRequest("POST", "/any/1", nil)
			req.TLS = &tls.ConnectionState{}
			assert.Nil(t, err)
			g.ServeHTTP(responseRecorder, req)
			assert.Equal(t, http.StatusOK, responseRecorder.Code)

			mockConfig.AssertExpectations(t)
		})
	}
}
