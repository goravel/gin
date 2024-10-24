package gin

import (
	"context"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/goravel/framework/contracts/http"
)

const goravelContextKey = "goravel_contextKey"

func Background() http.Context {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	return NewContext(ctx)
}

type Context struct {
	instance *gin.Context
	request  http.ContextRequest
}

func NewContext(ctx *gin.Context) http.Context {
	return &Context{instance: ctx}
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
	goravelCtx := c.getGoravelCtx()
	goravelCtx[key] = value
	c.instance.Set(goravelContextKey, goravelCtx)
}

func (c *Context) WithContext(ctx context.Context) {
	// Changing the request context to a new context
	c.instance.Request = c.instance.Request.WithContext(ctx)
}

func (c *Context) Context() context.Context { return c }

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
	return c.getGoravelCtx()[key]
}

func (c *Context) Instance() *gin.Context {
	return c.instance
}

func (c *Context) getGoravelCtx() map[any]any {
	if val, exist := c.instance.Get(goravelContextKey); exist {
		if goravelCtxVal, ok := val.(map[any]any); ok {
			return goravelCtxVal
		}
	}
	return make(map[any]any)
}
