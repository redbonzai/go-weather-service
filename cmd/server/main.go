package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/redbonzai/go-weather-service/internal/httpapi"
	"github.com/redbonzai/go-weather-service/internal/weather"
	"github.com/redbonzai/go-weather-service/internal/weather/services"
)

func main() {
	userAgent := os.Getenv("NWS_USER_AGENT")
	if userAgent == "" {
		log.Fatal("NWS_USER_AGENT env var is required (e.g. 'go-weather/1.0 (you@example.com)')")
	}

	coldLT := parseFloatEnv("COLD_LT_F", 50) // default thresholds: <50 cold
	hotGE := parseFloatEnv("HOT_GE_F", 80)   // >=80 hot

	client, err := weather.NewWeatherServiceClient(nil, userAgent)
	if err != nil {
		log.Fatal(err)
	}

	loc := time.Local
	sfSvc := services.NewShortForecastService(client, loc)
	classifier := services.NewTemperatureClassifier(coldLT, hotGE)

	handler := &httpapi.Handler{
		Forecast:   sfSvc,
		Client:     client,
		Classifier: classifier,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/weather", handler.Weather)

	addr := ":8080"
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// Parsing ENV variables for thresholds
func parseFloatEnv(key string, def float64) float64 {
	if envVal := os.Getenv(key); envVal != "" {
		if f, err := parseFloat(envVal); err == nil {
			return f
		}
	}
	return def
}

func parseFloat(string string) (float64, error) {
	var float float64
	_, err := fmt.Sscan(string, &float)
	return float, err
}
