package generator

import (
	"encoding/json"
	"fmt"
	"github.com/VaheMuradyan/Live2/db/models"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
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

	queueName := fmt.Sprintf("queue%v", scoreSnapshot.EventID)
	_, err := g.channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		panic(err)
	}

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

			body, err := json.Marshal(scoreSnapshot)
			if err != nil {
				log.Fatalf("Error marshalling simulation score %v", err)
			}

			message := amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			}

			err = g.channel.Publish("", queueName, false, false, message)
			if err != nil {
				log.Fatalf("Error publishing simulation score %v", err)
			}
		}
	}
}
