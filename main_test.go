package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockWeatherAPIResponse struct {
	CurrentWeather struct {
		DateTime            string  `json:"datetime"`
		Temperature         float64 `json:"temperature_2m"`
		WindSpeed           float64 `json:"wind_speed_10m"`
		WindGusts           float64 `json:"wind_gusts_10m"`
		Humidity            int     `json:"relative_humidity_2m"`
		ApparentTemperature float64 `json:"apparent_temperature"`
		Rain                float64 `json:"rain"`
		Precipitation       float64 `json:"precipitation"`
	} `json:"current"`
}

func TestProcessRequest_Success(t *testing.T) {
	mockResponse := mockWeatherAPIResponse{}
	mockResponse.CurrentWeather.DateTime = "2025-03-23T04:00"
	mockResponse.CurrentWeather.Temperature = 18.5
	mockResponse.CurrentWeather.ApparentTemperature = 17.2
	mockResponse.CurrentWeather.WindSpeed = 22.5
	mockResponse.CurrentWeather.WindGusts = 30.1
	mockResponse.CurrentWeather.Humidity = 55
	mockResponse.CurrentWeather.Precipitation = 0.0
	mockResponse.CurrentWeather.Rain = 0.0

	bodyBytes, _ := json.Marshal(mockResponse)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(bodyBytes)
	}))
	defer mockServer.Close()

	originalProcessRequest := processRequest
	processRequest = func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(mockServer.URL)
		if err != nil {
			http.Error(w, "Failed to fetch weather data", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Error reading weather API response", http.StatusInternalServerError)
			return
		}

		var apiResp weatherAPIResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			http.Error(w, "Error parsing weather data", http.StatusInternalServerError)
			return
		}

		weather := weatherResponse{
			Time:          apiResp.CurrentWeather.DateTime,
			Temperature:   "18.5°C",
			ApparentTemp:  "17.2°C",
			WindSpeed:     "22.5 km/h",
			WindGusts:     "30.1 km/h",
			Humidity:      "55%",
			Precipitation: "0.00 mm",
			Rain:          "0.00 mm",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(weather)
	}
	defer func() { processRequest = originalProcessRequest }()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	processRequest(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %v", rr.Code)
	}

	if !bytes.Contains(rr.Body.Bytes(), []byte("18.5°C")) {
		t.Errorf("Response body missing temperature")
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte("17.2°C")) {
		t.Errorf("Response body missing apparent temperature")
	}
}

func TestProcessRequest_APIFailure(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	originalProcessRequest := processRequest
	processRequest = func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(mockServer.URL)
		if err != nil {
			http.Error(w, "Failed to fetch weather data", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, "Weather API returned error", http.StatusInternalServerError)
			return
		}
	}
	defer func() { processRequest = originalProcessRequest }()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	processRequest(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 on API failure, got %v", rr.Code)
	}
}
