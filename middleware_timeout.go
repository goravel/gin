package gin

import (
	"time"

	gintimeout "github.com/gin-contrib/timeout"
	contractshttp "github.com/goravel/framework/contracts/http"
)

// Timeout creates middleware to set a timeout for a request
func Timeout(timeout time.Duration) contractshttp.Middleware {
	timeoutMiddleware := gintimeout.New(gintimeout.WithTimeout(timeout))

	return func(ctx contractshttp.Context) {
		if timeout <= 0 {
			ctx.Request().Next()
			return
		}

		timeoutMiddleware(ctx.(*Context).Instance())
	}
}
