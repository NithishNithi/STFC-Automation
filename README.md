<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>STFC Automation Service</title>
</head>
<body>
    <h1>STFC Automation Service</h1>

    <h2>Overview</h2>
    <p>
        This project automates the claiming of rewards for 
        <em>Star Trek Fleet Command</em> using scheduled cron jobs. 
        It interacts with the game’s API to claim gifts at specific intervals 
        (10 minutes, 4 hours, 24 hours, and daily). If a request fails, 
        the service sends notifications to a configured Slack webhook for visibility.
    </p>

    <h2>Features</h2>
    <ul>
        <li><strong>Cron-based Scheduling:</strong> Automates requests at specific intervals using cron.</li>
        <li><strong>Error Handling:</strong> Logs request outcomes (both success and failure).</li>
        <li><strong>Slack Notifications:</strong> Sends notifications on success or failure.</li>
        <li><strong>Syslog Integration:</strong> Logs messages to syslog and console.</li>
        <li><strong>Custom Logging:</strong> Uses a custom log formatter to simplify log output.</li>
    </ul>

    <h2>Table of Contents</h2>
    <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
        <li><a href="#configuration">Configuration</a></li>
        <li><a href="#how-to-get-bearer-token">How to Get Bearer Token</a></li>
        <li><a href="#usage">Usage</a></li>
        <li><a href="#project-structure">Project Structure</a></li>
        <li><a href="#cron-job-schedule">Cron Job Schedule</a></li>
        <li><a href="#logging">Logging</a></li>
        <li><a href="#troubleshooting">Troubleshooting</a></li>
        <li><a href="#license">License</a></li>
    </ul>

    <h2 id="prerequisites">Prerequisites</h2>
    <p>Ensure the following tools and dependencies are installed on your system:</p>
    <ul>
        <li><strong>Go 1.18+</strong> installed and configured.</li>
        <li><strong>Access to Star Trek Fleet Command API</strong> with a valid bearer token.</li>
        <li><strong>Slack Webhook URL</strong> for sending notifications.</li>
        <li>A <em>nix-based environment</em> (for syslog integration).</li>
    </ul>

    <h2 id="installation">Installation</h2>
    <ol>
        <li><strong>Clone the repository:</strong>
            <pre><code>git clone &lt;repository-url&gt;
cd &lt;repository-folder&gt;</code></pre>
        </li>
        <li><strong>Install dependencies:</strong>
            <pre><code>go get github.com/robfig/cron/v3
go get github.com/sirupsen/logrus</code></pre>
        </li>
        <li><strong>Compile the application:</strong>
            <pre><code>go build -o stfc-automation</code></pre>
        </li>
    </ol>

    <h2 id="configuration">Configuration</h2>
    <p>Create a <code>config.json</code> file in the project root with the following structure:</p>
    <pre><code>{
  "bearerToken": "&lt;your_bearer_token_here&gt;",
  "bundleId10m": 1786571320,
  "bundleId4h": 844758222,
  "bundleId24h": 1918154038,
  "DailyMissionKey": 787829412,
  "OpticalDiode": 1579845062,
  "ReplicatorRations": 1210188306,
  "TrailBells": 718968170,
  "NadionSupply": 1904351560,
  "TranswarpCell": 1438866306,
  "slackWebhookURL": "&lt;your_slack_webhook_url_here&gt;"
}</code></pre>

    <h2 id="how-to-get-bearer-token">How to Get Bearer Token</h2>
    <ol>
        <li>Go to the <a href="https://home.startrekfleetcommand.com/" target="_blank">Star Trek Fleet Command website</a>.</li>
        <li>Enter your email and password to log in.</li>
        <li>Open the browser's <strong>Developer Tools</strong> (usually by pressing <code>F12</code> or right-clicking and selecting <em>Inspect</em>).</li>
        <li>Navigate to the <strong>Network</strong> tab in Developer Tools.</li>
        <li>Submit the login form on the website.</li>
        <li>Look for a <strong>login</strong> request in the list of network requests.</li>
        <li>Click on the <strong>login</strong> request, and in the <strong>Response</strong> section, find the <code>access_token</code>.</li>
        <li>Use the <code>access_token</code> as the <strong>bearerToken</strong> value in your <code>config.json</code>.</li>
    </ol>

    <h2 id="usage">Usage</h2>
    <ol>
        <li><strong>Run the service:</strong>
            <pre><code>./stfc-automation</code></pre>
            <p>Output:</p>
            <pre><code>Engines to maximum, we're ready for launch</code></pre>
        </li>
        <li>The program will run indefinitely, executing scheduled jobs as per the cron configuration.</li>
    </ol>

    <h2 id="project-structure">Project Structure</h2>
    <pre><code>.
├── main.go          # Main application code
├── config.json      # Configuration file (user-defined)
└── README.md        # Project documentation</code></pre>

    <h2 id="cron-job-schedule">Cron Job Schedule</h2>
    <ul>
        <li><strong>Every 10 minutes + 30 seconds:</strong> Claims gift with <code>BundleId10m</code>.</li>
        <li><strong>Every 4 hours + 30 seconds:</strong> Claims gift with <code>BundleId4h</code>.</li>
        <li><strong>Every day at 10:00:30 AM:</strong> Claims all daily gifts.</li>
    </ul>

    <h3>Cron Expressions</h3>
    <ul>
        <li><strong>Every 10 minutes:</strong>
            <pre><code>30 */10 * * * *</code></pre>
        </li>
        <li><strong>Every 4 hours:</strong>
            <pre><code>30 0 */4 * * *</code></pre>
        </li>
        <li><strong>Daily at 10:00:30 AM:</strong>
            <pre><code>30 00 10 * * *</code></pre>
        </li>
    </ul>

    <h2 id="logging">Logging</h2>
    <p>The service uses <strong>Logrus</strong> for logging and sends output to both:</p>
    <ul>
        <li><strong>Syslog:</strong> Logs the information using the <code>syslog</code> package.</li>
        <li><strong>Console:</strong> Outputs logs to <code>stdout</code>.</li>
    </ul>

    <h3>Custom Formatter</h3>
    <pre><code>type CustomFormatter struct{}
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
    return []byte(entry.Message + "\n"), nil
}</code></pre>

    <h2 id="slack-notifications">Slack Notifications</h2>
    <p>
        The service sends a <strong>Slack notification</strong> whenever a request succeeds or fails.
    </p>
    <h3>Failure Messages Example:</h3>
    <ul>
        <li><strong>10 Minutes Chest:</strong> ❌ 10 Minutes Chest Failed</li>
        <li><strong>4 Hours Chest:</strong> ❌ 4 Hours Chest Failed</li>
    </ul>

    <h3>Example Slack Payload:</h3>
    <pre><code>{
  "text": "STFC Automation Error: ❌ 10 Minutes Chest Failed"
}</code></pre>

    <h2 id="troubleshooting">Troubleshooting</h2>
    <ul>
        <li><strong>Cannot connect to syslog:</strong> Ensure syslog is configured on your system and running correctly.</li>
        <li><strong>Invalid Bearer Token:</strong> Double-check the <code>bearerToken</code> in <code>config.json</code> and ensure it hasn’t expired.</li>
        <li><strong>Slack notifications not sent:</strong> Verify the <code>slackWebhookURL</code> in <code>config.json</code>. Check network connectivity to Slack's API.</li>
    </ul>

    <h2 id="license">License</h2>
    <p>This project is licensed under the MIT License. See the LICENSE file for more details.</p>

    <p>Enjoy automating your Star Trek Fleet Command rewards collection! 🚀</p>
</body>
</html>
