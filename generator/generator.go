package generator

import (
	"github.com/VaheMuradyan/Live2/centrifugoClient"
	"gorm.io/gorm"
)

type Generator struct {
	db     *gorm.DB
	client *centrifugoClient.CentrifugoClient
}

func NewGenerator(client *centrifugoClient.CentrifugoClient, db *gorm.DB) *Generator {
	return &Generator{
		db:     db,
		client: client,
	}
}

func (gen *Generator) Start() {

}
