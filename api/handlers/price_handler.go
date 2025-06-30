package handlers

import (
	"github.com/VaheMuradyan/Live2/api/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type body struct {
	Ank string `json:"ank"`
}

type response struct {
	Bar string `json:"bar"`
}
type PriceHandler struct {
	service *services.PriceService
}

func NewHandler(service *services.PriceService) *PriceHandler {
	return &PriceHandler{service: service}
}

func (h *PriceHandler) Start(c *gin.Context) {
	var reqBody body

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.InchvorBan(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "chi ashxatum"})
		return
	}

	res := &response{Bar: "sax tuyna ashxatuma"}

	c.IndentedJSON(http.StatusOK, res)
}
