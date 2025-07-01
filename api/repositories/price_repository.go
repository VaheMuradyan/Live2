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
	p.db.Model(&models.Event{}).Update("active", false)
	err := p.db.Model(&models.Event{}).Where("code IN ?", events).
		Update("active", true).Error
	return err
}

func (p *PriceRepository) ActivateMarkets(markets []string) error {
	p.db.Model(&models.Market{}).Update("active", false)
	err := p.db.Model(&models.Market{}).Where("code IN ?", markets).
		Update("active", true).Error
	return err
}
