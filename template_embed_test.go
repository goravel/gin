package gin

import (
	"bytes"
	"embed"
	"io/fs"
	"os"
	"testing"
	"testing/fstest"

	mockslog "github.com/goravel/framework/mocks/log"
	mocksview "github.com/goravel/framework/mocks/view"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/views
var embeddedViews embed.FS

// subFS roots an fs.FS the way the framework's View.LoadViewsFromFS does.
func subFS(t *testing.T, fsys fs.FS, root string) fs.FS {
	t.Helper()
	sub, err := fs.Sub(fsys, root)
	require.NoError(t, err)
	return sub
}

func renderView(t *testing.T, options RenderOptions, name string, data any) string {
	t.Helper()
	r, err := NewTemplate(options)
	require.NoError(t, err)
	require.NotNil(t, r)

	var buf bytes.Buffer
	require.NoError(t, r.Template.ExecuteTemplate(&buf, name, data))
	return buf.String()
}

func TestNewTemplate_EmbeddedViews(t *testing.T) {
	pkg := subFS(t, embeddedViews, "testdata/views")

	t.Run("embedded package views", func(t *testing.T) {
		defer func() { ViewFacade = nil }()

		mockView := mocksview.NewView(t)
		mockView.EXPECT().RegisteredViews().Return(nil).Once()
		mockView.EXPECT().RegisteredViewFS().Return([]fs.FS{pkg}).Once()
		ViewFacade = mockView

		assert.Equal(t, "Embedded Content", renderView(t, RenderOptions{}, "page.tmpl", nil))
	})

	t.Run("nested layout, partial and block", func(t *testing.T) {
		defer func() { ViewFacade = nil }()

		mockView := mocksview.NewView(t)
		mockView.EXPECT().RegisteredViews().Return(nil).Once()
		mockView.EXPECT().RegisteredViewFS().Return([]fs.FS{pkg}).Once()
		ViewFacade = mockView

		assert.Equal(t,
			"<html><body><nav>Menu</nav><main><h1>Home</h1></main></body></html>",
			renderView(t, RenderOptions{}, "pages/home.tmpl", map[string]any{"Title": "Home", "Nav": "Menu"}),
		)
	})

	t.Run("app overrides embedded package", func(t *testing.T) {
		defer func() {
			ViewFacade = nil
			assert.Nil(t, file.Remove(path.Resource()))
		}()
		assert.Nil(t, file.PutContent(path.Resource("views", "page.tmpl"), `{{ define "page.tmpl" }}App Content{{ end }}`))

		mockView := mocksview.NewView(t)
		mockView.EXPECT().RegisteredViews().Return(nil).Once()
		mockView.EXPECT().RegisteredViewFS().Return([]fs.FS{pkg}).Once()
		ViewFacade = mockView

		assert.Equal(t, "App Content", renderView(t, RenderOptions{}, "page.tmpl", nil))
	})

	t.Run("embedded package fallback for templates missing from app", func(t *testing.T) {
		defer func() {
			ViewFacade = nil
			assert.Nil(t, file.Remove(path.Resource()))
		}()
		assert.Nil(t, file.PutContent(path.Resource("views", "other.tmpl"), `{{ define "other.tmpl" }}Other{{ end }}`))

		mockView := mocksview.NewView(t)
		mockView.EXPECT().RegisteredViews().Return(nil).Once()
		mockView.EXPECT().RegisteredViewFS().Return([]fs.FS{pkg}).Once()
		ViewFacade = mockView

		r, err := NewTemplate(RenderOptions{})
		require.NoError(t, err)
		require.NotNil(t, r)

		var buf bytes.Buffer
		require.NoError(t, r.Template.ExecuteTemplate(&buf, "other.tmpl", nil))
		assert.Equal(t, "Other", buf.String())

		buf.Reset()
		require.NoError(t, r.Template.ExecuteTemplate(&buf, "page.tmpl", nil))
		assert.Equal(t, "Embedded Content", buf.String())
	})

	t.Run("filesystem package overrides embedded package", func(t *testing.T) {
		pkgDir := path.Resource("pkg_dir")
		defer func() {
			ViewFacade = nil
			assert.Nil(t, file.Remove(path.Resource()))
		}()
		assert.Nil(t, os.MkdirAll(pkgDir, os.ModePerm))
		assert.Nil(t, file.PutContent(path.Resource("pkg_dir", "page.tmpl"), `{{ define "page.tmpl" }}Dir Content{{ end }}`))

		mockLog := mockslog.NewLog(t)
		LogFacade = mockLog
		defer func() { LogFacade = nil }()
		mockLog.EXPECT().Warningf("view collision: %q defined in %q and %q, using first", "page.tmpl", path.Resource("pkg_dir", "page.tmpl"), "fs[0]/page.tmpl").Return().Once()

		mockView := mocksview.NewView(t)
		mockView.EXPECT().RegisteredViews().Return([]string{pkgDir}).Once()
		mockView.EXPECT().RegisteredViewFS().Return([]fs.FS{pkg}).Once()
		ViewFacade = mockView

		assert.Equal(t, "Dir Content", renderView(t, RenderOptions{}, "page.tmpl", nil))
	})

	t.Run("multiple embedded packages", func(t *testing.T) {
		defer func() { ViewFacade = nil }()

		pkgB := fstest.MapFS{
			"views/bar.tmpl":     {Data: []byte(`{{ define "bar.tmpl" }}Bar{{ end }}`)},
			"views/sub/baz.tmpl": {Data: []byte(`{{ define "sub/baz.tmpl" }}Baz{{ end }}`)},
			"other/nope.tmpl":    {Data: []byte(`{{ define "nope.tmpl" }}Outside root{{ end }}`)},
		}

		mockView := mocksview.NewView(t)
		mockView.EXPECT().RegisteredViews().Return(nil).Once()
		mockView.EXPECT().RegisteredViewFS().Return([]fs.FS{pkg, subFS(t, pkgB, "views")}).Once()
		ViewFacade = mockView

		r, err := NewTemplate(RenderOptions{})
		require.NoError(t, err)
		require.NotNil(t, r)

		for name, expected := range map[string]string{
			"page.tmpl":    "Embedded Content",
			"bar.tmpl":     "Bar",
			"sub/baz.tmpl": "Baz",
		} {
			var buf bytes.Buffer
			require.NoError(t, r.Template.ExecuteTemplate(&buf, name, nil))
			assert.Equal(t, expected, buf.String())
		}

		assert.Nil(t, r.Template.Lookup("nope.tmpl"), "files outside the registered root must not be loaded")
		assert.Nil(t, r.Template.Lookup("missing.tmpl"))
	})

	t.Run("collision between embedded packages uses first", func(t *testing.T) {
		defer func() {
			ViewFacade = nil
			LogFacade = nil
		}()

		second := fstest.MapFS{
			"page.tmpl": {Data: []byte(`{{ define "page.tmpl" }}Second{{ end }}`)},
		}

		mockLog := mockslog.NewLog(t)
		LogFacade = mockLog
		mockLog.EXPECT().Warningf("view collision: %q defined in %q and %q, using first", "page.tmpl", "fs[0]/page.tmpl", "fs[1]/page.tmpl").Return().Once()

		mockView := mocksview.NewView(t)
		mockView.EXPECT().RegisteredViews().Return(nil).Once()
		mockView.EXPECT().RegisteredViewFS().Return([]fs.FS{pkg, second}).Once()
		ViewFacade = mockView

		assert.Equal(t, "Embedded Content", renderView(t, RenderOptions{}, "page.tmpl", nil))
	})

	t.Run("custom delims", func(t *testing.T) {
		defer func() { ViewFacade = nil }()

		custom := fstest.MapFS{
			"delim.tmpl": {Data: []byte(`{[ define "delim.tmpl" ]}Custom{[ end ]}`)},
		}

		mockView := mocksview.NewView(t)
		mockView.EXPECT().RegisteredViews().Return(nil).Once()
		mockView.EXPECT().RegisteredViewFS().Return([]fs.FS{custom}).Once()
		ViewFacade = mockView

		options := RenderOptions{Delims: &Delims{Left: "{[", Right: "]}"}}
		assert.Equal(t, "Custom", renderView(t, options, "delim.tmpl", nil))
	})

	t.Run("unreadable root is skipped", func(t *testing.T) {
		defer func() { ViewFacade = nil }()

		missing, err := fs.Sub(fstest.MapFS{}, "does/not/exist")
		require.NoError(t, err)

		mockView := mocksview.NewView(t)
		mockView.EXPECT().RegisteredViews().Return(nil).Once()
		mockView.EXPECT().RegisteredViewFS().Return([]fs.FS{missing, pkg}).Once()
		ViewFacade = mockView

		assert.Equal(t, "Embedded Content", renderView(t, RenderOptions{}, "page.tmpl", nil))
	})

	t.Run("only empty filesystems yields no renderer", func(t *testing.T) {
		defer func() { ViewFacade = nil }()

		mockView := mocksview.NewView(t)
		mockView.EXPECT().RegisteredViews().Return(nil).Once()
		mockView.EXPECT().RegisteredViewFS().Return([]fs.FS{fstest.MapFS{}}).Once()
		ViewFacade = mockView

		r, err := NewTemplate(RenderOptions{})
		assert.Nil(t, err)
		assert.Nil(t, r)
	})

	t.Run("invalid template returns parse error", func(t *testing.T) {
		defer func() { ViewFacade = nil }()

		broken := fstest.MapFS{
			"broken.tmpl": {Data: []byte(`{{ define "broken.tmpl" }}{{ .Unclosed`)},
		}

		mockView := mocksview.NewView(t)
		mockView.EXPECT().RegisteredViews().Return(nil).Once()
		mockView.EXPECT().RegisteredViewFS().Return([]fs.FS{broken}).Once()
		ViewFacade = mockView

		r, err := NewTemplate(RenderOptions{})
		assert.Error(t, err)
		assert.Nil(t, r)
	})
}
