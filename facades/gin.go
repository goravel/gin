package facades

import (
	"log"

	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/gin"
)

func Route() route.Engine {
	instance, err := gin.App.Make(gin.RouteBinding)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	return instance.(*gin.Route)
}
