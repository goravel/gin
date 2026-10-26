package gin

import (
	"html/template"
	"maps"

	"github.com/gin-gonic/gin/render"
	contractsview "github.com/goravel/framework/contracts/view"
	frameworkview "github.com/goravel/framework/view"
)

type Delims struct {
	Left  string
	Right string
}

type RenderOptions struct {
	Delims  *Delims
	FuncMap template.FuncMap
}

// NewTemplate compiles the application's resources/views and every source
// registered on the view facade into a gin renderer. The functions registered
// on the view facade through View.AddFuncs, which is also where
// ApplicationBuilder.WithViewFuncs puts the application's, are installed first;
// options.FuncMap is added on top and overrides them by name. Precedence, naming
// and collision handling are those of view.Compile. It returns (nil, nil) when
// no view file exists, so callers keep gin's default renderer.
func NewTemplate(options RenderOptions) (*render.HTMLProduction, error) {
	compileOptions := frameworkview.CompileOptions{
		Log: LogFacade,
	}
	if options.Delims != nil {
		compileOptions.Delims = &frameworkview.Delims{Left: options.Delims.Left, Right: options.Delims.Right}
	}
	if viewFacade := resolveViewFacade(); viewFacade != nil {
		compileOptions.Paths = viewFacade.RegisteredViews()
		compileOptions.FS = viewFacade.RegisteredViewFS()
		compileOptions.FuncMap = viewFacade.Funcs()
	}
	if len(options.FuncMap) > 0 {
		if compileOptions.FuncMap == nil {
			compileOptions.FuncMap = make(template.FuncMap, len(options.FuncMap))
		}
		maps.Copy(compileOptions.FuncMap, options.FuncMap)
	}

	instance, err := frameworkview.Compile(compileOptions)
	if err != nil || instance == nil {
		return nil, err
	}

	return &render.HTMLProduction{Template: instance}, nil
}

// DefaultTemplate compiles every registered view source with the functions
// registered on the view facade and no custom delimiters. Custom delimiters
// still require the http.drivers.<driver>.template callback, which keeps the
// registered functions because NewTemplate installs them itself.
func DefaultTemplate() (*render.HTMLProduction, error) {
	return NewTemplate(RenderOptions{})
}

// resolveViewFacade returns the view facade ServiceProvider.Boot assigned, or
// resolves it from the container when the driver is used before Boot.
func resolveViewFacade() contractsview.View {
	if ViewFacade != nil {
		return ViewFacade
	}
	if App != nil {
		return App.MakeView()
	}

	return nil
}
