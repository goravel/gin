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

// holderMu makes the holder lookup and publish in NewContext atomic, so two
// wrappers of one gin.Context cannot install separate holders.
var holderMu sync.Mutex

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

func (rv *requestValues) load(key any) (any, bool) {
	rv.mu.RLock()
	defer rv.mu.RUnlock()

	value, ok := rv.values[key]
	return value, ok
}

func (rv *requestValues) store(key any, value any) {
	rv.mu.Lock()
	defer rv.mu.Unlock()

	if rv.values == nil {
		rv.values = make(map[any]any)
	}
	rv.values[key] = value
}

// snapshot copies the entries under the read lock so the context chain can be
// built without holding it.
func (rv *requestValues) snapshot() []requestValue {
	rv.mu.RLock()
	defer rv.mu.RUnlock()

	entries := make([]requestValue, 0, len(rv.values))
	for key, value := range rv.values {
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
	holderMu.Lock()
	defer holderMu.Unlock()
	existing, _ := c.Get(contextKey)
	ctx.values, _ = existing.(*requestValues)
	if ctx.values == nil {
		ctx.values = &requestValues{}
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
	// NewContext always sets values; a nil holder here is a programming error.
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
	if value, ok := valueOf(c.instance, key); ok {
		return value
	}

	return c.instance.Request.Context().Value(key)
}

// valueOf reads a request value from the holder shared through c.
func valueOf(c *gin.Context, key any) (any, bool) {
	if c == nil {
		return nil, false
	}

	v, ok := c.Get(contextKey)
	if !ok {
		return nil, false
	}
	values, ok := v.(*requestValues)
	if !ok {
		return nil, false
	}

	return values.load(key)
}

func (c *Context) Instance() *gin.Context {
	return c.instance
}
