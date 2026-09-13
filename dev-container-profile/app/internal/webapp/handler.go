package webapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var measurementIDPattern = regexp.MustCompile(`^G-[A-Z0-9]+$`)

var page = template.Must(template.New("page").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Runtime Conditions profile composition demo</title>
  <script async src="https://www.googletagmanager.com/gtag/js?id={{.MeasurementID}}"></script>
  <script>
    window.dataLayer = window.dataLayer || [];
    function gtag(){dataLayer.push(arguments);}
    gtag('js', new Date());
    gtag('config', '{{.MeasurementID}}');
  </script>
</head>
<body>
  <main>
    <h1>Runtime Conditions profile composition demo</h1>
    <p id="message">{{.Message}}</p>
  </main>
</body>
</html>
`))

type Config struct {
	ContentAPIURL string
	MeasurementID string
	Client        *http.Client
}

func New(config Config) (http.Handler, error) {
	config.ContentAPIURL = strings.TrimRight(config.ContentAPIURL, "/")
	if config.ContentAPIURL == "" {
		return nil, errors.New("CONTENT_API_URL is required")
	}
	if !measurementIDPattern.MatchString(config.MeasurementID) {
		return nil, errors.New("GA_MEASUREMENT_ID must match G-[A-Z0-9]+")
	}
	if config.Client == nil {
		config.Client = &http.Client{Timeout: 3 * time.Second}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		message, err := fetchMessage(r.Context(), config)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := page.Execute(w, map[string]string{
			"MeasurementID": config.MeasurementID,
			"Message":       message,
		}); err != nil {
			http.Error(w, "render page", http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fetchMessage(r.Context(), config); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	return mux, nil
}

func fetchMessage(ctx context.Context, config Config) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, config.ContentAPIURL+"/message", nil)
	if err != nil {
		return "", err
	}
	response, err := config.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("content API request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("content API returned %s", response.Status)
	}
	var payload struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode content API response: %w", err)
	}
	if payload.Message == "" {
		return "", errors.New("content API returned an empty message")
	}
	return payload.Message, nil
}
