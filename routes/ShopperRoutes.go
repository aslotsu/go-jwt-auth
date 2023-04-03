package routes

import (
	"github.com/gin-gonic/gin"
	"user-auth/controllers"
)

func ShopperRouter(routes *gin.Engine) {
	routes.POST("/users/signup", controllers.SignUpShopper)
	routes.POST("/users/login", controllers.LoginShopper)
	routes.POST("/users/logout", controllers.LogoutShopper)
	routes.GET("/users/get-user", controllers.GetUser)
}
