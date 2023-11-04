package gin

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/config"
	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/support"
	"github.com/savioxavier/termlink"
)

type Route struct {
	route.Router
	config   config.Config
	instance *gin.Engine
}

func NewRoute(config config.Config, parameters map[string]any) (*Route, error) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	if debugLog := getDebugLog(config); debugLog != nil {
		engine.Use(debugLog)
	}

	if driver, exist := parameters["driver"]; exist {
		htmlRender, ok := config.Get("http.drivers." + driver.(string) + ".template").(render.HTMLRender)
		if ok {
			engine.HTMLRender = htmlRender
		} else {
			htmlRenderCallback, ok := config.Get("http.drivers." + driver.(string) + ".template").(func() (render.HTMLRender, error))
			if ok {
				htmlRender, err := htmlRenderCallback()
				if err != nil {
					return nil, err
				}

				engine.HTMLRender = htmlRender
			}
		}
	}

	if engine.HTMLRender == nil {
		var err error
		engine.HTMLRender, err = DefaultTemplate()
		if err != nil {
			return nil, err
		}
	}

	return &Route{
		Router: NewGroup(
			config,
			engine.Group("/"),
			"",
			[]httpcontract.Middleware{},
			[]httpcontract.Middleware{ResponseMiddleware()},
		),
		config:   config,
		instance: engine,
	}, nil
}

func (r *Route) Fallback(handler httpcontract.HandlerFunc) {
	r.instance.NoRoute(handlerToGinHandler(handler))
}

func (r *Route) GlobalMiddleware(middlewares ...httpcontract.Middleware) {
	middlewares = append(middlewares, Cors(), Tls())
	r.instance.Use(middlewaresToGinHandlers(middlewares)...)
	r.Router = NewGroup(
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
	color.Greenln("[HTTP] Listening and serving HTTP on" + termlink.Link("", "http://"+host[0]))

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
	color.Greenln("[HTTPS] Listening and serving HTTPS on" + termlink.Link("", "https://"+host))

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
