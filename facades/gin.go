package facades

import (
	"log"

	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/gin"
)

func Route(driver string) route.Route {
	if gin.App == nil {
		panic("App is nil")
	}
	instance, err := gin.App.MakeWith(gin.RouteBinding, map[string]any{
		"driver": driver,
	})

	if err != nil {
		log.Fatalln(err)
		return nil
	}

	return instance.(*gin.Route)
}
