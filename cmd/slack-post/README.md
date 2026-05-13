# slack-post

`slack-post` posts a message to Slack using `chat.postMessage`.

## Usage

```bash
slack-post --prompt "run the e2e test"
slack-post --target claude --prompt "run the e2e test"
slack-post --format json --target-member-id "$SLACK_E2E_TARGET_BOT_MEMBER_ID" "run the e2e test"
```

## Required Inputs

Provide these values as flags or environment variables:

| Flag | Environment variable | Description |
| --- | --- | --- |
| `--token` | `SLACK_E2E_BOT_TOKEN` | Slack bot token. |
| `--channel` | `SLACK_E2E_TARGET_CHANNEL` | Slack channel ID. |
| `--prompt` | none | Message text. Positional text is also accepted. |

Optional mention target:

| Flag | Environment variable | Description |
| --- | --- | --- |
| `--target-member-id`, `--member-id` | `SLACK_E2E_TARGET_BOT_MEMBER_ID` | Slack member ID to mention. |
| `--target <name>` | `SLACK_E2E_<TARGET>_BOT_MEMBER_ID` | Resolves a named target, such as `SLACK_E2E_CLAUDE_BOT_MEMBER_ID`. |

See the local `.env.example` for the maintained environment template.

## Output

Default text output:

```text
Response Status: 200 OK
Response Body: {"ok":true,...}
```

Agent-readable output:

```bash
slack-post --format json --prompt "hello"
```

## Build

```bash
make build-tool TOOL=slack-post VERSION=v0.1.0
dist/slack-post --version
```

## Package

```bash
make package TOOL=slack-post VERSION=v0.1.0 GOOS=linux GOARCH=amd64
tar -tzf dist/slack-post_v0.1.0_linux_amd64.tar.gz
```
