# Agent CLI Tools

Agent CLI Tools is a home for small Go CLIs that should be easy for AI agents to install and call as executable tools.

## Shape

Use one repository and one Go module, but export each tool as its own binary under `cmd/<binary-name>`.

```text
cmd/
  slack-post/
    .env.example
    README.md
    main.go
internal/
  cli/
    slackpost/
      command.go
      command_test.go
```

This keeps unrelated tools from sharing a command namespace while still letting the repo share test, build, release, and helper code. If a future agent runtime benefits from one suite command, add a thin `cmd/agent-cli-tools` wrapper that reuses the same internal command constructors.

## Add a CLI

1. Add a new `cmd/<binary-name>/main.go`.
2. Put Cobra command behavior under `internal/cli/<tool>/`.
3. Keep required inputs available as flags and environment variables.
4. Support `--format json` when output may be consumed by an agent.
5. Add command tests under the same internal package.
6. Define `var version = "dev"` in the `main` package so release builds can stamp the tag into `--version`.
7. Add `cmd/<binary-name>/README.md` with usage, inputs, output, build, and package notes.
8. Add or update `cmd/<binary-name>/.env.example` for every environment variable the tool reads.

## Build

```bash
make build
```

Binaries are written to `dist/`.

## Test

```bash
make test
```

## Project Docs

- [Changelog](CHANGELOG.md)
- [Contributing](CONTRIBUTING.md)
- [License](LICENSE)

## Tool Docs

Each binary owns a README in its command path:

- `cmd/slack-post/README.md`

Each binary also owns its environment template next to that README:

- `cmd/slack-post/.env.example`

Leave secret values blank.

## Per-Binary SemVer

Each binary gets its own SemVer tag:

```text
slack-post-v0.1.0
```

The Makefile validates tool names and versions, stamps the version into the binary, and creates release archives with stable names.

```bash
make print-tag TOOL=slack-post VERSION=v0.1.0
make release-archives TOOL=slack-post VERSION=v0.1.0
make tag TOOL=slack-post VERSION=v0.1.0
make push-tag TOOL=slack-post VERSION=v0.1.0
```

Pushing a `*-v*` tag that matches a known `cmd/<tool>` binary runs `.github/workflows/release.yml`, builds that one binary for Linux and macOS on `amd64` and `arm64`, and publishes tarballs to the GitHub release.

Release notes are generated automatically from merged pull requests and commits, categorized by [.github/release.yml](.github/release.yml), and compared against the previous tag for the same binary.

For example, a tag `slack-post-v0.1.0` produces:

```text
slack-post_v0.1.0_linux_amd64.tar.gz
slack-post_v0.1.0_linux_arm64.tar.gz
slack-post_v0.1.0_darwin_amd64.tar.gz
slack-post_v0.1.0_darwin_arm64.tar.gz
```

## Docker Install Example

```dockerfile
# syntax=docker/dockerfile:1.7

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

Build with a token that can read this private repository:

```bash
docker build --secret id=github_token,env=GITHUB_TOKEN .
```

## Tools

- `slack-post`: see `cmd/slack-post/README.md`
