package centrifugoClient

import (
	"context"
	"encoding/json"
	apiproto "github.com/VaheMuradyan/Live2/centrifugo"
	"github.com/VaheMuradyan/Live2/db/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log"
	"strings"
	"time"
)

const (
	apiKey = "0957bfe1-5aa9-40c0-991f-d15150f91594"
)

type CentrifugoClient struct {
	cfConn   *grpc.ClientConn
	cfClient apiproto.CentrifugoApiClient
}

func NewCentrifugoClient() *CentrifugoClient {

	conn, err := grpc.NewClient("localhost:10000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Centrifugo: %v", err)
	}

	client := apiproto.NewCentrifugoApiClient(conn)

	server := &CentrifugoClient{
		cfConn:   conn,
		cfClient: client,
	}

	return server
}

func (s *CentrifugoClient) Close() {
	s.cfConn.Close()
}

func (s *CentrifugoClient) SendToCentrifugo(coefficient *models.Coefficient) error {
	md := metadata.New(map[string]string{
		"authorization": "apikey " + apiKey,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	price := coefficient.Price
	market := price.Market
	marketCollection := market.MarketCollection
	event := coefficient.Event
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
		"new_coefficient":        float32(coefficient.Coefficient),
		"timestamp":              time.Now().Format(time.RFC3339),
	}

	channelName := strings.ToLower(sport.Name) + "_" + strings.ToLower(marketCollection.Code)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req := &apiproto.PublishRequest{
		Channel: channelName,
		Data:    jsonData,
	}

	_, err = s.cfClient.Publish(ctx, req)

	return err
}
