package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/altiscope/platform-stack/pkg/schema/latest"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const kubectlApplyTemplate = `kubectl apply -f "{{ .YamlFile }}"`

const kubetplRenderTemplate = `kubetpl render {{if .Output}} -o {{.OutputFile}} {{end}} --allow-fs-access {{ .Manifest }} {{ range .TemplateConfig }} -i {{.}} {{end}} {{ range .Env }} -s {{.}} {{end}}`

type KubectlApplyRequest struct {
	YamlFile string
}

type KubetplRenderRequest struct {
	Manifest       string
	TemplateConfig []string
	Env            []string
	OutputFile     string
	Output         bool
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
		err = viper.BindPFlag("dryrun", cmd.Flags().Lookup("dryrun"))
		if err != nil {
			return err
		}
		return initK8s("")
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return configPreRunnerE(cmd, args)
	},
	RunE: upComponents,
}

// envsApply checks a list of environment names, and returns true or false if any members match currentEnvName
func envsApply(e []string, currentEnvName string) bool {
	if len(e) >= 1 {
		active := false
		for _, env := range e {
			if currentEnvName == env {
				active = true
				break
			}
		}
		return active
	}
	return true
}

func upComponents(cmd *cobra.Command, args []string) (err error) {

	if len(args) == 0 {
		return upAllComponents(cmd, args)
	}

	currentEnv, err := getEnvironment()
	if err != nil {
		return err
	}
	if currentEnv == (latest.EnvironmentDescription{}) {
		return fmt.Errorf("no active environment detected")
	}
	if currentEnv.Activation.ConfirmWithUser {
		confirmWithUser(fmt.Sprintf("You are about to deploy to environment `%v`", currentEnv.Name))
	}

	// Determine component list from config
	upComponents, err := parseComponentArgs(args, config.Components)
	if err != nil {
		return err
	}

	dryrun := viper.GetBool("dryrun")

	// Bring up each configured component
	for _, component := range upComponents {
		if !dryrun {
			if !envsApply(component.Environments, currentEnv.Name) {
				fmt.Printf("skipping `up` for component `%v`: not in active environment\n", component.Name)
				continue
			}
			fmt.Println("Bringing up", component.Name)
		}
		err := componentUpFunction(cmd, component, currentEnv)
		if err != nil {
			fmt.Printf("Bringing up `%v` failed", component.Name)
			return err
		}
	}

	wait := viper.GetInt("wait")
	if wait >= 0 {
		waitTime := wait * 1000
		api := clientset.CoreV1()
		err, ctx := waitForStackWithTimeout(api, cmd, time.Duration(waitTime))
		if err != nil {
			return err
		}
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("timed out waiting for stack")
		}
	}

	return nil
}

func upAllComponents(cmd *cobra.Command, args []string) (err error) {

	currentEnv, err := getEnvironment()
	if err != nil {
		return err
	}
	if currentEnv == (latest.EnvironmentDescription{}) {
		return fmt.Errorf("no active environment detected")
	}
	if currentEnv.Activation.ConfirmWithUser {
		confirmWithUser(fmt.Sprintf("You are about to deploy to environment `%v`", currentEnv.Name))
	}

	// Determine component list from config
	upComponents, err := parseComponentArgs(args, config.Components)
	if err != nil {
		return err
	}

	dryrun := viper.GetBool("dryrun")

	// Bring up each configured component
	for _, component := range upComponents {
		if !dryrun {
			if !envsApply(component.Environments, currentEnv.Name) {
				fmt.Printf("skipping `up` for component `%v`: not in active environment\n", component.Name)
				continue
			}
			fmt.Println("Bringing up", component.Name)
		}
		err := componentUpFunction(cmd, component, currentEnv)
		if err != nil {
			fmt.Printf("Bringing up `%v` failed", component.Name)
			return err
		}
	}

	wait := viper.GetInt("wait")
	if wait >= 0 {
		waitTime := wait * 1000
		api := clientset.CoreV1()
		err, ctx := waitForStackWithTimeout(api, cmd, time.Duration(waitTime))
		if err != nil {
			return err
		}
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("timed out waiting for stack")
		}
	}

	return nil
}

func componentUpFunction(cmd *cobra.Command, component latest.ComponentDescription, stackEnv latest.EnvironmentDescription) (err error) {

	absoluteProjectDirectory, _ := filepath.Abs(viper.GetString("stack_directory"))
	requiredVariables := component.RequiredVariables
	envOverrides, _ := cmd.Flags().GetStringSlice("env")

	for _, manifest := range component.Manifests {
		manifestName := strings.TrimSuffix(filepath.Base(manifest), filepath.Ext(manifest))
		manifestPath := filepath.Join(absoluteProjectDirectory, manifest)
		manifestDirectory := filepath.Dir(manifestPath)
		outputYamlFile := fmt.Sprintf("%v/%v-generated.yaml", manifestDirectory, manifestName)

		envs, err := generateEnvs(requiredVariables, os.Getenv)
		if err != nil {
			return err
		}
		envs = append(envs, envOverrides...)

		// if a componet does not have config specified, try to find the magic template config
		cf := component.TemplateConfig
		if len(cf) == 0 {
			cf = []string{fmt.Sprintf("%v/config-%v.env", manifestDirectory, stackEnv.Name)}
		}

		dryrun := viper.GetBool("dryrun")
		generateYamlCmd, err := GenerateCommand(kubetplRenderTemplate, KubetplRenderRequest{
			Manifest:       fmt.Sprintf("%v/%v.yaml", manifestDirectory, manifestName),
			TemplateConfig: cf,
			Env:            envs,
			OutputFile:     outputYamlFile,
			Output:         !dryrun,
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

		if !dryrun {
			applyYamlCmd, err := GenerateCommand(kubectlApplyTemplate, KubectlApplyRequest{
				YamlFile: outputYamlFile,
			})
			if err != nil {
				return err
			}
			applyYamlCmd.Stdout = os.Stdout
			applyYamlCmd.Stderr = os.Stderr
			if err := applyYamlCmd.Run(); err != nil {
				return err
			}
		}
	}

	return nil
}

// parseComponentArgs generates a list of ComponentDescriptions from the up command's arguments if provided, defaulting
// to all configured components if none are provided
func parseComponentArgs(args []string, configuredComponents []latest.ComponentDescription) (components []latest.ComponentDescription, err error) {
	if len(configuredComponents) < 1 {
		return components, fmt.Errorf("no components found - double check you are in a stack directory with configured components")
	}
	argMap := make(map[string]bool)
	for _, arg := range args {
		argMap[arg] = true
	}

	for _, component := range configuredComponents {
		if len(args) >= 1 {
			if argMap[component.Name] {
				components = append(components, component)
			}
		} else {
			components = append(components, component)
		}
	}
	return components, nil
}

// generateEnvs builds a list of environment key value pairs that are hydrated with values obtained from the provided getEnv function.
// Pairs are given in .env format as `key="value"`.
// You can provide `os.Getenv` as the argument to the getEnv parameter to access system variables
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
	upCmd.Flags().IntP("wait", "w", -1, "Stack readiness wait period in seconds")
	upCmd.Flags().BoolP("dryrun", "d", false, "Generate yaml only, do not kubectl apply")
	upCmd.Flags().StringSliceP("env", "e", []string{}, "Env variables")
	upCmd.Flags().Lookup("wait").NoOptDefVal = "300"
}
