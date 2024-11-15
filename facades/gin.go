package facades

import (
	"fmt"
	"log"

	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/gin"
)

func Route(driver string) route.Route {
	if gin.App != nil {
		instance, err := gin.App.MakeWith(gin.RouteBinding, map[string]any{
			"driver": driver,
		})

		if err != nil {
			log.Fatalln(err)
			return nil
		}

		return instance.(*gin.Route)
	} else {
		fmt.Println("App is nil")
		return nil
	}
}
