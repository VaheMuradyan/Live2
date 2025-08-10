package cache

import (
	"context"
	"encoding/json"
	"errors"
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
		return fmt.Errorf("failed to marshal event prices: %w", err)
	}

	err = r.client.Set(r.ctx, key, data, 30*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("failed to set event prices in Redis: %w", err)
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
		return nil, fmt.Errorf("failed to unmarshal event prices: %w", err)
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

func (r *RedisCache) SetScoreSnapshot(eventID uint, score models.ScoreSnapshot) error {
	key := fmt.Sprintf("score:%d", eventID)

	data, err := json.Marshal(score)
	if err != nil {
		return err
	}

	err = r.client.Set(r.ctx, key, data, 30*time.Minute).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisCache) GetScoreSnapshot(eventID uint) (models.ScoreSnapshot, error) {
	key := fmt.Sprintf("score:%d", eventID)

	data, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return models.ScoreSnapshot{}, err
	}

	var score models.ScoreSnapshot
	err = json.Unmarshal([]byte(data), &score)
	if err != nil {
		return models.ScoreSnapshot{}, err
	}

	return score, nil
}

func (r *RedisCache) SetAllScoreSnapshots(scores map[uint]models.ScoreSnapshot) error {
	pipe := r.client.Pipeline()

	for eventID, score := range scores {
		key := fmt.Sprintf("score:%d", eventID)
		data, err := json.Marshal(score)
		if err != nil {
			continue
		}
		pipe.Set(r.ctx, key, data, 30*time.Minute)
	}

	_, err := pipe.Exec(r.ctx)
	return err
}

func (r *RedisCache) GetAllScoreSnapshots(eventIDs []uint) (map[uint]models.ScoreSnapshot, error) {
	if len(eventIDs) == 0 {
		return make(map[uint]models.ScoreSnapshot), nil
	}

	pipe := r.client.Pipeline()

	cmds := make([]*redis.StringCmd, len(eventIDs))
	for i, eventID := range eventIDs {
		key := fmt.Sprintf("score:%d", eventID)
		cmds[i] = pipe.Get(r.ctx, key)
	}

	_, err := pipe.Exec(r.ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("failed to execute pipeline: %w", err)
	}

	result := make(map[uint]models.ScoreSnapshot)
	for i, cmd := range cmds {
		if errors.Is(cmd.Err(), redis.Nil) || cmd.Err() != nil {
			continue
		}

		var score models.ScoreSnapshot
		err = json.Unmarshal([]byte(cmd.Val()), &score)
		if err == nil {
			result[eventIDs[i]] = score
		}
	}

	return result, nil
}
