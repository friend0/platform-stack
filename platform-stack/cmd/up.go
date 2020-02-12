package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
	"time"

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
		initK8s()
		return viper.BindPFlag("wait", cmd.Flags().Lookup("wait"))
	},
	RunE: upAllComponents,
}

// todo: share this map with downAllComponents
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

	// todo: timeout
	// todo: easily parsible report
	wait := viper.GetBool("wait")

	if wait {
		healthDetail, err, ctx := waitForStackWithTimeout(cmd, 30000)
		if err != nil {
			return err
		}

		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("timed out waiting for stack:\n %v", healthDetail)
		}

	}

	return nil
}

func waitForStackWithTimeout(cmd *cobra.Command, timeoutMs time.Duration) (results []string, err error, ctx context.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), timeoutMs*time.Millisecond)
	defer cancel()  // releases resources if slowOperation completes before timeout elapses
	results, err =  waitForStack(cmd)
	return results, err, ctx
}

func waitForStack(cmd *cobra.Command) (results []string, err error) {

	api := clientset.CoreV1()

	ns, _ := cmd.Flags().GetString("namespace")
	label, _ := cmd.Flags().GetStringSlice("label")
	field, _ := cmd.Flags().GetStringSlice("field")

	podList, err := getPodsList(api, ns, label, field)
	if err != nil {
		return results, err
	}
	for {
		results = podHealth(podList)
		allReady := true
		for _, item := range results {
			notHealthy := strings.Contains(item, "not healthy")
			if notHealthy {
				allReady = false
			}
		}
		if allReady {
			return results, nil
		}
		time.Sleep(10 * time.Second)
	}
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

	// todo: up command hard-coded to config-local.env - taker flags
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
