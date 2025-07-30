package router

import (
	"github.com/VaheMuradyan/Live2/prices"
	"github.com/gin-gonic/gin"
)

func SetupRouter(router *gin.Engine, handler *prices.PriceHandler) {
	router.POST("/api/start", handler.Start)
	router.GET("/api/get-events", handler.GetEvenetList)
}
