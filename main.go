package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"user-auth/routes"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.Use(cors.Default())
	routes.BusinessRouter(router)
	routes.ShopperRouter(router)
	router.GET("/hi", func(context *gin.Context) {
		context.JSON(200, "The router is working")
	})

	if err := router.Run(":8901"); err != nil {
		log.Println(err)
		return
	}
}
