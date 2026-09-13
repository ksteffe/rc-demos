package webapp

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestPageUsesContentAPIAndGoogleAnalytics(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.String() != "http://content.example/message" {
			t.Fatalf("unexpected content URL %q", request.URL.String())
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"message":"hello from the content API"}`)),
		}, nil
	})}

	handler, err := New(Config{ContentAPIURL: "http://content.example", MeasurementID: "G-DEMO123", Client: client})
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d; body = %s", response.Code, response.Body.String())
	}
	for _, expected := range []string{"hello from the content API", "googletagmanager.com/gtag/js", "G-DEMO123"} {
		if !strings.Contains(response.Body.String(), expected) {
			t.Fatalf("page does not contain %q", expected)
		}
	}
}

func TestConfigurationIsRequired(t *testing.T) {
	if _, err := New(Config{MeasurementID: "G-DEMO123"}); err == nil {
		t.Fatal("expected missing content API URL to fail")
	}
	if _, err := New(Config{ContentAPIURL: "http://example.test", MeasurementID: "invalid"}); err == nil {
		t.Fatal("expected invalid measurement ID to fail")
	}
}
