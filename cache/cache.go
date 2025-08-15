package cache

import (
	"github.com/VaheMuradyan/Live2/db/models"
	"gorm.io/gorm"
	"log"
	"sync"
)

type Cache struct {
	db           *gorm.DB
	redis        *RedisCache
	mu           sync.RWMutex
	eventsMap    map[uint]models.Event
	staticLookup map[uint]StaticEventData
}

type StaticEventData struct {
	EventID         uint
	EventName       string
	EventCode       string
	CompetitionName string
	CountryName     string
	SportName       string
	PriceRelations  map[uint]PriceRelation
}

type PriceRelation struct {
	PriceID              uint
	PriceName            string
	PriceCode            string
	MarketCode           string
	MarketName           string
	MarketCollectionCode string
	MarketCollectionName string
}

func NewCache(db *gorm.DB) *Cache {
	cache := &Cache{
		db:           db,
		redis:        NewRedisCache(),
		eventsMap:    make(map[uint]models.Event),
		staticLookup: make(map[uint]StaticEventData),
	}
	return cache
}

func (c *Cache) LoadStaticEventData() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var events []models.Event
	err := c.db.Where("events.active = ?", true).
		Preload("Competition").
		Preload("Competition.Country").
		Preload("Competition.Country.Sport").
		Preload("Teams").
		Find(&events).Error

	if err != nil {
		log.Printf("Error loading events: %v", err)
		return
	}

	c.eventsMap = make(map[uint]models.Event)
	c.staticLookup = make(map[uint]StaticEventData)

	for _, event := range events {
		c.eventsMap[event.ID] = event

		staticData := StaticEventData{
			EventID:         event.ID,
			EventName:       event.Name,
			EventCode:       event.Code,
			CompetitionName: event.Competition.Name,
			CountryName:     event.Competition.Country.Name,
			SportName:       event.Competition.Country.Sport.Name,
			PriceRelations:  make(map[uint]PriceRelation),
		}

		c.staticLookup[event.ID] = staticData
	}

	c.loadPriceRelations()
}

func (c *Cache) loadPriceRelations() {
	var eventPrices []models.EventPrice
	err := c.db.Preload("Price").
		Preload("Price.Market").
		Preload("Price.Market.MarketCollection").
		Joins("JOIN events ON event_prices.event_id = events.id").
		Joins("JOIN prices ON event_prices.price_id = prices.id").
		Joins("JOIN markets ON prices.market_id = markets.id").
		Where("events.active = ? AND markets.active = ?",
			true, true).
		Find(&eventPrices).Error

	if err != nil {
		log.Printf("Error loading price relations: %v", err)
		return
	}

	eventPricesMap := make(map[uint][]models.EventPrice)

	for _, ep := range eventPrices {
		if staticData, exists := c.staticLookup[ep.EventID]; exists {
			staticData.PriceRelations[ep.PriceID] = PriceRelation{
				PriceID:              ep.PriceID,
				PriceName:            ep.Price.Name,
				PriceCode:            ep.Price.Code,
				MarketCode:           ep.Price.Market.Code,
				MarketName:           ep.Price.Market.Name,
				MarketCollectionCode: ep.Price.Market.MarketCollection.Code,
				MarketCollectionName: ep.Price.Market.MarketCollection.Name,
			}
			c.staticLookup[ep.EventID] = staticData
		}
		eventPricesMap[ep.EventID] = append(eventPricesMap[ep.EventID], ep)
	}

	for eventID, prices := range eventPricesMap {
		err = c.redis.SetEventPrices(eventID, prices)
		if err != nil {
			log.Printf("Error setting event prices in Redis for event %d: %v", eventID, err)
		}
	}
}

func (c *Cache) GetActiveEvents() []models.Event {
	c.mu.RLock()
	defer c.mu.RUnlock()

	events := make([]models.Event, 0, len(c.eventsMap))
	for _, event := range c.eventsMap {
		events = append(events, event)
	}
	return events
}

func (c *Cache) GetEventPrices(eventID uint, forCentrifugo bool) []models.EventPrice {
	eventPrices, err := c.redis.GetEventPrices(eventID)
	if err != nil {
		return []models.EventPrice{}
	}
	if forCentrifugo {
		c.enrichEventPricesForCentrifugo(eventPrices, eventID)
	}

	return eventPrices
}

func (c *Cache) enrichEventPricesForCentrifugo(eventPrices []models.EventPrice, eventID uint) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	staticData, exists := c.staticLookup[eventID]
	if !exists {
		return
	}

	event, eventExists := c.eventsMap[eventID]
	if !eventExists {
		return
	}

	for i := range eventPrices {
		eventPrices[i].Event = event

		if priceRel, exists2 := staticData.PriceRelations[eventPrices[i].PriceID]; exists2 {
			eventPrices[i].Price = models.Price{
				Model: gorm.Model{ID: priceRel.PriceID},
				Name:  priceRel.PriceName,
				Code:  priceRel.PriceCode,
				Market: models.Market{
					Name: priceRel.MarketName,
					Code: priceRel.MarketCode,
					MarketCollection: models.MarketCollection{
						Name: priceRel.MarketCollectionName,
						Code: priceRel.MarketCollectionCode,
					},
				},
			}
		}
	}
}

func (c *Cache) UpdateEventPriceCoefficient(eventID, priceID uint, newCoefficient float64) error {
	eventPrices, err := c.redis.GetEventPrices(eventID)
	if err != nil {
		return err
	}

	for i := range eventPrices {
		if eventPrices[i].PriceID == priceID {
			eventPrices[i].Coefficient = newCoefficient
			break
		}
	}

	return c.redis.SetEventPrices(eventID, eventPrices)
}

func (c *Cache) DeactivateEventPrices(eventID uint, priceIDs []uint) error {
	eventPrices, err := c.redis.GetEventPrices(eventID)
	if err != nil {
		return err
	}

	priceIDSet := make(map[uint]bool)
	for _, id := range priceIDs {
		priceIDSet[id] = true
	}

	for i := range eventPrices {
		if priceIDSet[eventPrices[i].PriceID] {
			eventPrices[i].Active = false
		}
	}

	return c.redis.SetEventPrices(eventID, eventPrices)
}

func (c *Cache) GetPriceIDsByCodes(codes []string) []uint {
	c.mu.RLock()
	defer c.mu.RUnlock()

	codeSet := make(map[string]bool)
	for _, code := range codes {
		codeSet[code] = true
	}

	var priceIDs []uint
	priceIDSet := make(map[uint]bool)

	for _, staticData := range c.staticLookup {
		for priceID, priceRel := range staticData.PriceRelations {
			if codeSet[priceRel.PriceCode] && !priceIDSet[priceID] {
				priceIDs = append(priceIDs, priceID)
				priceIDSet[priceID] = true
			}
		}
	}

	return priceIDs
}

func (c *Cache) GetAllScoreSnapshotsForSimulation() []models.ScoreSnapshot {
	c.mu.RLock()

	scores := make([]models.ScoreSnapshot, 0, len(c.eventsMap))
	for id := range c.eventsMap {
		scores = append(scores, models.ScoreSnapshot{
			EventID:    id,
			Team1Score: 0,
			Team2Score: 0,
			Total:      0,
		})
	}
	c.mu.RUnlock()

	return scores
}

func (c *Cache) SaveData() {
	events := c.GetActiveEvents()

	for _, event := range events {
		evPrices := c.GetEventPrices(event.ID, false)
		c.db.Save(evPrices)
	}
}
