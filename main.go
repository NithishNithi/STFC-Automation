package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/robfig/cron/v3"
)

const url = "https://storeapi.startrekfleetcommand.com/api/v2/offers/gifts/claim"

type Config struct {
	BearerToken       string `json:"bearerToken"`
	BundleId10m       int    `json:"bundleId10m"`
	BundleId4h        int    `json:"bundleId4h"`
	BundleId24h       int    `json:"bundleId24h"`
	DailyMissionKey   int    `json:"DailyMissionKey"`
	OpticalDiode      int    `json:"OpticalDiode"`
	ReplicatorRations int    `json:"ReplicatorRations"`
	TrailBells        int    `json:"TrailBells"`
	NadionSupply      int    `json:"NadionSupply"`
	TranswarpCell     int    `json:"TranswarpCell"`
}

func main() {
	c := cron.New(cron.WithSeconds()) // Enable seconds field

	// Open log file
	logFile, err := os.OpenFile("stfc.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.LstdFlags)

	// Read config file
	config, err := readConfig("config.json")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Schedule the first cron job (every 10 minutes and 30 seconds)
	_, err = c.AddFunc("30 */10 * * * *", func() {
		fmt.Println("Running cron job: every 10 minutes and 30 seconds")
		claimGift(config.BundleId10m, config.BearerToken, logger)
	})
	if err != nil {
		log.Fatalf("Error scheduling the first cron job: %v", err)
	}

	// Schedule the second cron job (every 4 hours and 30 seconds)
	_, err = c.AddFunc("30 0 */4 * * *", func() {
		fmt.Println("Running cron job: every 4 hours and 30 seconds")
		claimGift(config.BundleId4h, config.BearerToken, logger)
	})
	if err != nil {
		log.Fatalf("Error scheduling the second cron job: %v", err)
	}

	// Schedule the daily cron jobs at 10:00:30 AM
	bundleIDs := []int{
		config.BundleId24h,
		config.DailyMissionKey,
		config.OpticalDiode,
		config.ReplicatorRations,
		config.TrailBells,
		config.NadionSupply,
		config.TranswarpCell,
	}

	for _, bundleId := range bundleIDs {
		bundleId := bundleId // Capture loop variable
		_, err = c.AddFunc("30 0 10 * * *", func() {
			fmt.Printf("Running cron job: daily at 10:00:30 AM for bundle ID %d\n", bundleId)
			claimGift(bundleId, config.BearerToken, logger)
		})
		if err != nil {
			log.Fatalf("Error scheduling daily cron job for bundle ID %d: %v", bundleId, err)
		}
	}

	c.Start()
	fmt.Println("Cron jobs started. Press Ctrl+C to exit.")

	// Wait indefinitely
	select {}
}

func readConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func claimGift(bundleId int, bearerToken string, logger *log.Logger) {
	payload := map[string]int{"bundleId": bundleId}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Fatalf("Error marshalling payload: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		logger.Fatalf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bearerToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Fatalf("Error reading response body: %v", err)
	}

	// Log response status and body
	logger.Printf("Bundle ID: %d, Status: %s, Response: %s\n", bundleId, resp.Status, body)
}
