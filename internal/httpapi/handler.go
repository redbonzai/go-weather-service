package httpapi

// Handler orchestration of weather-related HTTP requests.
import (
	"context"
	"encoding/json"
	_ "errors"
	"net/http"
	"strconv"
	"time"

	"github.com/redbonzai/go-weather-service/internal/weather"
)

type Handler struct {
	Forecast   weather.ShortForecastService
	Client     weather.ForecastClient
	Classifier weather.TemperatureClassifier
}

type response struct {
	ShortForecast string `json:"shortForecast"`
	TempCategory  string `json:"tempCategory"`
}

// Weather GET /weather?lat=..&lon=..
func (handler *Handler) Weather(responseWriter http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	latStr := q.Get("lat")
	lonStr := q.Get("lon")

	lat, err := strconv.ParseFloat(latStr, 64)
	lon, err2 := strconv.ParseFloat(lonStr, 64)
	if latStr == "" || lonStr == "" || err != nil || err2 != nil {
		http.Error(responseWriter, "invalid or missing lat/lon", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	short, err := handler.Forecast.ShortForecastForToday(ctx, lat, lon)
	if err != nil {
		http.Error(responseWriter, "failed to get short forecast: "+err.Error(), http.StatusBadGateway)
		return
	}

	// We need a temperature to classify. Reuse client to fetch periods and pick today's first temp.
	periods, err := handler.Client.GetForecastPeriods(ctx, lat, lon)
	if err != nil || len(periods) == 0 {
		http.Error(responseWriter, "failed to get temperature: "+errMsg(err, "no periods"), http.StatusBadGateway)
		return
	}
	tempCat := handler.Classifier.Classify(periods[0].Temperature)

	writeJSON(responseWriter, response{
		ShortForecast: short,
		TempCategory:  tempCat,
	}, http.StatusOK)
}

func writeJSON(responseWriter http.ResponseWriter, v any, code int) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(code)
	_ = json.NewEncoder(responseWriter).Encode(v)
}

func errMsg(err error, fallback string) string {
	if err != nil {
		return err.Error()
	}
	return fallback
}
