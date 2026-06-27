package gin

import (
	"fmt"
	"time"

	gintimeout "github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
	contractshttp "github.com/goravel/framework/contracts/http"
)

type TimeoutMiddleware struct {
	Duration time.Duration
	tm       gin.HandlerFunc
}

func (t *TimeoutMiddleware) Signature() string {
	return fmt.Sprintf("goravel:timeout:%v", t.Duration)
}

func (t *TimeoutMiddleware) Handle(ctx contractshttp.Context) {
	if t.Duration <= 0 {
		ctx.Request().Next()
		return
	}

	t.tm(ctx.(*Context).Instance())
}

// Timeout creates middleware to set a timeout for a request
func Timeout(timeout time.Duration) contractshttp.Middleware {
	return &TimeoutMiddleware{
		Duration: timeout,
		tm:       gintimeout.New(gintimeout.WithTimeout(timeout)),
	}
}
