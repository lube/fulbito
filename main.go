package main

import (
	"fmt"
	"fulbito/rating"
	"fulbito/sheets"
	"html/template"
	"log"
	"net/http"
	"strings"
)

func main() {
	helpers := template.FuncMap{
		"FormatNumber": func(value float64) string {
			return fmt.Sprintf("%.4f", value)
		},
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("./static/index.tmpl")
		if err != nil {
			log.Fatal(err)
			return
		}
		err = t.Funcs(helpers).Execute(w, map[string]interface{}{
			"Elo":    rating.ProcessAllMatchResultsAndGetEloRating(sheets.RetrieveGoogleSheetsResults()),
			"Glicko": rating.ProcessAllMatchResultsAndGetGlickoRating(sheets.RetrieveGoogleSheetsResults()),
		})
		if err != nil {
			return
		}
	})
	http.HandleFunc("/teams", func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("./static/teams.tmpl")

		err := r.ParseForm()
		if err != nil {
			return
		}

		data := map[string][]string{
			"TeamA":      {},
			"TeamB":      {},
			"PlayerList": {},
		}

		if len(r.Form["playerList"]) > 0 {
			playerListRaw := r.FormValue("playerList")
			playerList := strings.Split(strings.ReplaceAll(playerListRaw, "\r\n", "\n"), "\n")
			for i := range playerList {
				playerList[i] = strings.Trim(playerList[i], " ")
				data["PlayerList"] = append(data["PlayerList"], playerList[i])
			}

			results := sheets.RetrieveGoogleSheetsResults()
			if r.FormValue("mode") == "ELO" {
				ratings := rating.ProcessAllMatchResultsAndGetEloRating(results)
				data["TeamA"], data["TeamB"] = rating.GenerateTeamsElo(playerList, ratings)
			}

			if r.FormValue("mode") == "GLICKO" {
				ratings := rating.ProcessAllMatchResultsAndGetGlickoRating(results)
				data["TeamA"], data["TeamB"] = rating.GenerateTeamsGlicko(playerList, ratings)
			}
		}

		if err = t.Funcs(helpers).Execute(w, data); err != nil {
			return
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
