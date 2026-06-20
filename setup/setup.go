package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/path"
)

func main() {
	setup := packages.Setup(os.Args)
	config := `map[string]any{
        // Optional, default is 4096 KB
        "body_limit": 4096,
        "header_limit": 4096,
        "route": func() (route.Route, error) {
            return ginfacades.Route("gin"), nil
        },
        // Optional, default is http/template
        "template": func() (render.HTMLRender, error) {
            return gin.DefaultTemplate()
        },
    }`
	ginServiceProvider := "&gin.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()
	httpConfigPath := path.Config("http.go")
	appConfigPath := path.Config("app.go")
	httpDriversConfig := match.Config("http.drivers")
	httpConfig := match.Config("http")
	routeContract := "github.com/goravel/framework/contracts/route"
	ginFacade := "github.com/goravel/gin/facades"
	ginRender := "github.com/gin-gonic/gin/render"

	setup.Install(
		// Add gin service provider to app.go if not using bootstrap setup
		modify.When(func(_ map[string]any) bool {
			return !env.IsBootstrapSetup()
		}, modify.GoFile(appConfigPath).
			Find(match.Imports()).Modify(modify.AddImport(moduleImport)).
			Find(match.Providers()).Modify(modify.Register(ginServiceProvider))),

		// Add gin service provider to providers.go if using bootstrap setup
		modify.When(func(_ map[string]any) bool {
			return env.IsBootstrapSetup()
		}, modify.RegisterProvider(moduleImport, ginServiceProvider)),

		// Add gin config to http.go
		modify.GoFile(httpConfigPath).
			Find(match.Imports()).
			Modify(
				modify.AddImport(routeContract), modify.AddImport(moduleImport),
				modify.AddImport(ginFacade, "ginfacades"), modify.AddImport(ginRender),
			).
			Find(httpDriversConfig).Modify(modify.AddConfig("gin", config)).
			Find(httpConfig).Modify(modify.AddConfig("default", `"gin"`)),
	).Uninstall(
		// Remove gin config from http.go
		modify.GoFile(httpConfigPath).
			Find(httpDriversConfig).Modify(modify.RemoveConfig("gin")).
			Find(httpConfig).Modify(modify.AddConfig("default", `""`)).
			Find(match.Imports()).
			Modify(
				modify.RemoveImport(routeContract), modify.RemoveImport(moduleImport),
				modify.RemoveImport(ginFacade, "ginfacades"), modify.RemoveImport(ginRender),
			),

		// Remove gin service provider from app.go if not using bootstrap setup
		modify.When(func(_ map[string]any) bool {
			return !env.IsBootstrapSetup()
		}, modify.GoFile(appConfigPath).
			Find(match.Providers()).Modify(modify.Unregister(ginServiceProvider)).
			Find(match.Imports()).Modify(modify.RemoveImport(moduleImport))),

		// Remove gin service provider from providers.go if using bootstrap setup
		modify.When(func(_ map[string]any) bool {
			return env.IsBootstrapSetup()
		}, modify.UnregisterProvider(moduleImport, ginServiceProvider)),
	).Execute()
}
