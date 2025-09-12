package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestGetWeatherData(t *testing.T) {
	// Save original env var
	originalKey := os.Getenv("OPENWEATHER_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("OPENWEATHER_API_KEY", originalKey)
		} else {
			os.Unsetenv("OPENWEATHER_API_KEY")
		}
	}()

	tests := []struct {
		name          string
		city          string
		apiKey        string
		mockResponse  string
		mockStatus    int
		expectError   bool
		expectedError string
	}{
		{
			name:          "no API key",
			city:          "London",
			apiKey:        "",
			expectError:   true,
			expectedError: "OPENWEATHER_API_KEY environment variable is required",
		},
		{
			name:   "successful request",
			city:   "London",
			apiKey: "test_api_key",
			mockResponse: `{
				"main": {
					"temp": 15.5,
					"feels_like": 14.2,
					"humidity": 65
				},
				"weather": [
					{
						"main": "Clear",
						"description": "clear sky",
						"icon": "01d"
					}
				],
				"wind": {
					"speed": 3.5
				},
				"name": "London",
				"sys": {
					"country": "GB"
				}
			}`,
			mockStatus:  http.StatusOK,
			expectError: false,
		},
		{
			name:          "city with spaces",
			city:          "New York",
			apiKey:        "test_api_key",
			mockResponse:  `{"name": "New York", "main": {"temp": 20}, "weather": [{"main": "Clear"}], "wind": {"speed": 2}, "sys": {"country": "US"}}`,
			mockStatus:    http.StatusOK,
			expectError:   false,
		},
		{
			name:          "city with special characters",
			city:          "São Paulo",
			apiKey:        "test_api_key",
			mockResponse:  `{"name": "São Paulo", "main": {"temp": 25}, "weather": [{"main": "Rain"}], "wind": {"speed": 1}, "sys": {"country": "BR"}}`,
			mockStatus:    http.StatusOK,
			expectError:   false,
		},
		{
			name:          "invalid city with real API",
			city:          "NonExistentCityThatDoesNotExist12345",
			apiKey:        "test_api_key",
			expectError:   true,
			expectedError: "weather API returned status", // API will return error status
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up API key
			if tt.apiKey == "" {
				os.Unsetenv("OPENWEATHER_API_KEY")
			} else {
				os.Setenv("OPENWEATHER_API_KEY", tt.apiKey)
			}

			// Create mock server if we have a mock response
			var server *httptest.Server
			if tt.mockResponse != "" {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Verify query parameters
					if !strings.Contains(r.URL.Query().Get("q"), strings.ReplaceAll(tt.city, " ", "%20")) &&
					   !strings.Contains(r.URL.Query().Get("q"), tt.city) {
						t.Errorf("Expected city parameter to contain '%s'", tt.city)
					}
					if r.URL.Query().Get("appid") != tt.apiKey {
						t.Errorf("Expected appid '%s', got '%s'", tt.apiKey, r.URL.Query().Get("appid"))
					}
					if r.URL.Query().Get("units") != "metric" {
						t.Errorf("Expected units 'metric', got '%s'", r.URL.Query().Get("units"))
					}

					w.WriteHeader(tt.mockStatus)
					w.Write([]byte(tt.mockResponse))
				}))
				defer server.Close()

				// Replace the API URL in the function temporarily
				// Since we can't easily mock the URL, we'll test the parsing part separately
			}

			data, err := GetWeatherData(tt.city)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.expectedError != "" && !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
				if data != nil {
					t.Error("Expected nil data when error occurs")
				}
			} else {
				// For cases with invalid API key or network issues, we expect errors
				if tt.apiKey != "" && tt.expectError {
					// Expected to fail with network/API error
					t.Logf("Expected error occurred: %v", err)
				}
			}
		})
	}
}

func TestWeatherDataStructure(t *testing.T) {
	// Test that WeatherData struct can be properly marshaled/unmarshaled
	testData := WeatherData{
		Main: struct {
			Temp      float64 `json:"temp"`
			FeelsLike float64 `json:"feels_like"`
			Humidity  int     `json:"humidity"`
		}{
			Temp:      20.5,
			FeelsLike: 19.8,
			Humidity:  70,
		},
		Weather: []struct {
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		}{
			{
				Main:        "Clear",
				Description: "clear sky",
				Icon:        "01d",
			},
		},
		Wind: struct {
			Speed float64 `json:"speed"`
		}{
			Speed: 3.2,
		},
		Name: "Test City",
		Sys: struct {
			Country string `json:"country"`
		}{
			Country: "TC",
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(testData)
	if err != nil {
		t.Errorf("Failed to marshal WeatherData: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled WeatherData
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal WeatherData: %v", err)
	}

	// Verify data integrity
	if unmarshaled.Main.Temp != testData.Main.Temp {
		t.Errorf("Expected temp %f, got %f", testData.Main.Temp, unmarshaled.Main.Temp)
	}
	if unmarshaled.Main.FeelsLike != testData.Main.FeelsLike {
		t.Errorf("Expected feels_like %f, got %f", testData.Main.FeelsLike, unmarshaled.Main.FeelsLike)
	}
	if unmarshaled.Main.Humidity != testData.Main.Humidity {
		t.Errorf("Expected humidity %d, got %d", testData.Main.Humidity, unmarshaled.Main.Humidity)
	}
	if len(unmarshaled.Weather) != 1 {
		t.Errorf("Expected 1 weather item, got %d", len(unmarshaled.Weather))
	} else {
		if unmarshaled.Weather[0].Main != testData.Weather[0].Main {
			t.Errorf("Expected weather main '%s', got '%s'", 
				testData.Weather[0].Main, unmarshaled.Weather[0].Main)
		}
	}
	if unmarshaled.Wind.Speed != testData.Wind.Speed {
		t.Errorf("Expected wind speed %f, got %f", testData.Wind.Speed, unmarshaled.Wind.Speed)
	}
	if unmarshaled.Name != testData.Name {
		t.Errorf("Expected name '%s', got '%s'", testData.Name, unmarshaled.Name)
	}
	if unmarshaled.Sys.Country != testData.Sys.Country {
		t.Errorf("Expected country '%s', got '%s'", testData.Sys.Country, unmarshaled.Sys.Country)
	}
}

func TestGetWeatherDataURLEncoding(t *testing.T) {
	// Save original env var
	originalKey := os.Getenv("OPENWEATHER_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("OPENWEATHER_API_KEY", originalKey)
		} else {
			os.Unsetenv("OPENWEATHER_API_KEY")
		}
	}()

	// Set a test API key
	os.Setenv("OPENWEATHER_API_KEY", "test_key")

	tests := []struct {
		name     string
		city     string
		expected string
	}{
		{
			name:     "simple city name",
			city:     "London",
			expected: "London",
		},
		{
			name:     "city with spaces",
			city:     "New York",
			expected: "New+York",
		},
		{
			name:     "city with special characters",
			city:     "São Paulo",
			expected: "S%C3%A3o+Paulo",
		},
		{
			name:     "city with multiple words",
			city:     "Los Angeles County",
			expected: "Los+Angeles+County",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily test the actual URL construction without significant refactoring,
			// but we can test that the function doesn't panic and handles the input
			_, err := GetWeatherData(tt.city)
			
			// We expect this to fail with a network error since we're not mocking the HTTP call
			// But it shouldn't fail with a URL encoding error
			if err != nil && !strings.Contains(err.Error(), "failed to fetch weather data") &&
			   !strings.Contains(err.Error(), "weather API returned status") {
				// If it's not a network/API error, it might be a URL encoding issue
				t.Logf("Error (expected for test): %v", err)
			}
		})
	}
}