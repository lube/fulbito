package rating

import (
	"fulbito/domain"
	"math"
	"sort"
	"strconv"
)

const ()

func getPlayer(player string, players []domain.EloPlayerRating) (bool, domain.EloPlayerRating) {
	for _, p := range players {
		if p.Name == player {
			return true, p
		}
	}

	return false, domain.EloPlayerRating{}
}

func updatePlayer(player domain.EloPlayerRating, players []domain.EloPlayerRating) []domain.EloPlayerRating {
	for i, p := range players {
		if p.Name == player.Name {
			players[i] = player
			return players
		}
	}

	return players
}

func getPlayerG(player string, players []domain.GlickoPlayerRating) (bool, domain.GlickoPlayerRating) {
	for _, p := range players {
		if p.Name == player {
			return true, p
		}
	}

	return false, domain.GlickoPlayerRating{}
}

func updatePlayerG(player domain.GlickoPlayerRating, players []domain.GlickoPlayerRating) []domain.GlickoPlayerRating {
	for i, p := range players {
		if p.Name == player.Name {
			players[i] = player
			return players
		}
	}

	return players
}

func ProcessAllMatchResultsAndGetEloRating(matchResults []domain.MatchResult) []domain.EloPlayerRating {
	players := []domain.EloPlayerRating{}

	for _, match := range matchResults {
		players = processMatchResultElo(players, match)
	}

	// sort by rating
	for i := 0; i < len(players); i++ {
		for j := i + 1; j < len(players); j++ {
			if players[i].Rating < players[j].Rating {
				players[i], players[j] = players[j], players[i]
			}
		}
	}

	return players
}

func ProcessAllMatchResultsAndGetGlickoRating(matchResults []domain.MatchResult) []domain.GlickoPlayerRating {
	players := []domain.GlickoPlayerRating{}

	for _, match := range matchResults {
		players = processMatchResultGlicko(players, match)
	}

	// sort by rating
	for i := 0; i < len(players); i++ {
		for j := i + 1; j < len(players); j++ {
			if players[i].Rating < players[j].Rating {
				players[i], players[j] = players[j], players[i]
			}
		}
	}

	return players
}

// elo
func processMatchResultElo(playerRatings []domain.EloPlayerRating, result domain.MatchResult) []domain.EloPlayerRating {

	var (
		k             = 32.0
		initialRating = 1000.0
	)

	// calculate the expected score for each team
	team1ExpectedScore := 0.0
	team2ExpectedScore := 0.0
	for _, player1 := range result.WinnerTeam {
		for _, player2 := range result.LoserTeam {
			ok, p1 := getPlayer(player1, playerRatings)
			if !ok {
				p1 = domain.EloPlayerRating{Name: player1, Rating: initialRating}
				playerRatings = append(playerRatings, p1)
			}

			ok, p2 := getPlayer(player2, playerRatings)
			if !ok {
				p2 = domain.EloPlayerRating{Name: player2, Rating: initialRating}
				playerRatings = append(playerRatings, p2)
			}

			team1ExpectedScore += 1 / (1 + math.Pow(10, (p2.Rating-p1.Rating)/400))
			team2ExpectedScore += 1 / (1 + math.Pow(10, (p1.Rating-p2.Rating)/400))
		}
	}
	if result.Draw {
		team1ExpectedScore = 0.5
		team2ExpectedScore = 0.5
	}

	//update the rating for each player
	for i := range result.WinnerTeam {
		_, p := getPlayer(result.WinnerTeam[i], playerRatings)
		p.Rating = p.Rating + (k * (1 - team1ExpectedScore/float64(len(result.WinnerTeam))))
		p.GamesPlayed += 1
		updatePlayer(p, playerRatings)
	}
	for i := range result.LoserTeam {
		_, p := getPlayer(result.LoserTeam[i], playerRatings)
		p.Rating = p.Rating + (k * (0 - team2ExpectedScore/float64(len(result.LoserTeam))))
		p.GamesPlayed += 1
		updatePlayer(p, playerRatings)
	}

	return playerRatings
}

