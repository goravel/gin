package gin

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/goravel/framework/contracts/config"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	contractsview "github.com/goravel/framework/contracts/view"
	"github.com/goravel/framework/http/middleware"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/color"
	supporthttp "github.com/goravel/framework/support/http"
	"github.com/goravel/framework/support/str"
	"github.com/spf13/cast"
)

// map[domain+path]map[method]info
var routes = make(map[string]map[string]contractshttp.Info)

var globalRecoverCallback func(ctx contractshttp.Context, err any) = defaultRecoverCallback

type recoverMiddleware struct{}

func (r *recoverMiddleware) Signature() string {
	return "goravel:recover"
}

func (r *recoverMiddleware) Handle(ctx contractshttp.Context) {
	defer func() {
		if err := recover(); err != nil {
			globalRecoverCallback(ctx, err)
		}
	}()
	ctx.Request().Next()
}

type Route struct {
	route.Router
	config           config.Config
	driver           string
	globalMiddleware []contractshttp.Middleware
	instance         *gin.Engine
	domains          []*domainRouter
	engineHandlers   []gin.HandlerFunc // what every engine runs first: logs, body limit and global middleware
	server           *http.Server
	tlsServer        *http.Server
	templateOnce     sync.Once
	templateErr      error
	templateCallback func() (render.HTMLRender, error) // configured callback, invoked by ensureTemplate(), not init()
	htmlRender       render.HTMLRender                 // compiled renderer, re-attached by init() on engine rebuilds
}

// Compile-time assertion that Route satisfies view.Compiler.
var _ contractsview.Compiler = (*Route)(nil)

func NewRoute(config config.Config, parameters map[string]any) (*Route, error) {
	driver := cast.ToString(parameters["driver"])
	if driver == "" {
		return nil, errors.New("please set the driver")
	}

	timeout := time.Duration(config.GetInt("http.request_timeout", 3)) * time.Second
	globalMiddleware := []contractshttp.Middleware{Timeout(timeout), Cors(), Tls(), middleware.CheckForMaintenanceMode()}

	route := &Route{
		config:           config,
		driver:           cast.ToString(parameters["driver"]),
		globalMiddleware: globalMiddleware,
	}
	route.init(globalMiddleware)

	return route, nil
}

func (r *Route) Fallback(handler contractshttp.HandlerFunc) {
	r.instance.NoRoute(handlerToGinHandler(handler))
}

func (r *Route) GetGlobalMiddleware() []contractshttp.Middleware {
	return r.globalMiddleware
}

func (r *Route) GetRoutes() []contractshttp.Info {
	paths := []string{}
	for path := range routes {
		paths = append(paths, path)
	}

	sort.Strings(paths)
	methods := []string{contractshttp.MethodGet + "|" + contractshttp.MethodHead, contractshttp.MethodHead, contractshttp.MethodGet, contractshttp.MethodPost, contractshttp.MethodPut, contractshttp.MethodDelete, contractshttp.MethodPatch, contractshttp.MethodOptions, contractshttp.MethodAny, contractshttp.MethodResource, contractshttp.MethodStatic, contractshttp.MethodStaticFile, contractshttp.MethodStaticFS}

	var infos []contractshttp.Info
	for _, path := range paths {
		for _, method := range methods {
			if info, ok := routes[path][method]; ok {
				infos = append(infos, info)
			}
		}
	}

	return infos
}

func (r *Route) GlobalMiddleware(middleware ...contractshttp.Middleware) {
	r.globalMiddleware = append(r.globalMiddleware, middleware...)
	r.init(r.globalMiddleware)
}

