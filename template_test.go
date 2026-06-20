package gin

import (
	"bytes"
	"html/template"
	"os"
	"testing"

	mockslog "github.com/goravel/framework/mocks/log"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
				assert.Nil(t, os.MkdirAll(path.Resource("views"), os.ModePerm))
			},
		},
		{
			name: "resources/views directory is not empty",
			setup: func() {
				assert.Nil(t, file.PutContent(path.Resource("views", "index.html"), "Hello, World!"))
			},
			expectRender: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()
			r, err := NewTemplate(RenderOptions{}, nil)
			assert.Equal(t, test.expectRender, r != nil)
			assert.Equal(t, test.expectError, err)
			assert.Nil(t, file.Remove(path.Resource()))
		})
	}
}

func TestNewTemplate_PackageViews(t *testing.T) {
	pkgDir := path.Resource("pkg_custom")
	defer func() {
		assert.Nil(t, file.Remove(path.Resource()))
	}()

	assert.Nil(t, os.MkdirAll(pkgDir, os.ModePerm))
	assert.Nil(t, file.PutContent(path.Resource("pkg_custom", "welcome.tmpl"), `{{ define "welcome.tmpl" }}Welcome!{{ end }}`))

	r, err := NewTemplate(RenderOptions{}, []string{pkgDir})
	assert.Nil(t, err)
	assert.NotNil(t, r)

	var buf bytes.Buffer
	assert.Nil(t, r.Template.ExecuteTemplate(&buf, "welcome.tmpl", nil))
	assert.Equal(t, "Welcome!", buf.String())
}

func TestNewTemplate_AppOverridesPackage(t *testing.T) {
	viewsDir := path.Resource("views")
	pkgDir := path.Resource("pkg_override")
	defer func() {
		assert.Nil(t, file.Remove(path.Resource()))
	}()

	assert.Nil(t, os.MkdirAll(viewsDir, os.ModePerm))
	assert.Nil(t, os.MkdirAll(pkgDir, os.ModePerm))

	assert.Nil(t, file.PutContent(path.Resource("views", "page.tmpl"), `{{ define "page.tmpl" }}App Content{{ end }}`))
	assert.Nil(t, file.PutContent(path.Resource("pkg_override", "page.tmpl"), `{{ define "page.tmpl" }}Package Content{{ end }}`))

	r, err := NewTemplate(RenderOptions{}, []string{pkgDir})
	assert.Nil(t, err)
	assert.NotNil(t, r)

	var buf bytes.Buffer
	assert.Nil(t, r.Template.ExecuteTemplate(&buf, "page.tmpl", nil))
	assert.Equal(t, "App Content", buf.String())
}

func TestNewTemplate_PackageCollision(t *testing.T) {
	mockLog := mockslog.NewLog(t)
	LogFacade = mockLog

	// str has escape chars, so match on the format prefix only
	mockLog.EXPECT().Warningf("view collision: %q defined in %q and %q, using first", mock.Anything, mock.Anything, mock.Anything).Return().Once()

	dir1 := path.Resource("pkg1")
	dir2 := path.Resource("pkg2")
	defer func() {
		LogFacade = nil
		assert.Nil(t, file.Remove(path.Resource()))
	}()

	assert.Nil(t, os.MkdirAll(dir1, os.ModePerm))
	assert.Nil(t, os.MkdirAll(dir2, os.ModePerm))

	assert.Nil(t, file.PutContent(path.Resource("pkg1", "layout.tmpl"), `{{ define "layout.tmpl" }}First{{ end }}`))
	assert.Nil(t, file.PutContent(path.Resource("pkg2", "layout.tmpl"), `{{ define "layout.tmpl" }}Second{{ end }}`))

	r, err := NewTemplate(RenderOptions{}, []string{dir1, dir2})
	assert.Nil(t, err)
	assert.NotNil(t, r)

	var buf bytes.Buffer
	assert.Nil(t, r.Template.ExecuteTemplate(&buf, "layout.tmpl", nil))
	assert.Equal(t, "First", buf.String())

	mockLog.AssertExpectations(t)
}

func TestNewTemplate_ExtraViewsNilOrEmpty(t *testing.T) {
	viewsDir := path.Resource("views")
	defer func() {
		assert.Nil(t, file.Remove(path.Resource()))
	}()

	assert.Nil(t, os.MkdirAll(viewsDir, os.ModePerm))
	assert.Nil(t, file.PutContent(path.Resource("views", "hello.tmpl"), `{{ define "hello.tmpl" }}Hello{{ end }}`))

	t.Run("nil extraViews", func(t *testing.T) {
		r, err := NewTemplate(RenderOptions{}, nil)
		assert.Nil(t, err)
		assert.NotNil(t, r)

		var buf bytes.Buffer
		assert.Nil(t, r.Template.ExecuteTemplate(&buf, "hello.tmpl", nil))
		assert.Equal(t, "Hello", buf.String())
	})

	t.Run("empty extraViews", func(t *testing.T) {
		r, err := NewTemplate(RenderOptions{}, []string{})
		assert.Nil(t, err)
		assert.NotNil(t, r)

		var buf bytes.Buffer
		assert.Nil(t, r.Template.ExecuteTemplate(&buf, "hello.tmpl", nil))
		assert.Equal(t, "Hello", buf.String())
	})
}

