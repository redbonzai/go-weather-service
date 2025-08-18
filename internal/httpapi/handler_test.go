package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/redbonzai/go-weather-service/internal/httpapi"
	"github.com/redbonzai/go-weather-service/internal/weather"
	"golang.org/x/net/context"
)

type fakeSFS struct {
	out string
	err error
}

func (f *fakeSFS) ShortForecastForToday(ctx context.Context, lat, lon float64) (string, error) {
	return f.out, f.err
}

type fakeFC struct {
	periods []weather.ForecastPeriod
	err     error
}

func (f *fakeFC) GetForecastPeriods(ctx context.Context, lat, lon float64) ([]weather.ForecastPeriod, error) {
	return f.periods, f.err
}

type fakeClassifier struct{}

func (fakeClassifier) Classify(t float64) string { return "moderate" }

func TestHandler_HappyPath(t *testing.T) {
	h := &httpapi.Handler{
		Forecast:   &fakeSFS{out: "Partly Cloudy"},
		Client:     &fakeFC{periods: []weather.ForecastPeriod{{Temperature: 72}}},
		Classifier: fakeClassifier{},
	}
	req := httptest.NewRequest(http.MethodGet, "/weather?lat=1&lon=2", nil)
	w := httptest.NewRecorder()
	h.Weather(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		ShortForecast string `json:"shortForecast"`
		TempCategory  string `json:"tempCategory"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.ShortForecast == "" || resp.TempCategory == "" {
		t.Fatalf("bad response: %+v", resp)
	}
}

func TestHandler_BadQuery(t *testing.T) {
	h := &httpapi.Handler{}
	req := httptest.NewRequest(http.MethodGet, "/weather?lat=abc&lon=2", nil)
	w := httptest.NewRecorder()
	h.Weather(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400; got %d", w.Code)
	}
}
