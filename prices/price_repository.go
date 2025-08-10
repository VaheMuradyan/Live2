package prices

import (
	"github.com/VaheMuradyan/Live2/db/models"
	"gorm.io/gorm"
)

type PriceRepository struct {
	db *gorm.DB
}

func NewPriceRepository(db *gorm.DB) *PriceRepository {
	return &PriceRepository{db: db}
}

func (p *PriceRepository) ActivateEvents(events []string) error {
	if err := p.db.Model(&models.Event{}).Where("1 = 1").Update("active", false).Error; err != nil {
		return err
	}

	if err := p.db.Model(&models.Event{}).Where("code IN ?", events).
		Update("active", true).Error; err != nil {
		return err
	}

	//todo hanel subquery-n
	if err := p.db.Model(&models.Score{}).
		Where("event_id IN (?)",
			p.db.Model(&models.Event{}).Select("id").Where("code IN ? AND active = ?", events, true)).
		Updates(map[string]interface{}{
			"team1_score": 0,
			"team2_score": 0,
			"total":       0,
		}).Error; err != nil {
		return err
	}

	return nil
}

func (p *PriceRepository) ActivateMarkets(markets []string) error {
	if err := p.db.Model(&models.Market{}).Where("1 = 1").Update("active", false).Error; err != nil {
		return err
	}
	err := p.db.Model(&models.Market{}).Where("code IN ?", markets).
		Update("active", true).Error
	return err
}

func (p *PriceRepository) GetEventList() []models.Event {
	var events []models.Event
	err := p.db.Model(&models.Event{}).Where("active = ?", true).Find(&events).Error
	if err != nil {
		return nil
	}
	return events
}
