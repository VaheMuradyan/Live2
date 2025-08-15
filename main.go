package main

import (
	"github.com/VaheMuradyan/Live2/centrifugoClient"
	db2 "github.com/VaheMuradyan/Live2/db"
	"github.com/VaheMuradyan/Live2/generator"
	"github.com/VaheMuradyan/Live2/prices"
	"github.com/VaheMuradyan/Live2/router"
	"github.com/gin-gonic/gin"
	"os"
)

func main() {
	db := db2.Connect()

	client := centrifugoClient.NewCentrifugoClient(db)
	generator2 := generator.NewGenerator(client, db)

	repo := prices.NewPriceRepository(db)
	service := prices.NewPriceService(repo, generator2)
	handler := prices.NewHandler(service)

	r := gin.Default()

	router.SetupRouter(r, handler)

	defer client.Close()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
