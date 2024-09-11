package gin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	httpCtx := Background()
	httpCtx.WithValue("Hello", "world")
	httpCtx.WithValue("Hi", "Goravel")
	ctx := httpCtx.Context()
	assert.Equal(t, ctx.Value("Hello").(string), "world")
	assert.Equal(t, ctx.Value("Hi").(string), "Goravel")
}

func TestContextWithCustomKeyType(t *testing.T) {
	type customKeyType struct{}
	var customKey customKeyType
	var customKeyTwo customKeyType

	httpCtx := Background()
	httpCtx.WithValue(customKey, "hello")
	httpCtx.WithValue(customKeyTwo, "world")

	assert.Equal(t, httpCtx.Value(customKey), "hello")
	assert.Equal(t, httpCtx.Value(customKeyTwo), "world")
}
