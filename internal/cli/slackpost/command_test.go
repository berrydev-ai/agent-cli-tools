package slackpost

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestCommandPostsPositionalPromptWithTargetEnv(t *testing.T) {
	var gotAuth string
	var gotPayload messagePayload
	var decodeErr atomic.Value
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			decodeErr.Store(err)
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"channel":"C123","ts":"123.456"}`))
	}))
	defer server.Close()

	out, _, err := execute(t, map[string]string{
		"SLACK_E2E_BOT_TOKEN":            "xoxb-test",
		"SLACK_E2E_TARGET_CHANNEL":       "C123",
		"SLACK_E2E_CLAUDE_BOT_MEMBER_ID": "UCLAUDE",
	}, "--url", server.URL, "--target", "claude", "--format", "json", "run", "the", "e2e", "test")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if value := decodeErr.Load(); value != nil {
		t.Fatalf("decode request body: %v", value)
	}

	if gotAuth != "Bearer xoxb-test" {
		t.Fatalf("authorization header = %q", gotAuth)
	}
	if gotPayload.Channel != "C123" {
		t.Fatalf("channel = %q", gotPayload.Channel)
	}
	if gotPayload.Text != "<@UCLAUDE> run the e2e test" {
		t.Fatalf("text = %q", gotPayload.Text)
	}

	var result postResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("decode command output: %v\noutput: %s", err, out)
	}
	if !result.OK || result.StatusCode != http.StatusOK {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestCommandRequiresNoninteractiveInputs(t *testing.T) {
	_, _, err := execute(t, nil, "--prompt", "hello")
	if err == nil {
		t.Fatal("expected missing input error")
	}

	errText := err.Error()
	for _, want := range []string{"--token or SLACK_E2E_BOT_TOKEN", "--channel or SLACK_E2E_TARGET_CHANNEL"} {
		if !strings.Contains(errText, want) {
			t.Fatalf("error %q does not contain %q", errText, want)
		}
	}
}

func TestCommandReportsSlackAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":false,"error":"channel_not_found"}`))
	}))
	defer server.Close()

	_, _, err := execute(t, map[string]string{
		"SLACK_E2E_BOT_TOKEN":      "xoxb-test",
		"SLACK_E2E_TARGET_CHANNEL": "C123",
	}, "--url", server.URL, "--prompt", "hello")
	if err == nil {
		t.Fatal("expected Slack API error")
	}
	if !strings.Contains(err.Error(), "Slack API error: channel_not_found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func execute(t *testing.T, env map[string]string, args ...string) (string, string, error) {
	t.Helper()

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewCommand(Options{
		Out: &out,
		Err: &errOut,
		Env: func(key string) string {
			return env[key]
		},
	})
	cmd.SetArgs(args)
	err := cmd.ExecuteContext(context.Background())
	return out.String(), errOut.String(), err
}
