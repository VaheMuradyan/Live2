package generator

import (
	"fmt"
	"github.com/VaheMuradyan/Live2/db/models"
	"math/rand"
	"sync"
	"time"
)

func (g *Generator) startEventsSimulation() {
	scores := g.cache.GetAllScoreSnapshotsForSimulation()

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

func (g *Generator) startEvent(scoreSnapshot models.ScoreSnapshot, stopChan <-chan bool, wg *sync.WaitGroup) {
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
				scoreSnapshot.Team1Score++
				scoreSnapshot.Total++
			case 1:
				scoreSnapshot.Team2Score++
				scoreSnapshot.Total++
			}

			g.cache.UpdateScoreSnapshot(scoreSnapshot.EventID, scoreSnapshot)
		}
	}
}
