package gin

import (
	nethttp "net/http"

	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/rs/cors"
)

func Cors() httpcontract.Middleware {
	return func(ctx httpcontract.Context) {
		allowedMethods := ConfigFacade.Get("cors.allowed_methods").([]string)
		if len(allowedMethods) == 1 && allowedMethods[0] == "*" {
			allowedMethods = []string{nethttp.MethodPost, nethttp.MethodGet, nethttp.MethodOptions, nethttp.MethodPut, nethttp.MethodDelete}
		}

		instance := cors.New(cors.Options{
			AllowedMethods:      allowedMethods,
			AllowedOrigins:      ConfigFacade.Get("cors.allowed_origins").([]string),
			AllowedHeaders:      ConfigFacade.Get("cors.allowed_headers").([]string),
			ExposedHeaders:      ConfigFacade.Get("cors.exposed_headers").([]string),
			MaxAge:              ConfigFacade.GetInt("cors.max_age"),
			AllowCredentials:    ConfigFacade.GetBool("cors.supports_credentials"),
			AllowPrivateNetwork: true,
		})

		instance.HandlerFunc(ctx.Response().Writer(), ctx.Request().Origin())

		if ctx.Request().Origin().Method == nethttp.MethodOptions &&
			ctx.Request().Header("Access-Control-Request-Method") != "" {
			ctx.Request().AbortWithStatus(nethttp.StatusNoContent)
		}

		ctx.Request().Next()
	}
}
