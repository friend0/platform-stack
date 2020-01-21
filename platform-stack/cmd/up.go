package cmd

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"path/filepath"

	//"github.com/spf13/viper"
	"os"
)

var env string

const kubectlApplyTemplate = `kubectl apply -f "{{ .YamlFile }}"`

const kubetplRenderTemplate = `kubetpl render --allow-fs-access {{ .Manifest }} {{ range .EnvFrom }} -i {{.}} {{end}} {{ range .Env }} -s {{.}} {{end}}`

type KubectlApplyRequest struct {
	YamlFile string
}

type KubetplRenderRequest struct {
	Manifest string
	EnvFrom      []string
	Env 		 []string
}

type ComponentDescription struct {
	Name string `json:"name"`
	RequiredVariables []string `json:"required_variables"`
	Exposable bool	`json:"exposable"`
}

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up [<component>...]",
	Short: "Brings up components of the stack.",
	Long: `Brings up components of the stack.

If no components are provided as arguments, all components will be brought up.`,
	RunE: upAllComponents,
}

// todo: share this map with downAllComponents
func upAllComponents(cmd *cobra.Command, args []string) (err error) {

	components, err := parseComponentArgs(args)
	if err != nil {
		return err
	}

	for _, component := range components {
		fmt.Println("Building component", component.Name)
		err := componentUpFunction(cmd, component)
		if err != nil {
			fmt.Printf("Build %v component failed", component)
			return err
		}
	}
	return nil
}

func componentUpFunction(cmd *cobra.Command, component ComponentDescription) (err error) {

	projectDirectory, _ := cmd.Flags().GetString("project_directory")
	projectDirectory, _ = filepath.Abs(projectDirectory)

	deploymentsDirectory := viper.GetString("deployment_directory")
	if projectDirectory != "" {
		deploymentsDirectory = filepath.Join(projectDirectory, deploymentsDirectory)
	}

	outputYamlFile := fmt.Sprintf("%v/%v-generated.yaml", deploymentsDirectory, component.Name)

	requiredVariables := component.RequiredVariables
	envs, err := generateEnvs(requiredVariables)
	if err != nil {
		return err
	}

	generateYamlCmd, err := GenerateCommand(kubetplRenderTemplate, KubetplRenderRequest{
		Manifest: fmt.Sprintf("%v/%v.yaml", deploymentsDirectory, component.Name),
		EnvFrom: []string{fmt.Sprintf("%v/config-%v.env", deploymentsDirectory, "local")},
		Env: envs,
	})

	applyYamlCmd, err := GenerateCommand(kubectlApplyTemplate, KubectlApplyRequest{
		YamlFile: outputYamlFile,
	})

	if err != nil {
		return err
	}

	var generatedYaml bytes.Buffer
	generateYamlCmd.Env = append(os.Environ())
	generateYamlCmd.Stdout = &generatedYaml
	generateYamlCmd.Stderr = os.Stderr
	if err := generateYamlCmd.Run(); err != nil {
		return err
	}

	err = ioutil.WriteFile(outputYamlFile, generatedYaml.Bytes(), 0664)
	if err != nil {
		return err
	}

	applyYamlCmd.Stdout = os.Stdout
	applyYamlCmd.Stderr = os.Stderr
	if err := applyYamlCmd.Run(); err != nil {
		return err
	}

	return nil

}

func parseComponentArgs(args []string) (components []ComponentDescription, err error) {
	if len(args) < 1 {
		configuredComponents := viper.Get("components").([]interface{})
		for _, component := range configuredComponents {

			if componentString, ok := component.(string); ok {
				components = append(components, ComponentDescription{
					Name: componentString,
				})
			}

			if componentMap, ok := component.(map[interface{}]interface{}); ok {
				if _, ok := componentMap["name"].(string); ok {
					if _, ok := componentMap["required_variables"].([]string); ok {
						components = append(components, ComponentDescription{
							Name: componentMap["name"].(string),
							RequiredVariables: componentMap["required_variables"].([]string),
						})
					} else {
						components = append(components, ComponentDescription{
							Name: componentMap["name"].(string),
						})
					}
				}
			}

		}
	} else {
		for _, arg := range args {
			components = append(components, ComponentDescription{Name: arg})
		}
	}
	return components, nil
}

func generateEnvs(requiredVariables []string) (envs []string, err error) {

	for _, variable := range requiredVariables {
		if os.Getenv(variable) != "" {
			envs = append(envs, fmt.Sprintf(`%s="%s"`, variable, os.Getenv(variable)))
		} else {
			return envs, fmt.Errorf("missing environment variable: %v", variable)
		}
	}
	return envs, nil

}

func init() {
	rootCmd.AddCommand(upCmd)
	upCmd.Flags().StringVar(&env, "environment", "local","Select deployment environment")
}
