package generator

import (
	"fmt"
	"github.com/VaheMuradyan/Live2/db/models"
	"log"
	"math/rand"
	"sync"
	"time"
)

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
			fmt.Println("Stopping event simulation")
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
