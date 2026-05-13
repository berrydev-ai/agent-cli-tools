package slackpost

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const defaultAPIURL = "https://slack.com/api/chat.postMessage"

type Options struct {
	Out        io.Writer
	Err        io.Writer
	HTTPClient *http.Client
	Env        func(string) string
	Version    string
}

type messagePayload struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

type slackAPIResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type postResult struct {
	Status     string          `json:"status"`
	StatusCode int             `json:"status_code"`
	OK         bool            `json:"ok"`
	Response   json.RawMessage `json:"response,omitempty"`
}

func Execute(ctx context.Context, opts Options) error {
	return NewCommand(opts).ExecuteContext(ctx)
}

func NewCommand(opts Options) *cobra.Command {
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	if opts.Err == nil {
		opts.Err = os.Stderr
	}
	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{Timeout: 15 * time.Second}
	}
	if opts.Env == nil {
		opts.Env = os.Getenv
	}

	var (
		token          string
		channel        string
		target         string
		targetMemberID string
		prompt         string
		apiURL         string
		format         string
	)

	cmd := &cobra.Command{
		Use:           "slack-post [message]",
		Short:         "Post a message to Slack",
		Version:       opts.Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ArbitraryArgs,
		Example: strings.TrimSpace(`
slack-post --prompt "run the e2e test"
slack-post --target claude --prompt "run the e2e test"
slack-post --format json --target-member-id "$SLACK_E2E_TARGET_BOT_MEMBER_ID" "run the e2e test"`),
		RunE: func(cmd *cobra.Command, args []string) error {
			token = firstNonEmpty(token, opts.Env("SLACK_E2E_BOT_TOKEN"))
			channel = firstNonEmpty(channel, opts.Env("SLACK_E2E_TARGET_CHANNEL"))

			if prompt == "" && len(args) > 0 {
				prompt = strings.Join(args, " ")
			}

			var targetEnvName string
			if targetMemberID == "" && target != "" {
				targetEnvName = "SLACK_E2E_" + normalizeEnvPart(target) + "_BOT_MEMBER_ID"
				targetMemberID = opts.Env(targetEnvName)
			}
			if targetMemberID == "" {
				targetMemberID = opts.Env("SLACK_E2E_TARGET_BOT_MEMBER_ID")
			}

			if format != "text" && format != "json" {
				return fmt.Errorf("--format must be text or json")
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
				return fmt.Errorf("missing required value(s): %s", strings.Join(missing, ", "))
			}

			text := prompt
			if targetMemberID != "" {
				text = fmt.Sprintf("<@%s> %s", targetMemberID, prompt)
			}

			result, err := postMessage(cmd.Context(), opts.HTTPClient, apiURL, token, channel, text)
			if err != nil {
				return err
			}

			return writeResult(cmd.OutOrStdout(), format, result)
		},
	}

	cmd.SetOut(opts.Out)
	cmd.SetErr(opts.Err)
	cmd.Flags().StringVar(&token, "token", "", "Slack bot token. Defaults to SLACK_E2E_BOT_TOKEN.")
	cmd.Flags().StringVar(&channel, "channel", "", "Slack channel ID. Defaults to SLACK_E2E_TARGET_CHANNEL.")
	cmd.Flags().StringVar(&target, "target", "", "Target name used to resolve SLACK_E2E_<TARGET>_BOT_MEMBER_ID.")
	cmd.Flags().StringVar(&targetMemberID, "target-member-id", "", "Slack member ID to mention, for example U123ABC456.")
	cmd.Flags().StringVar(&targetMemberID, "member-id", "", "Alias for --target-member-id.")
	cmd.Flags().StringVar(&prompt, "prompt", "", "Prompt/message text to send.")
	cmd.Flags().StringVar(&apiURL, "url", defaultAPIURL, "Slack API URL.")
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text or json.")

	return cmd
}

func postMessage(ctx context.Context, client *http.Client, apiURL string, token string, channel string, text string) (postResult, error) {
	payload := messagePayload{
		Channel: channel,
		Text:    text,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return postResult{}, fmt.Errorf("encode JSON payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return postResult{}, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		return postResult{}, fmt.Errorf("post Slack message: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return postResult{}, fmt.Errorf("read response body: %w", err)
	}

	result := postResult{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Response:   json.RawMessage(respBody),
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return result, fmt.Errorf("Slack API HTTP error: %s", resp.Status)
	}

	var slackResp slackAPIResponse
	if err := json.Unmarshal(respBody, &slackResp); err != nil {
		return result, fmt.Errorf("decode Slack API response: %w", err)
	}
	result.OK = slackResp.OK
	if !slackResp.OK {
		if slackResp.Error != "" {
			return result, fmt.Errorf("Slack API error: %s", slackResp.Error)
		}
		return result, fmt.Errorf("Slack API returned ok=false")
	}

	return result, nil
}

func writeResult(out io.Writer, format string, result postResult) error {
	if format == "json" {
		encoder := json.NewEncoder(out)
		return encoder.Encode(result)
	}

	_, err := fmt.Fprintf(out, "Response Status: %s\nResponse Body: %s\n", result.Status, string(result.Response))
	return err
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
