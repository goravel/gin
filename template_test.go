package gin

import (
	"os"
	"testing"

	"github.com/goravel/framework/support/file"
	"github.com/stretchr/testify/assert"
)

func TestNewTemplate(t *testing.T) {
	tests := []struct {
		name         string
		setup        func()
		expectError  error
		expectRender bool
	}{
		{
			name:  "resources/views directory not found",
			setup: func() {},
		},
		{
			name: "resources/views directory is empty",
			setup: func() {
				assert.Nil(t, os.MkdirAll("resources/views", os.ModePerm))
			},
		},
		{
			name: "resources/views directory is not empty",
			setup: func() {
				assert.Nil(t, file.Create("resources/views/index.html", "Hello, World!"))
			},
			expectRender: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()
			r, err := NewTemplate(RenderOptions{})
			assert.Equal(t, test.expectRender, r != nil)
			assert.Equal(t, test.expectError, err)
			assert.Nil(t, file.Remove("resources"))
		})
	}
}
