//go:build integration

package weather_test

import (
	"context"
	"os"
	"testing"

	"github.com/redbonzai/go-weather-service/internal/weather"
)

func TestNWSClient_Live(t *testing.T) {
	userAgent := os.Getenv("NWS_USER_AGENT")
	if userAgent == "" {
		t.Skip("NWS_USER_AGENT not set; skipping live test")
	}
	client, err := weather.NewWeatherServiceClientWithBaseURL(nil, userAgent, "https://api.weather.gov")
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.GetForecastPeriods(context.Background(), 38.8894, -77.0352)
	if err != nil {
		t.Fatalf("live NWS call failed: %v", err)
	}
}
