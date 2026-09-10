package gin

import (
	"bytes"
	"embed"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mockslog "github.com/goravel/framework/mocks/log"
	mocksview "github.com/goravel/framework/mocks/view"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/views
var embeddedViews embed.FS

// subFS roots an fs.FS at root, the shape View.LoadViewsFromFS() stores after
// its own validation (it additionally rejects a root that is missing or is not a
// directory, which subFS does not replicate).
func subFS(t *testing.T, fsys fs.FS, root string) fs.FS {
	t.Helper()
	sub, err := fs.Sub(fsys, root)
	require.NoError(t, err)
	return sub
}

// failingFS serves base except for failPath, where every open fails. It models a
// filesystem that turns unreadable after the root has been stat'ed: failPath may
// be a directory (the walk fails to list it) or a file (the walk fails to read it).
type failingFS struct {
	base     fs.FS
	failPath string
}

func (f failingFS) Open(name string) (fs.File, error) {
	if name == f.failPath {
		return nil, &fs.PathError{Op: "open", Path: name, Err: errors.New("permission denied")}
	}
	return f.base.Open(name)
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
		require.NoError(t, file.PutContent(path.Resource("views", "page.tmpl"), `{{ define "page.tmpl" }}App Content{{ end }}`))

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
		require.NoError(t, file.PutContent(path.Resource("views", "other.tmpl"), `{{ define "other.tmpl" }}Other{{ end }}`))

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
		require.NoError(t, os.MkdirAll(pkgDir, os.ModePerm))
		require.NoError(t, file.PutContent(path.Resource("pkg_dir", "page.tmpl"), `{{ define "page.tmpl" }}Dir Content{{ end }}`))

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

	t.Run("unreadable root is skipped and warned about", func(t *testing.T) {
		defer func() {
			ViewFacade = nil
			LogFacade = nil
		}()

		missing, err := fs.Sub(fstest.MapFS{}, "does/not/exist")
		require.NoError(t, err)

		mockLog := mockslog.NewLog(t)
		mockLog.EXPECT().Warningf("view source fs[%d] is unreadable, skipping: %v", 0, mock.Anything).Return().Once()
		LogFacade = mockLog

		mockView := mocksview.NewView(t)
		mockView.EXPECT().RegisteredViews().Return(nil).Once()
		mockView.EXPECT().RegisteredViewFS().Return([]fs.FS{missing, pkg}).Once()
		ViewFacade = mockView

		assert.Equal(t, "Embedded Content", renderView(t, RenderOptions{}, "page.tmpl", nil))
	})

	t.Run("nil filesystem is skipped and warned about", func(t *testing.T) {
		defer func() {
			ViewFacade = nil
			LogFacade = nil
		}()

		mockLog := mockslog.NewLog(t)
		mockLog.EXPECT().Warningf("view source fs[%d] is nil, skipping", 0).Return().Once()
		LogFacade = mockLog

		mockView := mocksview.NewView(t)
		mockView.EXPECT().RegisteredViews().Return(nil).Once()
		mockView.EXPECT().RegisteredViewFS().Return([]fs.FS{nil, pkg}).Once()
		ViewFacade = mockView

		assert.Equal(t, "Embedded Content", renderView(t, RenderOptions{}, "page.tmpl", nil))
	})

	t.Run("only empty filesystems yields no renderer", func(t *testing.T) {
		defer func() { ViewFacade = nil }()

		// A readable root holding no files: the source passes the stat guard and
		// is dropped because loadSource() finds nothing to parse.
		empty := fstest.MapFS{"sub": {Mode: fs.ModeDir}}
		require.NoError(t, fstest.TestFS(empty, "sub"))

		mockView := mocksview.NewView(t)
		mockView.EXPECT().RegisteredViews().Return(nil).Once()
		mockView.EXPECT().RegisteredViewFS().Return([]fs.FS{empty}).Once()
		ViewFacade = mockView

		r, err := NewTemplate(RenderOptions{})
		assert.Nil(t, err)
		assert.Nil(t, r)
	})

	t.Run("non-template files in an embedded filesystem are harmless", func(t *testing.T) {
		defer func() { ViewFacade = nil }()

		// //go:embed of a views directory picks up whatever else lives there.
		// Files without a define block are still parsed, under their base name.
		assets := fstest.MapFS{
			"README.md":       {Data: []byte("# Package views\n")},
			"assets/site.css": {Data: []byte("body { margin: 0 }")},
			"widget.tmpl":     {Data: []byte(`{{ define "widget.tmpl" }}Widget{{ end }}`)},
		}

		mockView := mocksview.NewView(t)
		mockView.EXPECT().RegisteredViews().Return(nil).Once()
		mockView.EXPECT().RegisteredViewFS().Return([]fs.FS{assets}).Once()
		ViewFacade = mockView

		r, err := NewTemplate(RenderOptions{})
		require.NoError(t, err)
		require.NotNil(t, r)

		var buf bytes.Buffer
		require.NoError(t, r.Template.ExecuteTemplate(&buf, "widget.tmpl", nil))
		assert.Equal(t, "Widget", buf.String())

		buf.Reset()
		require.NoError(t, r.Template.ExecuteTemplate(&buf, "README.md", nil))
		assert.Equal(t, "# Package views\n", buf.String())
	})

	t.Run("names containing glob metacharacters are loaded literally", func(t *testing.T) {
		// Regression test: template names are resolved literally. Resolving them
		// as globs turns "page[1].tmpl" into the pattern "page1.tmpl", which either
		// loads the wrong file or fails the whole compile.
		pkgDir := path.Resource("pkg_glob")
		defer func() {
			ViewFacade = nil
			assert.Nil(t, file.Remove(path.Resource()))
		}()
		require.NoError(t, os.MkdirAll(pkgDir, os.ModePerm))
		require.NoError(t, file.PutContent(filepath.Join(pkgDir, "dir[1].tmpl"), `{{ define "dir[1].tmpl" }}Dir Bracket{{ end }}`))

		embedded := fstest.MapFS{
			"page[1].tmpl": {Data: []byte(`{{ define "page[1].tmpl" }}Bracket{{ end }}`)},
			"page1.tmpl":   {Data: []byte(`{{ define "page1.tmpl" }}Decoy{{ end }}`)},
			"draft?.tmpl":  {Data: []byte(`{{ define "draft?.tmpl" }}Question{{ end }}`)},
		}

		mockView := mocksview.NewView(t)
		mockView.EXPECT().RegisteredViews().Return([]string{pkgDir}).Once()
		mockView.EXPECT().RegisteredViewFS().Return([]fs.FS{embedded}).Once()
		ViewFacade = mockView

		r, err := NewTemplate(RenderOptions{})
		require.NoError(t, err)
		require.NotNil(t, r)

		for name, expected := range map[string]string{
			"dir[1].tmpl":  "Dir Bracket",
			"page[1].tmpl": "Bracket",
			"page1.tmpl":   "Decoy",
			"draft?.tmpl":  "Question",
		} {
			var buf bytes.Buffer
			require.NoError(t, r.Template.ExecuteTemplate(&buf, name, nil), name)
			assert.Equal(t, expected, buf.String(), name)
		}
	})

	t.Run("walk errors are propagated", func(t *testing.T) {
		base := fstest.MapFS{
			"page.tmpl":     {Data: []byte(`{{ define "page.tmpl" }}Page{{ end }}`)},
			"sub/deep.tmpl": {Data: []byte(`{{ define "sub/deep.tmpl" }}Deep{{ end }}`)},
		}

		for _, failPath := range []string{"sub", "page.tmpl"} {
			t.Run(failPath, func(t *testing.T) {
				defer func() { ViewFacade = nil }()

				mockView := mocksview.NewView(t)
				mockView.EXPECT().RegisteredViews().Return(nil).Once()
				mockView.EXPECT().RegisteredViewFS().Return([]fs.FS{failingFS{base: base, failPath: failPath}}).Once()
				ViewFacade = mockView

				r, err := NewTemplate(RenderOptions{})
				assert.ErrorContains(t, err, "permission denied")
				assert.Nil(t, r)
			})
		}
	})

	t.Run("view facade resolved from the container", func(t *testing.T) {
		// route.init() can run before ViewFacade is assigned; NewTemplate() then
		// resolves the facade through the application container.
		pkgDir := path.Resource("pkg_app")
		defer func() {
			App = nil
			ViewFacade = nil
			assert.Nil(t, file.Remove(path.Resource()))
		}()
		require.NoError(t, os.MkdirAll(pkgDir, os.ModePerm))
		require.NoError(t, file.PutContent(filepath.Join(pkgDir, "app.tmpl"), `{{ define "app.tmpl" }}From Container{{ end }}`))

		mockView := mocksview.NewView(t)
		mockView.EXPECT().RegisteredViews().Return([]string{pkgDir}).Once()
		mockView.EXPECT().RegisteredViewFS().Return(nil).Once()

		mockApp := mocksfoundation.NewApplication(t)
		mockApp.EXPECT().MakeView().Return(mockView).Once()
		App = mockApp
		ViewFacade = nil

		assert.Equal(t, "From Container", renderView(t, RenderOptions{}, "app.tmpl", nil))
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
