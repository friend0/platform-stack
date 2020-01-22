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
	RunE: downAllComponents,
}

func downAllComponents(cmd *cobra.Command, args []string) (err error) {

	components, err := parseComponentArgs(args)
	if err != nil {
		return err
	}

	for _, component := range components {
		fmt.Printf("Tearing down components at %v...\n", component.Name)
		deploymentsDirectory, _ := filepath.Abs(viper.Get("deployments_directory").(string))
		generatedYaml := filepath.Join(deploymentsDirectory, fmt.Sprintf("%v-generated", component.Name))
		err := downComponent(generatedYaml)
		if err != nil {
			fmt.Printf("Tear down %v component failed: %v", component.Name, err.Error())
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
