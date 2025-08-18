package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// defaultBaseURL is used in production; tests can override via constructor.
const defaultBaseURL = "https://api.weather.gov"

type weatherServiceClient struct {
	baseURL    string
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

// NewWeatherServiceClient: production constructor (uses api.weather.gov)
func NewWeatherServiceClient(httpClient *http.Client, userAgent string) (ForecastClient, error) {
	return NewWeatherServiceClientWithBaseURL(httpClient, userAgent, defaultBaseURL)
}

// NewWeatherServiceClientWithBaseURL: test/alt-env constructor
func NewWeatherServiceClientWithBaseURL(httpClient *http.Client, userAgent, baseURL string) (ForecastClient, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	if userAgent == "" {
		return nil, errors.New("user agent required for NWS API")
	}
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &weatherServiceClient{
		baseURL:    baseURL,
		httpClient: httpClient,
		userAgent:  userAgent,
	}, nil
}

func (c *weatherServiceClient) GetForecastPeriods(ctx context.Context, lat, lon float64) ([]ForecastPeriod, error) {
	pointsURL := fmt.Sprintf("%s/points/%.4f,%.4f", c.baseURL, lat, lon)

	var pr pointsResp
	if err := c.getJSON(ctx, pointsURL, &pr); err != nil {
		return nil, fmt.Errorf("points request failed: %w", err)
	}
	if pr.Properties.Forecast == "" {
		return nil, errors.New("no forecast URL returned by points endpoint")
	}

	var forecastResponse forecastResp
	if err := c.getJSON(ctx, pr.Properties.Forecast, &forecastResponse); err != nil {
		return nil, fmt.Errorf("forecast request failed: %w", err)
	}

	out := make([]ForecastPeriod, 0, len(forecastResponse.Properties.Periods))
	for _, p := range forecastResponse.Properties.Periods {
		out = append(out, ForecastPeriod{
			Name:          p.Name,
			ShortForecast: p.ShortForecast,
			Temperature:   p.Temperature,
			StartTimeISO:  p.StartTime,
		})
	}
	return out, nil
}

func (c *weatherServiceClient) getJSON(ctx context.Context, url string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	// Use GeoJSON (or omit Accept) so /points returns properties.forecast
	req.Header.Set("Accept", "application/geo+json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("http %d for %s", resp.StatusCode, url)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}
