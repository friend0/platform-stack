package cmd

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
)

const kubectlDeleteTemplate = `kubectl delete -f "{{ .YamlFile }}"`

type KubectlDeleteRequest struct {
	YamlFile string
}

// downCmd represents the down command
var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Tears down the Logflights stack.",
	Long: `Tears down the Logflights stack.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		fmt.Println("Tearing down logflights objects")
		return downAllComponents()
	},
}

func downAllComponents() (err error) {

	componentDownMap := map[string]string {
		"config": "deployments/config-generated.yaml",
		"database": "deployments/database-generated.yaml",
		"logflights": "deployments/logflights-generated.yaml",
		"celery": "deployments/celery-generated.yaml",
		"frontend": "deployments/frontend-generated.yaml",
		"redis": "deployments/redis-generated.yaml",
	}

	for component, componentYaml := range componentDownMap {
		fmt.Printf("Tearing down components at %v...\n", componentYaml)
		err := downComponent(componentYaml)
		if err != nil {
			fmt.Printf("Tear down %v component failed: %v", component, err.Error())
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
