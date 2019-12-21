package cmd

import (
	"fmt"
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
	Use:   "build [component]",
	Args: cobra.RangeArgs(1, 2),
	Short: "Builds an image from the Dockerfile at the determined context.",
	Long: `Builds an image from the Dockerfile at the determined context.

Example:

	stack build  myComponent 	# builds Dockerfile of myComponent at the configured context
	stack build -c ./myContext -d Dockerfile.env myComponent 	# builds Dockerfile "Dockerfile.env" of 'myComponent' using the context "./myContext" 

If a configuration file is present, the context is inferred from the 'build_directory' env variable and the name of the component.
For example, for build_directory=./containers, the first example command would build the Dockerfile at ./containers/[component name]

`,
	RunE: buildComponent,
}

func buildComponent(cmd *cobra.Command, args []string) (err error) {

	tag, _ := cmd.Flags().GetString("tag")
	context, _ := cmd.Flags().GetString("context")
	dockerfile, _ := cmd.Flags().GetString("dockerfile")

	image := args[0]
	if len(args) == 2 {
		tag = args[1]
	}

	fmt.Println(image, ":", tag)

	buildDirectory, _ := filepath.Abs(context)
	componentDirectory := filepath.Join(buildDirectory, image)


	dockerBuildCommand, err := GenerateCommand(dockerBuildTemplate, DockerBuildRequest{
		Dockerfile: filepath.Join(componentDirectory, dockerfile),
		Image:      image,
		Tag:        tag,
		Context:    componentDirectory,
		NoCache:    noCache,
	})

	fmt.Println(dockerBuildCommand)
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

	buildCmd.PersistentFlags().StringP("tag", "t", "latest", "tag for images to be built")

	buildCmd.PersistentFlags().BoolVar(&noCache, "noCache", false, "build images without cache")

	buildCmd.Flags().StringP("context", "c", ".", "Select Docker build context")
	viper.BindPFlag("build_directory", buildCmd.Flags().Lookup("context"))

	buildCmd.Flags().StringP("dockerfile", "d", "Dockerfile","Select name of Dockerfile")


}
