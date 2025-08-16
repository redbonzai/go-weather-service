package services

import (
	"context"
	"errors"
	"time"

	"github.com/redbonzai/go-weather-service/internal/util"
	"github.com/redbonzai/go-weather-service/internal/weather"
)

// ---- Short forecast ----

type shortForecastSvc struct {
	client weather.ForecastClient
	loc    *time.Location
}

// NewShortForecastService returns a service that extracts today's short forecast.
func NewShortForecastService(client weather.ForecastClient, loc *time.Location) weather.ShortForecastService {
	if loc == nil {
		loc = time.Local
	}
	return &shortForecastSvc{client: client, loc: loc}
}

func (s *shortForecastSvc) ShortForecastForToday(ctx context.Context, lat, lon float64) (string, error) {
	periods, err := s.client.GetForecastPeriods(ctx, lat, lon)
	if err != nil {
		return "", err
	}
	if len(periods) == 0 {
		return "", errors.New("no forecast periods")
	}
	now := time.Now().In(s.loc)
	for _, p := range periods {
		// choose first period starting today (local). Period names vary ("Today", "This Afternoon", etc)
		start, err := time.Parse(time.RFC3339, p.StartTimeISO)
		if err != nil {
			continue
		}
		if util.IsSameLocalDay(start, now, s.loc) {
			return p.ShortForecast, nil
		}
	}
	// Fallback: if nothing starts today, return first period
	return periods[0].ShortForecast, nil
}

// ---- Temperature classifier ----

type thresholds struct {
	ColdLT float64 // temp < ColdLT => "cold"
	HotGE  float64 // temp >= HotGE => "hot"
}

type defaultTempClassifier struct {
	thr thresholds
}

// NewTemperatureClassifier returns a classifier with provided thresholds.
// Example: ColdLT=50, HotGE=80.
func NewTemperatureClassifier(coldLT, hotGE float64) weather.TemperatureClassifier {
	return &defaultTempClassifier{thr: thresholds{ColdLT: coldLT, HotGE: hotGE}}
}

func (c *defaultTempClassifier) Classify(tempF float64) string {
	if tempF < c.thr.ColdLT {
		return "cold"
	}
	if tempF >= c.thr.HotGE {
		return "hot"
	}
	return "moderate"
}
