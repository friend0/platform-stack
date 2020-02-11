package cmd

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return viper.Unmarshal(&config)
	},
	RunE: downAllComponents,
}

func downAllComponents(cmd *cobra.Command, args []string) (err error) {

	// parses ComponentDescriptions from input args, uses all configured components if none are provided
	components, err := parseComponentArgs(args)
	if err != nil {
		return err
	}

	projectDirectory, _ := cmd.Flags().GetString("project_directory")
	absoluteProjectDirectory, _ := filepath.Abs(projectDirectory)

	deploymentsDirectory := filepath.Join(absoluteProjectDirectory, viper.GetString("deployment_directory"))

	for _, component := range components {
		fmt.Printf("Tearing down components at %v...\n", component.Name)
		generatedYaml := filepath.Join(deploymentsDirectory, fmt.Sprintf("%v-generated.yaml", component.Name))
		err := downComponent(generatedYaml)
		if err != nil {
			fmt.Printf("`%v` component failed teardown. You may need to delete it manually.\n", component.Name)
		}
	}
	return nil
}

func downComponent(yamlFile string) (err error) {

	deleteYamlCmd, err := GenerateCommand(kubectlDeleteTemplate, KubectlDeleteRequest{
		YamlFile: yamlFile,
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
	return nil
}

func init() {
	rootCmd.AddCommand(downCmd)
}
