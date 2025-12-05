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
	modulePath := packages.GetModulePath()

	packages.Setup(os.Args).
		Install(
			// Add gin service provider to app.go if not using bootstrap setup
			modify.When(func(_ map[string]any) bool {
				return !env.IsBootstrapSetup()
			}, modify.GoFile(path.Config("app.go")).
				Find(match.Imports()).Modify(modify.AddImport(modulePath)).
				Find(match.Providers()).Modify(modify.Register(ginServiceProvider))),

			// Add gin service provider to providers.go if using bootstrap setup
			modify.When(func(_ map[string]any) bool {
				return env.IsBootstrapSetup()
			}, modify.AddProviderApply(modulePath, ginServiceProvider)),

			// Add gin config to http.go
			modify.GoFile(path.Config("http.go")).
				Find(match.Imports()).
				Modify(
					modify.AddImport("github.com/goravel/framework/contracts/route"), modify.AddImport(modulePath),
					modify.AddImport("github.com/goravel/gin/facades", "ginfacades"), modify.AddImport("github.com/gin-gonic/gin/render"),
				).
				Find(match.Config("http.drivers")).Modify(modify.AddConfig("gin", config)).
				Find(match.Config("http")).Modify(modify.AddConfig("default", `"gin"`)),
		).
		Uninstall(
			// Remove gin config from http.go
			modify.GoFile(path.Config("http.go")).
				Find(match.Config("http.drivers")).Modify(modify.RemoveConfig("gin")).
				Find(match.Config("http")).Modify(modify.AddConfig("default", `""`)).
				Find(match.Imports()).
				Modify(
					modify.RemoveImport("github.com/goravel/framework/contracts/route"), modify.RemoveImport(packages.GetModulePath()),
					modify.RemoveImport("github.com/goravel/gin/facades", "ginfacades"), modify.RemoveImport("github.com/gin-gonic/gin/render"),
				),

			// Remove gin service provider to app.go if not using bootstrap setup
			modify.When(func(_ map[string]any) bool {
				return !env.IsBootstrapSetup()
			}, modify.GoFile(path.Config("app.go")).
				Find(match.Providers()).Modify(modify.Unregister(ginServiceProvider)).
				Find(match.Imports()).Modify(modify.RemoveImport(modulePath))),

			// Remove gin service provider to providers.go if using bootstrap setup
			modify.When(func(_ map[string]any) bool {
				return env.IsBootstrapSetup()
			}, modify.RemoveProviderApply(modulePath, ginServiceProvider)),
		).
		Execute()
}
