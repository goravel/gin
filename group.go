package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goravel/framework/contracts/config"
	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
)

type Group struct {
	config            config.Config
	instance          gin.IRouter
	originPrefix      string
	prefix            string
	originMiddlewares []httpcontract.Middleware
	middlewares       []httpcontract.Middleware
	lastMiddlewares   []httpcontract.Middleware
}

func NewGroup(config config.Config, instance gin.IRouter, prefix string, originMiddlewares []httpcontract.Middleware, lastMiddlewares []httpcontract.Middleware) route.Route {
	return &Group{
		config:            config,
		instance:          instance,
		originPrefix:      prefix,
		originMiddlewares: originMiddlewares,
		lastMiddlewares:   lastMiddlewares,
	}
}

func (r *Group) Group(handler route.GroupFunc) {
	var middlewares []httpcontract.Middleware
	middlewares = append(middlewares, r.originMiddlewares...)
	middlewares = append(middlewares, r.middlewares...)
	r.middlewares = []httpcontract.Middleware{}
	prefix := pathToGinPath(r.originPrefix + "/" + r.prefix)
	r.prefix = ""

	handler(NewGroup(r.config, r.instance, prefix, middlewares, r.lastMiddlewares))
}

func (r *Group) Prefix(addr string) route.Route {
	r.prefix += "/" + addr

	return r
}

func (r *Group) Middleware(middlewares ...httpcontract.Middleware) route.Route {
	r.middlewares = append(r.middlewares, middlewares...)

	return r
}

func (r *Group) Any(relativePath string, handler httpcontract.HandlerFunc) {
	r.getRoutesWithMiddlewares().Any(pathToGinPath(relativePath), []gin.HandlerFunc{handlerToGinHandler(handler)}...)
	r.clearMiddlewares()
}

func (r *Group) Get(relativePath string, handler httpcontract.HandlerFunc) {
	r.getRoutesWithMiddlewares().GET(pathToGinPath(relativePath), []gin.HandlerFunc{handlerToGinHandler(handler)}...)
	r.clearMiddlewares()
}

func (r *Group) Post(relativePath string, handler httpcontract.HandlerFunc) {
	r.getRoutesWithMiddlewares().POST(pathToGinPath(relativePath), []gin.HandlerFunc{handlerToGinHandler(handler)}...)
	r.clearMiddlewares()
}

func (r *Group) Delete(relativePath string, handler httpcontract.HandlerFunc) {
	r.getRoutesWithMiddlewares().DELETE(pathToGinPath(relativePath), []gin.HandlerFunc{handlerToGinHandler(handler)}...)
	r.clearMiddlewares()
}

func (r *Group) Patch(relativePath string, handler httpcontract.HandlerFunc) {
	r.getRoutesWithMiddlewares().PATCH(pathToGinPath(relativePath), []gin.HandlerFunc{handlerToGinHandler(handler)}...)
	r.clearMiddlewares()
}

func (r *Group) Put(relativePath string, handler httpcontract.HandlerFunc) {
	r.getRoutesWithMiddlewares().PUT(pathToGinPath(relativePath), []gin.HandlerFunc{handlerToGinHandler(handler)}...)
	r.clearMiddlewares()
}

func (r *Group) Options(relativePath string, handler httpcontract.HandlerFunc) {
	r.getRoutesWithMiddlewares().OPTIONS(pathToGinPath(relativePath), []gin.HandlerFunc{handlerToGinHandler(handler)}...)
	r.clearMiddlewares()
}

func (r *Group) Resource(relativePath string, controller httpcontract.ResourceController) {
	r.getRoutesWithMiddlewares().GET(pathToGinPath(relativePath), []gin.HandlerFunc{handlerToGinHandler(controller.Index)}...)
	r.getRoutesWithMiddlewares().POST(pathToGinPath(relativePath), []gin.HandlerFunc{handlerToGinHandler(controller.Store)}...)
	r.getRoutesWithMiddlewares().GET(pathToGinPath(relativePath+"/{id}"), []gin.HandlerFunc{handlerToGinHandler(controller.Show)}...)
	r.getRoutesWithMiddlewares().PUT(pathToGinPath(relativePath+"/{id}"), []gin.HandlerFunc{handlerToGinHandler(controller.Update)}...)
	r.getRoutesWithMiddlewares().PATCH(pathToGinPath(relativePath+"/{id}"), []gin.HandlerFunc{handlerToGinHandler(controller.Update)}...)
	r.getRoutesWithMiddlewares().DELETE(pathToGinPath(relativePath+"/{id}"), []gin.HandlerFunc{handlerToGinHandler(controller.Destroy)}...)
	r.clearMiddlewares()
}

func (r *Group) Static(relativePath, root string) {
	r.getRoutesWithMiddlewares().Static(pathToGinPath(relativePath), root)
	r.clearMiddlewares()
}

func (r *Group) StaticFile(relativePath, filepath string) {
	r.getRoutesWithMiddlewares().StaticFile(pathToGinPath(relativePath), filepath)
	r.clearMiddlewares()
}

func (r *Group) StaticFS(relativePath string, fs http.FileSystem) {
	r.getRoutesWithMiddlewares().StaticFS(pathToGinPath(relativePath), fs)
	r.clearMiddlewares()
}

func (r *Group) getRoutesWithMiddlewares() gin.IRoutes {
	prefix := pathToGinPath(r.originPrefix + "/" + r.prefix)

	r.prefix = ""
	ginGroup := r.instance.Group(prefix)

	var middlewares []gin.HandlerFunc
	ginOriginMiddlewares := middlewaresToGinHandlers(r.originMiddlewares)
	ginMiddlewares := middlewaresToGinHandlers(r.middlewares)
	ginLastMiddlewares := middlewaresToGinHandlers(r.lastMiddlewares)

	middlewares = append(middlewares, ginOriginMiddlewares...)
	middlewares = append(middlewares, ginMiddlewares...)
	middlewares = append(middlewares, ginLastMiddlewares...)

	if len(middlewares) > 0 {
		return ginGroup.Use(middlewares...)
	}

	return ginGroup
}

func (r *Group) clearMiddlewares() {
	r.middlewares = []httpcontract.Middleware{}
}
