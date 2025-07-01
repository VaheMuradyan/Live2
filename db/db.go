package db

import (
	"github.com/VaheMuradyan/Live2/db/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

var db *gorm.DB

func Connect() *gorm.DB {
	var err error

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	dsn := "vahe:java@tcp(127.0.0.1:3306)/sport2?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&models.Sport{}, &models.Country{}, &models.Competition{}, &models.Team{}, &models.Event{},
		&models.MarketCollection{}, &models.Market{}, &models.Price{}, &models.Score{}, &models.Coefficient{})
	if err != nil {
		panic(err)
	}
	return db
}
