package generator

import (
	"github.com/VaheMuradyan/Live2/cache"
	"github.com/VaheMuradyan/Live2/centrifugoClient"
	"gorm.io/gorm"
	"sync"
)

type Generator struct {
	db             *gorm.DB
	client         *centrifugoClient.CentrifugoClient
	scoreSnapshots sync.Map
	cache          *cache.Cache
	stopChan       chan bool
}

func NewGenerator(client *centrifugoClient.CentrifugoClient, db *gorm.DB) *Generator {
	return &Generator{
		db:             db,
		client:         client,
		scoreSnapshots: sync.Map{},
		cache:          cache.NewCache(db),
		stopChan:       make(chan bool),
	}
}

func (gen *Generator) Start() {
	gen.cache.LoadStaticEventData()
	go gen.startScoreMonitoring()
	gen.startEventsSimulation()
}
