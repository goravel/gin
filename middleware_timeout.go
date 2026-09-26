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

	// gin-contrib/timeout swaps in a buffered writer and does not restore the
	// original one after the chain finishes. gin writes its default 404 after
	// the global middleware return, so without restoring it the status is lost
	// and the client gets an empty 200.
	instance := ctx.(*Context).Instance()
	writer := instance.Writer
	t.tm(instance)
	instance.Writer = writer
}

// Timeout creates middleware to set a timeout for a request
func Timeout(timeout time.Duration) contractshttp.Middleware {
	return &TimeoutMiddleware{
		Duration: timeout,
		tm:       gintimeout.New(gintimeout.WithTimeout(timeout)),
	}
}
