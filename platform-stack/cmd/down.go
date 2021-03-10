package cmd

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
	"github.com/altiscope/platform-stack/pkg/schema/latest"
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
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return configPreRunnerE(cmd, args)
	},
	RunE: downAllComponents,
}

func downAllComponents(cmd *cobra.Command, args []string) (err error) {

	currentEnv, err := getEnvironment()
	if err != nil {
		return err
	}
	if currentEnv == (latest.EnvironmentDescription{}) {
		return fmt.Errorf("no active environment detected")
	}
	if currentEnv.Activation.ConfirmWithUser {
		confirmWithUser(fmt.Sprintf("You are about to destroy pods in `%v`", currentEnv.Name))
	}

	components, err := parseComponentArgs(args, config.Components)
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

func downComponent(cmd *cobra.Command, component latest.ComponentDescription) (err error) {
	absoluteProjectDirectory, _ := filepath.Abs(viper.GetString("stack_directory"))

	for _, manifest := range component.Manifests {

		manifestName := strings.TrimSuffix(filepath.Base(manifest), filepath.Ext(manifest))
		manifestPath := filepath.Join(absoluteProjectDirectory, manifest)
		manifestDirectory := filepath.Dir(manifestPath)
		generatedYamlFile := fmt.Sprintf("%v/%v-generated.yaml", manifestDirectory, manifestName)

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
