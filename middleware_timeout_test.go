package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	contractshttp "github.com/goravel/framework/contracts/http"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mockslog "github.com/goravel/framework/mocks/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeoutMiddleware(t *testing.T) {
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	mockConfig.EXPECT().GetInt("http.drivers.gin.body_limit", 4096).Return(4096).Once()

	route, err := NewRoute(mockConfig, nil)
	require.NoError(t, err)

	route.Middleware(Timeout(1*time.Second)).Get("/timeout", func(ctx contractshttp.Context) contractshttp.Response {
		time.Sleep(2 * time.Second)
		return nil
	})

	route.Middleware(Timeout(1*time.Second)).Get("/normal", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().Success().String("normal")
	})

	route.Middleware(Timeout(1*time.Second)).Get("/panic", func(ctx contractshttp.Context) contractshttp.Response {
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

	// Test with default recover callback
	mockLog := mockslog.NewLog(t)
	mockLog.EXPECT().Error(1).Once()
	LogFacade = mockLog

	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/panic", nil)
	require.NoError(t, err)

	route.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "{\"error\":\"Internal Server Error\"}", w.Body.String())

	// Test with custom recover callback
	globalRecover := func(ctx contractshttp.Context, err any) {
		ctx.Request().AbortWithStatusJson(http.StatusInternalServerError, gin.H{"error": "Internal Panic"})
	}
	route.Recover(globalRecover)

	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/panic", nil)
	require.NoError(t, err)

	route.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "{\"error\":\"Internal Panic\"}", w.Body.String())
}
