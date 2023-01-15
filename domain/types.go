package domain

type EloPlayerRating struct {
	Name        string
	Rating      float64
	GamesPlayed int
}

type GlickoPlayerRating struct {
	Name            string
	Rating          float64
	RatingDeviation float64
	Volatility      float64
	GamesPlayed     int
}

type MatchResult struct {
	Date       string
	WinnerTeam []string
	LoserTeam  []string
	Draw       bool
}

type MatchSheetStructure struct {
	Date     string
	RowBegin int
	RowEnd   int
}
