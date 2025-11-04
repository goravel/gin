package gin

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	contractshttp "github.com/goravel/framework/contracts/http"
	mocksconfig "github.com/goravel/framework/mocks/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTls(t *testing.T) {
	var (
		mockConfig       *mocksconfig.Config
		responseRecorder *httptest.ResponseRecorder
	)
	beforeEach := func() {
		mockConfig = mocksconfig.NewConfig(t)
	}

	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "not use tls",
			setup: func() {
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

			mockConfig.EXPECT().GetInt("http.drivers.gin.body_limit", 4096).Return(4096).Once()
			mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
			mockConfig.EXPECT().Get("http.drivers.gin.template").Return(nil).Once()

			route := &Route{
				config: mockConfig,
				driver: "gin",
			}
			err := route.init([]contractshttp.Middleware{Tls()})
			require.Nil(t, err)

			route.Any("/any/{id}", func(ctx contractshttp.Context) contractshttp.Response {
				return ctx.Response().Success().Json(contractshttp.Json{
					"id": ctx.Request().Input("id"),
				})
			})

			responseRecorder = httptest.NewRecorder()
			req, err := http.NewRequest("POST", "/any/1", nil)
			req.TLS = &tls.ConnectionState{}
			assert.Nil(t, err)

			route.ServeHTTP(responseRecorder, req)
			assert.Equal(t, http.StatusOK, responseRecorder.Code)
		})
	}
}
