package weather_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/redbonzai/go-weather-service/internal/weather"
)

func TestNWSClient_Flow(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/points/38.0000,-77.0000", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Fatalf("missing UA")
		}
		w.Header().Set("Content-Type", "application/geo+json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"properties": map[string]any{
				"forecast": testSrvURL(r, "/gridpoints/XX/1,2/forecast"),
			},
		})
	})
	mux.HandleFunc("/gridpoints/XX/1,2/forecast", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/geo+json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"properties": map[string]any{
				"periods": []map[string]any{
					{"name": "Today", "startTime": time.Now().Format(time.RFC3339), "temperature": 70.0, "shortForecast": "Sunny"},
				},
			},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Point the client at our test server by overriding baseURL via a helper or a test-only setter.
	// If baseURL is const, consider refactoring to make it configurable for tests.
	client, err := weather.NewWeatherServiceClientWithBaseURL(
		&http.Client{Timeout: 2 * time.Second},
		"test/1.0 (test@example.com)",
		srv.URL,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Hack: if baseURL is not configurable, skip; otherwise set it to srv.URL

	// Assuming you modified NewNWSClient to accept baseURL in tests, then:
	periods, err := client.GetForecastPeriods(context.Background(), 38.0, -77.0)
	if err != nil {
		t.Fatal(err)
	}
	if len(periods) != 1 || periods[0].ShortForecast != "Sunny" {
		t.Fatalf("unexpected periods: %+v", periods)
	}
}

func testSrvURL(r *http.Request, path string) string {
	return "http://" + r.Host + path
}
