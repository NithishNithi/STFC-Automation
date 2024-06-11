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
	"gopkg.in/gomail.v2" // New SMTP package
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
}

// EmailConfig struct to hold email configuration
type EmailConfig struct {
	SenderName     string `json:"sender_name"`
	SenderEmail    string `json:"sender_email"`
	SenderPassword string `json:"sender_password"`
	RecipientEmail string `json:"recipient_email"`
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
		_, err = c.AddFunc("35 0 12 * * *", func() {
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
	if resp.StatusCode != http.StatusOK {
		err := SendEmailNotification(bundleId, "email.config")
		if err != nil {
			logger.Printf("Error sending email: %v\n", err)
		}
	}
}

func SendEmailNotification(bundleid int, configFilePath string) error {
	// Map bundle IDs to failure messages
	failureMessages := map[int]string{
		1786571320: "bundleId10m Chest Failed",
		844758222:  "bundleId4h Chest Failed",
		000000000:  "24 hour Chest Failed",
		787829412:  "dailymission Chest Failed",
		1579845062: "OpticalDiode Chest Failed",
		1250837343: "ReplicatorRations Chest Failed",
		718968170:  "TrailBells Chest Failed",
		1904351560: "NadionSupply Chest Failed",
		71216663:   "TranswarpCell Chest Failed",
	}

	// Check if the bundle ID corresponds to a failure message
	failureMessage, found := failureMessages[bundleid]
	if !found {
		return fmt.Errorf("bundle ID %d does not correspond to a known failure", bundleid)
	}

	// Read the email configuration from the JSON file
	file, err := os.Open(configFilePath)
	if err != nil {
		return fmt.Errorf("error opening config file: %v", err)
	}
	defer file.Close()

	var config EmailConfig
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return fmt.Errorf("error decoding JSON: %v", err)
	}

	// Create a new message
	m := gomail.NewMessage()
	m.SetHeader("From", config.SenderEmail)
	m.SetHeader("To", config.RecipientEmail)
	m.SetHeader("Subject", "STFC - AutomationError")
	// Include failure message in the email body
	m.SetBody("text/plain", fmt.Sprintf("STFC automation error: %s", failureMessage))

	// Create new SMTP dialer
	d := gomail.NewDialer("smtp.gmail.com", 587, config.SenderEmail, config.SenderPassword)

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("error sending email: %v", err)
	}

	fmt.Println("Email sent successfully!")
	return nil
}
