package gin

import (
	"testing"

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
