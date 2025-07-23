package main

import (
	"github.com/VaheMuradyan/Live2/centrifugoClient"
	db2 "github.com/VaheMuradyan/Live2/db"
	"github.com/VaheMuradyan/Live2/generator"
	"github.com/VaheMuradyan/Live2/prices"
	"github.com/VaheMuradyan/Live2/router"
	"github.com/gin-gonic/gin"
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

	r.Run(":8080")
}
