package gin

import (
	"context"
	"errors"
	"fmt"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goravel/framework/contracts/http"
)

const goravelContextKey = "goravel_contextKey"

func Background() http.Context {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	return &Context{instance: ctx}
}

var contextPool = sync.Pool{New: func() any {
	return &Context{}
}}

type Context struct {
	instance *gin.Context
	request  http.ContextRequest
	response http.ContextResponse
}

func NewContext() http.Context {
	return contextPool.Get().(*Context)
}

func (c *Context) Request() http.ContextRequest {
	if c.request == nil {
		request := NewContextRequest().(*ContextRequest)
		httpBody, err := getHttpBody(c)
		if err != nil {
			LogFacade.Error(fmt.Sprintf("%+v", errors.Unwrap(err)))
		}
		request.ctx = c
		request.instance = c.instance
		request.httpBody = httpBody
		c.request = request
	}

	return c.request
}

func (c *Context) Response() http.ContextResponse {
	if c.response == nil {
		response := NewContextResponse().(*ContextResponse)
		response.instance = c.instance
		c.response = response
	}

	responseOrigin := c.Value("responseOrigin")
	if responseOrigin != nil {
		c.response.(*ContextResponse).origin = responseOrigin.(http.ResponseOrigin)
	} else {
		c.response.(*ContextResponse).origin = &BodyWriter{ResponseWriter: c.instance.Writer}
	}

	return c.response
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
