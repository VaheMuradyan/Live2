package router

import (
	"github.com/VaheMuradyan/Live2/api/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRouter(handler *handlers.PriceHandler, router *gin.Engine) {
	router.POST("/api/start", handler.Start)
}
