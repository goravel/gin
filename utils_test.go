package gin

import (
	"testing"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/stretchr/testify/assert"
)

func TestBracketToColon(t *testing.T) {
	assert.Equal(t, "/:id/:name", bracketToColon("/{id}/{name}"))
}

func TestColonToBracket(t *testing.T) {
	assert.Equal(t, "/{id}/{name}", colonToBracket("/:id/:name"))
}

type testMw struct {
	sig string
}

func (m *testMw) Signature() string     { return m.sig }
func (m *testMw) Handle(ctx contractshttp.Context) {}

func TestIsSameMiddleware(t *testing.T) {
	mw1 := &testMw{sig: "auth"}
	mw2 := &testMw{sig: "auth"}
	assert.True(t, isSameMiddleware(mw1, mw2))

	mw3 := &testMw{sig: "log"}
	assert.False(t, isSameMiddleware(mw1, mw3))

	assert.False(t, isSameMiddleware(nil, mw1))
	assert.False(t, isSameMiddleware(nil, nil))

	assert.False(t, isSameMiddleware("not a middleware", mw1))
}
