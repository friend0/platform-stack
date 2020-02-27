package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

const dockerBuildTemplate = `docker build {{if .NoCache}} --no-cache {{end}} --build-arg GIT_TOKEN="$GIT_TOKEN" -t {{.Image}}:{{.Tag}} -f {{.Dockerfile}} {{.Context}}`

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
	Use:   "build <component> [container]",
	Args:  cobra.RangeArgs(1, 2),
	Short: "Builds images for the given component using containers defined in config.",
	Long: `Builds images for the given component using containers defined in config.
This command can also be used to build a specific container for a specific component instead of building and tagging them all at once.
An optional tag can be provided as a flag, or 'latest' will be used.

For example:

	stack build app -t v0.1.0-alpha		# builds the images for all the containers defined by the app component in the project's config' with the tag v0.1.0-alpha 

	stack build app app-image			# build the image 'app:latest' for the container 'app' defined by the component 'app'
`,
	RunE: runBuildComponent,
}

func runBuildComponent(cmd *cobra.Command, args []string) (err error) {
	tag, _ := cmd.Flags().GetString("tag")
	for _, component := range config.Components {
		if args[0] == component.Name {
			for _, container := range component.Containers {
				if len(args) == 2 {
					if args[1] == container.Image {
						return buildComponent(container.Context, container.Dockerfile, container.Image, tag)
					} else {
						continue
					}
				}
				err = buildComponent(container.Context, container.Dockerfile, container.Image, tag)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func buildComponent(context, dockerfile, image, tag string) (err error) {
	configDirectory := viper.GetString("project_directory")
	contextPath, _ := filepath.Abs(filepath.Join(configDirectory, context))
	dockerfilePath, _ := filepath.Abs(filepath.Join(configDirectory, dockerfile))

	dockerBuildCommand, err := GenerateCommand(dockerBuildTemplate, DockerBuildRequest{
		Dockerfile: dockerfilePath,
		Image:      image,
		Tag:        tag,
		Context:    contextPath,
		NoCache:    noCache,
	})
	if err != nil {
		return err
	}

	dockerBuildCommand.Stdout = os.Stdout
	dockerBuildCommand.Stderr = os.Stderr
	if err := dockerBuildCommand.Run(); err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.PersistentFlags().StringP("tag", "t", "latest", "Image tag. Tag parameter will override this.")
	buildCmd.PersistentFlags().BoolVar(&noCache, "noCache", false, "Build images without cache")
}
