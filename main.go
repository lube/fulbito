package main

import (
	"crypto/tls"
	"fulbito/rating"
	"fulbito/sheets"
	"golang.org/x/crypto/acme/autocert"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("./static/index.tmpl")
		err := t.Execute(w, map[string]interface{}{
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

		if err = t.Execute(w, data); err != nil {
			return
		}
	})

	if os.Getenv("USE_TLS") == "true" {
		srv := &http.Server{Addr: "fulbitodelosviernes.com.ar:https"}
		setupTLS(srv, "sebastianlube@gmail.com", []string{"fulbitodelosviernes.com.ar", "www.fulbitodelosviernes.com.ar"})
		log.Fatal(srv.ListenAndServeTLS("", ""))
	} else {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}

func setupTLS(srv *http.Server, email string, domains []string) {
	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domains...),
		Email:      email,
		Cache:      autocert.DirCache("path/to/cache"),
	}
	srv.TLSConfig = &tls.Config{GetCertificate: m.GetCertificate}
	srv.Handler = m.HTTPHandler(srv.Handler)
}
