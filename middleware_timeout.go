package gin

import (
	"context"
	"net/http"
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/errors"
)

// Timeout creates middleware to set a timeout for a request
func Timeout(timeout time.Duration) contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		timeoutCtx, cancel := context.WithTimeout(ctx.Context(), timeout)
		defer cancel()

		ctx.WithContext(timeoutCtx)

		done := make(chan struct{})

		go func() {
			defer func() {
				if err := recover(); err != nil {
					globalRecoverCallback(ctx, err)
				}

				close(done)
			}()
			ctx.Request().Next()
		}()

		select {
		case <-done:
		case <-timeoutCtx.Done():
			if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
				ctx.Request().AbortWithStatus(http.StatusRequestTimeout)
			}
		}
	}
}
