package gin

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/config"
	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/support"
)

type Route struct {
	route.Route
	config   config.Config
	instance *gin.Engine
}

func NewRoute(config config.Config) *Route {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	if debugLog := getDebugLog(config); debugLog != nil {
		engine.Use(debugLog)
	}

	return &Route{
		Route: NewGroup(
			config,
			engine.Group("/"),
			"",
			[]httpcontract.Middleware{},
			[]httpcontract.Middleware{ResponseMiddleware()},
		),
		config:   config,
		instance: engine,
	}
}

func (r *Route) Fallback(handler httpcontract.HandlerFunc) {
	r.instance.NoRoute(handlerToGinHandler(handler))
}

func (r *Route) GlobalMiddleware(middlewares ...httpcontract.Middleware) {
	middlewares = append(middlewares, Tls())

	if len(middlewares) > 0 {
		r.instance.Use(middlewaresToGinHandlers(middlewares)...)
	}
	r.Route = NewGroup(
		r.config,
		r.instance.Group("/"),
		"",
		[]httpcontract.Middleware{},
		[]httpcontract.Middleware{ResponseMiddleware()},
	)
}

func (r *Route) Run(host ...string) error {
	if len(host) == 0 {
		defaultHost := r.config.GetString("http.host")
		if defaultHost == "" {
			return errors.New("host can't be empty")
		}

		defaultPort := r.config.GetString("http.port")
		if defaultPort == "" {
			return errors.New("port can't be empty")
		}
		completeHost := defaultHost + ":" + defaultPort
		host = append(host, completeHost)
	}

	r.outputRoutes()
	color.Greenln("[HTTP] Listening and serving HTTP on " + host[0])

	server := &http.Server{
		Addr:    host[0],
		Handler: http.AllowQuerySemicolons(r.instance),
	}

	return server.ListenAndServe()
}

func (r *Route) RunTLS(host ...string) error {
	if len(host) == 0 {
		defaultHost := r.config.GetString("http.tls.host")
		if defaultHost == "" {
			return errors.New("host can't be empty")
		}

		defaultPort := r.config.GetString("http.tls.port")
		if defaultPort == "" {
			return errors.New("port can't be empty")
		}
		completeHost := defaultHost + ":" + defaultPort
		host = append(host, completeHost)
	}

	certFile := r.config.GetString("http.tls.ssl.cert")
	keyFile := r.config.GetString("http.tls.ssl.key")

	return r.RunTLSWithCert(host[0], certFile, keyFile)
}

func (r *Route) RunTLSWithCert(host, certFile, keyFile string) error {
	if host == "" {
		return errors.New("host can't be empty")
	}
	if certFile == "" || keyFile == "" {
		return errors.New("certificate can't be empty")
	}
	if strings.HasPrefix(certFile, "/") {
		certFile = "." + certFile
	}
	if strings.HasPrefix(keyFile, "/") {
		keyFile = "." + keyFile
	}

	r.outputRoutes()
	color.Greenln("[HTTPS] Listening and serving HTTPS on " + host)

	return r.instance.RunTLS(host, certFile, keyFile)
}

func (r *Route) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	r.instance.ServeHTTP(writer, request)
}

func (r *Route) outputRoutes() {
	if r.config.GetBool("app.debug") && support.Env != support.EnvArtisan {
		for _, item := range r.instance.Routes() {
			fmt.Printf("%-10s %s\n", item.Method, colonToBracket(item.Path))
		}
	}
}
