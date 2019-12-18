package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// buildDatadogCmd represents the buildDatadog command
var buildDatadogCmd = &cobra.Command{
	Use:   "datadog",
    Aliases: []string{"dog", "dd"},
	Short: "Builds Datadog Components.",
	Long: `Builds Datadog Components.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		fmt.Println("Building Datadog Components")
		return buildDatadogComponents()
	},
}

func buildDatadogComponents() (err error) {
	fmt.Println("Datadog refers to an image. Skipping.")
	fmt.Println("Build Datadog Succeeded")
	return nil
}

func init() {
	buildCmd.AddCommand(buildDatadogCmd)
}
