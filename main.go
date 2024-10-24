package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log/syslog"
	"net/http"
	"os"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// CustomFormatter to format logs with only the message
type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(entry.Message + "\n"), nil
}

const url = "https://storeapi.startrekfleetcommand.com/api/v2/offers/gifts/claim"

// Config struct to hold configuration values
type Config struct {
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
	fmt.Println("Engines to maximum, we're ready for launch")
	c := cron.New(cron.WithSeconds()) // Enable seconds field

	// Read config file
	config, err := ReadConfig("config.json")
	if err != nil {
		logrus.Fatalf("Error reading config file: %v", err)
	}

	// Get Bearer Token and Slack Webhook URL from environment variables
	bearerToken := os.Getenv("BEARER_TOKEN")
	if bearerToken == "" {
		logrus.Fatal("Environment variable BEARER_TOKEN not set")
	}

	slackWebhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if slackWebhookURL == "" {
		logrus.Fatal("Environment variable SLACK_WEBHOOK_URL not set")
	}

	// Configure logrus to output logs to both syslog and console
	sysLog, err := syslog.New(syslog.LOG_INFO, "stfc-automation")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to syslog: %v", err)
		os.Exit(1)
	}

	// Create a multi-writer to write to both os.Stdout and syslog
	mw := io.MultiWriter(os.Stdout, sysLog)
	logrus.SetOutput(mw)
	logrus.SetFormatter(&CustomFormatter{})

	// Schedule the first cron job (every 10 minutes and 30 seconds)
	_, err = c.AddFunc("30 */10 * * * *", func() {
		logrus.Info("Running cron job: every 10 minutes and 30 seconds")
		ClaimGift(config.BundleId10m, bearerToken, slackWebhookURL)
	})
	if err != nil {
		logrus.Fatalf("Error scheduling the first cron job: %v", err)
	}

	// Schedule the second cron job (every 4 hours and 30 seconds)
	_, err = c.AddFunc("30 0 */4 * * *", func() {
		logrus.Info("Running cron job: every 4 hours and 30 seconds")
		ClaimGift(config.BundleId4h, bearerToken, slackWebhookURL)
	})
	if err != nil {
		logrus.Fatalf("Error scheduling the second cron job: %v", err)
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
			logrus.Infof("Running cron job: daily at 10:00:30 AM for bundle ID %d\n", bundleId)
			ClaimGift(bundleId, bearerToken, slackWebhookURL)
		})
		if err != nil {
			logrus.Fatalf("Error scheduling daily cron job for bundle ID %d: %v", bundleId, err)
		}
	}

	c.Start()
	logrus.Warn("Engines to maximum, we're ready for launch")

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

func ClaimGift(bundleId int, bearerToken string, slackWebhookURL string) {
	payload := map[string]int{"bundleId": bundleId}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logrus.Errorf("Error marshalling payload: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		logrus.Errorf("Error creating request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bearerToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("Error making request: %v\n", err)
		go SendSlackNotification(bundleId, true, slackWebhookURL)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("Error reading response body: %v\n", err)
		go SendSlackNotification(bundleId, true, slackWebhookURL)
		return
	}

	logrus.Printf("Bundle ID: %d, Status: %s, Response: %s\n", bundleId, resp.Status, body)
	if resp.StatusCode != http.StatusOK {
		go SendSlackNotification(bundleId, true, slackWebhookURL) // Notify Slack about failure
	} else {
		go SendSlackNotification(bundleId, false, slackWebhookURL) // Notify Slack about success
	}
}

func SendSlackNotification(bundleId int, isFailure bool, webhookURL string) {
	var message string
	if isFailure {
		FailureMessages := map[int]string{
			1786571320: "❌ 10 Minutes Chest Failed",
			844758222:  "❌ 4 Hours Chest Failed",
			1918154038: "❌ 24 hour Chest Failed",
			787829412:  "❌ DailyMission Chest Failed",
			1579845062: "❌ OpticalDiode Chest Failed",
			1210188306: "❌ ReplicatorRations Chest Failed",
			718968170:  "❌ TrailBells Chest Failed",
			1904351560: "❌ NadionSupply Chest Failed",
			1438866306: "❌ TranswarpCell Chest Failed",
		}
		failureMessage, found := FailureMessages[bundleId]
		if !found {
			logrus.Printf("Bundle ID %d does not correspond to a known failure\n", bundleId)
			return
		}
		message = fmt.Sprintf("STFC Automation Error: %s", failureMessage)
	} else {
		SuccessMessages := map[int]string{
			// 1786571320: "✅ 10 Minutes Chest Successful",
			844758222:  "✅ 4 Hours Chest Successful",
			1918154038: "✅ 24 hour Chest Successful",
			787829412:  "✅ DailyMission Chest Successful",
			1579845062: "✅ OpticalDiode Chest Successful",
			1210188306: "✅ ReplicatorRations Chest Successful",
			718968170:  "✅ TrailBells Chest Successful",
			1904351560: "✅ NadionSupply Chest Successful",
			1438866306: "✅ TranswarpCell Chest Successful",
		}
		successMessage, found := SuccessMessages[bundleId]
		if !found {
			logrus.Printf("Bundle ID %d does not correspond to a known success\n", bundleId)
			return
		}
		message = fmt.Sprintf("STFC Automation Success: %s", successMessage)
	}

	slackMessage := map[string]string{"text": message}
	messageBytes, err := json.Marshal(slackMessage)
	if err != nil {
		logrus.Errorf("Error marshalling Slack message: %v\n", err)
		return
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(messageBytes))
	if err != nil {
		logrus.Errorf("Error sending Slack notification: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logrus.Printf("Received non-OK response code: %d\n", resp.StatusCode)
		return
	}

	logrus.Println("Slack notification sent successfully!")
}
