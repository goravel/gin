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
				if r := recover(); r != nil {
					Recover(func(ctx *gin.Context, err interface{}) {
						if err != nil {
							ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
						}
					})(ctx.Request().Origin().(*gin.Context), r)
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
