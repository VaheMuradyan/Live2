package generator

import (
	"github.com/VaheMuradyan/Live2/db/models"
	"github.com/VaheMuradyan/Live2/generator/markets"
	"gorm.io/gorm"
	"log"
	"time"
)

func (g *Generator) startScoreMonitoring() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-g.stopChan:
			return
		case <-ticker.C:
			g.checkScoreUpdate()
		}
	}
}

func (g *Generator) checkScoreUpdate() {
	var events []models.Event

	err := g.db.Where("events.active = ?", true).
		Joins("Competition").
		Joins("Competition.Country").
		Joins("Competition.Country.Sport").
		Preload("Score").
		Preload("Teams").
		Preload("Coefficients", func(db *gorm.DB) *gorm.DB {
			return db.Where("coefficients.active = ?", true)
		}).
		Preload("Coefficients.Price", func(db *gorm.DB) *gorm.DB {
			return db.Where("prices.active = ?", true)
		}).
		Preload("Coefficients.Price.Market", func(db *gorm.DB) *gorm.DB {
			return db.Where("markets.active = ?", true)
		}).
		Preload("Coefficients.Price.Market.MarketCollection").
		Find(&events).Error

	if err != nil {
		return
	}

	for _, event := range events {
		if event.Score == nil {
			continue
		}

		currentScore := models.ScoreSnapshot{
			EventID:    event.ID,
			Team1Score: event.Score.Team1Score,
			Team2Score: event.Score.Team2Score,
			Total:      event.Score.Total,
		}

		if prev, exists := g.scoreSnapshots.Load(event.ID); !exists {
			g.handleScoreChange(event, currentScore)
			g.scoreSnapshots.Store(event.ID, currentScore)
		} else {
			previousScore := prev.(models.ScoreSnapshot)
			if g.scoreHasChanged(previousScore, currentScore) {
				g.handleScoreChange(event, currentScore)
				g.scoreSnapshots.Store(event.ID, currentScore)
			}
		}
	}
}

func (g *Generator) scoreHasChanged(previous, current models.ScoreSnapshot) bool {
	return previous.Team1Score != current.Team1Score || previous.Team2Score != current.Team2Score
}

func (g *Generator) handleScoreChange(event models.Event, currentScore models.ScoreSnapshot) {
	g.checkAndStopMarkets(event, currentScore)
	g.sendActiveCoefficients(event, currentScore)
}

func (g *Generator) checkAndStopMarkets(event models.Event, scoreSnapshot models.ScoreSnapshot) {
	eventID := event.ID
	totalGoals := scoreSnapshot.Total

	var allCoeffs []models.Coefficient
	err := g.db.Preload("Price").Where("event_id = ?", eventID).Find(&allCoeffs).Error
	if err != nil {
		return
	}

	priceCodesToDeactivate := []string{}

	if totalGoals >= 1 {
		priceCodesToDeactivate = append(priceCodesToDeactivate, "U5", "O5")
	}
	if totalGoals >= 2 {
		priceCodesToDeactivate = append(priceCodesToDeactivate, "U15", "O15")
	}
	if totalGoals >= 3 {
		priceCodesToDeactivate = append(priceCodesToDeactivate, "U25", "O25")
	}
	if totalGoals >= 4 {
		priceCodesToDeactivate = append(priceCodesToDeactivate, "U35", "O35")
	}
	if totalGoals >= 5 {
		priceCodesToDeactivate = append(priceCodesToDeactivate, "U45", "O45")
	}

	if scoreSnapshot.Team1Score > 0 && scoreSnapshot.Team2Score > 0 {
		priceCodesToDeactivate = append(priceCodesToDeactivate, "BTTS_N", "BTTS_Y")
	}

	if len(priceCodesToDeactivate) == 0 {
		return
	}

	var coeffIDsToDeactivate []uint
	err = g.db.Model(&models.Coefficient{}).
		Joins("JOIN prices ON coefficients.price_id = prices.id").
		Where("coefficients.event_id = ? AND prices.code IN ? AND coefficients.active = ?",
			eventID, priceCodesToDeactivate, true).
		Select("coefficients.id").
		Find(&coeffIDsToDeactivate).Error

	if err != nil {
		return
	}

	if len(coeffIDsToDeactivate) == 0 {
		return
	}

	result := g.db.Model(&models.Coefficient{}).
		Where("id IN (?)", coeffIDsToDeactivate).
		Update("active", false)

	if result.Error != nil {
		log.Fatalf("Error deactivating coefficients for event %d: %v", eventID, result.Error)
	} else {
		log.Printf("Successfully deactivated %d coefficients for event %d", result.RowsAffected, eventID)
	}
}

func (g *Generator) sendActiveCoefficients(event models.Event, scoreSnapshot models.ScoreSnapshot) {
	eventID := event.ID

	var coefficients []models.Coefficient
	err := g.db.Preload("Price").Preload("Price.Market").
		Joins("JOIN prices ON coefficients.price_id = prices.id").
		Joins("JOIN markets ON prices.market_id = markets.id").
		Where("coefficients.event_id = ? AND coefficients.active = ? AND markets.active = ? AND prices.active = ?",
			eventID, true, true, true).
		Find(&coefficients).Error

	if err != nil {
		return
	}

	for _, coefficient := range coefficients {
		newCoeff := g.calculateNewCoefficient(coefficient, scoreSnapshot)

		coefficient.Coefficient = newCoeff

		if err = g.db.Save(&coefficient).Error; err != nil {
			continue
		}

		if err = g.client.SendToCentrifugo(coefficient); err != nil {
			continue
		}

	}
}

func (g *Generator) calculateNewCoefficient(coefficient models.Coefficient, score models.ScoreSnapshot) float64 {
	market := coefficient.Price.Market
	price := coefficient.Price

	switch market.Code {
	case "1X2":
		return markets.Calculate1x2Coefficient(price.Code, score)
	case "OU5":
		return markets.CalculateOverUnderCoefficient(price.Code, score)
	case "OU15":
		return markets.CalculateOverUnderCoefficient(price.Code, score)
	case "OU25":
		return markets.CalculateOverUnderCoefficient(price.Code, score)
	case "OU35":
		return markets.CalculateOverUnderCoefficient(price.Code, score)
	case "OU45":
		return markets.CalculateOverUnderCoefficient(price.Code, score)
	case "BTTS":
		return markets.CalculateBTTSCoefficient(price.Code, score)
	}
	return 0
}
