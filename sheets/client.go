package sheets

import (
	"context"
	"fulbito/domain"
	"log"

	"google.golang.org/api/option"
	sheets "google.golang.org/api/sheets/v4"
)

func RetrieveGoogleSheetsResults() []domain.MatchResult {
	// create a new context
	ctx := context.Background()

	client, err := sheets.NewService(ctx, option.WithCredentialsFile("./fulbitodelosviernes.json"))
	if err != nil {
		log.Fatalf("Unable to create sheets client: %v", err)
	}
	// define the spreadsheet ID
	spreadsheetID := "1ZXpQcICXxrGOUpC7rQ7bZb4Owf6XKLxu8q7OhMTJOGA"

	// define the range of cells to read
	rangeStr := "partidos!A2:B"

	// read the values from the sheet
	resp, err := client.Spreadsheets.Values.Get(spreadsheetID, rangeStr).Do()
	if err != nil {
		log.Fatalf("Unable to read values from sheet: %v", err)
	}

	matchsStructures := []domain.MatchSheetStructure{}
	for i, row := range resp.Values {
		if len(row) > 0 && row[0] != "" {
			cell := row[0].(string)
			if len(matchsStructures) == 0 {
				matchsStructures = append(matchsStructures, domain.MatchSheetStructure{
					Date:     cell,
					RowBegin: i + 2,
				})
			} else {
				matchsStructures[len(matchsStructures)-1].RowEnd = i + 1
				matchsStructures = append(
					matchsStructures,
					domain.MatchSheetStructure{Date: cell, RowBegin: i + 2},
				)
			}
		}
	}
	matchsStructures[len(matchsStructures)-1].RowEnd = len(resp.Values)

	// Partidos ganados y perdidos
	rangeStr = "partidos!B2:C"
	resp, err = client.Spreadsheets.Values.Get(spreadsheetID, rangeStr).Do()
	if err != nil {
		log.Fatalf("Unable to read values from sheet: %v", err)
	}

	winResults := []domain.MatchResult{}
	for i, match := range matchsStructures {

		if len(resp.Values[i]) == 0 {
			continue
		}

		ok, result := getMatchResultFromSheet(resp, match, domain.MatchResult{Draw: false,
			Date: match.Date, WinnerTeam: []string{}, LoserTeam: []string{}})
		if ok {
			winResults = append(winResults, result)
		}
	}

	// Partidos empatados
	rangeStr = "partidos!D2:E"
	resp, err = client.Spreadsheets.Values.Get(spreadsheetID, rangeStr).Do()
	if err != nil {
		log.Fatalf("Unable to read values from sheet: %v", err)
	}
	drawResults := []domain.MatchResult{}
	for i, match := range matchsStructures {

		if len(resp.Values[i]) == 0 {
			continue
		}

		ok, result := getMatchResultFromSheet(resp, match, domain.MatchResult{Draw: true,
			Date: match.Date, WinnerTeam: []string{}, LoserTeam: []string{}})
		if ok {
			drawResults = append(drawResults, result)
		}
	}

	return append(drawResults, winResults...)
}

func getMatchResultFromSheet(resp *sheets.ValueRange, match domain.MatchSheetStructure, result domain.MatchResult) (bool, domain.MatchResult) {
	for j := match.RowBegin; j < match.RowEnd; j++ {
		if len(resp.Values[j]) == 0 {
			continue
		}
		if resp.Values[j][0] != "" {
			result.WinnerTeam = append(result.WinnerTeam, resp.Values[j][0].(string))
		}
		if resp.Values[j][1] != "" {
			result.LoserTeam = append(result.LoserTeam, resp.Values[j][1].(string))
		}
	}

	return len(result.WinnerTeam) != 0 || len(result.LoserTeam) != 0, result
}
