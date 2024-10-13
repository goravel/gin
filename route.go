package gin

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/goravel/framework/contracts/config"
	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/color"
	"github.com/savioxavier/termlink"
)

type Route struct {
	route.Router
	config    config.Config
	instance  *gin.Engine
	server    *http.Server
	tlsServer *http.Server
}

func NewRoute(config config.Config, parameters map[string]any) (*Route, error) {
	gin.SetMode(gin.ReleaseMode)
	gin.DisableBindValidation()
	engine := gin.New()
	engine.MaxMultipartMemory = int64(config.GetInt("http.drivers.gin.body_limit", 4096)) << 10
	engine.Use(gin.Recovery()) // recovery middleware
	// timeoutMiddleware
	// timeout := TimeoutMiddleware(config.GetInt("http.request_timeout", 3) * time.Second)
	// engine.Use(middlewaresToGinHandlers([]httpcontract.Middleware{timeout})...)
	
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
	middlewares = append(middlewares, Cors(), Tls(), TImeoutMiddleware())
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
		defaultPort := r.config.GetString("http.port")
		if defaultPort == "" {
			return errors.New("port can't be empty")
		}
		completeHost := defaultHost + ":" + defaultPort
		host = append(host, completeHost)
	}

	r.outputRoutes()
	color.Green().Println(termlink.Link("[HTTP] Listening and serving HTTP on", "http://"+host[0]))

	r.server = &http.Server{
		Addr:           host[0],
		Handler:        http.AllowQuerySemicolons(r.instance),
		MaxHeaderBytes: r.config.GetInt("http.drivers.gin.header_limit", 4096) << 10,
	}

	if err := r.server.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
		return nil
	} else {
		return err
	}
}

func (r *Route) RunTLS(host ...string) error {
	if len(host) == 0 {
		defaultHost := r.config.GetString("http.tls.host")
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

	r.outputRoutes()
	color.Green().Println(termlink.Link("[HTTPS] Listening and serving HTTPS on", "https://"+host))

	r.tlsServer = &http.Server{
		Addr:           host,
		Handler:        http.AllowQuerySemicolons(r.instance),
		MaxHeaderBytes: r.config.GetInt("http.drivers.gin.header_limit", 4096) << 10,
	}

	if err := r.tlsServer.ListenAndServeTLS(certFile, keyFile); errors.Is(err, http.ErrServerClosed) {
		return nil
	} else {
		return err
	}
}

func (r *Route) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	r.instance.ServeHTTP(writer, request)
}

// DEPRECATED Use Stop instead.
func (r *Route) Shutdown(ctx ...context.Context) error {
	return r.Stop(ctx...)
}

func (r *Route) Stop(ctx ...context.Context) error {
	c := context.Background()
	if len(ctx) > 0 {
		c = ctx[0]
	}

	if r.server != nil {
		return r.server.Shutdown(c)
	}
	if r.tlsServer != nil {
		return r.tlsServer.Shutdown(c)
	}
	return nil
}

func (r *Route) outputRoutes() {
	if r.config.GetBool("app.debug") && support.Env != support.EnvArtisan {
		for _, item := range r.instance.Routes() {
			fmt.Printf("%-10s %s\n", item.Method, colonToBracket(item.Path))
		}
	}
}
