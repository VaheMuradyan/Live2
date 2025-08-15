package generator

import (
	"encoding/json"
	"fmt"
	"github.com/VaheMuradyan/Live2/db/models"
	"github.com/VaheMuradyan/Live2/generator/markets"
	"log"
)

func (g *Generator) startScoreMonitoring() {
	events := g.cache.GetActiveEvents()

	for _, event := range events {
		queueName := fmt.Sprintf("queue%v", event.ID)

		_, err := g.channel.QueueDeclare(queueName, true, false, false, false, nil)
		if err != nil {
			log.Fatalf("Failed to declare queue %s: %v", queueName, err)
		}

		go g.consumeQueue(queueName)
	}

	<-g.stopChan
	log.Println("Stopping score monitoring...")
	g.cache.SaveData()
}

func (g *Generator) consumeQueue(queueName string) {
	messages, err := g.channel.Consume(queueName, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to consume from %s: %v", queueName, err)
	}

	for msg := range messages {
		var score models.ScoreSnapshot
		if err := json.Unmarshal(msg.Body, &score); err != nil {
			log.Printf("Invalid message in %s: %v", queueName, err)
			continue
		}
		g.handleScoreChange(score.EventID, score)
	}
}

func (g *Generator) handleScoreChange(eventID uint, currentScore models.ScoreSnapshot) {
	g.checkAndStopMarkets(eventID, currentScore)
	g.sendActiveCoefficients(eventID, currentScore)
}

func (g *Generator) checkAndStopMarkets(eventID uint, scoreSnapshot models.ScoreSnapshot) {
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

func (g *Generator) sendActiveCoefficients(eventID uint, scoreSnapshot models.ScoreSnapshot) {
	eventPrices := g.cache.GetEventPrices(eventID, true)

	for _, eventPrice := range eventPrices {

		if !eventPrice.Active {
			continue
		}

		newCoeff := g.calculateNewCoefficient(eventPrice, scoreSnapshot)

		err := g.cache.UpdateEventPriceCoefficient(eventID, eventPrice.PriceID, newCoeff)
		if err != nil {
			log.Printf("Error updating event price coefficient for event %d: %v", eventID, err)
		}

		eventPrice.Coefficient = newCoeff

		if err = g.client.SendToCentrifugo(eventPrice); err != nil {
			log.Fatalf("Chexav centriguon")
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
