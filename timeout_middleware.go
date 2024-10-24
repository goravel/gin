package gin

import (
	"context"
	"net/http"
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/config"
)

// TimeoutMiddleware creates middleware to set a timeout for a request
func TimeoutMiddleware(config config.Config) contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		timeout := time.Duration(config.GetInt("http.timeout_request", 1)) * time.Second
		timeoutCtx, cancel := context.WithTimeout(ctx.Context(), timeout)
		defer cancel()

		ctx.WithContext(timeoutCtx)
		
		go ctx.Request().Next()

		select {
		case <-ctx.Request().Origin().Context().Done():
			if timeoutCtx.Err() == context.DeadlineExceeded {
				ctx.Response().Writer().WriteHeader(http.StatusGatewayTimeout)
				_, _ = ctx.Response().Writer().Write([]byte("Request timed out"))
				ctx.Request().AbortWithStatus(http.StatusGatewayTimeout)
			}
		}
	}
}
