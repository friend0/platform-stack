package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// buildFrontendCmd represents the buildFrontend command
var buildFrontendCmd = &cobra.Command{
	Use:   "frontend",
    Aliases: []string{"fe"},
	Short: "Builds Frontend Components.",
	Long: `Builds frontend Components.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		fmt.Println("Building Frontend Components")
		return buildFrontendComponents()
	},
}

func buildFrontendComponents() (err error) {
	cmd, err := GenerateCommand(dockerBuildTemplate, DockerBuildRequest{
		Dockerfile: "./containers/frontend/Dockerfile",
		Image: "frontend",
		Tag: "latest",
		Context: "./containers/frontend",
		NoCache: noCache,
	})

	if err != nil {
		return err
	}

	//cmd.Env = append(os.Environ())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func init() {
	buildCmd.AddCommand(buildFrontendCmd)
}
