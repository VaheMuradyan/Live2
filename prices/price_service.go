package prices

import (
	"errors"
	"github.com/VaheMuradyan/Live2/db/models"
	"github.com/VaheMuradyan/Live2/generator"
)

type PriceService struct {
	repo      *PriceRepository
	generator *generator.Generator
}

func NewPriceService(repo *PriceRepository, generator *generator.Generator) *PriceService {
	return &PriceService{
		repo:      repo,
		generator: generator,
	}
}

func (s *PriceService) ActivateData(data models.RequestData) error {
	if err := s.repo.ActivateCoefficients(); err != nil {
		return errors.New("failed to activate coefficients")
	}
	if err := s.repo.ActivateMarkets(data.MarketCodes); err != nil {
		return errors.New("failed to activate markets")
	}
	if err := s.repo.ActivateEvents(data.EventCodes); err != nil {
		return errors.New("failed to activate events")
	}

	go s.generator.Start()

	return nil
}

func (s *PriceService) GetEventList() []models.GetEventListResponse {
	var res []models.GetEventListResponse

	events := s.repo.GetEventList()
	if events == nil {
		return nil
	}

	for i := 0; i < len(events); i++ {
		res = append(res, models.GetEventListResponse{
			EventID:   events[i].ID,
			EventName: events[i].Name,
			EventCode: events[i].Code,
		})
	}

	return res
}
