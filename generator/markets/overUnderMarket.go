package markets

import (
	"github.com/VaheMuradyan/Live2/db/models"
	"math/rand"
)

func CalculateOverUnderCoefficient(priceCode string, score models.ScoreSnapshot) float64 {
	currentTotal := float64(score.Total)

	switch priceCode {
	case "O5":
		if currentTotal > 0.5 {
			return 1.1 + float64(rand.Intn(30))/100
		}
		return 1.8 + float64(rand.Intn(60))/100
	case "U5":
		if currentTotal > 0.5 {
			return 2.5 + float64(rand.Intn(150))/100
		}
		return 1.6 + float64(rand.Intn(40))/100
	case "O15":
		if currentTotal > 1.5 {
			return 1.1 + float64(rand.Intn(30))/100
		}
		return 1.8 + float64(rand.Intn(60))/100
	case "U15":
		if currentTotal > 1.5 {
			return 2.5 + float64(rand.Intn(150))/100
		}
		return 1.6 + float64(rand.Intn(40))/100
	case "O25":
		if currentTotal > 2.5 {
			return 1.1 + float64(rand.Intn(30))/100
		}
		return 1.8 + float64(rand.Intn(60))/100
	case "U25":
		if currentTotal > 2.5 {
			return 2.5 + float64(rand.Intn(150))/100
		}
		return 1.6 + float64(rand.Intn(40))/100
	case "O35":
		if currentTotal > 3.5 {
			return 1.1 + float64(rand.Intn(30))/100
		}
		return 1.8 + float64(rand.Intn(60))/100
	case "U35":
		if currentTotal > 3.5 {
			return 2.5 + float64(rand.Intn(150))/100
		}
		return 1.6 + float64(rand.Intn(40))/100
	case "O45":
		if currentTotal > 4.5 {
			return 1.1 + float64(rand.Intn(30))/100
		}
		return 1.8 + float64(rand.Intn(60))/100
	case "U45":
		if currentTotal > 4.5 {
			return 2.5 + float64(rand.Intn(150))/100
		}
		return 1.6 + float64(rand.Intn(40))/100
	}
	return 1.9
}
