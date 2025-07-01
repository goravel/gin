package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

var config = `map[string]any{
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

func main() {
	packages.Setup(os.Args).
		Install(
			modify.GoFile(path.Config("app.go")).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register("&gin.ServiceProvider{}")),
			modify.GoFile(path.Config("http.go")).
				Find(match.Imports()).
				Modify(
					modify.AddImport("github.com/goravel/framework/contracts/route"), modify.AddImport(packages.GetModulePath()),
					modify.AddImport("github.com/goravel/gin/facades", "ginfacades"), modify.AddImport("github.com/gin-gonic/gin/render"),
				).
				Find(match.Config("http.drivers")).Modify(modify.AddConfig("gin", config)),
		).
		Uninstall(
			modify.GoFile(path.Config("app.go")).
				Find(match.Providers()).Modify(modify.Unregister("&gin.ServiceProvider{}")).
				Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
			modify.GoFile(path.Config("http.go")).
				Find(match.Config("http.drivers")).Modify(modify.RemoveConfig("gin")).
				Find(match.Imports()).
				Modify(
					modify.RemoveImport("github.com/goravel/framework/contracts/route"), modify.RemoveImport(packages.GetModulePath()),
					modify.RemoveImport("github.com/goravel/gin/facades", "ginfacades"), modify.RemoveImport("github.com/gin-gonic/gin/render"),
				),
		).
		Execute()
}
