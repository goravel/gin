package gin

import (
	"context"
	"time"

	gintimeout "github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
	contractshttp "github.com/goravel/framework/contracts/http"
)

// Timeout creates middleware to set a timeout for a request
func Timeout(timeout time.Duration) contractshttp.Middleware {
	timeoutResponse := func(c *gin.Context) {
		c.Status(contractshttp.StatusRequestTimeout)
	}

	timeoutMiddleware := gintimeout.New(
		gintimeout.WithTimeout(timeout),
		gintimeout.WithResponse(timeoutResponse),
	)

	return func(ctx contractshttp.Context) {
		if timeout <= 0 {
			ctx.Request().Next()
			return
		}

		goravelCtx, ok := ctx.(*Context)
		if !ok {
			timeoutCtx, cancel := context.WithTimeout(ctx.Context(), timeout)
			defer cancel()

			ctx.WithContext(timeoutCtx)
			ctx.Request().Next()
			return
		}

		timeoutMiddleware(goravelCtx.Instance())
	}
}
