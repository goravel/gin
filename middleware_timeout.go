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
            defer close(done)

            defer func() {
                if err := recover(); err != nil {
                    if globalRecoverCallback != nil {
                        globalRecoverCallback(ctx.Context(), err)
                    } else {
                        ctx.Request().AbortWithStatusJson(http.StatusInternalServerError, map[string]interface{}{
                            "error": "Internal Server Error",
                        })
                    }
                }
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