func TestNewTemplate_NoViews(t *testing.T) {
	r, err := NewTemplate(RenderOptions{}, nil)
	assert.Nil(t, err)
	assert.Nil(t, r)
}

func TestDefaultTemplate(t *testing.T) {
	viewsDir := path.Resource("views")
	defer func() {
		assert.Nil(t, file.Remove(path.Resource()))
	}()

	assert.Nil(t, os.MkdirAll(viewsDir, os.ModePerm))
	assert.Nil(t, file.PutContent(path.Resource("views", "home.tmpl"), `{{ define "home.tmpl" }}Home{{ end }}`))

	r, err := DefaultTemplate(nil)
	assert.Nil(t, err)
	assert.NotNil(t, r)

	var buf bytes.Buffer
	assert.Nil(t, r.Template.ExecuteTemplate(&buf, "home.tmpl", nil))
	assert.Equal(t, "Home", buf.String())
}

func TestNewTemplate_CustomDelims(t *testing.T) {
	pkgDir := path.Resource("pkg_custom")
	defer func() {
		assert.Nil(t, file.Remove(path.Resource()))
	}()

	assert.Nil(t, os.MkdirAll(pkgDir, os.ModePerm))
	assert.Nil(t, file.PutContent(path.Resource("pkg_custom", "delim.tmpl"), `{[ define "delim.tmpl" ]}Custom{[ end ]}`))

	options := RenderOptions{
		Delims: &Delims{Left: "{[", Right: "]}"},
	}

	r, err := NewTemplate(options, []string{pkgDir})
	assert.Nil(t, err)
	assert.NotNil(t, r)

	var buf bytes.Buffer
	assert.Nil(t, r.Template.ExecuteTemplate(&buf, "delim.tmpl", nil))
	assert.Equal(t, "Custom", buf.String())
}

func TestNewTemplate_CustomFuncMap(t *testing.T) {
	pkgDir := path.Resource("pkg_func")
	defer func() {
		assert.Nil(t, file.Remove(path.Resource()))
	}()

	assert.Nil(t, os.MkdirAll(pkgDir, os.ModePerm))
	assert.Nil(t, file.PutContent(path.Resource("pkg_func", "func.tmpl"), `{{ define "func.tmpl" }}{{ greet "User" }}{{ end }}`))

	options := RenderOptions{
		FuncMap: template.FuncMap{
			"greet": func(name string) string { return "Hello, " + name + "!" },
		},
	}

	r, err := NewTemplate(options, []string{pkgDir})
	assert.Nil(t, err)
	assert.NotNil(t, r)

	var buf bytes.Buffer
	assert.Nil(t, r.Template.ExecuteTemplate(&buf, "func.tmpl", nil))
	assert.Equal(t, "Hello, User!", buf.String())
}

func TestExtractDefineName(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "single line define",
			content:  `{{ define "index.tmpl" }}hello{{ end }}`,
			expected: "index.tmpl",
		},
		{
			name:     "define with extra whitespace",
			content:  `{{ define    "page.tmpl" }}`,
			expected: "page.tmpl",
		},
		{
			name:     "no define",
			content:  `hello world`,
			expected: "",
		},
		{
			name:     "block instead of define",
			content:  `{{ block "content" . }}{{ end }}`,
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := extractDefineName(test.content)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestNewTemplate_AppFuncMapApplied(t *testing.T) {
	viewsDir := path.Resource("views")
	defer func() {
		assert.Nil(t, file.Remove(path.Resource()))
	}()

	assert.Nil(t, os.MkdirAll(viewsDir, os.ModePerm))
	assert.Nil(t, file.PutContent(path.Resource("views", "greet.tmpl"), `{{ define "greet.tmpl" }}{{ upper "hello" }}{{ end }}`))

	options := RenderOptions{
		FuncMap: template.FuncMap{
			"upper": func(s string) string { return "UPPER" },
		},
	}

	r, err := NewTemplate(options, nil)
	assert.Nil(t, err)
	assert.NotNil(t, r)

	var buf bytes.Buffer
	assert.Nil(t, r.Template.ExecuteTemplate(&buf, "greet.tmpl", nil))
	assert.Equal(t, "UPPER", buf.String())
}
