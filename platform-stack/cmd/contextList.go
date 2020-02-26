package cmd

import (
	"github.com/spf13/cobra"
)

// contextListCmd represents the contextList command
var contextListCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: runContextCommandFunction("get-contexts"),
}

func init() {
	contextCmd.AddCommand(contextListCmd)
}
