package generator

import "math/rand"

func (gen *Generator) calculateBTTSCoefficient(priceCode string, score ScoreSnapshot) float64 {
	bothTeamsScored := score.Team1Score > 0 && score.Team2Score > 0

	switch priceCode {
	case "BTTS_Y":
		if bothTeamsScored {
			return 1.2 + float64(rand.Intn(30))/100
		}
		return 1.8 + float64(rand.Intn(60))/100
	case "BTTS_N":
		if bothTeamsScored {
			return 4.0 + float64(rand.Intn(200))/100
		}
		return 1.5 + float64(rand.Intn(40))/100
	}
	return 1.9
}
