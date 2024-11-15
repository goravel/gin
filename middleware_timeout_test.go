package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mockslog "github.com/goravel/framework/mocks/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestiTmeoutMiddleware(t *testing.T) {
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	mockConfig.EXPECT().GetInt("http.drivers.gin.body_limit", 4096).Return(4096).Once()

	route, err := NewRoute(mockConfig, nil)
	require.NoError(t, err)

	route.Middleware(Timeout(1*time.Second, nil)).Get("/timeout", func(ctx contractshttp.Context) contractshttp.Response {
		time.Sleep(2 * time.Second)

		return ctx.Response().Success().String("timeout")
	})
	route.Middleware(Timeout(1*time.Second, nil)).Get("/normal", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Success().String("normal")
	})
	route.Middleware(Timeout(1*time.Second, nil)).Get("/panic", func(ctx contractshttp.Context) contractshttp.Response {
		panic(1)
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/timeout", nil)
	require.NoError(t, err)

	route.ServeHTTP(w, req)
	assert.Equal(t, http.StatusGatewayTimeout, w.Code)

	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/normal", nil)
	require.NoError(t, err)

	route.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "normal", w.Body.String())

	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/panic", nil)
	require.NoError(t, err)

	mockLog := mockslog.NewLog(t)
	mockLog.EXPECT().Request(mock.Anything).Return(mockLog).Once()
	mockLog.EXPECT().Error(mock.Anything).Once()
	LogFacade = mockLog

	route.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Internal Server Error", w.Body.String())
}
