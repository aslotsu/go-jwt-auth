package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"time"
	"user-auth/routes"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"127.0.0.1:3000", "localhost:3000"},
		AllowMethods:     []string{"POST", "PATCH"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "127.0.0.1:3000" || origin == "localhost:3000"
		},
		MaxAge: 12 * time.Hour,
	}))

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
