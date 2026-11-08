package gin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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

	// gin-contrib/timeout picks between "handler finished" and "timer fired"
	// with a select. When the handler returns right after the deadline both
	// cases are ready and the finish branch can win: it flushes an empty buffer
	// and returns without writing the timeout response, leaving the client with
	// gin's implicit 200 instead of 408. Write that response ourselves; the
	// Written and Status checks keep a response the application already
	// produced (body or status only) untouched.
	if errors.Is(instance.Request.Context().Err(), context.DeadlineExceeded) &&
		!writer.Written() && writer.Status() == http.StatusOK {
		instance.String(http.StatusRequestTimeout, http.StatusText(http.StatusRequestTimeout))
	}
}

// Timeout creates middleware to set a timeout for a request
func Timeout(timeout time.Duration) contractshttp.Middleware {
	return &TimeoutMiddleware{
		Duration: timeout,
		tm:       gintimeout.New(gintimeout.WithTimeout(timeout)),
	}
}
