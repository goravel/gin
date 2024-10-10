package gin

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	// even with the same underlying empty struct, Go will still distinguish between
	//  each type declaration
	type customType struct{}
	type anotherCustomType struct{}
	var customTypeKey customType
	var anotherCustomTypeKey anotherCustomType

	httpCtx := Background()
	httpCtx.WithValue("Hello", "world")
	httpCtx.WithValue("Hi", "Goravel")
	httpCtx.WithValue(customTypeKey, "halo")
	httpCtx.WithValue(anotherCustomTypeKey, "hola")
	httpCtx.WithValue(1, "one")
	httpCtx.WithValue(2.2, "two point two")

	ctx := httpCtx.Context()
	assert.Equal(t, "world", ctx.Value("Hello"))
	assert.Equal(t, "Goravel", ctx.Value("Hi"))
	assert.Equal(t, "halo", ctx.Value(customTypeKey))
	assert.Equal(t, "hola", ctx.Value(anotherCustomTypeKey))
	assert.Equal(t, "one", ctx.Value(1))
	assert.Equal(t, "two point two", ctx.Value(2.2))
}

func TestWithContext(t *testing.T) {
	httpCtx := Background()

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	httpCtx.WithContext(timeoutCtx)

	ctx := httpCtx.Context()
	assert.Equal(t, timeoutCtx, ctx)

	deadline, ok := ctx.Deadline()
	assert.True(t, ok, "Deadline should be set")
	assert.WithinDuration(t, time.Now().Add(2*time.Second), deadline, 50*time.Millisecond, "Deadline should be approximately 2 seconds from now")

	select {
	case <-ctx.Done():
		assert.Fail(t, "context should not be done yet")
	default:
		
	}
	time.Sleep(2 * time.Second)

	select {
	case <-ctx.Done():
		assert.Equal(t, context.DeadlineExceeded, ctx.Err(), "context should be exceeded")
	default:
		assert.Fail(t, "context should be done after timeout")
	}
}
