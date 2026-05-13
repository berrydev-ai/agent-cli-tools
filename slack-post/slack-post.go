package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type SlackMessagePayload struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

type SlackAPIResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

func main() {
	var (
		token          string
		channel        string
		target         string
		targetMemberID string
		prompt         string
		apiURL         string
	)

	flag.StringVar(&token, "token", "", "Slack bot token. Defaults to SLACK_E2E_BOT_TOKEN.")
	flag.StringVar(&channel, "channel", "", "Slack channel ID. Defaults to SLACK_E2E_TARGET_CHANNEL.")
	flag.StringVar(&target, "target", "", "Target name used to resolve SLACK_E2E_<TARGET>_BOT_MEMBER_ID.")
	flag.StringVar(&targetMemberID, "target-member-id", "", "Slack member ID to mention, for example U123ABC456.")
	flag.StringVar(&targetMemberID, "member-id", "", "Alias for --target-member-id.")
	flag.StringVar(&prompt, "prompt", "", "Prompt/message text to send.")
	flag.StringVar(&apiURL, "url", "https://slack.com/api/chat.postMessage", "Slack API URL.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage:
  slack-post --prompt "your message"

Examples:
  slack-post \
    --token "$SLACK_E2E_BOT_TOKEN" \
    --channel "$SLACK_E2E_TARGET_CHANNEL" \
    --target-member-id "$SLACK_E2E_TARGET_BOT_MEMBER_ID" \
    --prompt "run the e2e test"

  slack-post --target claude --prompt "run the e2e test"

If --target claude is provided, this command will look for:
  SLACK_E2E_CLAUDE_BOT_MEMBER_ID

Flags:
`)
		flag.PrintDefaults()
	}

	flag.Parse()

	token = firstNonEmpty(token, os.Getenv("SLACK_E2E_BOT_TOKEN"))
	channel = firstNonEmpty(channel, os.Getenv("SLACK_E2E_TARGET_CHANNEL"))

	if prompt == "" && flag.NArg() > 0 {
		prompt = strings.Join(flag.Args(), " ")
	}

	var targetEnvName string
	if targetMemberID == "" && target != "" {
		targetEnvName = "SLACK_E2E_" + normalizeEnvPart(target) + "_BOT_MEMBER_ID"
		targetMemberID = os.Getenv(targetEnvName)
	}

	if targetMemberID == "" {
		targetMemberID = os.Getenv("SLACK_E2E_TARGET_BOT_MEMBER_ID")
	}

	var missing []string

	if token == "" {
		missing = append(missing, "--token or SLACK_E2E_BOT_TOKEN")
	}

	if channel == "" {
		missing = append(missing, "--channel or SLACK_E2E_TARGET_CHANNEL")
	}

	if prompt == "" {
		missing = append(missing, "--prompt or positional message text")
	}

	if target != "" && targetMemberID == "" {
		missing = append(missing, "--target-member-id or "+targetEnvName)
	}

	if len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "Missing required value(s): %s\n\n", strings.Join(missing, ", "))
		flag.Usage()
		os.Exit(2)
	}

	text := prompt
	if targetMemberID != "" {
		text = fmt.Sprintf("<@%s> %s", targetMemberID, prompt)
	}

	payload := SlackMessagePayload{
		Channel: channel,
		Text:    text,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON payload: %s\n", err)
		os.Exit(1)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %s\n", err)
		os.Exit(1)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error making request: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response body: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Response Status:", resp.Status)
	fmt.Println("Response Body:", string(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		os.Exit(1)
	}

	var slackResp SlackAPIResponse
	if err := json.Unmarshal(respBody, &slackResp); err == nil {
		if !slackResp.OK {
			if slackResp.Error != "" {
				fmt.Fprintf(os.Stderr, "Slack API error: %s\n", slackResp.Error)
			} else {
				fmt.Fprintln(os.Stderr, "Slack API returned ok=false")
			}
			os.Exit(1)
		}
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func normalizeEnvPart(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))

	var b strings.Builder

	for _, r := range value {
		if r >= 'A' && r <= 'Z' {
			b.WriteRune(r)
		} else if r >= '0' && r <= '9' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}

	return b.String()
}