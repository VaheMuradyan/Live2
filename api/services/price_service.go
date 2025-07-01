package services

import (
	"errors"
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

func (s *PriceService) ActivateData(data models.RequestData) error {
	if err := s.repo.ActivateMarkets(data.MarketCodes); err != nil {
		return errors.New("failed to activate markets")
	}
	if err := s.repo.ActivateEvents(data.EventCodes); err != nil {
		return errors.New("failed to activate events")
	}

	return nil
}
