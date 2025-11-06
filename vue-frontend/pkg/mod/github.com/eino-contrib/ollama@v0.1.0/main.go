package main

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/eino-contrib/ollama/cmd"
)

func main() {
	cobra.CheckErr(cmd.NewCLI().ExecuteContext(context.Background()))
}
