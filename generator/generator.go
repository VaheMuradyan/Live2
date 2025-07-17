package generator

import (
	"github.com/VaheMuradyan/Live2/centrifugoClient"
	"gorm.io/gorm"
	"sync"
)

type ScoreSnapshot struct {
	EventID    uint
	Team1Score int
	Team2Score int
	Total      int
}

type Generator struct {
	db             *gorm.DB
	client         *centrifugoClient.CentrifugoClient
	scoreSnapshots map[uint]ScoreSnapshot
	snapshotMutex  sync.RWMutex
	stopChan       chan bool
	wg             sync.WaitGroup
}

func NewGenerator(client *centrifugoClient.CentrifugoClient, db *gorm.DB) *Generator {
	return &Generator{
		db:             db,
		client:         client,
		scoreSnapshots: make(map[uint]ScoreSnapshot),
		stopChan:       make(chan bool),
	}
}

func (gen *Generator) Start() {
	go gen.startScoreMonitoring()
	gen.startEventsSimulation()
}
