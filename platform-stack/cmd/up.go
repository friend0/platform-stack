package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"path/filepath"
	//"github.com/spf13/viper"
	"os"
)

var env string

const kubectlApplyTemplate = `kubectl apply -f "{{ .YamlFile }}"`

const kubetplRenderTemplate = `kubetpl render -o {{.OutputFile}} --allow-fs-access {{ .Manifest }} {{ range .EnvFrom }} -i {{.}} {{end}} {{ range .Env }} -s {{.}} {{end}}`

type KubectlApplyRequest struct {
	YamlFile string
}

type KubetplRenderRequest struct {
	Manifest   string
	EnvFrom    []string
	Env        []string
	OutputFile string
}

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up [<component>...]",
	Short: "Brings up components of the stack.",
	Long: `Brings up components of the stack.

If no components are provided as arguments, all configured components will be brought up.'`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := viper.BindPFlag("wait", cmd.Flags().Lookup("wait"))
		if err != nil {
			return err
		}
		return initK8s()
	},
	RunE: upAllComponents,
}

func upAllComponents(cmd *cobra.Command, args []string) (err error) {

	// Determine component list from config
	upComponents, err := parseComponentArgs(args)
	if err != nil {
		return err
	}

	// Bring up each configured component
	for _, component := range upComponents {
		fmt.Println("Bringing up", component.Name)
		err := componentUpFunction(cmd, component)
		if err != nil {
			fmt.Printf("Bringing up `%v` failed", component)
			return err
		}
	}

	wait := viper.GetBool("wait")
	if wait {
		api := clientset.CoreV1()
		err, ctx := waitForStackWithTimeout(api, cmd, 60000)
		if err != nil {
			return err
		}
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("timed out waiting for stack")
		}
	}

	return nil
}

func componentUpFunction(cmd *cobra.Command, component ComponentDescription) (err error) {

	projectDirectory, _ := cmd.Flags().GetString("project_directory")
	absoluteProjectDirectory, _ := filepath.Abs(projectDirectory)

	deploymentsDirectory := filepath.Join(absoluteProjectDirectory, viper.GetString("deployment_directory"))

	outputYamlFile := fmt.Sprintf("%v/%v-generated.yaml", deploymentsDirectory, component.Name)

	requiredVariables := component.RequiredVariables
	envs, err := generateEnvs(requiredVariables, os.Getenv)
	if err != nil {
		return err
	}

	generateYamlCmd, err := GenerateCommand(kubetplRenderTemplate, KubetplRenderRequest{
		Manifest:   fmt.Sprintf("%v/%v.yaml", deploymentsDirectory, component.Name),
		EnvFrom:    []string{fmt.Sprintf("%v/config-%v.env", deploymentsDirectory, viper.Get("env"))},
		Env:        envs,
		OutputFile: outputYamlFile,
	})

	applyYamlCmd, err := GenerateCommand(kubectlApplyTemplate, KubectlApplyRequest{
		YamlFile: outputYamlFile,
	})

	if err != nil {
		return err
	}

	// todo: remove env append if able
	generateYamlCmd.Env = append(os.Environ())
	generateYamlCmd.Stdout = os.Stdout
	generateYamlCmd.Stderr = os.Stderr
	if err := generateYamlCmd.Run(); err != nil {
		return err
	}

	applyYamlCmd.Stdout = os.Stdout
	applyYamlCmd.Stderr = os.Stderr
	if err := applyYamlCmd.Run(); err != nil {
		return err
	}

	return nil

}

// parseComponentArgs generates a list of ComponentDescriptions from the `up`, and will
// use the configured components in the project's stack.yaml if no arguments are provided.
func parseComponentArgs(args []string) (components []ComponentDescription, err error) {

	if len(args) >= 1 {
		// todo: cross reference with config to get additional component description items
		for _, arg := range args {
			components = append(components, ComponentDescription{Name: arg})
		}
		return components, nil
	} else {

		if len(config.Components) < 1 {
			return components, fmt.Errorf("no components found - double check you are in a stack directory with configured components")
		}

		for _, component := range config.Components {
			if len(component.RequiredVariables) >= 1 {
				components = append(components, ComponentDescription{
					Name:              component.Name,
					RequiredVariables: component.RequiredVariables,
				})
			} else {
				components = append(components, ComponentDescription{
					Name: component.Name,
				})
			}
		}
	}
	return components, nil
}

// generateEnvs builds a list of environment key value pairs to hydrate required variables with system values.
// Pairs are given in .env format as `key="value"`.
func generateEnvs(requiredVariables []string, getEnv func(string) string) (envs []string, err error) {
	for _, variable := range requiredVariables {
		if getEnv(variable) != "" {
			envs = append(envs, fmt.Sprintf(`%s="%s"`, variable, getEnv(variable)))
		} else {
			return envs, fmt.Errorf("missing environment variable: %v", variable)
		}
	}
	return envs, nil
}

func init() {
	rootCmd.AddCommand(upCmd)
	upCmd.Flags().StringVarP(&env, "environment", "e", "local", "Select deployment environment")
	upCmd.Flags().BoolP("wait", "w", false, "Wait until stack is ready before exit")

}
