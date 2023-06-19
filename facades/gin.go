package facades

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/gin"
	"log"
)

func Http() http.Context {
	instance, err := gin.App.Make(gin.HttpBinding)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	return instance.(*gin.GinContext)
}

func Route() route.Engine {
	instance, err := gin.App.Make(gin.RouteBinding)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	return instance.(*gin.GinRoute)
}
