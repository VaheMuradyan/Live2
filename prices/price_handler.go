package prices

import (
	"github.com/VaheMuradyan/Live2/db/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type PriceHandler struct {
	service *PriceService
}

func NewHandler(service *PriceService) *PriceHandler {
	return &PriceHandler{service: service}
}

func (h *PriceHandler) Start(c *gin.Context) {
	var req models.RequestData

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cant bind request"})
		return
	}

	if err := h.service.ActivateData(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "successfully activated", "data": req})
}
