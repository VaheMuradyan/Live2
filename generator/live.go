package generator

import (
	"github.com/VaheMuradyan/Live2/db/models"
	"gorm.io/gorm"
	"log"
	"math/rand"
	"sync"
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

	g.snapshotMutex.Lock()
	defer g.snapshotMutex.Unlock()

	for _, event := range events {
		if event.Score == nil {
			continue
		}

		currentScore := ScoreSnapshot{
			EventID:    event.ID,
			Team1Score: event.Score.Team1Score,
			Team2Score: event.Score.Team2Score,
			Total:      event.Score.Total,
		}

		if previousScore, exists := g.scoreSnapshots[event.ID]; !exists || g.scoreHasChanged(previousScore, currentScore) {
			g.handleScoreChange(event, currentScore)
			g.scoreSnapshots[event.ID] = currentScore
		}
	}
}

func (g *Generator) scoreHasChanged(previous, current ScoreSnapshot) bool {
	return previous.Team1Score != current.Team1Score ||
		previous.Team2Score != current.Team2Score ||
		previous.Total != current.Total
}

func (g *Generator) handleScoreChange(event models.Event, currentScore ScoreSnapshot) {
	g.checkAndStopMarkets(event, currentScore)
	g.sendActiveCoefficients(event, currentScore)
}

func (g *Generator) checkAndStopMarkets(event models.Event, scoreSnapshot ScoreSnapshot) {
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
		log.Printf("Error deactivating coefficients for event %d: %v", eventID, result.Error)
	} else {
		log.Printf("Successfully deactivated %d coefficients for event %d", result.RowsAffected, eventID)
	}
}

func (g *Generator) sendActiveCoefficients(event models.Event, scoreSnapshot ScoreSnapshot) {
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

		if newCoeff > 0 {
			coefficient.Coefficient = newCoeff

			if err = g.db.Save(&coefficient).Error; err != nil {
				continue
			}

			if err = g.client.SendToCentrifugo(coefficient); err != nil {
				continue
			}

		}
	}
}

func (g *Generator) startEventsSimulation() {
	var scores []models.Score
	if err := g.db.Joins("Event").Where("Event.active = ?", true).Find(&scores).Error; err != nil {
		log.Fatal(err)
	}

	num := len(scores)
	stopChan := make(chan bool, num)

	var wg sync.WaitGroup
	wg.Add(num)

	for _, score := range scores {
		go g.startEvent(score, stopChan, &wg)
	}

	time.Sleep(55 * time.Second)
	for i := 0; i < num; i++ {
		stopChan <- true
	}
	g.stopChan <- true

	wg.Wait()
}

func (g *Generator) startEvent(score models.Score, stopChan <-chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			rand.Seed(time.Now().UnixNano())
			x := rand.Intn(2)

			switch x {
			case 0:
				score.Team1Score++
				score.Total++
			case 1:
				score.Team2Score++
				score.Total++
			}

			err := g.db.Save(&score).Error
			if err != nil {
				return
			}

		}
	}
}

func (g *Generator) calculateNewCoefficient(coefficient models.Coefficient, score ScoreSnapshot) float64 {
	market := coefficient.Price.Market
	price := coefficient.Price

	switch market.Code {
	case "1X2":
		return g.calculate1x2Coefficient(price.Code, score)
	case "OU5":
		return g.calculateOverUnderCoefficient(price.Code, score)
	case "OU15":
		return g.calculateOverUnderCoefficient(price.Code, score)
	case "OU25":
		return g.calculateOverUnderCoefficient(price.Code, score)
	case "OU35":
		return g.calculateOverUnderCoefficient(price.Code, score)
	case "OU45":
		return g.calculateOverUnderCoefficient(price.Code, score)
	case "BTTS":
		return g.calculateBTTSCoefficient(price.Code, score)
	}
	return 0
}
