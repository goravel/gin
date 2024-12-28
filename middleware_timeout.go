package gin

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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
					if globalRecoverCallback != nil {
						globalRecoverCallback(ctx, err)
					} else {
						ctx.Request().AbortWithStatusJson(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
					}
				}
				close(done)
			}()
			ctx.Request().Next()
		}()

		select {
		case <-done:
		case <-timeoutCtx.Done():
			if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
				ctx.Request().AbortWithStatus(http.StatusGatewayTimeout)
			}
		}
	}
}
