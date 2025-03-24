package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

const weatherAPI = "https://api.open-meteo.com/v1/forecast?latitude=37.3996&longitude=144.5884&current=temperature_2m,precipitation,wind_speed_10m,relative_humidity_2m,rain,apparent_temperature,wind_gusts_10m&timezone=Australia%2FSydney"

type weatherAPIResponse struct {
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

type weatherResponse struct {
	Time          string `json:"time"`
	Temperature   string `json:"temperature"`
	ApparentTemp  string `json:"apparent_temperature"`
	WindSpeed     string `json:"wind_speed_kmh"`
	WindGusts     string `json:"wind_gusts_kmh"`
	Humidity      string `json:"humidity_percent"`
	Precipitation string `json:"precipitation_mm"`
	Rain          string `json:"rain_mm"`
}

var processRequest = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Processing request")
	resp, err := http.Get(weatherAPI)
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
		Temperature:   fmt.Sprintf("%.1f°C", apiResp.CurrentWeather.Temperature),
		ApparentTemp:  fmt.Sprintf("%.1f°C", apiResp.CurrentWeather.ApparentTemperature),
		WindSpeed:     fmt.Sprintf("%.1f km/h", apiResp.CurrentWeather.WindSpeed),
		WindGusts:     fmt.Sprintf("%.1f km/h", apiResp.CurrentWeather.WindGusts),
		Humidity:      fmt.Sprintf("%d%%", apiResp.CurrentWeather.Humidity),
		Precipitation: fmt.Sprintf("%.2f mm", apiResp.CurrentWeather.Precipitation),
		Rain:          fmt.Sprintf("%.2f mm", apiResp.CurrentWeather.Rain),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(weather); err != nil {
		log.Printf("Failed to write response: %v", err)
	}

}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received request for unhandled path: %s\n", r.URL.Path)
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "404 - The path %s was not found on this server.", r.URL.Path)
}

func main() {

	listenAdddr := os.Getenv("LISTEN_ADDR")
	if listenAdddr == "" {
		listenAdddr = "0.0.0.0:8080"
	}

	router := mux.NewRouter()
	router.Methods("GET").Path("/").HandlerFunc(processRequest)
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	webServer := &http.Server{
		Addr:         listenAdddr,
		Handler:      router,
		ReadTimeout:  10 * time.Minute,
		WriteTimeout: 20 * time.Minute,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		log.Printf("Server listening on %s", webServer.Addr)
		if err := webServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", webServer.Addr, err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := webServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed: %v", err)
	}
	log.Println("Server exited cleanly.")
}
