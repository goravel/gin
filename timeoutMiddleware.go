package gin

import (
	"context"
	"net/http"
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"
)

// TimeoutMiddleware creates middleware to set a timeout for a request
func TimeoutMiddleware(timeout time.Duration) contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		timeoutCtx, cancel := context.WithTimeout(ctx.Context(), timeout)
		defer cancel()

		// Refreshing the request context with a timeout
		ctx.WithContext(timeoutCtx)

		// We start executing the request in a new goroutine
		done := make(chan struct{})
		go func() {
			defer close(done)
			ctx.Request().Next()
		}()

		select {
		case <-done:
			// The request completed before the timeout expired
		case <-timeoutCtx.Done():
			// After the timeout expires, return the status 504 Gateway Timeout
			if timeoutCtx.Err() == context.DeadlineExceeded {
				ctx.Response().Writer().WriteHeader(http.StatusGatewayTimeout)
				_, _ = ctx.Response().Writer().Write([]byte("Request timed out"))
			}
		}
	}
}
