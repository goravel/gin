package gin

import (
	"context"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/goravel/framework/contracts/http"
)

func Background() http.Context {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	return NewContext(ctx)
}

type Ctx context.Context

type Context struct {
	Ctx
	instance *gin.Context
	request  http.ContextRequest
}

func NewContext(ctx *gin.Context) http.Context {
	return &Context{instance: ctx, Ctx: context.Background()}
}

func (c *Context) Request() http.ContextRequest {
	if c.request == nil {
		c.request = NewContextRequest(c, LogFacade, ValidationFacade)
	}

	return c.request
}

func (c *Context) Response() http.ContextResponse {
	responseOrigin := c.Value("responseOrigin")
	if responseOrigin != nil {
		return NewContextResponse(c.instance, responseOrigin.(http.ResponseOrigin))
	}

	return NewContextResponse(c.instance, &BodyWriter{ResponseWriter: c.instance.Writer})
}

func (c *Context) WithValue(key any, value any) {
	//nolint:all
	c.Ctx = context.WithValue(c.Ctx, key, value)
}

func (c *Context) Context() context.Context {
	for key, value := range c.instance.Keys {
		// nolint
		c.Ctx = context.WithValue(c.Ctx, key, value)
	}

	return c.Ctx
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.instance.Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.instance.Done()
}

func (c *Context) Err() error {
	return c.instance.Err()
}

func (c *Context) Value(key any) any {
	return c.Ctx.Value(key)
}

func (c *Context) Instance() *gin.Context {
	return c.instance
}
