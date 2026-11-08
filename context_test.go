package gin

import (
	"context"
	"net/http/httptest"
	"runtime"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContext(t *testing.T) {
	type customType struct{}
	var customTypeKey customType

	ginCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ginCtx.Request = httptest.NewRequest("GET", "/", nil)
	httpCtx := NewContext(ginCtx)
	httpCtx.WithValue("Hello", "world")
	httpCtx.WithValue("Hi", "Goravel")
	httpCtx.WithValue(customTypeKey, "halo")

	userContext := context.Background()
	//nolint:all
	userContext = context.WithValue(userContext, "user_a", "b")
	httpCtx.WithContext(userContext)

	httpCtx.WithValue(1, "one")
	httpCtx.WithValue(2.2, "two point two")

	assert.Equal(t, "world", httpCtx.Value("Hello"))
	assert.Equal(t, "Goravel", httpCtx.Value("Hi"))
	assert.Equal(t, "halo", httpCtx.Value(customTypeKey))
	assert.Equal(t, "one", httpCtx.Value(1))
	assert.Equal(t, "two point two", httpCtx.Value(2.2))

	// The value of UserContext can't be covered.
	assert.Equal(t, "b", httpCtx.Value("user_a"))

	ctx := httpCtx.Context()

	assert.Equal(t, "world", ctx.Value("Hello"))
	assert.Equal(t, "Goravel", ctx.Value("Hi"))
	assert.Equal(t, "halo", ctx.Value(customTypeKey))
	assert.Equal(t, "b", ctx.Value("user_a"))
	assert.Equal(t, "one", ctx.Value(1))
	assert.Equal(t, "two point two", ctx.Value(2.2))
}

func TestContext_Concurrent(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T) (writer, reader *Context)
	}{
		{
			name: "same context",
			setup: func(t *testing.T) (*Context, *Context) {
				httpCtx := Background().(*Context)
				return httpCtx, httpCtx
			},
		},
		{
			name: "two wrappers over one gin.Context",
			setup: func(t *testing.T) (*Context, *Context) {
				base := Background().(*Context)
				writer := NewContext(base.Instance())
				reader := NewContext(base.Instance())

				require.Same(t, writer.values, reader.values)

				writer.WithValue("shared", "value")
				require.Equal(t, "value", reader.Value("shared"))

				return writer, reader
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer, reader := tt.setup(t)
			runConcurrentContext(t, writer, reader)
		})
	}

	t.Run("both writes are kept", func(t *testing.T) {
		ginCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ginCtx.Request = httptest.NewRequest("GET", "/", nil)

		writer := NewContext(ginCtx)
		reader := NewContext(ginCtx)
		require.Same(t, writer.values, reader.values)

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			writer.WithValue("first", "writer")
		}()
		go func() {
			defer wg.Done()
			reader.WithValue("second", "reader")
		}()

		wg.Wait()

		// Before the fix each wrapper could publish its own map on its first
		// write, so one of these values was lost.
		require.Equal(t, "writer", reader.Value("first"))
		require.Equal(t, "reader", writer.Value("second"))
	})

	t.Run("concurrent construction shares one holder", func(t *testing.T) {
		ginCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ginCtx.Request = httptest.NewRequest("GET", "/", nil)

		const goroutines = 10

		contexts := make([]*Context, goroutines)
		var wg sync.WaitGroup
		start := make(chan struct{})

		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				<-start
				ctx := NewContext(ginCtx)
				ctx.WithValue(idx, "v")
				contexts[idx] = ctx
			}(i)
		}

		close(start)
		wg.Wait()

		require.NotNil(t, contexts[0].values)
		for _, ctx := range contexts {
			require.Same(t, contexts[0].values, ctx.values)
			for i := 0; i < goroutines; i++ {
				require.Equal(t, "v", ctx.Value(i))
				require.Equal(t, "v", ctx.Context().Value(i))
			}
		}
	})

	t.Run("pooled Context does not keep the old request's values", func(t *testing.T) {
		oldGinCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
		oldGinCtx.Request = httptest.NewRequest("GET", "/old", nil)
		newGinCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
		newGinCtx.Request = httptest.NewRequest("GET", "/new", nil)

		stale := &Context{instance: oldGinCtx, values: &requestValues{}}
		stale.values.store("old", "value")

		oldPool := contextPool
		contextPool = &sync.Pool{New: func() any { return stale }}
		defer func() { contextPool = oldPool }()

		ctx := NewContext(newGinCtx)
		require.Same(t, stale, ctx)
		require.Nil(t, ctx.Value("old"))

		ctx.WithValue("new", "value")
		holder, ok := newGinCtx.Get(contextKey)
		require.True(t, ok)
		require.Same(t, ctx.values, holder)
		require.Equal(t, "value", ctx.Value("new"))
	})
}

// runConcurrentContext writes values through WithValue while other goroutines
// read the context the way a handler or a view render would.
func runConcurrentContext(t *testing.T, writer, reader *Context) {
	t.Helper()

	const (
		goroutines = 10
		iterations = 100
	)

	var wg sync.WaitGroup
	start := make(chan struct{})

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			<-start
			for j := 0; j < iterations; j++ {
				writer.WithValue(idx, "value")
				runtime.Gosched()
			}
		}(i)
	}

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			<-start
			for j := 0; j < iterations; j++ {
				_ = reader.Context()
				_ = reader.Done()
				_ = reader.Err()
				_, _ = reader.Deadline()
				_ = reader.Value(0)
			}
		}()
	}

	close(start)
	wg.Wait()

	// Each writer stores under its own key; every key must be readable.
	for i := 0; i < goroutines; i++ {
		require.Equal(t, "value", reader.Value(i))
		require.Equal(t, "value", reader.Context().Value(i))
	}
}
