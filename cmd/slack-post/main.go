package main

import (
	"context"
	"fmt"
	"os"

	"github.com/berrydev-ai/agent-cli-tools/internal/cli/slackpost"
)

var version = "dev"

func main() {
	if err := slackpost.Execute(context.Background(), slackpost.Options{Version: version}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
