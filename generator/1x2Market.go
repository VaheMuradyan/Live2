package generator

import (
	"math/rand"
)

func (g *Generator) calculate1x2Coefficient(priceCode string, score ScoreSnapshot) float64 {
	scoreDiff := score.Team1Score - score.Team2Score

	switch priceCode {
	case "1":
		if scoreDiff > 0 {
			return 1.1 + float64(rand.Intn(50))/100
		} else if scoreDiff < 0 {
			return 3.0 + float64(rand.Intn(200))/100
		}
		return 2.0 + float64(rand.Intn(100))/100
	case "X":
		if scoreDiff == 0 {
			return 2.5 + float64(rand.Intn(100))/100
		}
		return 3.0 + float64(rand.Intn(150))/100
	case "2":
		if scoreDiff < 0 {
			return 1.1 + float64(rand.Intn(50))/100
		} else if scoreDiff > 0 {
			return 3.0 + float64(rand.Intn(200))/100
		}
		return 2.0 + float64(rand.Intn(100))/100
	}
	return 2.0
}
