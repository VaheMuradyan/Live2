package generator

import (
	"github.com/VaheMuradyan/Live2/centrifugoClient"
	"gorm.io/gorm"
	"sync"
)

type Generator struct {
	db             *gorm.DB
	client         *centrifugoClient.CentrifugoClient
	scoreSnapshots sync.Map
	stopChan       chan bool
}

func NewGenerator(client *centrifugoClient.CentrifugoClient, db *gorm.DB) *Generator {
	return &Generator{
		db:             db,
		client:         client,
		scoreSnapshots: sync.Map{},
		stopChan:       make(chan bool),
	}
}

func (gen *Generator) Start() {
	go gen.startScoreMonitoring()
	gen.startEventsSimulation()
}
