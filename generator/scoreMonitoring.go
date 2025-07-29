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
		Preload("EventPrices", func(db *gorm.DB) *gorm.DB {
			return db.Where("event_prices.active = ?", true)
		}).
		Preload("EventPrices.Price", func(db *gorm.DB) *gorm.DB {
			return db.Where("prices.active = ?", true)
		}).
		Preload("EventPrices.Price.Market", func(db *gorm.DB) *gorm.DB {
			return db.Where("markets.active = ?", true)
		}).
		Preload("EventPrices.Price.Market.MarketCollection").
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

	priceCodesToDeactivate := []string{}

	if totalGoals >= 5 {
		priceCodesToDeactivate = append(priceCodesToDeactivate, "U45", "O45")
	} else if totalGoals >= 4 {
		priceCodesToDeactivate = append(priceCodesToDeactivate, "U35", "O35")
	} else if totalGoals >= 3 {
		priceCodesToDeactivate = append(priceCodesToDeactivate, "U25", "O25")
	} else if totalGoals >= 2 {
		priceCodesToDeactivate = append(priceCodesToDeactivate, "U15", "O15")
	} else if totalGoals >= 1 {
		priceCodesToDeactivate = append(priceCodesToDeactivate, "U5", "O5")
	}

	if scoreSnapshot.Team1Score > 0 && scoreSnapshot.Team2Score > 0 {
		priceCodesToDeactivate = append(priceCodesToDeactivate, "BTTS_N", "BTTS_Y")
	}

	if len(priceCodesToDeactivate) == 0 {
		return
	}

	result := g.db.Model(&models.EventPrice{}).
		Where("event_id = ? AND active = ? AND price_id IN (?)",
			eventID,
			true,
			g.db.Model(&models.Price{}).Select("id").Where("code IN ?", priceCodesToDeactivate)).
		Update("active", false)

	if result.Error != nil {
		log.Fatalf("Error deactivating event prices for event %d: %v", eventID, result.Error)
	} else {
		log.Printf("Successfully deactivated %d event prices for event %d", result.RowsAffected, eventID)
	}
}

func (g *Generator) sendActiveCoefficients(event models.Event, scoreSnapshot models.ScoreSnapshot) {
	eventID := event.ID

	var eventPrices []models.EventPrice
	err := g.db.Preload("Price").Preload("Price.Market").
		Preload("Price.Market.MarketCollection").
		Preload("Event").
		Preload("Event.Competition").
		Preload("Event.Competition.Country").
		Preload("Event.Competition.Country.Sport").
		Joins("JOIN prices ON event_prices.price_id = prices.id").
		Joins("JOIN markets ON prices.market_id = markets.id").
		Where("event_prices.event_id = ? AND event_prices.active = ? AND markets.active = ? AND prices.active = ?",
			eventID, true, true, true).
		Find(&eventPrices).Error

	if err != nil {
		return
	}

	for _, eventPrice := range eventPrices {
		newCoeff := g.calculateNewCoefficient(eventPrice, scoreSnapshot)

		eventPrice.Coefficient = newCoeff

		if err = g.client.SendToCentrifugo(eventPrice); err != nil {
			continue
		}
	}
}

func (g *Generator) calculateNewCoefficient(eventPrice models.EventPrice, score models.ScoreSnapshot) float64 {
	market := eventPrice.Price.Market
	price := eventPrice.Price

	switch market.Code {
	case "1X2":
		return markets.Calculate1x2Coefficient(price.Code, score)
	case "BTTS":
		return markets.CalculateBTTSCoefficient(price.Code, score)
	default:
		return markets.CalculateOverUnderCoefficient(price.Code, score)
	}
}
