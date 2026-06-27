package gin

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/goravel/framework/contracts/config"
	contractshttp "github.com/goravel/framework/contracts/http"
	contractsroute "github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/support/debug"
	"github.com/goravel/framework/support/str"
)

type Group struct {
	config              config.Config
	instance            gin.IRouter
	prefix              string
	middlewares         []contractshttp.Middleware
	lastMiddlewares     []contractshttp.Middleware
	excludedMiddlewares []contractshttp.Middleware
}

func NewGroup(config config.Config, instance gin.IRouter, prefix string, middlewares []contractshttp.Middleware, lastMiddlewares []contractshttp.Middleware) contractsroute.Router {
	return &Group{
		config:          config,
		instance:        instance,
		prefix:          prefix,
		middlewares:     middlewares,
		lastMiddlewares: lastMiddlewares,
	}
}

func (r *Group) Group(handler contractsroute.GroupFunc) {
	handler(&Group{
		config:              r.config,
		instance:            r.instance,
		prefix:              r.getFullPath(""),
		middlewares:         r.middlewares,
		lastMiddlewares:     r.lastMiddlewares,
		excludedMiddlewares: r.excludedMiddlewares,
	})
}

func (r *Group) Prefix(path string) contractsroute.Router {
	return &Group{
		config:              r.config,
		instance:            r.instance,
		prefix:              r.getFullPath(path),
		middlewares:         r.middlewares,
		lastMiddlewares:     r.lastMiddlewares,
		excludedMiddlewares: r.excludedMiddlewares,
	}
}

func (r *Group) Middleware(middlewares ...contractshttp.Middleware) contractsroute.Router {
	return &Group{
		config:              r.config,
		instance:            r.instance,
		prefix:              r.getFullPath(""),
		middlewares:         append(r.middlewares, middlewares...),
		lastMiddlewares:     r.lastMiddlewares,
		excludedMiddlewares: r.excludedMiddlewares,
	}
}

func (r *Group) WithoutMiddleware(middlewares ...contractshttp.Middleware) contractsroute.Router {
	return &Group{
		config:              r.config,
		instance:            r.instance,
		prefix:              r.getFullPath(""),
		middlewares:         r.middlewares,
		lastMiddlewares:     r.lastMiddlewares,
		excludedMiddlewares: append(r.excludedMiddlewares, middlewares...),
	}
}

func (r *Group) Any(path string, handler contractshttp.HandlerFunc) contractsroute.Action {
	r.WithMiddlewares().Any(r.getGinFullPath(path), []gin.HandlerFunc{handlerToGinHandler(handler)}...)

	return NewAction(contractshttp.MethodAny, r.getFullPath(path), r.getHandlerName(handler))
}

func (r *Group) Get(path string, handler contractshttp.HandlerFunc) contractsroute.Action {
	ginFullPath := r.getGinFullPath(path)
	r.WithMiddlewares().GET(ginFullPath, []gin.HandlerFunc{handlerToGinHandler(handler)}...)
	r.WithMiddlewares().HEAD(ginFullPath, []gin.HandlerFunc{handlerToGinHandler(handler)}...)

	return NewAction(contractshttp.MethodGet, r.getFullPath(path), r.getHandlerName(handler))
}

func (r *Group) Post(path string, handler contractshttp.HandlerFunc) contractsroute.Action {
	r.WithMiddlewares().POST(r.getGinFullPath(path), []gin.HandlerFunc{handlerToGinHandler(handler)}...)

	return NewAction(contractshttp.MethodPost, r.getFullPath(path), r.getHandlerName(handler))
}

func (r *Group) Delete(path string, handler contractshttp.HandlerFunc) contractsroute.Action {
	r.WithMiddlewares().DELETE(r.getGinFullPath(path), []gin.HandlerFunc{handlerToGinHandler(handler)}...)

	return NewAction(contractshttp.MethodDelete, r.getFullPath(path), r.getHandlerName(handler))
}

func (r *Group) Patch(path string, handler contractshttp.HandlerFunc) contractsroute.Action {
	r.WithMiddlewares().PATCH(r.getGinFullPath(path), []gin.HandlerFunc{handlerToGinHandler(handler)}...)

	return NewAction(contractshttp.MethodPatch, r.getFullPath(path), r.getHandlerName(handler))
}

