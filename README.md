
# STFC Automation Service

## Overview
This project automates the claiming of rewards for **Star Trek Fleet Command** using scheduled cron jobs. It interacts with the game’s API to claim gifts at specific intervals (10 minutes, 4 hours, 24 hours, and daily). If a request fails, the service sends notifications to a configured Slack webhook for visibility.

## Features
- **Cron-based Scheduling:** Automates requests at specific intervals using cron.
- **Error Handling:** Logs request outcomes (both success and failure).
- **Slack Notifications:** Sends notifications on success or failure.
- **Syslog Integration:** Logs messages to syslog and console.
- **Custom Logging:** Uses a custom log formatter to simplify log output.

## Table of Contents
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [How to Get Bearer Token](#how-to-get-bearer-token)
- [Usage](#usage)
- [Project Structure](#project-structure)
- [Cron Job Schedule](#cron-job-schedule)
- [Logging](#logging)
- [Slack Notifications](#slack-notifications)
- [Troubleshooting](#troubleshooting)
- [License](#license)

## Prerequisites
Make sure the following are installed on your system:
- **Go 1.18+** installed and configured.
- **Access to Star Trek Fleet Command API** with a valid bearer token.
- **Slack Webhook URL** for sending notifications.
- A **nix-based environment** (for syslog integration).

## Installation
1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd <repository-folder>
   ```
2. **Install dependencies:**
   ```bash
   go get github.com/robfig/cron/v3
   go get github.com/sirupsen/logrus
   ```
3. **Compile the application:**
   ```bash
   go build -o stfc-automation
   ```

## Configuration
Create a `config.json` file in the project root with the following structure:
```json
{
  "bearerToken": "<your_bearer_token_here>",
  "bundleId10m": 1786571320,
  "bundleId4h": 844758222,
  "bundleId24h": 1918154038,
  "DailyMissionKey": 787829412,
  "OpticalDiode": 1579845062,
  "ReplicatorRations": 1210188306,
  "TrailBells": 718968170,
  "NadionSupply": 1904351560,
  "TranswarpCell": 1438866306,
  "slackWebhookURL": "<your_slack_webhook_url_here>"
}
```

## How to Get Bearer Token
1. Visit [Star Trek Fleet Command](https://home.startrekfleetcommand.com/).
2. Enter your **email** and **password** to log in.
3. Open **Developer Tools** (press `F12` or right-click and select **Inspect**).
4. Go to the **Network** tab in Developer Tools.
5. Submit the login form on the website.
6. Look for a **login** request in the network requests list.
7. Click on the **login** request, and in the **Response** section, locate the `access_token`.
8. Use the `access_token` as the **bearerToken** in your `config.json` file.

## Usage
1. **Run the service:**
   ```bash
   ./stfc-automation
   ```
   Output:
   ```
   Engines to maximum, we're ready for launch
   ```
2. The program will run indefinitely, executing scheduled jobs according to the cron configuration.

## Project Structure
```
.
├── main.go          # Main application code
├── config.json      # Configuration file (user-defined)
└── README.md        # Project documentation
```

## Cron Job Schedule
- **Every 10 minutes + 30 seconds:** Claims gift with `BundleId10m`.
- **Every 4 hours + 30 seconds:** Claims gift with `BundleId4h`.
- **Every day at 10:00:30 AM:** Claims all daily gifts.

### Cron Expressions
- **Every 10 minutes:**
  ```
  30 */10 * * * *
  ```
- **Every 4 hours:**
  ```
  30 0 */4 * * *
  ```
- **Daily at 10:00:30 AM:**
  ```
  30 00 10 * * *
  ```

## Logging
The service uses **Logrus** for logging and sends output to both:
- **Syslog:** Logs the information using the `syslog` package.
- **Console:** Outputs logs to `stdout`.

### Custom Formatter
```go
type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
    return []byte(entry.Message + "
"), nil
}
```

## Slack Notifications
The service sends a **Slack notification** whenever a request succeeds or fails.

### Failure Messages Example:
- **10 Minutes Chest:** ❌ 10 Minutes Chest Failed
- **4 Hours Chest:** ❌ 4 Hours Chest Failed

### Example Slack Payload:
```json
{
  "text": "STFC Automation Error: ❌ 10 Minutes Chest Failed"
}
```

## Troubleshooting
- **Cannot connect to syslog:** Ensure syslog is configured on your system and running correctly.
- **Invalid Bearer Token:** Double-check the `bearerToken` in `config.json` and ensure it hasn’t expired.
- **Network issues:** Ensure your machine can connect to the Star Trek Fleet Command API.

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
