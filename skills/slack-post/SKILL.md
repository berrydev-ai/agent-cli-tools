---
name: slack-post
description: Install and use the slack-post CLI from berrydev-ai/agent-cli-tools to send Slack chat.postMessage prompts from an AI agent. Use when the user asks to post, resend, smoke-test, or automate a Slack message with the slack-post binary, especially when targeting a channel, bot member ID, or named Slack bot target.
---

# Slack Post

## Overview

Use `slack-post` to send a noninteractive Slack `chat.postMessage` request from an agent workflow. Treat every invocation as externally visible: only send after the user explicitly asks for the message to be posted or approves the exact target and text.

## Skill Installation

Install this skill from the repository with the Skills CLI:

```bash
npx skills add berrydev-ai/agent-cli-tools --skill slack-post
```

List the available skills in the repository before installing:

```bash
npx skills add berrydev-ai/agent-cli-tools --list
```

## Binary Installation

Install the `slack-post` binary before trying to send a message. Prefer Homebrew for released tools:

```bash
brew install berrydev-ai/tap/slack-post
```

Use a release archive when Homebrew is not available:

```bash
VERSION=v0.1.0
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64) ARCH=amd64 ;;
  arm64|aarch64) ARCH=arm64 ;;
esac

curl -fsSL -H "Authorization: Bearer ${GITHUB_TOKEN}" -L \
  "https://github.com/berrydev-ai/agent-cli-tools/releases/download/slack-post-${VERSION}/slack-post_${VERSION}_${OS}_${ARCH}.tar.gz" \
  | tar -xz -C /usr/local/bin slack-post
```

Use a GitHub token that can read the repository if the release is private. If `/usr/local/bin` is not writable, extract to a user-owned directory on `PATH`. Homebrew installs the latest released binary; for unreleased source changes, build from a checkout.

Build from a checkout when changing or verifying the tool locally:

```bash
git clone git@github.com:berrydev-ai/agent-cli-tools.git
cd agent-cli-tools
make build-tool TOOL=slack-post VERSION=v0.1.0
install -m 0755 dist/slack-post /usr/local/bin/slack-post
```

Verify the binary:

```bash
slack-post --version
slack-post --help
```

## Required Configuration

Set required values with flags or environment variables:

| Flag | Environment variable | Purpose |
| --- | --- | --- |
| `--token` | `SLACK_POST_BOT_TOKEN` | Slack bot token for `chat.postMessage`. |
| `--channel` | `SLACK_POST_TARGET_CHANNEL` | Slack channel ID to post into. |
| `--prompt` or positional text | none | Message text to send. |

Optional mention targeting:

| Flag | Environment variable | Purpose |
| --- | --- | --- |
| `--target-member-id`, `--member-id` | `SLACK_POST_TARGET_BOT_MEMBER_ID` | Slack member ID to mention before the message. |
| `--target <name>` | `SLACK_POST_<TARGET>_BOT_MEMBER_ID` | Named target lookup, for example `--target claude` reads `SLACK_POST_CLAUDE_BOT_MEMBER_ID`. |

Never print or persist Slack tokens. Prefer channel IDs and member IDs over human-readable names so the command is deterministic.

## Usage

Post a simple message:

```bash
slack-post --prompt "run the e2e test"
```

Mention a configured target:

```bash
slack-post --target claude --prompt "run the e2e test"
```

Use agent-readable JSON output:

```bash
slack-post --format json --target-member-id "$SLACK_POST_TARGET_BOT_MEMBER_ID" "run the e2e test"
```

Pass all required values as flags when environment variables are not available:

```bash
slack-post \
  --token "<your-slack-bot-token>" \
  --channel "<channel-id>" \
  --prompt "hello from slack-post"
```

## Workflow

1. Confirm the user wants an actual Slack message sent.
2. Resolve the channel ID, token source, and optional target member ID without exposing secrets.
3. Run `slack-post --help` if the local binary version or flags are uncertain.
4. Send the message with `--format json` when the result will be parsed by an agent.
5. Report the Slack API status, `ok` value, and any actionable error text.
