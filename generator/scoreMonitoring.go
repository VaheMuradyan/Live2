package generator

import (
	"github.com/VaheMuradyan/Live2/db/models"
	"github.com/VaheMuradyan/Live2/generator/markets"
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
	events := g.cache.GetActiveEvents()

	for _, event := range events {
		currentScore, ok := g.cache.GetScoreSnapshot(event.ID)
		if !ok {
			continue
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
	}
	if totalGoals >= 4 {
		priceCodesToDeactivate = append(priceCodesToDeactivate, "U35", "O35")
	}
	if totalGoals >= 3 {
		priceCodesToDeactivate = append(priceCodesToDeactivate, "U25", "O25")
	}
	if totalGoals >= 2 {
		priceCodesToDeactivate = append(priceCodesToDeactivate, "U15", "O15")
	}
	if totalGoals >= 1 {
		priceCodesToDeactivate = append(priceCodesToDeactivate, "U5", "O5")
	}

	if scoreSnapshot.Team1Score > 0 && scoreSnapshot.Team2Score > 0 {
		priceCodesToDeactivate = append(priceCodesToDeactivate, "BTTS_N", "BTTS_Y")
	}

	if len(priceCodesToDeactivate) == 0 {
		return
	}

	priceIDs := g.cache.GetPriceIDsByCodes(priceCodesToDeactivate)
	if len(priceIDs) == 0 {
		log.Printf("No price IDs found for codes: %v", priceCodesToDeactivate)
		return
	}

	err := g.cache.DeactivateEventPrices(eventID, priceIDs)
	if err != nil {
		log.Printf("Error deactivating event prices in Redis cache for event %d: %v", eventID, err)
	}
}

func (g *Generator) sendActiveCoefficients(event models.Event, scoreSnapshot models.ScoreSnapshot) {
	eventPrices := g.cache.GetEventPrices(event.ID)

	for _, eventPrice := range eventPrices {

		if !eventPrice.Active {
			continue
		}

		newCoeff := g.calculateNewCoefficient(eventPrice, scoreSnapshot)

		err := g.cache.UpdateEventPriceCoefficient(event.ID, eventPrice.PriceID, newCoeff)
		if err != nil {
			log.Printf("Error updating event price coefficient for event %d: %v", event.ID, err)
		}

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
	}
	return 0
}
