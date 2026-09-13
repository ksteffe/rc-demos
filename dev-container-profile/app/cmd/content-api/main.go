package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /message", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Profiles can compose without changing the Profile specification."})
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}
	log.Printf("content API listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
