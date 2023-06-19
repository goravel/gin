package facades

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/gin"
	"log"
)

func Route() route.Engine {
	instance, err := gin.App.Make(gin.Binding)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	return instance.(*gin.GinRoute)
}

func Http() http.Context {
	instance, err := gin.App.Make(gin.Binding)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	return instance.(*gin.GinContext)
}
