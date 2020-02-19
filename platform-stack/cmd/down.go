package cmd

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"path/filepath"
)

const kubectlDeleteTemplate = `kubectl delete -f "{{ .YamlFile }}"`

type KubectlDeleteRequest struct {
	YamlFile string
}

// downCmd represents the down command
var downCmd = &cobra.Command{
	Use:   "down [<component>...]",
	Short: "Tears down the stack.",
	Long: `Tears down the stack.

If no arguments are provided, all configured objects will be taken down.`,
	RunE: downAllComponents,
}

func downAllComponents(cmd *cobra.Command, args []string) (err error) {
	components, err := parseComponentArgs(args)
	if err != nil {
		return err
	}

	for _, component := range components {
		fmt.Printf("Tearing down components at %v...\n", component.Name)
		err := downComponent(cmd, component)
		if err != nil {
			fmt.Printf("`%v` component failed teardown. You may need to delete it manually.\n", component.Name)
		}
	}
	return nil
}

func downComponent(cmd *cobra.Command, component ComponentDescription) (err error) {
	projectDirectory, _ := cmd.Flags().GetString("project_directory")
	absoluteProjectDirectory, _ := filepath.Abs(projectDirectory)

	for _, manifest := range component.Manifests {
		manifestPath := filepath.Join(absoluteProjectDirectory, manifest)
		manifestDirectory := filepath.Dir(manifestPath)
		generatedYamlFile := fmt.Sprintf("%v/%v-generated.yaml", manifestDirectory, component.Name)

		deleteYamlCmd, err := GenerateCommand(kubectlDeleteTemplate, KubectlDeleteRequest{
			YamlFile: generatedYamlFile,
		})
		if err != nil {
			return err
		}
		var stdoutBytes, errorBytes bytes.Buffer
		deleteYamlCmd.Stdout = &stdoutBytes
		deleteYamlCmd.Stderr = &errorBytes
		if err := deleteYamlCmd.Run(); err != nil {
			return fmt.Errorf(errorBytes.String())
		}
		fmt.Println(stdoutBytes.String())
	}
	return nil
}

func init() {
	rootCmd.AddCommand(downCmd)
}
