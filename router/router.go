package router

import (
	"github.com/VaheMuradyan/Live2/api/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRouter(router *gin.Engine, handler *handlers.PriceHandler) {
	router.POST("/api/start", handler.Start)
}
