package generator

import (
	"github.com/VaheMuradyan/Live2/cache"
	"github.com/VaheMuradyan/Live2/centrifugoClient"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

type Generator struct {
	db       *gorm.DB
	client   *centrifugoClient.CentrifugoClient
	cache    *cache.Cache
	channel  *amqp.Channel
	conn     *amqp.Connection
	stopChan chan bool
}

func NewGenerator(client *centrifugoClient.CentrifugoClient, db *gorm.DB) *Generator {
	conn, err := amqp.Dial("amqp://localhost:5672")
	if err != nil {
		panic(err)
	}

	channel, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	return &Generator{
		db:       db,
		client:   client,
		cache:    cache.NewCache(db),
		channel:  channel,
		conn:     conn,
		stopChan: make(chan bool),
	}
}

func (gen *Generator) Start() {
	gen.cache.LoadStaticEventData()
	go gen.startScoreMonitoring()
	gen.startEventsSimulation()
}
