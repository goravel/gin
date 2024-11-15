package gin

import (
	"context"
	"net/http"
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/errors"
)

// Timeout creates middleware to set a timeout for a request
func Timeout(timeout time.Duration, recoverFunc func(ctx contractshttp.Context, err interface{})) contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		timeoutCtx, cancel := context.WithTimeout(ctx.Context(), timeout)
		defer cancel()

		ctx.WithContext(timeoutCtx)

		done := make(chan struct{})

		go func() {
			defer func() {
				if r := recover(); r != nil {
					// Use the custom recoverFunc if provided, otherwise fallback to default behavior
					if recoverFunc != nil {
						recoverFunc(ctx, r)
					} else {
						if LogFacade != nil {
							LogFacade.Request(ctx.Request()).Error(r)
						}
						_ = ctx.Response().Status(http.StatusInternalServerError).String("Internal Server Error").Render()
					}
				}

				close(done)
			}()

			ctx.Request().Next()
		}()

		select {
		case <-done:
		case <-ctx.Request().Origin().Context().Done():
			if errors.Is(ctx.Request().Origin().Context().Err(), context.DeadlineExceeded) {
				ctx.Request().AbortWithStatus(http.StatusGatewayTimeout)
			}
		}
	}
}
