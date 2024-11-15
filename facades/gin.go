package facades

import (
	"log"

	"github.com/goravel/framework/contracts/route"

	"github.com/goravel/gin"
)

func Route(driver string) route.Route {
	if  gin.App != nil {
		instance, err := gin.App.MakeWith(gin.RouteBinding, map[string]any{
			"driver": driver,
		})
	} else {
    		fmt.Println("App is nil")
	}
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	return instance.(*gin.Route)
}
