package main

import (
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
)

func handleStatus(w http.ResponseWriter, r *http.Request) {
	regions, err := scrape(WorldStatusURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = json.NewEncoder(w).Encode(regions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleSubscribe(conn *sqlx.DB) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		email := r.FormValue("email")
		if email == "" {
			http.Error(w, "empty form submission", http.StatusBadRequest)
			return
		}
		server := r.FormValue("server")
		if server == "" {
			http.Error(w, "empty form submission", http.StatusBadRequest)
			return
		}

		q := "INSERT INTO subscribers (email, server) VALUES (?, ?)"
		_, err = conn.Exec(q, email, server)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	return fn

}
