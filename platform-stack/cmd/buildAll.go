package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"path/filepath"
)

// buildAllCmd represents the buildAll command
var buildAllCmd = &cobra.Command{
	Use:   "all [component]",
	Args: cobra.MaximumNArgs(1),
	Short: "Builds all containers for all components of the stack, or for the component provided [optional].",
	Long: `Builds all containers for all components of the stack, or for the component provided [optional].`,
	RunE: buildAllComponents,
}


func buildAllComponents(cmd *cobra.Command, args []string) (err error) {

	tag, _ := cmd.Flags().GetString("tag")

	if len(config.Components) == 0 {
		return  fmt.Errorf("no components found - double check you are in a configured stack directory")
	}

	var componentsToBuild []ComponentDescription
	for _, component := range config.Components {
		if len(args) == 1 {
			if args[0] == component.Name {
				componentsToBuild = append(componentsToBuild, component)
				break
			} else {
				continue
			}
		}
		componentsToBuild = append(componentsToBuild, component)
	}

	if len(componentsToBuild) == 0 {
		return fmt.Errorf("component `%v` not found", args[0])
	}

	// todo: confirmWithUser that they are going to build multiple containers with the same tag
	for _, component := range componentsToBuild {
		fmt.Printf("Building all containers for component `%v`", component.Name)
		for _, container := range component.Containers {

			var contextPath, dockerfilePath string
			configDirectory := viper.GetString("project_directory")
			if configDirectory != "" {
				contextPath, _ = filepath.Abs(filepath.Join(configDirectory, container.Context))
				dockerfilePath, _ = filepath.Abs(filepath.Join(configDirectory, container.Dockerfile))
			} else {
				contextPath, _ = filepath.Abs(container.Context)
				dockerfilePath, _ = filepath.Abs(container.Dockerfile)
			}

			fmt.Println(dockerfilePath, contextPath)
			err := buildComponent(contextPath, dockerfilePath, container.Image, tag)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func init() {
	buildCmd.AddCommand(buildAllCmd)
}