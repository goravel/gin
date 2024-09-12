package gin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	type customKeyType struct{}
	var customKey customKeyType

	httpCtx := Background()
	httpCtx.WithValue("Hello", "world")
	httpCtx.WithValue("Hi", "Goravel")
	httpCtx.WithValue(1, "one")
	httpCtx.WithValue(1.1, "one point one")
	httpCtx.WithValue(customKey, "hello")
	ctx := httpCtx.Context()
	assert.Equal(t, ctx.Value("Hello").(string), "world")
	assert.Equal(t, ctx.Value("Hi").(string), "Goravel")
	assert.Equal(t, ctx.Value(1).(string), "one")
	assert.Equal(t, ctx.Value(1.1).(string), "one point one")
	assert.Equal(t, ctx.Value(customKey), "hello")
}
