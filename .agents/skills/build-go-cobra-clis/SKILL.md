---
name: build-go-cobra-clis
description: Use when building, extending, testing, or packaging Go command-line tools with Cobra, especially AI-agent-facing CLIs that may expose many tools as subcommands or individually exportable binaries.
---

# Build Go Cobra CLIs

## Overview

Build Cobra CLIs as stable tool contracts for agents first, and human-friendly terminals second. Choose one suite binary with subcommands when agents benefit from one installed tool; split binaries only when distribution, permissions, dependencies, or invocation ergonomics justify it.

## Pick the Export Shape

Prefer a single agent-facing suite binary when:

- The tools are usually installed together.
- Agent tool configuration is simpler with one executable.
- Shared auth, config, output conventions, and logging matter.
- Commands form a coherent namespace, such as `agent-tools github issue create`.

Prefer separate binaries when:

- Tools have unrelated dependencies, credentials, or release cadence.
- One tool should be safely exposed without the others.
- Names are short and ergonomic as standalone commands.
- A target agent runtime can export individual executables more cleanly than subcommands.

It is acceptable to support both shapes: one `cmd/agent-tools` suite binary and selected `cmd/<tool>` wrapper binaries that call the same internal command constructors.

## Default Layout

Use this shape unless the repo already has a stronger convention:

```text
go.mod
cmd/
  agent-tools/
    main.go
  github-issue-create/
    main.go
internal/
  cli/
    suite/
      root.go
    github/
      root.go
      issue.go
      issue_create.go
      issue_create_test.go
    shared/
      output.go
      errors.go
dist/
```

Rules:

- Keep `cmd/<binary>/main.go` limited to context setup, command execution, error printing, and process exit.
- Put Cobra constructors and behavior under `internal/cli/...`.
- Let standalone binaries wrap the same command package used by the suite command.
- Keep reusable non-CLI behavior outside Cobra packages so it can be tested without CLI wiring.
- Use `pkg/` only when external import is intentional.

## Agent Contract

Design every command so an AI agent can call it predictably:

- Provide noninteractive flags for every required input. Do not prompt by default.
- Add `--format json` for structured output when results may be parsed by an agent.
- Send machine-readable results to stdout and diagnostics to stderr.
- Use stable exit codes: `0` success, nonzero for failure, and consistent error text.
- Disable color, spinners, pagers, and TTY-only behavior unless explicitly requested.
- Keep command paths explicit and resource/action oriented: `issue create`, `repo search`, `gmail thread-summary`.
- Make `--help` complete enough for an agent to infer required flags, examples, and output format.

## Workflow

1. Define the tool contract: command path, flags, args, stdout schema, stderr behavior, exit behavior, and credential/config inputs.
2. Write a failing Go test for the command behavior before adding implementation.
3. Add command constructors that return `*cobra.Command`; avoid global command state and `init()` registration in new code.
4. Use `RunE`, `Args`, `cmd.Context()`, and returned errors. Reserve `os.Exit` for `main.go`.
5. Add local flags by default. Use persistent flags only for concerns inherited by every child command.
6. Inject I/O with `cmd.SetOut`, `cmd.SetErr`, and constructor options. Avoid `fmt.Println` in command logic.
7. Verify with tests, a binary build, `--help`, and at least one realistic command invocation.

## Command Pattern

This single-package sketch shows the constructor style. In implementation, split feature constructors into packages matching the repo layout when that keeps dependencies cleaner.

```go
package suite

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

type Options struct {
	Out     io.Writer
	Err     io.Writer
	Version string
}

func NewRootCommand(opts Options) *cobra.Command {
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	if opts.Err == nil {
		opts.Err = os.Stderr
	}

	cmd := &cobra.Command{
		Use:           "agent-tools",
		Short:         "Agent-oriented command-line tools",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.SetOut(opts.Out)
	cmd.SetErr(opts.Err)
	cmd.AddCommand(NewGithubCommand(opts))
	return cmd
}

func Execute(ctx context.Context, opts Options) error {
	return NewRootCommand(opts).ExecuteContext(ctx)
}

func NewGithubCommand(opts Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "github",
		Short: "Work with GitHub",
	}
	cmd.AddCommand(newGithubIssueCommand(opts))
	return cmd
}

func newGithubIssueCommand(opts Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Work with GitHub issues",
	}
	cmd.AddCommand(NewGithubIssueCreateCommand(opts))
	return cmd
}

func NewGithubIssueCreateCommand(opts Options) *cobra.Command {
	var title string
	var format string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a GitHub issue",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if title == "" {
				return fmt.Errorf("--title is required")
			}
			if format == "json" {
				_, err := fmt.Fprintf(cmd.OutOrStdout(), `{"title":%q}`+"\n", title)
				return err
			}
			_, err := fmt.Fprintln(cmd.OutOrStdout(), title)
			return err
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "issue title")
	cmd.Flags().StringVar(&format, "format", "text", "output format: text or json")
	return cmd
}
```

`cmd/agent-tools/main.go` should stay thin:

```go
package main

import (
	"context"
	"fmt"
	"os"

	"example.com/agent-cli-tools/internal/cli/suite"
)

func main() {
	if err := suite.Execute(context.Background(), suite.Options{}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

A standalone wrapper can reuse the same command constructor instead of duplicating behavior. If needed, set the wrapper command's `Use` to the standalone binary name before executing it.

## Tests

Test constructors directly. Do not shell out for ordinary command behavior tests.

```go
func execute(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	var out, errOut bytes.Buffer
	cmd := NewRootCommand(Options{Out: &out, Err: &errOut})
	cmd.SetArgs(args)
	err := cmd.ExecuteContext(context.Background())
	return out.String(), errOut.String(), err
}
```

Add tests for successful output, JSON output, invalid args, required flags, error text, and help text when it is part of the public contract. Use an actual built binary only for smoke tests and release packaging checks.

## Packaging

Build the suite binary:

```bash
go build -o dist/agent-tools ./cmd/agent-tools
```

Build selected standalone exports:

```bash
go build -o dist/github-issue-create ./cmd/github-issue-create
```

For all binaries:

```bash
mkdir -p dist
for dir in ./cmd/*; do
  tool="${dir##*/}"
  go build -o "dist/${tool}" "${dir}"
done
```

When using `cobra-cli`, use it for quick bootstrapping only. In this repo shape, prefer hand-written constructors over generated global `cmd` package state once behavior is added.

## Completion Checklist

- `go test ./...`
- `go build -o dist/<binary> ./cmd/<binary>`
- `./dist/<binary> --help`
- `./dist/<binary> <representative command>` with expected stdout, stderr, and exit behavior
- JSON output validates with `jq` when a command supports `--format json`

## Official References

- https://cobra.dev/docs/tutorials/getting-started/
- https://cobra.dev/docs/how-to-guides/working-with-commands/
- https://cobra.dev/docs/how-to-guides/working-with-flags/
- https://cobra.dev/docs/how-to-guides/shell-completion/
- https://github.com/spf13/cobra
