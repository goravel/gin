package gin

import (
	"github.com/gin-gonic/gin"
	supporthttp "github.com/goravel/framework/support/http"
)

// routeDomainKey keeps the domain of the matched route in the gin context, so
// the route's Info is found among the routes of that domain.
const routeDomainKey = "goravel_route_domain"

// domainRouter holds the routes of a domain in an engine of their own, so the
// same path can be registered on several domains. Domains that differ only in
// their parameter names share it.
type domainRouter struct {
	domain *supporthttp.Domain
	engine *gin.Engine
}

// domainMiddleware runs first on every route of the domain: it records the
// domain, and adds the domain's parameters to the route parameters, so they
// are read like the ones of the path. A path parameter with the same name
// wins, so it isn't added.
func domainMiddleware(domain *supporthttp.Domain) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(routeDomainKey, domain.String())

		if params, ok := domain.Match(ctx.Request.Host); ok {
			for _, name := range domain.Params() {
				if _, exists := ctx.Params.Get(name); exists {
					continue
				}
				ctx.Params = append(ctx.Params, gin.Param{Key: name, Value: params[name]})
			}
		}

		ctx.Next()
	}
}

// routeKey returns the key of a route in the routes map: the routes of each
// domain are kept apart, so the same path can be on several domains.
func routeKey(domain, path string) string {
	return domain + path
}
