package cmd

import (
	"github.com/spf13/viper"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
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
	Use:   "build <component> [tag]",
	Args:  cobra.RangeArgs(1, 2),
	Short: "Builds an image for the given component using the Dockerfile at the configured build directory.",
	Long: `Builds an image for the given component using the Dockerfile at the configured build directory.
An optional tag can be provided after the component name, or as an option.

We assume deployable objects are organized in a single build directory.
The 'build_directory' variable must be configured in the current project's stack configuration file.

For example, where build_directory=./containers:

	stack build <component>		# builds the Dockerfile at ./containers/component with a context at that directory

You can also set your own context or dockerfile, and provide tags if needed: 

	stack build -c <context> -d <dockerfile> -t <tag> <component>
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("build_directory", cmd.Flags().Lookup("context"))
	},
	RunE: buildComponent,
}

func buildComponent(cmd *cobra.Command, args []string) (err error) {

	project_directory, _ := cmd.Flags().GetString("project_directory")
	project_directory, _ = filepath.Abs(project_directory)

	tag, _ := cmd.Flags().GetString("tag")
	context := viper.GetString("context")

	if project_directory != "" {
		context = filepath.Join(project_directory, context)
	}
	dockerfile, _ := cmd.Flags().GetString("dockerfile")

	image := args[0]
	if len(args) == 2 {
		tag = args[1]
	}

	buildDirectory, _ := filepath.Abs(context)
	componentDirectory := filepath.Join(buildDirectory, image)

	dockerBuildCommand, err := GenerateCommand(dockerBuildTemplate, DockerBuildRequest{
		Dockerfile: filepath.Join(componentDirectory, dockerfile),
		Image:      image,
		Tag:        tag,
		Context:    componentDirectory,
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

	buildCmd.Flags().StringP("context", "c", "", "Select Docker build context")

	buildCmd.Flags().StringP("dockerfile", "d", "Dockerfile", "Select Dockerfile")

}