func (r *Group) Put(path string, handler contractshttp.HandlerFunc) contractsroute.Action {
	r.WithMiddlewares().PUT(r.getGinFullPath(path), []gin.HandlerFunc{handlerToGinHandler(handler)}...)

	return NewAction(contractshttp.MethodPut, r.getFullPath(path), r.getHandlerName(handler))
}

func (r *Group) Options(path string, handler contractshttp.HandlerFunc) contractsroute.Action {
	r.WithMiddlewares().OPTIONS(r.getGinFullPath(path), []gin.HandlerFunc{handlerToGinHandler(handler)}...)

	return NewAction(contractshttp.MethodOptions, r.getFullPath(path), r.getHandlerName(handler))
}

func (r *Group) Resource(path string, controller contractshttp.ResourceController) contractsroute.Action {
	ginFullPath := r.getGinFullPath(path)
	r.WithMiddlewares().GET(ginFullPath, []gin.HandlerFunc{handlerToGinHandler(controller.Index)}...)
	r.WithMiddlewares().POST(ginFullPath, []gin.HandlerFunc{handlerToGinHandler(controller.Store)}...)

	ginFullPathWithID := r.getGinFullPath(path + "/{id}")
	r.WithMiddlewares().GET(ginFullPathWithID, []gin.HandlerFunc{handlerToGinHandler(controller.Show)}...)
	r.WithMiddlewares().PUT(ginFullPathWithID, []gin.HandlerFunc{handlerToGinHandler(controller.Update)}...)
	r.WithMiddlewares().PATCH(ginFullPathWithID, []gin.HandlerFunc{handlerToGinHandler(controller.Update)}...)
	r.WithMiddlewares().DELETE(ginFullPathWithID, []gin.HandlerFunc{handlerToGinHandler(controller.Destroy)}...)

	return NewAction(contractshttp.MethodResource, r.getFullPath(path), r.getHandlerName(controller))
}

func (r *Group) Static(path, root string) contractsroute.Action {
	fullPath := r.getFullPath(path)
	r.WithMiddlewares().Static(pathToGinPath(fullPath), root)

	return NewAction(contractshttp.MethodStatic, fullPath, r.getHandlerName(nil))
}

func (r *Group) StaticFile(path, filepath string) contractsroute.Action {
	r.WithMiddlewares().StaticFile(r.getGinFullPath(path), filepath)

	return NewAction(contractshttp.MethodStaticFile, r.getFullPath(path), r.getHandlerName(nil))
}

func (r *Group) StaticFS(path string, fs http.FileSystem) contractsroute.Action {
	r.WithMiddlewares().StaticFS(r.getGinFullPath(path), fs)

	return NewAction(contractshttp.MethodStaticFS, r.getFullPath(path), r.getHandlerName(nil))
}

func (r *Group) getFullPath(path string) string {
	if path == "" {
		return r.prefix
	}

	return r.prefix + str.Of(path).Start("/").String()
}

func (r *Group) getGinFullPath(path string) string {
	return pathToGinPath(r.getFullPath(path))
}

func (r *Group) WithMiddlewares() gin.IRoutes {
	ginGroup := r.instance.Group("")
	ginMiddlewares := middlewaresToGinHandlers(r.excludeMiddlewares(append(r.middlewares, r.lastMiddlewares...)))

	if len(ginMiddlewares) > 0 {
		return ginGroup.Use(ginMiddlewares...)
	}

	return ginGroup
}

func (r *Group) excludeMiddlewares(middlewares []contractshttp.Middleware) []contractshttp.Middleware {
	if len(r.excludedMiddlewares) == 0 {
		return middlewares
	}

	var result []contractshttp.Middleware
	for _, middleware := range middlewares {
		excluded := false
		for _, ex := range r.excludedMiddlewares {
			if isSameMiddleware(ex, middleware) {
				excluded = true
				break
			}
		}
		if !excluded {
			result = append(result, middleware)
		}
	}

	return result
}

func (r *Group) getHandlerName(handler any) string {
	if handler == nil {
		return ""
	}

	if res, ok := handler.(contractshttp.ResourceController); ok {
		var (
			prefix string
			t      = reflect.TypeOf(res)
		)
		if t.Kind() == reflect.Pointer {
			prefix = "*"
			t = t.Elem()
		}

		return fmt.Sprintf("%s.(%s%s)", t.PkgPath(), prefix, t.Name())
	}

	return debug.GetFuncInfo(handler).Name
}
