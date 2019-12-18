package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// buildLogflightsCmd represents the buildLogflights command
var buildLogflightsCmd = &cobra.Command{
	Use:   "logflights",
    Aliases: []string{"app", "lf"},
	Short: "Builds Logflights Application Components.",
	Long: `Builds Logflights Application Components.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		fmt.Println("Building Logflight Application Components")
		return buildLogflightsComponents()
	},
}

func buildLogflightsComponents() (err error) {
	cmd, err := GenerateCommand(dockerBuildTemplate, DockerBuildRequest{
		Dockerfile: "./containers/logflights/Dockerfile",
		Image: "logflights",
		Tag: "latest",
		Context: "./containers/logflights",
		NoCache: noCache,
	})

	if err != nil {
		return err
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}


func init() {
	buildCmd.AddCommand(buildLogflightsCmd)
}
