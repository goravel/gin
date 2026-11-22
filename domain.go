package gin

import (
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// domainParamRegex matches the parameters of a domain, such as {account}.
var domainParamRegex = regexp.MustCompile(`{([^}]+)}`)

// domainRouter holds the routes of a domain in an engine of their own, so the
// same path can be registered on several domains.
type domainRouter struct {
	domain  string
	pattern *regexp.Regexp
	params  []string
	engine  *gin.Engine
}

func newDomainRouter(domain string, engine *gin.Engine) *domainRouter {
	router := &domainRouter{
		domain: strings.ToLower(domain),
		engine: engine,
	}

	if matches := domainParamRegex.FindAllStringSubmatch(router.domain, -1); len(matches) > 0 {
		for _, match := range matches {
			router.params = append(router.params, match[1])
		}

		// Each parameter matches one label of the host.
		parts := domainParamRegex.Split(router.domain, -1)
		for i := range parts {
			parts[i] = regexp.QuoteMeta(parts[i])
		}
		router.pattern = regexp.MustCompile("^" + strings.Join(parts, "([^.]+)") + "$")
	}

	return router
}

// match reports whether the host belongs to the domain, with the values of its
// parameters.
func (r *domainRouter) match(host string) (gin.Params, bool) {
	if r.pattern == nil {
		return nil, host == r.domain
	}

	matches := r.pattern.FindStringSubmatch(host)
	if matches == nil {
		return nil, false
	}

	params := make(gin.Params, 0, len(r.params))
	for i, name := range r.params {
		params = append(params, gin.Param{Key: name, Value: matches[i+1]})
	}

	return params, true
}

// domainParamsMiddleware adds the parameters of the domain to the route
// parameters, so they are read like the ones of the path.
func (r *domainRouter) domainParamsMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if params, ok := r.match(requestHost(ctx.Request)); ok {
			ctx.Params = append(ctx.Params, params...)
		}

		ctx.Next()
	}
}

// requestHost returns the host of the request in lower case, without its port.
func requestHost(request *http.Request) string {
	host := request.Host
	if parsed, _, err := net.SplitHostPort(host); err == nil {
		host = parsed
	}

	return strings.ToLower(strings.TrimSuffix(host, "."))
}
