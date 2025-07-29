package centrifugoClient

import (
	"context"
	"encoding/json"
	apiproto "github.com/VaheMuradyan/Live2/centrifugo"
	"github.com/VaheMuradyan/Live2/db/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
	"log"
	"strings"
	"time"
)

const (
	apiKey = "0957bfe1-5aa9-40c0-991f-d15150f91594"
)

type CentrifugoClient struct {
	cfConn   *grpc.ClientConn
	CfClient apiproto.CentrifugoApiClient
	Ctx      context.Context
	db       *gorm.DB
}

func NewCentrifugoClient(db *gorm.DB) *CentrifugoClient {

	conn, err := grpc.NewClient("localhost:10000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Centrifugo: %v", err)
	}

	client := apiproto.NewCentrifugoApiClient(conn)

	md := metadata.New(map[string]string{
		"authorization": "apikey " + apiKey,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	server := &CentrifugoClient{
		cfConn:   conn,
		CfClient: client,
		Ctx:      ctx,
		db:       db,
	}

	return server
}

func (s *CentrifugoClient) Close() {
	s.cfConn.Close()
}

func (s *CentrifugoClient) SendToCentrifugo(eventPrice models.EventPrice) error {
	price := eventPrice.Price
	market := price.Market
	marketCollection := market.MarketCollection
	event := eventPrice.Event
	competition := event.Competition
	country := competition.Country
	sport := country.Sport

	data := map[string]interface{}{
		"sport":                  sport.Name,
		"country":                country.Name,
		"competition":            competition.Name,
		"event":                  event.Name,
		"market":                 market.Code,
		"market_collection_code": marketCollection.Code,
		"price":                  price.Name,
		"new_coefficient":        eventPrice.Coefficient,
		"old_coefficient":        float64(5),
		"timestamp":              time.Now().Format(time.RFC3339),
		"coefficient_id":         eventPrice.ID,
		"active":                 eventPrice.Active,
	}

	lower := strings.ToLower(event.Name)

	channelName := strings.ReplaceAll(lower, " ", "") + "_" + strings.ToLower(marketCollection.Code) + "_" + strings.ToLower(market.Code)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req := &apiproto.PublishRequest{
		Channel: channelName,
		Data:    jsonData,
	}

	_, err = s.CfClient.Publish(s.Ctx, req)
	if err != nil {
		return err
	}

	return err
}
