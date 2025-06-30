package services

import (
	"github.com/VaheMuradyan/Live2/api/repositories"
	"github.com/VaheMuradyan/Live2/centrifugoClient"
	"github.com/VaheMuradyan/Live2/db/models"
)

type PriceService struct {
	repo   *repositories.PriceRepository
	client *centrifugoClient.CentrifugoClient
}

func NewPriceService(repo *repositories.PriceRepository, client *centrifugoClient.CentrifugoClient) *PriceService {
	return &PriceService{
		repo:   repo,
		client: client,
	}
}

func (s *PriceService) InchvorBan() error {
	err := s.client.SendToCentrifugo(&models.Price{})
	return err
}
