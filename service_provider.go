package gin

import (
	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
)

const RouteBinding = "goravel.gin.route"

var (
	App              foundation.Application
	ConfigFacade     config.Config
	LogFacade        log.Log
	ValidationFacade validation.Validation
	ViewFacade       http.View
)

type ServiceProvider struct{}

func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings: []string{
			RouteBinding,
		},
		Dependencies: []string{
			binding.Config,
			binding.Log,
			binding.Validation,
			binding.View,
		},
		ProvideFor: []string{
			binding.Route,
		},
	}
}

func (r *ServiceProvider) Register(app foundation.Application) {
	App = app

	app.BindWith(RouteBinding, func(app foundation.Application, parameters map[string]any) (any, error) {
		return NewRoute(app.MakeConfig(), parameters)
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	module := "gin"

	if ConfigFacade = app.MakeConfig(); ConfigFacade == nil {
		color.Errorln(errors.ConfigFacadeNotSet.SetModule(module))
	}

	if LogFacade = app.MakeLog(); LogFacade == nil {
		color.Errorln(errors.LogFacadeNotSet.SetModule(module))
	}

	if ValidationFacade = app.MakeValidation(); ValidationFacade == nil {
		color.Errorln(errors.New("validation facade is not initialized").SetModule(module))
	}

	if ViewFacade = app.MakeView(); ViewFacade == nil {
		color.Errorln(errors.New("view facade is not initialized").SetModule(module))
	}

	app.Publishes("github.com/goravel/gin", map[string]string{
		"config/cors.go": app.ConfigPath("cors.go"),
	})
}