func (r *Route) Listen(l net.Listener) error {
	if err := r.ensureTemplate(); err != nil {
		return err
	}

	r.outputRoutes()
	color.Green().Println("[HTTP] Listening on: " + str.Of(l.Addr().String()).Start("http://").String())

	r.server = &http.Server{
		Addr:           l.Addr().String(),
		Handler:        http.AllowQuerySemicolons(http.HandlerFunc(r.serveEngine)),
		MaxHeaderBytes: r.config.GetInt(fmt.Sprintf("http.drivers.%s.header_limit", r.driver), 4096) << 10,
	}

	if err := r.server.Serve(l); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (r *Route) ListenTLS(l net.Listener) error {
	return r.ListenTLSWithCert(l, r.config.GetString("http.tls.ssl.cert"), r.config.GetString("http.tls.ssl.key"))
}

func (r *Route) ListenTLSWithCert(l net.Listener, certFile, keyFile string) error {
	if err := r.ensureTemplate(); err != nil {
		return err
	}

	r.outputRoutes()
	color.Green().Println("[HTTPS] Listening on: " + str.Of(l.Addr().String()).Start("https://").String())

	r.tlsServer = &http.Server{
		Addr:           l.Addr().String(),
		Handler:        http.AllowQuerySemicolons(http.HandlerFunc(r.serveEngine)),
		MaxHeaderBytes: r.config.GetInt(fmt.Sprintf("http.drivers.%s.header_limit", r.driver), 4096) << 10,
	}

	if err := r.tlsServer.ServeTLS(l, certFile, keyFile); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (r *Route) Info(name string) contractshttp.Info {
	routes := r.GetRoutes()

	for _, route := range routes {
		if route.Name == name {
			return route
		}
	}

	return contractshttp.Info{}
}

func (r *Route) Recover(callback func(ctx contractshttp.Context, err any)) {
	globalRecoverCallback = callback
	r.init(r.globalMiddleware)
}

func (r *Route) Run(host ...string) error {
	if err := r.ensureTemplate(); err != nil {
		return err
	}

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
	color.Green().Println("[HTTP] Listening on: " + str.Of(host[0]).Start("http://").String())

	r.server = &http.Server{
		Addr:           host[0],
		Handler:        http.AllowQuerySemicolons(http.HandlerFunc(r.serveEngine)),
		MaxHeaderBytes: r.config.GetInt(fmt.Sprintf("http.drivers.%s.header_limit", r.driver), 4096) << 10,
	}

	if err := r.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
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
	if err := r.ensureTemplate(); err != nil {
		return err
	}

	if host == "" {
		return errors.New("host can't be empty")
	}
	if certFile == "" || keyFile == "" {
		return errors.New("certificate can't be empty")
	}

	r.outputRoutes()
	color.Green().Println("[HTTPS] Listening on: " + str.Of(host).Start("https://").String())

	r.tlsServer = &http.Server{
		Addr:           host,
		Handler:        http.AllowQuerySemicolons(http.HandlerFunc(r.serveEngine)),
		MaxHeaderBytes: r.config.GetInt(fmt.Sprintf("http.drivers.%s.header_limit", r.driver), 4096) << 10,
	}

	if err := r.tlsServer.ListenAndServeTLS(certFile, keyFile); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (r *Route) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if err := r.ensureTemplate(); err != nil {
		// ensureTemplate() runs before the recover middleware, so a template compile
		// error surfaces at request time. Log it and return a 500 instead of
		// panicking: when embedded as an http.Handler, a panic here would be
		// recovered by net/http without any HTTP response reaching the client.
		// The response body carries the specific compile error, so the failure stays
		// visible to the caller even when the log is not reachable.
		if LogFacade != nil {
			LogFacade.WithContext(request.Context()).Error(err)
		}
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}
	r.serveEngine(writer, request)
}

func (r *Route) SetGlobalMiddleware(middlewares []contractshttp.Middleware) {
	r.globalMiddleware = middlewares
	r.init(r.globalMiddleware)
}

func (r *Route) Shutdown(ctx ...context.Context) error {
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

func (r *Route) Test(request *http.Request) (*http.Response, error) {
	recorder := httptest.NewRecorder()

	r.ServeHTTP(recorder, request)

	return recorder.Result(), nil
}

func (r *Route) init(globalMiddleware []contractshttp.Middleware) {
	gin.SetMode(gin.ReleaseMode)
	gin.DisableBindValidation()
	bodyLimit := int64(r.config.GetInt(fmt.Sprintf("http.drivers.%s.body_limit", r.driver), defaultBodyLimitKB)) << 10
	if bodyLimit <= 0 {
		bodyLimit = defaultBodyLimit
	}
	engine := newEngine(bodyLimit, nil)

	ginMiddleware := []gin.HandlerFunc{}
	if r.config.GetBool("app.debug") {
		ginMiddleware = append(ginMiddleware, logMiddleware())
	}
	ginMiddleware = append(ginMiddleware, bodyLimitMiddleware(bodyLimit))

	globalMiddleware = append([]contractshttp.Middleware{&recoverMiddleware{}}, globalMiddleware...)
	r.engineHandlers = append(ginMiddleware, middlewaresToGinHandlers(globalMiddleware)...)
	engine.Use(r.engineHandlers...)

	template := r.config.Get("http.drivers." + r.driver + ".template")
	switch t := template.(type) {
	case render.HTMLRender:
		engine.HTMLRender = t
	case func() (render.HTMLRender, error):
		// Deferred to ensureTemplate(): init() may run before providers boot, so
		// compiling now would miss their views and funcs (goravel/goravel#989).
		r.templateCallback = t
	}
	// Carry a renderer compiled before this engine rebuild.
	if engine.HTMLRender == nil && r.htmlRender != nil {
		engine.HTMLRender = r.htmlRender
	}

	group := NewGroup(
		r.config,
		engine.Group("/"),
		"",
		[]contractshttp.Middleware{},
		[]contractshttp.Middleware{ResponseMiddleware()},
	).(*Group)
	group.domains = r
	r.Router = group
	r.instance = engine
	r.domains = nil
}

// domainRouter returns the router of the domain, creating its engine on first
// use, and the domain as the routes record it. The engine runs the same
// handlers as the main one, and passes a request it has no route for to the
// main engine, whose routes match any host. Domains that differ only in their
// parameter names share the engine, and each route reads its own names. An
// invalid domain panics, as an invalid route does.
func (r *Route) domainRouter(domain string) (gin.IRouter, string) {
	parsed, err := supporthttp.ParseDomain(domain)
	if err != nil {
		panic(err)
	}

	var router *domainRouter
	for _, existing := range r.domains {
		if existing.domain.Key() == parsed.Key() {
			router = existing
			break
		}
	}

	if router == nil {
		engine := newEngine(r.instance.MaxMultipartMemory, r.instance.HTMLRender)
		engine.NoRoute(func(ctx *gin.Context) {
			// A static route of the domain that doesn't find the file ends up
			// here after the domain's handlers ran, so it answers 404 rather
			// than running them again on the main engine.
			if _, ok := ctx.Get(routeDomainKey); ok {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			r.instance.ServeHTTP(ctx.Writer, ctx.Request)
			ctx.Abort()
		})

		router = &domainRouter{domain: parsed, engine: engine}
		r.domains = append(r.domains, router)
	}

	// The handlers go on the group rather than the engine, so they don't run
	// twice for a request passed to the main engine. The domain middleware is
	// first, so the global middleware finds the route's Info.
	handlers := append([]gin.HandlerFunc{domainMiddleware(parsed)}, r.engineHandlers...)

	return router.engine.Group("/", handlers...), parsed.String()
}

// newEngine returns an engine with the settings the main engine and the
// engines of the domains share.
func newEngine(maxMultipartMemory int64, htmlRender render.HTMLRender) *gin.Engine {
	engine := gin.New()
	engine.MaxMultipartMemory = maxMultipartMemory
	engine.HTMLRender = htmlRender

	return engine
}

// serveEngine serves the request with the engine of its domain.
func (r *Route) serveEngine(writer http.ResponseWriter, request *http.Request) {
	r.engineFor(request).ServeHTTP(writer, request)
}

// engineFor returns the engine of the request's domain: a domain without
// parameters first, then the ones with parameters in the order they were
// first registered, then the main engine.
func (r *Route) engineFor(request *http.Request) *gin.Engine {
	if len(r.domains) == 0 {
		return r.instance
	}

	for _, router := range r.domains {
		if router.domain.HasParams() {
			continue
		}
		if _, ok := router.domain.Match(request.Host); ok {
			return router.engine
		}
	}
	for _, router := range r.domains {
		if !router.domain.HasParams() {
			continue
		}
		if _, ok := router.domain.Match(request.Host); ok {
			return router.engine
		}
	}

	return r.instance
}

// ensureTemplate compiles the view set once, from the configured callback or
// DefaultTemplate() when there is none or it returns a nil renderer. It runs on CompileViews() or the first serve call, whichever
// comes first, so it sees every view and func registered during provider Boot.
// With no view files HTMLRender stays nil. A compile error is sticky.
func (r *Route) ensureTemplate() error {
	r.templateOnce.Do(func() {
		if r.instance.HTMLRender != nil {
			return
		}

		var htmlRender render.HTMLRender
		if r.templateCallback != nil {
			htmlRender, r.templateErr = r.templateCallback()
		}
		// A callback returning no renderer falls back to the default set, as it
		// did when init() invoked the callback.
		if r.templateErr == nil && htmlRender == nil {
			// Keep a nil *render.HTMLProduction from becoming a non-nil interface.
			production, err := DefaultTemplate()
			if production != nil {
				htmlRender = production
			}
			r.templateErr = err
		}
		if r.templateErr != nil {
			return
		}

		r.htmlRender = htmlRender
		if htmlRender != nil {
			r.instance.HTMLRender = htmlRender
			for _, router := range r.domains {
				router.engine.HTMLRender = htmlRender
			}
		}
	})
	return r.templateErr
}

// CompileViews implements view.Compiler; the framework calls it after every
// provider has booted so compile errors surface at startup.
func (r *Route) CompileViews() error {
	return r.ensureTemplate()
}

func (r *Route) outputRoutes() {
	if r.config.GetBool("app.debug") && support.RuntimeMode != support.RuntimeArtisan && support.RuntimeMode != support.RuntimeTest {
		if err := App.MakeArtisan().Call("route:list"); err != nil {
			color.Errorln(fmt.Errorf("print route list failed: %w", err))
		}
	}
}

func defaultRecoverCallback(ctx contractshttp.Context, err any) {
	LogFacade.WithContext(ctx).Request(ctx.Request()).Error(err)
	ctx.Request().Abort(http.StatusInternalServerError)
}
