package main

import (
	"log"
	"net/http"
	"os"

	"github.com/runtimeconditions/rc-demos/dev-container-profile/app/internal/webapp"
)

func main() {
	handler, err := webapp.New(webapp.Config{
		ContentAPIURL: os.Getenv("CONTENT_API_URL"),
		MeasurementID: os.Getenv("GA_MEASUREMENT_ID"),
	})
	if err != nil {
		log.Fatal(err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("web demo listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
