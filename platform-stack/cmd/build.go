package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const dockerBuildTemplate = `docker build {{if .NoCache}} --no-cache {{end}} --build-arg GIT_TOKEN="$GIT_TOKEN" -t {{.Image}}:{{.Tag}} -f {{.Dockerfile}} {{.Context}}`

var tag string
var noCache bool

type DockerBuildRequest struct {
	Dockerfile string
	Image      string
	Tag        string
	Context    string
	NoCache    bool
}

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds all images for Logflights.",
	Long: `Build is used to generate or regenerate the images necessary to run the Logflights stack in any environment.

All components, or individual components, may be built and tagged using this comand.

Available Components: 
- database
- datadog
- frontend
- logflights
- proxy
- redis`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Please specify which components to build")
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.PersistentFlags().StringVarP(&tag, "tag", "t", "latest", "tag for images to be built")

	buildCmd.PersistentFlags().BoolVarP(&noCache, "noCache", "c", false, "build images without cache")

}
