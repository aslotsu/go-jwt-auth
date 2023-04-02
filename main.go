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
		AllowOrigins:     []string{"https://go-jwt-auth-production.up.railway.app"},
		AllowMethods:     []string{"POST", "PATCH"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://go-jwt-auth-production.up.railway.app"
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
