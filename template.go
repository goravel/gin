package gin

import (
	"html/template"

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
// registered on the view facade into a gin renderer. Precedence, naming and
// collision handling are those of view.Compile. It returns (nil, nil) when no
// view file exists, so callers keep gin's default renderer.
func NewTemplate(options RenderOptions) (*render.HTMLProduction, error) {
	compileOptions := frameworkview.CompileOptions{
		FuncMap: options.FuncMap,
		Log:     LogFacade,
	}
	if options.Delims != nil {
		compileOptions.Delims = &frameworkview.Delims{Left: options.Delims.Left, Right: options.Delims.Right}
	}
	if viewFacade := resolveViewFacade(); viewFacade != nil {
		compileOptions.Paths = viewFacade.RegisteredViews()
		compileOptions.FS = viewFacade.RegisteredViewFS()
	}

	instance, err := frameworkview.Compile(compileOptions)
	if err != nil || instance == nil {
		return nil, err
	}

	return &render.HTMLProduction{Template: instance}, nil
}

// DefaultTemplate compiles every registered view source with the functions
// registered on the view facade through View.AddFuncs, which is also where
// ApplicationBuilder.WithViewFuncs puts the application's functions. Custom
// delimiters still require the http.drivers.<driver>.template callback.
func DefaultTemplate() (*render.HTMLProduction, error) {
	options := RenderOptions{}
	if viewFacade := resolveViewFacade(); viewFacade != nil {
		options.FuncMap = viewFacade.Funcs()
	}

	return NewTemplate(options)
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
