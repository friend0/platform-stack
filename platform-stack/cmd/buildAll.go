package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"path/filepath"
)

// buildAllCmd represents the buildAll command
var buildAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Builds all containers for all components of the stack.",
	Long:  `Builds all containers for all components of the stack.`,
	RunE:  buildAllComponents,
}

func buildAllComponents(cmd *cobra.Command, args []string) (err error) {
	if len(config.Components) == 0 {
		return fmt.Errorf("no components found - double check you are in a configured stack directory")
	}
	// todo: confirmWithUser that they are going to build multiple components, multiple containers with the same tag
	for _, component := range config.Components {
		fmt.Printf("Building all containers for component `%v`", component.Name)
		for _, container := range component.Containers {
			tag, _ := cmd.Flags().GetString("tag")
			configDirectory := viper.GetString("project_directory")
			contextPath, _ := filepath.Abs(filepath.Join(configDirectory, container.Context))
			dockerfilePath, _ := filepath.Abs(filepath.Join(configDirectory, container.Dockerfile))

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
