package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

// WeatherData represents the response from OpenWeatherMap current weather API
type WeatherData struct {
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
	Name string `json:"name"`
	Sys  struct {
		Country string `json:"country"`
	} `json:"sys"`
}

// ForecastData represents the response from OpenWeatherMap forecast API
type ForecastData struct {
	List []ForecastEntry `json:"list"`
	City struct {
		Name    string `json:"name"`
		Country string `json:"country"`
		Coord   struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"coord"`
	} `json:"city"`
}

// ForecastEntry represents a single forecast data point
type ForecastEntry struct {
	Dt   int64 `json:"dt"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
	DtTxt string `json:"dt_txt"`
}

// DailyForecast represents aggregated daily forecast data
type DailyForecast struct {
	Date        time.Time
	TempMin     float64
	TempMax     float64
	Condition   string
	Description string
	Humidity    int
	WindSpeed   float64
}

// GetWeatherData fetches weather data from OpenWeatherMap API
func GetWeatherData(city string) (*WeatherData, error) {
	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENWEATHER_API_KEY environment variable is required")
	}

	// URL encode the city name to handle spaces and special characters
	encodedCity := url.QueryEscape(city)
	apiURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric", encodedCity, apiKey)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}

	var weatherData WeatherData
	if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
		return nil, fmt.Errorf("failed to decode weather data: %w", err)
	}

	return &weatherData, nil
}

// GetForecastData fetches forecast data from OpenWeatherMap API
func GetForecastData(city string, days int) (*ForecastData, error) {
	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENWEATHER_API_KEY environment variable is required")
	}

	// URL encode the city name to handle spaces and special characters
	encodedCity := url.QueryEscape(city)
	
	// Calculate count based on days (forecast API provides data every 3 hours)
	// For 1 day: 8 entries (24 hours / 3 hours per entry)
	// For 7 days: we'll get 5 days max from the free API
	count := days * 8
	if count > 40 { // 5-day forecast API limit
		count = 40
	}

	apiURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?q=%s&appid=%s&units=metric&cnt=%d", encodedCity, apiKey, count)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch forecast data: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("forecast API returned status %d", resp.StatusCode)
	}

	var forecastData ForecastData
	if err := json.NewDecoder(resp.Body).Decode(&forecastData); err != nil {
		return nil, fmt.Errorf("failed to decode forecast data: %w", err)
	}

	return &forecastData, nil
}

// ProcessDailyForecasts converts forecast entries to daily summaries
func ProcessDailyForecasts(forecastData *ForecastData) []DailyForecast {
	dailyMap := make(map[string]*DailyForecast)

	for _, entry := range forecastData.List {
		// Convert Unix timestamp to time
		date := time.Unix(entry.Dt, 0)
		dateKey := date.Format("2006-01-02")

		// Initialize daily forecast if not exists
		if _, exists := dailyMap[dateKey]; !exists {
			dailyMap[dateKey] = &DailyForecast{
				Date:      date,
				TempMin:   entry.Main.TempMin,
				TempMax:   entry.Main.TempMax,
				Condition: entry.Weather[0].Main,
				Description: entry.Weather[0].Description,
				Humidity:  entry.Main.Humidity,
				WindSpeed: entry.Wind.Speed,
			}
		} else {
			// Update min/max temperatures
			if entry.Main.TempMin < dailyMap[dateKey].TempMin {
				dailyMap[dateKey].TempMin = entry.Main.TempMin
			}
			if entry.Main.TempMax > dailyMap[dateKey].TempMax {
				dailyMap[dateKey].TempMax = entry.Main.TempMax
			}
			// Use the most recent weather condition (noon entry preferred)
			hour := date.Hour()
			if hour >= 12 && hour < 15 { // Use midday weather as representative
				dailyMap[dateKey].Condition = entry.Weather[0].Main
				dailyMap[dateKey].Description = entry.Weather[0].Description
			}
		}
	}

	// Convert map to sorted slice
	var dailyForecasts []DailyForecast
	for _, daily := range dailyMap {
		dailyForecasts = append(dailyForecasts, *daily)
	}

	// Sort by date
	for i := 0; i < len(dailyForecasts)-1; i++ {
		for j := 0; j < len(dailyForecasts)-i-1; j++ {
			if dailyForecasts[j].Date.After(dailyForecasts[j+1].Date) {
				dailyForecasts[j], dailyForecasts[j+1] = dailyForecasts[j+1], dailyForecasts[j]
			}
		}
	}

	return dailyForecasts
}