package prices

import (
	"github.com/VaheMuradyan/Live2/db/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type PriceHandler struct {
	service     *PriceService
	eventCodes  []string
	marketCodes []string
}

func NewHandler(service *PriceService) *PriceHandler {
	return &PriceHandler{
		service:     service,
		eventCodes:  []string{"MA", "BB", "JM", "PM", "RB"},
		marketCodes: []string{"1X2", "BTTS", "OU5", "OU15", "OU25", "OU35", "OU45"},
	}
}

func (h *PriceHandler) Start(c *gin.Context) {
	var req models.RequestData

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cant bind request"})
		return
	}

	if h.validate(req) == false {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cant validate request"})
		return
	}

	if err := h.service.ActivateData(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "successfully activated", "data": req})
}

func (h *PriceHandler) GetEvenetList(c *gin.Context) {
	list := h.service.GetEventList()
	if list == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cant get events list"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *PriceHandler) validate(req models.RequestData) bool {
	validEvents := make(map[string]struct{})
	for _, code := range h.eventCodes {
		validEvents[code] = struct{}{}
	}

	validMarkets := make(map[string]struct{})
	for _, code := range h.marketCodes {
		validMarkets[code] = struct{}{}
	}

	for _, code := range req.EventCodes {
		if _, ok := validEvents[code]; !ok {
			return false
		}
	}

	for _, code := range req.MarketCodes {
		if _, ok := validMarkets[code]; !ok {
			return false
		}
	}

	return true
}
