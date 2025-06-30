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

func (s *CentrifugoClient) SendToCentrifugo(price *models.Price) error {
	md := metadata.New(map[string]string{
		"authorization": "apikey " + apiKey,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	//sport := price.Market.MarketCollection.Event.Competition.Country.Sport.Name
	//marketCollectionCode := price.Market.MarketCollection.Code
	//data := map[string]interface{}{
	//	"sport":                  sport,
	//	"country":                price.Market.MarketCollection.Event.Competition.Country.Name,
	//	"competition":            price.Market.MarketCollection.Event.Competition.Name,
	//	"event":                  price.Market.MarketCollection.Event.Name,
	//	"market":                 price.Code,
	//	"market_type":            price.Market.Type,
	//	"market_collection_code": price.Market.MarketCollection.Code,
	//	"price":                  price.Name,
	//	"old_coefficient":        float32(price.PreviousCoefficient),
	//	"new_coefficient":        float32(price.CurrentCoefficient),
	//	"timestamp":              time.Now().Format(time.RFC3339),
	//	"change":                 float32(price.CurrentCoefficient) - float32(price.PreviousCoefficient),
	//}
	//
	//channelName := strings.ToLower(sport) + "_" + strings.ToLower(marketCollectionCode)

	data := map[string]interface{}{
		"ank": "ank",
	}

	channelName := "ban"

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
