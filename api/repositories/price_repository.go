package repositories

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
	err := p.db.Model(&models.Event{}).Where("code IN ?", events).
		Update("active", true).Error
	return err
}

func (p *PriceRepository) ActivateMarkets(markets []string) error {
	if err := p.db.Model(&models.Market{}).Where("1 = 1").Update("active", false).Error; err != nil {
		return err
	}
	err := p.db.Model(&models.Market{}).Where("code IN ?", markets).
		Update("active", true).Error
	return err
}