// glicko
func processMatchResultGlicko(playerRatings []domain.GlickoPlayerRating, result domain.MatchResult) []domain.GlickoPlayerRating {
	// Define the system constant
	var (
		initialRanking           float64 = 1500
		initialRankingDeviation  float64 = 700
		initialRankingVolatility float64 = 0.12
		tau                      float64 = 0.3
		k                                = 24
	)
	// calculate the expected score for each team
	team1ExpectedScore := 0.0
	team2ExpectedScore := 0.0

	for _, player1 := range result.WinnerTeam {
		for _, player2 := range result.LoserTeam {
			ok, p1 := getPlayerG(player1, playerRatings)
			if !ok {
				p1 = domain.GlickoPlayerRating{
					Name:            player1,
					Rating:          initialRanking,
					RatingDeviation: initialRankingDeviation,
					Volatility:      initialRankingVolatility,
				}
				playerRatings = append(playerRatings, p1)
			}

			ok, p2 := getPlayerG(player2, playerRatings)
			if !ok {
				p2 = domain.GlickoPlayerRating{
					Name:            player2,
					Rating:          initialRanking,
					RatingDeviation: initialRankingDeviation,
					Volatility:      initialRankingVolatility,
				}
				playerRatings = append(playerRatings, p2)
			}

			// Calculate the expected score
			team1ExpectedScore += expectedScore(p1.Rating, p1.RatingDeviation, p2.Rating, p2.RatingDeviation)
			team2ExpectedScore += 1 - expectedScore(p1.Rating, p1.RatingDeviation, p2.Rating, p2.RatingDeviation)
		}
	}

	// update the rating for each player
	for i := range result.WinnerTeam {
		_, p := getPlayerG(result.WinnerTeam[i], playerRatings)

		if result.Draw {
			p = processGlickoUpdate(p, 0.5, team1ExpectedScore/float64(len(result.WinnerTeam)), tau, k)
		} else {
			p = processGlickoUpdate(p, 1, team1ExpectedScore/float64(len(result.WinnerTeam)), tau, k)
		}
		p.GamesPlayed += 1
		updatePlayerG(p, playerRatings)
	}
	for i := range result.LoserTeam {
		_, p := getPlayerG(result.LoserTeam[i], playerRatings)

		if result.Draw {
			p = processGlickoUpdate(p, 0.5, team2ExpectedScore/float64(len(result.LoserTeam)), tau, k)
		} else {
			p = processGlickoUpdate(p, 0, team2ExpectedScore/float64(len(result.LoserTeam)), tau, k)
		}
		p.GamesPlayed += 1
		updatePlayerG(p, playerRatings)
	}

	return playerRatings
}

func GenerateTeamsGlicko(playerList []string, playerRatings []domain.GlickoPlayerRating) ([]string, []string) {
	players := []domain.GlickoPlayerRating{}
	for _, playerName := range playerList {
		ok, gp := getPlayerG(playerName, playerRatings)
		if !ok {
			players = append(players, domain.GlickoPlayerRating{
				Name:            playerName,
				Rating:          1500,
				RatingDeviation: 0,
				Volatility:      0,
			})
		} else {
			players = append(players, gp)
		}
	}

	sort.Slice(players, func(i, j int) bool {
		return players[i].Rating > players[j].Rating
	})

	team1 := []string{}
	team2 := []string{}
	for i := range players {
		if i%2 == 0 {
			team1 = append(team1, players[i].Name+" "+strconv.FormatFloat(players[i].Rating, 'f', 5, 64))
		} else {
			team2 = append(team2, players[i].Name+" "+strconv.FormatFloat(players[i].Rating, 'f', 5, 64))
		}
	}

	return team1, team2
}

func GenerateTeamsElo(playerList []string, playerRatings []domain.EloPlayerRating) ([]string, []string) {
	players := []domain.EloPlayerRating{}
	for _, playerName := range playerList {
		ok, gp := getPlayer(playerName, playerRatings)
		if !ok {
			players = append(players, domain.EloPlayerRating{
				Name:   playerName,
				Rating: 1500,
			})
		} else {
			players = append(players, gp)
		}
	}

	sort.Slice(players, func(i, j int) bool {
		return players[i].Rating > players[j].Rating
	})

	team1 := []string{}
	team2 := []string{}
	for i := range players {
		if i%2 == 0 {
			team1 = append(team1, players[i].Name+" "+strconv.FormatFloat(players[i].Rating, 'f', 2, 64))
		} else {
			team2 = append(team2, players[i].Name+" "+strconv.FormatFloat(players[i].Rating, 'f', 2, 64))
		}
	}

	return team1, team2
}

func processGlickoUpdate(p domain.GlickoPlayerRating, result float64, expected float64, tau float64, k int) domain.GlickoPlayerRating {
	var (
		delta  float64
		newVol float64
	)
	delta = float64(k) * (result - expected)
	newVol = 1 / math.Sqrt((1/math.Pow(p.Volatility, 2))+(1/math.Pow(p.RatingDeviation, 2)))
	p.Rating = p.Rating + newVol*g(p.RatingDeviation)*delta
	p.RatingDeviation = math.Sqrt(1 / (1/math.Pow(p.RatingDeviation, 2) + 1/math.Pow(newVol, 2)))
	p.Volatility = math.Sqrt((1/math.Pow(p.RatingDeviation, 2) + 1/math.Pow(newVol, 2)) * math.Pow(tau, 2))

	return p
}

func expectedScore(r1, _, r2, rd2 float64) float64 {
	return 1 / (1 + math.Exp(-g(rd2)*(r1-r2)/400))
}

func g(rd float64) float64 {
	return 1 / math.Sqrt(1+(3*math.Pow(rd, 2))/math.Pow(math.Pi, 2))
}
