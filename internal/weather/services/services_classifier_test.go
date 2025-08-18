package services_test

import (
	"testing"

	"github.com/redbonzai/go-weather-service/internal/weather/services"
)

func TestTemperatureClassifier(t *testing.T) {
	//What to verify:
	//
	//< ColdLT → cold
	//
	//== HotGE and > HotGE → hot
	//
	//in-between → moderate

	classifier := services.NewTemperatureClassifier(50, 80)

	tests := []struct {
		name string
		in   float64
		want string
	}{
		{"below cold", 49.9, "cold"},
		{"at cold boundary is moderate", 50.0, "moderate"},
		{"moderate mid", 65.0, "moderate"},
		{"at hot boundary", 80.0, "hot"},
		{"above hot", 99.0, "hot"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(testing *testing.T) {
			if got := classifier.Classify(tt.in); got != tt.want {
				testing.Fatalf("Classify(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
