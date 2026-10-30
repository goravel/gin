package gin

import (
	"bytes"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin/render"
	contractshttp "github.com/goravel/framework/contracts/http"
	configmocks "github.com/goravel/framework/mocks/config"
	mocksview "github.com/goravel/framework/mocks/view"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTemplateTestRoute builds a Route the way NewRoute does, with the given
// http.drivers.gin.template value, without booting any facade.
func newTemplateTestRoute(t *testing.T, templateConfig any) (*Route, *configmocks.Config) {
	t.Helper()

	mockConfig := configmocks.NewConfig(t)
	mockConfig.EXPECT().GetBool("app.debug").Return(false).Once()
	mockConfig.EXPECT().GetInt("http.drivers.gin.body_limit", 4096).Return(4096).Once()
	mockConfig.EXPECT().Get("http.drivers.gin.template").Return(templateConfig).Once()
	ConfigFacade = mockConfig

	route := &Route{
		config: mockConfig,
		driver: "gin",
	}
	route.init(nil)

	return route, mockConfig
}

// expectReinit adds the config reads a GlobalMiddleware, SetGlobalMiddleware or
// Recover call performs when it rebuilds the engine through init().
func expectReinit(mockConfig *configmocks.Config, templateConfig any) {
	mockConfig.EXPECT().GetBool("app.debug").Return(false).Once()
	mockConfig.EXPECT().GetInt("http.drivers.gin.body_limit", 4096).Return(4096).Once()
	mockConfig.EXPECT().Get("http.drivers.gin.template").Return(templateConfig).Once()
}

func serve(t *testing.T, route *Route, url string) *httptest.ResponseRecorder {
	t.Helper()

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	route.ServeHTTP(w, req)

	return w
}

// TestRoute_TemplateCallbackDeferred guards the ordering fix for the custom
// template callback: it must not run while the driver is constructed, because
// that can happen before any service provider has booted. It runs once, from
// CompileViews() or the first serve call, and therefore sees views and functions
// registered during Boot.
func TestRoute_TemplateCallbackDeferred(t *testing.T) {
	pkgDir := path.Resource("pkg_deferred")
	require.NoError(t, os.MkdirAll(pkgDir, os.ModePerm))
	require.NoError(t, file.PutContent(path.Resource("pkg_deferred", "shout.tmpl"), `{{ define "shout.tmpl" }}{{ upper .word }}{{ end }}`))
	defer func() {
		ConfigFacade = nil
		ViewFacade = nil
		assert.Nil(t, file.Remove(path.Resource()))
	}()

	calls := 0
	callback := func() (render.HTMLRender, error) {
		calls++

		// The callback reads the registry the way an application would, so it
		// only works if it runs after the package registered its views.
		return NewTemplate(RenderOptions{
			FuncMap: template.FuncMap{"upper": strings.ToUpper},
		})
	}

	route, mockConfig := newTemplateTestRoute(t, callback)
	assert.Equal(t, 0, calls, "init() must not invoke the template callback")
	assert.Nil(t, route.instance.HTMLRender)

	// A provider's Boot() registers the package views after the driver exists.
	mockView := mocksview.NewView(t)
	mockView.EXPECT().RegisteredViews().Return([]string{pkgDir}).Once()
	mockView.EXPECT().RegisteredViewFS().Return(nil).Once()
	mockView.EXPECT().Funcs().Return(nil).Once()
	mockView.EXPECT().GetShared().Return(nil).Times(3)
	ViewFacade = mockView

	route.Get("/shout", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().View().Make("shout.tmpl", map[string]any{"word": "hi"})
	})

	// The framework compiles after Boot through the view.Compiler bridge.
	require.NoError(t, route.CompileViews())
	assert.Equal(t, 1, calls)

	w := serve(t, route, "/shout")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "HI", w.Body.String())
	assert.Equal(t, 1, calls, "serving must not compile again")

	// Rebuilding the engine after the first serve keeps the compiled renderer and
	// does not invoke the callback a second time.
	expectReinit(mockConfig, callback)
	route.GlobalMiddleware()
	route.Get("/shout", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().View().Make("shout.tmpl", map[string]any{"word": "again"})
	})
	w = serve(t, route, "/shout")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "AGAIN", w.Body.String())
	assert.Equal(t, 1, calls)

	// Without the bridge the first serve call compiles; the result is identical.
	route2, _ := newTemplateTestRoute(t, callback)
	route2.Get("/shout", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().View().Make("shout.tmpl", map[string]any{"word": "late"})
	})
	mockView.EXPECT().RegisteredViews().Return([]string{pkgDir}).Once()
	mockView.EXPECT().RegisteredViewFS().Return(nil).Once()
	mockView.EXPECT().Funcs().Return(nil).Once()
	w = serve(t, route2, "/shout")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "LATE", w.Body.String())
	assert.Equal(t, 2, calls)
}

