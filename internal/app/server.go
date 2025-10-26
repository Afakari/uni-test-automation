package app

import "github.com/gin-gonic/gin"

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/register", RegisterHandler)
	r.POST("/login", LoginHandler)

	protected := r.Group("/todos")
	protected.Use(AuthMiddleware())
	{
		protected.GET("", GetTodosHandler)
		protected.POST("", CreateTodoHandler)
		protected.GET("/:id", GetTodoHandler)
		protected.PUT("/:id", UpdateTodoHandler)
		protected.DELETE("/:id", DeleteTodoHandler)
	}

	return r
}
