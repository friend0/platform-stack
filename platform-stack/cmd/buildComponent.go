package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var context string
var dockerfile string

// buildLogflightsCmd represents the buildLogflights command
var buildComponentsCommand = &cobra.Command{
	Use:   "component",
	Aliases: []string{"app", "lf"},
	Args: cobra.RangeArgs(1, 2),
	Short: "Builds the component with the context at the current working directory.",
	Long: `Builds the component with the context at the current working directory.

Navigate to a directory with a Dockerfile, and build the image. `,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		fmt.Println("Building Logflight Application Components")
		return buildComponent(args)
	},
}

// stack build _component_ v0.0.1

func buildComponent(args []string) (err error) {

	image := args[0]
	tag := "latest"
	if len(args) == 2 {
		tag = args[1]
	}

	cmd, err := GenerateCommand(dockerBuildTemplate, DockerBuildRequest{
		Dockerfile: strings.Join([]string{context, dockerfile}, "/"),
		Image: image,
		Tag: tag,
		Context: context,
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
	buildCmd.AddCommand(buildComponentsCommand)

	buildComponentsCommand.Flags().StringVarP(&context, "context", "c",".","Select Docker build context")
	buildComponentsCommand.Flags().StringVarP(&dockerfile, "dockerfile", "d","Dockerfile","Select name of Dockerfile")

}
