package generator

import (
	"github.com/VaheMuradyan/Live2/centrifugoClient"
	"github.com/VaheMuradyan/Live2/db/models"
	"gorm.io/gorm"
	"sync"
)

type Generator struct {
	db             *gorm.DB
	client         *centrifugoClient.CentrifugoClient
	scoreSnapshots map[uint]models.ScoreSnapshot
	snapshotMutex  sync.RWMutex
	stopChan       chan bool
}

func NewGenerator(client *centrifugoClient.CentrifugoClient, db *gorm.DB) *Generator {
	return &Generator{
		db:             db,
		client:         client,
		scoreSnapshots: make(map[uint]models.ScoreSnapshot),
		stopChan:       make(chan bool),
	}
}

func (gen *Generator) Start() {
	go gen.startScoreMonitoring()
	gen.startEventsSimulation()
}
