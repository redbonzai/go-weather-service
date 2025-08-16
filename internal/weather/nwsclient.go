package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const baseAPIURL = "https://api.weather.gov"

// NewWeatherServiceClient creates a ForecastClient backed by api.weather.gov.
// userAgent must include app id and contact per NWS requirements.
func NewWeatherServiceClient(httpClient *http.Client, userAgent string) (ForecastClient, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	if userAgent == "" {
		return nil, errors.New("user agent required for NWS API")
	}
	return &weatherServiceClient{httpClient: httpClient, userAgent: userAgent}, nil
}

type weatherServiceClient struct {
	httpClient *http.Client
	userAgent  string
}

type pointsResp struct {
	Properties struct {
		Forecast string `json:"forecast"`
	} `json:"properties"`
}

type forecastResp struct {
	Properties struct {
		Periods []struct {
			Name          string  `json:"name"`
			StartTime     string  `json:"startTime"`
			Temperature   float64 `json:"temperature"`
			ShortForecast string  `json:"shortForecast"`
		} `json:"periods"`
	} `json:"properties"`
}

func (client *weatherServiceClient) GetForecastPeriods(ctx context.Context, lat, lon float64) ([]ForecastPeriod, error) {
	// 1) points -> forecast URL
	pointsURL := fmt.Sprintf("%s/points/%.4f,%.4f", baseAPIURL, lat, lon)

	var pr pointsResp
	if err := client.getJSON(ctx, pointsURL, &pr); err != nil {
		return nil, fmt.Errorf("points request failed: %w", err)
	}
	if pr.Properties.Forecast == "" {
		return nil, errors.New("no forecast URL returned by points endpoint")
	}

	// 2) GET forecast URL
	var fr forecastResp
	if err := client.getJSON(ctx, pr.Properties.Forecast, &fr); err != nil {
		return nil, fmt.Errorf("forecast request failed: %w", err)
	}

	out := make([]ForecastPeriod, 0, len(fr.Properties.Periods))
	for _, p := range fr.Properties.Periods {
		out = append(out, ForecastPeriod{
			Name:          p.Name,
			ShortForecast: p.ShortForecast,
			Temperature:   p.Temperature,
			StartTimeISO:  p.StartTime,
		})
	}
	return out, nil
}

func (client *weatherServiceClient) getJSON(ctx context.Context, url string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	// ✅ Prefer Geo+JSON (or just omit Accept entirely)
	req.Header.Set("Accept", "application/geo+json")
	req.Header.Set("User-Agent", client.userAgent)

	// (Optional but helpful) Avoid stale caches while debugging:
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("http %d for %s", resp.StatusCode, url)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}
