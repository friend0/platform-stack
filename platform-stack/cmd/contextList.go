package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

// contextListCmd represents the contextList command
var contextListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available kubectx.",
	Long:  `List all available kubectx.`,
	RunE:  runContextCommandFunction("get-contexts", os.Stdout),
}

func init() {
	contextCmd.AddCommand(contextListCmd)
}
