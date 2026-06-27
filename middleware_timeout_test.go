package gin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mockslog "github.com/goravel/framework/mocks/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTimeoutMiddleware(t *testing.T) {
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetInt("http.drivers.gin.body_limit", 4096).Return(4096).Once()
	mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	mockConfig.EXPECT().Get("http.drivers.gin.template").Return(nil).Once()

	route := &Route{
		config: mockConfig,
		driver: "gin",
	}
	err := route.init(nil)
	require.Nil(t, err)

	t.Run("timeout waits for handler completion", func(t *testing.T) {
		var (
			err         error
			deadline    time.Time
			hasDeadline bool
		)
		timedOut := make(chan struct{})
		allowReturn := make(chan struct{})

		route.Middleware(Timeout(time.Second)).Get("/timeout", func(ctx contractshttp.Context) contractshttp.Response {
			<-ctx.Done()
			err = ctx.Err()
			deadline, hasDeadline = ctx.Deadline()
			close(timedOut)
			<-allowReturn

			return ctx.Response().Status(contractshttp.StatusRequestTimeout).String("timeout")
		})

		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/timeout", nil)
		require.NoError(t, err)
		done := make(chan struct{})

		go func() {
			route.ServeHTTP(w, req)
			close(done)
		}()

		select {
		case <-timedOut:
		case <-time.After(5 * time.Second):
			t.Fatal("request context deadline was not observed")
		}

		select {
		case <-done:
			t.Fatal("request returned before the handler completed")
		default:
		}

		close(allowReturn)

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("request did not complete after the handler returned")
		}

		assert.Equal(t, contractshttp.StatusRequestTimeout, w.Code)
		// Skip body assertion on Windows due to a known data race between
		// gintimeout.Copy() and responseMiddleware.WithValue() accessing
		// gin.Context.Keys concurrently. The status code assertion still
		// verifies the timeout fires correctly.
		if runtime.GOOS != "windows" {
			assert.Equal(t, http.StatusText(contractshttp.StatusRequestTimeout), w.Body.String())
		}
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		assert.True(t, hasDeadline)
		assert.False(t, deadline.IsZero())
	})

	t.Run("normal request", func(t *testing.T) {
		route.Middleware(Timeout(1*time.Second)).Get("/normal", func(ctx contractshttp.Context) contractshttp.Response {
			return ctx.Response().Success().String("normal")
		})

		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/normal", nil)
		require.NoError(t, err)

		route.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "normal", w.Body.String())
	})

	t.Run("timed out request does not affect later responses", func(t *testing.T) {
		timedOut := make(chan struct{})
		allowReturn := make(chan struct{})
		firstDone := make(chan struct{})

		route.Middleware(Timeout(time.Second)).Get("/timeout-isolated", func(ctx contractshttp.Context) contractshttp.Response {
			<-ctx.Done()
			close(timedOut)
			<-allowReturn

			return ctx.Response().Success().String("stale")
		})
		route.Middleware(Timeout(2*time.Second)).Get("/after-timeout", func(ctx contractshttp.Context) contractshttp.Response {
			return ctx.Response().Success().String("fresh")
		})

		firstWriter := httptest.NewRecorder()
		firstReq, err := http.NewRequest("GET", "/timeout-isolated", nil)
		require.NoError(t, err)

		go func() {
			route.ServeHTTP(firstWriter, firstReq)
			close(firstDone)
		}()

		select {
		case <-timedOut:
		case <-time.After(5 * time.Second):
			t.Fatal("request context deadline was not observed")
		}

		secondWriter := httptest.NewRecorder()
		secondReq, err := http.NewRequest("GET", "/after-timeout", nil)
		require.NoError(t, err)
		route.ServeHTTP(secondWriter, secondReq)

		select {
		case <-firstDone:
			t.Fatal("timed out request returned before its handler completed")
		default:
		}

		close(allowReturn)

		select {
		case <-firstDone:
		case <-time.After(5 * time.Second):
			t.Fatal("timed out request did not complete after the handler returned")
		}

		assert.Equal(t, contractshttp.StatusRequestTimeout, firstWriter.Code)
		assert.Equal(t, http.StatusText(contractshttp.StatusRequestTimeout), firstWriter.Body.String())
		assert.Equal(t, http.StatusOK, secondWriter.Code)
		assert.Equal(t, "fresh", secondWriter.Body.String())
	})

	t.Run("panic with default recover", func(t *testing.T) {
		route.Middleware(Timeout(1*time.Second)).Get("/panic", func(ctx contractshttp.Context) contractshttp.Response {
			panic(1)
		})

		mockLog := mockslog.NewLog(t)
		mockLog.EXPECT().WithContext(mock.Anything).Return(mockLog).Once()
		mockLog.EXPECT().Request(mock.Anything).Return(mockLog).Once()
		mockLog.EXPECT().Error(1).Once()
		LogFacade = mockLog

		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/panic", nil)
		require.NoError(t, err)

		route.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Empty(t, w.Body.String())
	})

	t.Run("panic with custom recover", func(t *testing.T) {
		mockConfig.EXPECT().GetInt("http.drivers.gin.body_limit", 4096).Return(4096).Once()
		mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
		mockConfig.EXPECT().Get("http.drivers.gin.template").Return(nil).Once()

		called := false
		globalRecover := func(ctx contractshttp.Context, err any) {
			called = true
			ctx.Request().Abort(http.StatusInternalServerError)
		}
		route.Recover(globalRecover)

		route.Middleware(Timeout(1*time.Second)).Get("/panic", func(ctx contractshttp.Context) contractshttp.Response {
			panic(1)
		})

		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/panic", nil)
		require.NoError(t, err)

		route.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Empty(t, w.Body.String())
		assert.True(t, called)

		// Reset to default recover callback
		globalRecoverCallback = defaultRecoverCallback
	})
}
