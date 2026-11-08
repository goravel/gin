package gin

import (
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	contractshttp "github.com/goravel/framework/contracts/http"
)

const (
	contextKey        = "goravel_contextKey"
	responseOriginKey = "goravel_responseOrigin"
	sessionKey        = "goravel_session"
)

var (
	internalContextKeys = []any{
		contextKey,
		responseOriginKey,
		sessionKey,
	}
)

func Background() contractshttp.Context {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(nethttp.MethodGet, "/", nil)
	return NewContext(ctx)
}

var contextPool = sync.Pool{New: func() any {
	return &Context{}
}}

// requestValues holds the values shared by every *Context wrapping one
// gin.Context. The lock lives here because middleware and handler each get
// their own *Context over the same gin.Context.
type requestValues struct {
	mu     sync.RWMutex
	values map[any]any
}

type requestValue struct {
	key   any
	value any
}

func newRequestValues() *requestValues {
	return &requestValues{}
}

func (s *requestValues) load(key any) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.values[key]
	return value, ok
}

func (s *requestValues) store(key any, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.values == nil {
		s.values = make(map[any]any)
	}
	s.values[key] = value
}

// snapshot copies the entries under the read lock so the context chain can be
// built without holding it.
func (s *requestValues) snapshot() []requestValue {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := make([]requestValue, 0, len(s.values))
	for key, value := range s.values {
		entries = append(entries, requestValue{key: key, value: value})
	}
	return entries
}

type Context struct {
	instance *gin.Context
	request  contractshttp.ContextRequest
	response contractshttp.ContextResponse
	values   *requestValues
}

func NewContext(c *gin.Context) *Context {
	ctx := contextPool.Get().(*Context)
	// Some test-created gin contexts do not carry a request, but Goravel's
	// context accessors always read through Request.Context().
	if c.Request == nil {
		c.Request = httptest.NewRequest(nethttp.MethodGet, "/", nil)
	}
	ctx.instance = c
	// The pool may hand back a Context from another request, and every wrapper
	// of one gin.Context must share a single holder.
	ctx.values = nil
	if existing, ok := c.Get(contextKey); ok {
		ctx.values, _ = existing.(*requestValues)
	}
	if ctx.values == nil {
		ctx.values = newRequestValues()
		c.Set(contextKey, ctx.values)
	}
	return ctx
}

func (c *Context) Request() contractshttp.ContextRequest {
	if c.request == nil {
		request := NewContextRequest(c, LogFacade, ValidationFacade)
		c.request = request
	}

	return c.request
}

func (c *Context) Response() contractshttp.ContextResponse {
	if c.response == nil {
		response := NewContextResponse(c.instance, &BodyWriter{ResponseWriter: c.instance.Writer})
		c.response = response
	}

	responseOrigin := c.Value(responseOriginKey)
	if responseOrigin != nil {
		c.response.(*ContextResponse).origin = responseOrigin.(contractshttp.ResponseOrigin)
	}

	return c.response
}

func (c *Context) WithValue(key any, value any) {
	if c.values == nil {
		// Only NewContext produces a usable Context: c.instance is nil on a zero value.
		c.values = newRequestValues()
		if c.instance != nil {
			c.instance.Set(contextKey, c.values)
		}
	}
	c.values.store(key, value)
}

func (c *Context) WithContext(ctx context.Context) {
	// Changing the request context to a new context
	c.instance.Request = c.instance.Request.WithContext(ctx)
}

func (c *Context) Context() context.Context {
	ctx := c.instance.Request.Context()
	if c.values == nil {
		return ctx
	}
	for _, entry := range c.values.snapshot() {
		skip := false
		for _, internalContextKey := range internalContextKeys {
			if entry.key == internalContextKey {
				skip = true
			}
		}

		if !skip {
			ctx = context.WithValue(ctx, entry.key, entry.value)
		}
	}

	return ctx
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.instance.Request.Context().Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.instance.Request.Context().Done()
}

func (c *Context) Err() error {
	return c.instance.Request.Context().Err()
}

func (c *Context) Value(key any) any {
	if c.values != nil {
		if value, ok := c.values.load(key); ok {
			return value
		}
	}

	return c.instance.Request.Context().Value(key)
}

func (c *Context) Instance() *gin.Context {
	return c.instance
}