// TestRoute_TemplateCallbackError guards that a failing callback no longer aborts
// driver construction. The error surfaces from CompileViews() and every serve
// entry point, and stays sticky, exactly like a DefaultTemplate() failure.
func TestRoute_TemplateCallbackError(t *testing.T) {
	defer func() {
		ConfigFacade = nil
		LogFacade = nil
	}()
	LogFacade = nil

	calls := 0
	route, _ := newTemplateTestRoute(t, func() (render.HTMLRender, error) {
		calls++

		return nil, errors.New("callback failed")
	})

	err := route.CompileViews()
	require.EqualError(t, err, "callback failed")
	assert.Equal(t, 1, calls)

	// Sticky: later calls report the same error without invoking the callback again.
	assert.Equal(t, err, route.ensureTemplate())
	assert.Equal(t, 1, calls)

	w := serve(t, route, "/")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "callback failed\n", w.Body.String())
}

// TestRoute_DefaultTemplateUsesViewFuncs guards that functions registered on the
// view facade (View.AddFuncs, which ApplicationBuilder.WithViewFuncs feeds) are
// installed on the default template set.
func TestRoute_DefaultTemplateUsesViewFuncs(t *testing.T) {
	require.NoError(t, os.MkdirAll(path.Resource("views"), os.ModePerm))
	require.NoError(t, file.PutContent(path.Resource("views", "greet.tmpl"), `{{ define "greet.tmpl" }}{{ greet "Ada" }}{{ end }}`))
	defer func() {
		ConfigFacade = nil
		ViewFacade = nil
		assert.Nil(t, file.Remove(path.Resource()))
	}()

	route, _ := newTemplateTestRoute(t, nil)

	mockView := mocksview.NewView(t)
	mockView.EXPECT().RegisteredViews().Return(nil).Once()
	mockView.EXPECT().RegisteredViewFS().Return(nil).Once()
	mockView.EXPECT().Funcs().Return(template.FuncMap{
		"greet": func(name string) string { return "Hello, " + name + "!" },
	}).Once()
	mockView.EXPECT().GetShared().Return(nil).Once()
	ViewFacade = mockView

	route.Get("/greet", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().View().Make("greet.tmpl")
	})

	w := serve(t, route, "/greet")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Hello, Ada!", w.Body.String())
}

// TestRoute_TemplateCallbackNilRender guards that a callback returning no
// renderer falls back to the default set instead of leaving HTMLRender nil,
// which would fail every render after a successful CompileViews().
func TestRoute_TemplateCallbackNilRender(t *testing.T) {
	require.NoError(t, os.MkdirAll(path.Resource("views"), os.ModePerm))
	require.NoError(t, file.PutContent(path.Resource("views", "home.tmpl"), `{{ define "home.tmpl" }}Home{{ end }}`))
	defer func() {
		ConfigFacade = nil
		ViewFacade = nil
		assert.Nil(t, file.Remove(path.Resource()))
	}()

	calls := 0
	route, _ := newTemplateTestRoute(t, func() (render.HTMLRender, error) {
		calls++

		return nil, nil
	})

	mockView := mocksview.NewView(t)
	mockView.EXPECT().RegisteredViews().Return(nil).Once()
	mockView.EXPECT().RegisteredViewFS().Return(nil).Once()
	mockView.EXPECT().Funcs().Return(nil).Once()
	mockView.EXPECT().GetShared().Return(nil).Once()
	ViewFacade = mockView

	require.NoError(t, route.CompileViews())
	assert.Equal(t, 1, calls)
	require.NotNil(t, route.instance.HTMLRender)

	route.Get("/home", func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().View().Make("home.tmpl")
	})
	w := serve(t, route, "/home")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Home", w.Body.String())
	assert.Equal(t, 1, calls)
}

