package weather

import "context"

type ForecastPeriod struct {
	Name          string  // e.g "today"
	ShortForecast string  // e.g "Partly cloudy"
	Temperature   float64 // Fahrenheit
	StartTimeISO  string  // e.g "2023-10-01T00:00:00-04:00"
}

type ForecastClient interface {
	GetForecastPeriods(ctx context.Context, lat, lon float64) ([]ForecastPeriod, error)
}

type ShortForecastService interface {
	// ShortForecastForToday returns the short forecast for "today".
	ShortForecastForToday(ctx context.Context, lat, lon float64) (string, error)
}

type TemperatureClassifier interface {
	// Classify returns "hot", "cold", or "moderate" for a given Fahrenheit temperature.
	Classify(tempF float64) string
}
