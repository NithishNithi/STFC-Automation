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

// Config struct to hold configuration values
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
	SlackWebhookURL   string `json:"slackWebhookURL"`
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
	config, err := ReadConfig("config.json")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Schedule the first cron job (every 10 minutes and 30 seconds)
	_, err = c.AddFunc("30 */10 * * * *", func() {
		fmt.Println("Running cron job: every 10 minutes and 30 seconds")
		ClaimGift(config.BundleId10m, config.BearerToken, logger, config.SlackWebhookURL)
	})
	if err != nil {
		log.Fatalf("Error scheduling the first cron job: %v", err)
	}

	// Schedule the second cron job (every 4 hours and 30 seconds)
	_, err = c.AddFunc("30 0 */4 * * *", func() {
		fmt.Println("Running cron job: every 4 hours and 30 seconds")
		ClaimGift(config.BundleId4h, config.BearerToken, logger, config.SlackWebhookURL)
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
		bundleId := bundleId
		_, err = c.AddFunc("30 00 10 * * *", func() {
			fmt.Printf("Running cron job: daily at 10:00:30 AM for bundle ID %d\n", bundleId)
			ClaimGift(bundleId, config.BearerToken, logger, config.SlackWebhookURL)
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

func ReadConfig(filename string) (*Config, error) {
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

func ClaimGift(bundleId int, bearerToken string, logger *log.Logger, slackWebhookURL string) {
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
	if resp.StatusCode != http.StatusOK {
		err := SendSlackNotification(bundleId, true, slackWebhookURL) // Notify Slack about failure
		if err != nil {
			logger.Printf("Error sending Slack notification: %v\n", err)
		}
	} else {
		err := SendSlackNotification(bundleId, false, slackWebhookURL) // Notify Slack about success
		if err != nil {
			logger.Printf("Error sending Slack notification: %v\n", err)
		}
	}
}
func SendSlackNotification(bundleId int, isFailure bool, webhookURL string) error {
	message := map[string]string{}
	if isFailure {
		// Map bundle IDs to failure messages
		FailureMessages := map[int]string{
			1786571320: "❌ 10 Minutes Chest Failed",
			844758222:  "❌ 4 Hours Chest Failed",
			1918154038: "❌ 24 hour Chest Failed",
			787829412:  "❌ DailyMission Chest Failed",
			1579845062: "❌ OpticalDiode Chest Failed",
			1250837343: "❌ ReplicatorRations Chest Failed",
			718968170:  "❌ TrailBells Chest Failed",
			1904351560: "❌ NadionSupply Chest Failed",
			71216663:   "❌ TranswarpCell Chest Failed",
		}
		// Check if the bundle ID corresponds to a failure message
		failureMessage, found := FailureMessages[bundleId]
		if !found {
			return fmt.Errorf("bundle ID %d does not correspond to a known failure", bundleId)
		}
		message["text"] = fmt.Sprintf("STFC Automation Error: %s", failureMessage)
	} else {
		// Map bundle IDs to success messages
		SuccessMessages := map[int]string{
			// Add your success messages here
			1786571320: "✅ 10 Minutes Chest Successful",
			844758222:  "✅ 4 Hours Chest Successful",
			1918154038: "✅ 24 hour Chest Successful",
			787829412:  "✅ DailyMission Chest Successful",
			1579845062: "✅ OpticalDiode Chest Successful",
			1250837343: "✅ ReplicatorRations Chest Successful",
			718968170:  "✅ TrailBells Chest Successful",
			1904351560: "✅ NadionSupply Chest Successful",
			71216663:   "✅ TranswarpCell Chest Successful",
		}
		// Check if the bundle ID corresponds to a success message
		successMessage, found := SuccessMessages[bundleId]
		if !found {
			return fmt.Errorf("bundle ID %d does not correspond to a known success", bundleId)
		}
		message["text"] = fmt.Sprintf("STFC Automation Success: %s", successMessage)
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshalling Slack message: %v", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(messageBytes))
	if err != nil {
		return fmt.Errorf("error sending Slack notification: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response code: %d", resp.StatusCode)
	}

	fmt.Println("Slack notification sent successfully!")
	return nil
}