// TestRoute_CompileViews guards the view.Compiler contract: CompileViews compiles
// the default set once, later serve calls reuse it, and an engine rebuild after
// the compile keeps the renderer.
func TestRoute_CompileViews(t *testing.T) {
	require.NoError(t, os.MkdirAll(path.Resource("views"), os.ModePerm))
	require.NoError(t, file.PutContent(path.Resource("views", "home.tmpl"), `{{ define "home.tmpl" }}Home{{ end }}`))
	defer func() {
		ConfigFacade = nil
		ViewFacade = nil
		assert.Nil(t, file.Remove(path.Resource()))
	}()

	route, mockConfig := newTemplateTestRoute(t, nil)

	mockView := mocksview.NewView(t)
	mockView.EXPECT().RegisteredViews().Return(nil).Once()
	mockView.EXPECT().RegisteredViewFS().Return(nil).Once()
	mockView.EXPECT().Funcs().Return(nil).Once()
	mockView.EXPECT().GetShared().Return(nil).Twice()
	ViewFacade = mockView

	require.NoError(t, route.CompileViews())
	require.NotNil(t, route.instance.HTMLRender)
	require.NoError(t, route.CompileViews(), "a second call is a no-op")

	handler := func(ctx contractshttp.Context) contractshttp.Response {
		return ctx.Response().View().Make("home.tmpl")
	}
	route.Get("/home", handler)
	w := serve(t, route, "/home")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Home", w.Body.String())

	// SetGlobalMiddleware rebuilds the engine (and drops the routes, by design);
	// the renderer must survive the rebuild.
	expectReinit(mockConfig, nil)
	route.SetGlobalMiddleware(nil)
	require.NotNil(t, route.instance.HTMLRender)
	route.Get("/home", handler)
	w = serve(t, route, "/home")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Home", w.Body.String())
}

// TestRoute_CompileViewsWithoutViews guards that an application without any view
// file keeps gin's default nil renderer instead of a non-nil interface holding a
// nil *render.HTMLProduction.
func TestRoute_CompileViewsWithoutViews(t *testing.T) {
	defer func() {
		ConfigFacade = nil
		ViewFacade = nil
	}()

	route, _ := newTemplateTestRoute(t, nil)

	mockView := mocksview.NewView(t)
	mockView.EXPECT().RegisteredViews().Return(nil).Once()
	mockView.EXPECT().RegisteredViewFS().Return(nil).Once()
	mockView.EXPECT().Funcs().Return(nil).Once()
	ViewFacade = mockView

	require.NoError(t, route.CompileViews())
	assert.Nil(t, route.instance.HTMLRender)
	assert.Nil(t, route.htmlRender)
}

// TestRoute_TemplateInstance guards that a ready renderer configured as an
// instance is used as is and never replaced by a compile.
func TestRoute_TemplateInstance(t *testing.T) {
	defer func() {
		ConfigFacade = nil
	}()

	tmpl, err := template.New("inline.tmpl").Parse(`inline`)
	require.NoError(t, err)
	instance := &render.HTMLProduction{Template: tmpl}

	route, _ := newTemplateTestRoute(t, instance)
	assert.Same(t, instance, route.instance.HTMLRender)

	require.NoError(t, route.CompileViews())
	assert.Same(t, instance, route.instance.HTMLRender)

	var buf bytes.Buffer
	require.NoError(t, instance.Template.ExecuteTemplate(&buf, "inline.tmpl", nil))
	assert.Equal(t, "inline", buf.String())
}
