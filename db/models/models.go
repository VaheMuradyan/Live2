package models

import (
	"gorm.io/gorm"
)

type Sport struct {
	gorm.Model
	Name      string
	Code      string    `gorm:"unique"`
	Countries []Country `gorm:"foreignKey:SportID"`
	Teams     []Team    `gorm:"many2many:sport_teams;"`
}

type Country struct {
	gorm.Model
	Name         string
	Code         string `gorm:"unique;size:3"`
	SportID      uint
	Sport        Sport         `gorm:"foreignKey:SportID"`
	Competitions []Competition `gorm:"foreignKey:CountryID"`
}

type Competition struct {
	gorm.Model
	Name      string
	CountryID uint
	Country   Country `gorm:"foreignKey:CountryID"`
	Teams     []Team  `gorm:"many2many:competition_teams;"`
	Events    []Event `gorm:"foreignKey:CompetitionID"`
}

type Team struct {
	gorm.Model
	Name         string
	Rating       int
	CountryID    uint
	Country      Country       `gorm:"foreignKey:CountryID"`
	Competitions []Competition `gorm:"many2many:competition_teams;"`
	Events       []Event       `gorm:"many2many:event_teams;"`
	Sports       []Sport       `gorm:"many2many:sport_teams;"`
}

type MarketCollection struct {
	gorm.Model
	Name    string   `gorm:"unique"`
	Code    string   `gorm:"unique"`
	Markets []Market `gorm:"foreignKey:MarketCollectionID"`
	Events  []Event  `gorm:"many2many:event_market_collections;"`
}

type Market struct {
	gorm.Model
	Name               string `gorm:"unique"`
	Code               string `gorm:"unique"`
	MarketCollectionID uint
	MarketCollection   MarketCollection `gorm:"foreignKey:MarketCollectionID" `
	Prices             []Price          `gorm:"foreignKey:MarketID"`
	Active             bool             `gorm:"default:false"`
}

type Price struct {
	gorm.Model
	Name        string `gorm:"unique"`
	Code        string `gorm:"unique"`
	MarketID    uint
	Market      Market       `gorm:"foreignKey:MarketID"`
	EventPrices []EventPrice `gorm:"foreignKey:PriceID"`
	Active      bool         `gorm:"default:true"`
}

type EventPrice struct {
	gorm.Model
	EventID     uint    `gorm:"index"`
	Event       Event   `gorm:"foreignKey:EventID"`
	PriceID     uint    `gorm:"index"`
	Price       Price   `gorm:"foreignKey:PriceID"`
	Coefficient float64 `gorm:"type:decimal(9,4);"`
	Active      bool    `gorm:"default:true"`
}

type Event struct {
	gorm.Model
	Name              string
	CompetitionID     uint
	Competition       Competition        `gorm:"foreignKey:CompetitionID"`
	Teams             []Team             `gorm:"many2many:event_teams;"`
	MarketCollections []MarketCollection `gorm:"many2many:event_market_collections;"`
	EventPrices       []EventPrice       `gorm:"foreignKey:EventID"`
	Active            bool               `gorm:"default:false"`
	Code              string             `gorm:"unique"`
}

type RequestData struct {
	EventCodes  []string `json:"event_codes"`
	MarketCodes []string `json:"market_codes"`
}

type GetEventListResponse struct {
	EventID   uint   `json:"event_id"`
	EventName string `json:"event_name"`
	EventCode string `json:"event_code"`
}

type ScoreSnapshot struct {
	EventID    uint
	Team1Score int
	Team2Score int
	Total      int
}
