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
		AllowOrigins:     []string{"https://trains-git-main-aslotsu.vercel.app"},
		AllowMethods:     []string{"POST", "OPTIONS", "GET", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "content-type"},
		ExposeHeaders:    []string{"Origin", "Content-Length", "Content-Type", "content-type"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://trains-git-main-aslotsu.vercel.app"
		},
		MaxAge: 12 * time.Hour,
	}))

	routes.BusinessRouter(router)
	routes.ShopperRouter(router)
	router.GET("/hi", func(context *gin.Context) {
		context.JSON(200, "The router is working")
	})

	if err := router.Run(":9000"); err != nil {
		log.Println(err)
		return
	}
}
