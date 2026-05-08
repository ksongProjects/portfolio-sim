package handlers

import (
	"encoding/json"
	"net/http"
)

func HandleListPortfolios(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode([]map[string]interface{}{})
}

func HandleCreatePortfolio(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": "new"})
}

func HandleListWatchlists(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode([]map[string]interface{}{})
}

func HandleCreateWatchlist(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": "new"})
}

func HandleListTickers(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode([]map[string]interface{}{})
}

func HandleListJobs(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode([]map[string]interface{}{})
}

func HandleCreateJob(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": "new"})
}
