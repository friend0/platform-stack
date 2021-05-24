package cmd

import (
	"fmt"
	"github.com/altiscope/platform-stack/pkg/schema/latest"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)


const dockerBuildTemplate = `DOCKER_BUILDKIT=1 docker build {{if .NoCache}} --no-cache {{end}} --build-arg GIT_TOKEN="$GIT_TOKEN" -t {{.Tag}} -f {{.Dockerfile}} {{.Context}}`

var noCache bool

type DockerBuildRequest struct {
	Dockerfile string
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
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return configPreRunnerE(cmd, args)
	},
	RunE: runBuildComponent,
}

func buildForCurrentEnvironment(cd latest.ContainerDescription, currentEnvName string) bool {
	e := cd.Environments
	if len(e) >= 1 {
		active := false
		for _, env := range e {
			if currentEnvName == env {
				active = true
				break
			}
		}
		if !active {
			fmt.Printf("skipping build for image `%v`: not in active environment\n", cd.Image)
			return false
		}
	}
	return true
}

func runBuildComponent(cmd *cobra.Command, args []string) (err error) {
	for _, component := range config.Components {
		if args[0] == component.Name {
			for _, container := range component.Containers {
				env, err := getBuildEnvironment()
				if err != nil {
					return err
				}
				environmentEnabled := buildForCurrentEnvironment(container, env.Name)
				if !environmentEnabled {
					continue
				}
				if len(args) == 2 {
					if args[1] != container.Image { // todo: || container.ShortName
						continue
					}
				}

				tag, _ := cmd.Flags().GetString("tag")
				if tag == "" {
					imageTag, _ := cmd.Flags().GetString("imageTag")
					if imageTag == "" {
						imageTag = "latest"
					}
					tag = fmt.Sprintf("%v:%v", container.Image, imageTag)
				}
				err = buildComponent(container.Context, container.Dockerfile, tag)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func buildComponent(context, dockerfile, tag string) (err error) {
	configDirectory, _ := filepath.Abs(viper.GetString("stack_directory"))
	contextPath := filepath.Join(configDirectory, context)
	dockerfilePath := filepath.Join(configDirectory, dockerfile)

	dockerBuildCommand, err := GenerateCommand(dockerBuildTemplate, DockerBuildRequest{
		Dockerfile: dockerfilePath,
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
	buildCmd.PersistentFlags().StringP("tag", "t", "", "Name and optionally a tag in the 'name:tag' format (same as docker flag). Defaults to image:latest based on stack config.")
	buildCmd.PersistentFlags().StringP("imageTag", "i", "", "Set the tag only of the 'name:tag' format and use the stack configured image name as the name.")
	buildCmd.PersistentFlags().BoolVar(&noCache, "noCache", false, "Build images without cache")
}
