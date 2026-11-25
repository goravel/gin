package gin

import (
	contractshttp "github.com/goravel/framework/contracts/http"
	contractsroute "github.com/goravel/framework/contracts/route"
)

type Action struct {
	method string
	// key is the route's key in the routes map: its domain and path.
	key string
}

func NewAction(method, path, handler string) contractsroute.Action {
	return newAction("", method, path, handler)
}

// newAction records a route on the domain, empty for any host. The routes of
// each domain are kept apart, so the same path can be on several domains.
func newAction(domain, method, path, handler string) contractsroute.Action {
	key := routeKey(domain, path)
	if _, ok := routes[key]; !ok {
		routes[key] = make(map[string]contractshttp.Info)
	}

	if method == contractshttp.MethodGet {
		method = contractshttp.MethodGet + "|" + contractshttp.MethodHead
	}

	routes[key][method] = contractshttp.Info{
		Domain:  domain,
		Handler: handler,
		Method:  method,
		Path:    path,
	}

	return &Action{
		method: method,
		key:    key,
	}
}

func (r *Action) Name(name string) contractsroute.Action {
	info := routes[r.key][r.method]
	info.Name = name
	routes[r.key][r.method] = info

	return r
}

func (r *Action) WithoutMiddleware(middleware ...contractshttp.Middleware) contractsroute.Action {
	info := routes[r.key][r.method]
	info.ExcludedMiddleware = append(info.ExcludedMiddleware, middleware...)
	routes[r.key][r.method] = info

	return r
}
