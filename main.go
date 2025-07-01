package main

import (
	"github.com/VaheMuradyan/Live2/api/handlers"
	"github.com/VaheMuradyan/Live2/api/repositories"
	"github.com/VaheMuradyan/Live2/api/services"
	"github.com/VaheMuradyan/Live2/centrifugoClient"
	db2 "github.com/VaheMuradyan/Live2/db"
	"github.com/VaheMuradyan/Live2/generator"
	"github.com/VaheMuradyan/Live2/router"
	"github.com/gin-gonic/gin"
)

func main() {
	db := db2.Connect()

	client := centrifugoClient.NewCentrifugoClient()
	generator2 := generator.NewGenerator(client, db)

	repo := repositories.NewPriceRepository(db)
	service := services.NewPriceService(repo, generator2)
	handler := handlers.NewHandler(service)

	r := gin.Default()

	router.SetupRouter(handler, r)

	r.Run(":8080")
}
