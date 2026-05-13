---
name: release-agent-binaries
description: Version, package, tag, push, and verify standalone Go CLI binaries in the agent-cli-tools repo. Use when the user asks to release, deploy, publish, version, bump, package, or install one of the binaries under cmd/*, especially with per-binary SemVer tags such as slack-post-v0.1.0 and Makefile targets.
---

# Release Agent Binaries

## Overview

Release each binary independently from this repo. Use per-binary SemVer tags, not repo-wide `v*` tags:

```text
<tool>-v<major>.<minor>.<patch>
slack-post-v0.1.0
```

Use the Makefile as the source of truth for validation, version stamping, archive naming, tag creation, and tag push commands.

## Workflow

1. Identify the binary:

```bash
make list
test -d "cmd/<tool>"
```

If the requested tool is not listed, stop and tell the user which binaries are available.

2. Determine the version:

- If the user gives an exact version, require a leading `v`, for example `v0.1.0`.
- If the user asks for `major`, `minor`, or `patch`, inspect existing tags for that binary and calculate the next SemVer.
- If there is no existing tag for that binary, default the first release to `v0.1.0` unless the user requested a different version.

Useful commands:

```bash
git fetch --tags
git tag -l "<tool>-v*" --sort=-version:refname
make print-tag TOOL=<tool> VERSION=<version>
```

3. Confirm the release points at the intended code:

```bash
git status --short --branch
git log --oneline -5
```

Do not create a release tag on stale or unintended code. If there are uncommitted changes that should be part of the release, finish and verify them first. If unrelated local changes exist, leave them alone and explain the release boundary.

4. Verify before tagging:

```bash
make test
make build-tool TOOL=<tool> VERSION=<version>
dist/<tool> --version
make package TOOL=<tool> VERSION=<version> GOOS=linux GOARCH=amd64
tar -tzf dist/<tool>_<version>_linux_amd64.tar.gz
```

Expected archive contents: one executable named exactly `<tool>`.

5. Create and push the per-binary tag when the user has asked to release or deploy:

```bash
make tag TOOL=<tool> VERSION=<version>
make push-tag TOOL=<tool> VERSION=<version>
```

The pushed tag should trigger `.github/workflows/release.yml`, which builds that one binary for Linux and macOS on `amd64` and `arm64`.

6. Verify the GitHub release:

```bash
gh release view "<tool>-<version>" --json tagName,name,assets,url
gh run list --workflow release.yml --limit 5
```

Confirm the release contains these assets:

```text
<tool>_<version>_linux_amd64.tar.gz
<tool>_<version>_linux_arm64.tar.gz
<tool>_<version>_darwin_amd64.tar.gz
<tool>_<version>_darwin_arm64.tar.gz
```

## Makefile Commands

Use these targets instead of hand-written build/tag commands:

```bash
make help
make list
make test
make build
make build-tool TOOL=slack-post VERSION=v0.1.0
make package TOOL=slack-post VERSION=v0.1.0 GOOS=linux GOARCH=amd64
make release-archives TOOL=slack-post VERSION=v0.1.0
make print-tag TOOL=slack-post VERSION=v0.1.0
make tag TOOL=slack-post VERSION=v0.1.0
make push-tag TOOL=slack-post VERSION=v0.1.0
```

`make package` and `make release-archives` require valid SemVer with a leading `v`. Bad examples such as `0.1.0` should fail before any archive or tag is created.

## Docker Install Updates

When the user asks to update a Docker install command, use this private-release URL shape:

```dockerfile
ARG AGENT_CLI_TOOLS_VERSION=v0.1.0
ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN --mount=type=secret,id=github_token \
  GITHUB_TOKEN="$(cat /run/secrets/github_token)" \
 && curl -fsSL -H "Authorization: Bearer ${GITHUB_TOKEN}" -L \
    "https://github.com/berrydev-ai/agent-cli-tools/releases/download/slack-post-${AGENT_CLI_TOOLS_VERSION}/slack-post_${AGENT_CLI_TOOLS_VERSION}_${TARGETOS}_${TARGETARCH}.tar.gz" \
  | tar -xz -C /usr/local/bin slack-post \
 && chmod +x /usr/local/bin/slack-post
```

Replace `slack-post` with the requested tool name.

## Response Checklist

When finished, report:

- Tool and version released.
- Tag name.
- Verification commands that passed.
- Release URL or the exact blocker if GitHub release publication failed.
