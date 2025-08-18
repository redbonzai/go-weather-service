package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/redbonzai/go-weather-service/internal/weather"
	"github.com/redbonzai/go-weather-service/internal/weather/services"
)

type fakeClient struct {
	periods []weather.ForecastPeriod
	err     error
}

//Approach:
//- Inject a fake ForecastClient that returns controlled periods. Construct period StartTimeISO timestamps on today and yesterday to exercise selection.
//
//Edge cases:
//- No periods → error
//- First “today” period picked correctly (even if its Name is “Tonight” or “This Afternoon”)
//- Invalid StartTimeISO strings should be skipped; fallback to first period if none match “today”

func (f *fakeClient) GetForecastPeriods(ctx context.Context, lat, lon float64) ([]weather.ForecastPeriod, error) {
	return f.periods, f.err
}

func iso(t time.Time) string { return t.Format(time.RFC3339) }

func TestShortForecast_TodayPick(t *testing.T) {
	loc := time.Local
	now := time.Now().In(loc)
	yesterday := now.Add(-24 * time.Hour)

	periods := []weather.ForecastPeriod{
		{Name: "Yesterday PM", ShortForecast: "Rain", StartTimeISO: iso(yesterday)},
		{Name: "This Afternoon", ShortForecast: "Partly Cloudy", StartTimeISO: iso(now)},
		{Name: "Tonight", ShortForecast: "Clear", StartTimeISO: iso(now.Add(6 * time.Hour))},
	}
	svc := services.NewShortForecastService(&fakeClient{periods: periods}, loc)

	got, err := svc.ShortForecastForToday(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != "Partly Cloudy" {
		t.Fatalf("got %q, want %q", got, "Partly Cloudy")
	}
}

func TestShortForecast_NoPeriods(t *testing.T) {
	svc := services.NewShortForecastService(&fakeClient{periods: nil}, time.Local)
	_, err := svc.ShortForecastForToday(context.Background(), 0, 0)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestShortForecast_AllBadTimes_FallbackFirst(t *testing.T) {
	periods := []weather.ForecastPeriod{
		{Name: "P1", ShortForecast: "SF1", StartTimeISO: "bad"},
		{Name: "P2", ShortForecast: "SF2", StartTimeISO: "also-bad"},
	}
	svc := services.NewShortForecastService(&fakeClient{periods: periods}, time.Local)
	got, err := svc.ShortForecastForToday(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != "SF1" {
		t.Fatalf("got %q, want SF1", got)
	}
}
