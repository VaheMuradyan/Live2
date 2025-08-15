package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/VaheMuradyan/Live2/db/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"time"
)

type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisCache() *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	return &RedisCache{
		client: rdb,
		ctx:    ctx,
	}
}

type EventPriceRedis struct {
	ID          uint    `json:"id"`
	EventID     uint    `json:"event_id"`
	PriceID     uint    `json:"price_id"`
	Coefficient float64 `json:"coefficient"`
	Active      bool    `json:"active"`
}

func (r *RedisCache) SetEventPrices(eventID uint, eventPrices []models.EventPrice) error {
	key := fmt.Sprintf("event_prices:%d", eventID)

	simplifiedPrices := make([]EventPriceRedis, len(eventPrices))
	for i, ep := range eventPrices {
		simplifiedPrices[i] = EventPriceRedis{
			ID:          ep.ID,
			EventID:     ep.EventID,
			PriceID:     ep.PriceID,
			Coefficient: ep.Coefficient,
			Active:      ep.Active,
		}
	}

	data, err := json.Marshal(simplifiedPrices)
	if err != nil {
		return err
	}

	err = r.client.Set(r.ctx, key, data, 5*time.Minute).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisCache) GetEventPrices(eventID uint) ([]models.EventPrice, error) {
	key := fmt.Sprintf("event_prices:%d", eventID)
	data, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var simplifiedPrices []EventPriceRedis
	err = json.Unmarshal([]byte(data), &simplifiedPrices)
	if err != nil {
		return nil, err
	}

	eventPrices := make([]models.EventPrice, len(simplifiedPrices))
	for i, sp := range simplifiedPrices {
		eventPrices[i] = models.EventPrice{
			Model:       gorm.Model{ID: sp.ID},
			EventID:     sp.EventID,
			PriceID:     sp.PriceID,
			Coefficient: sp.Coefficient,
			Active:      sp.Active,
		}
	}

	return eventPrices, nil
}
