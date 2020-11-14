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

const kubetplRenderTemplate = `kubetpl render {{if .Output}} -o {{.OutputFile}} {{end}} --allow-fs-access {{ .Manifest }} {{ range .TemplateConfig }} -i {{.}} {{end}} {{ range .Env }} -s {{.}} {{end}}`

type KubectlApplyRequest struct {
	YamlFile string
}

type KubetplRenderRequest struct {
	Manifest       string
	TemplateConfig []string
	Env            []string
	OutputFile     string
	Output 		   bool
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
	RunE: upAllComponents,
	Args: cobra.MaximumNArgs(1),
}

func upAllComponents(cmd *cobra.Command, args []string) (err error) {

	currentEnv, err := getEnvironment()
	if err != nil {
		return err
	}
	if currentEnv == (EnvironmentDescription{}) {
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

func componentUpFunction(cmd *cobra.Command, component ComponentDescription, stackEnv EnvironmentDescription) (err error) {
	projectDirectory, _ := cmd.Flags().GetString("project_directory")
	absoluteProjectDirectory, _ := filepath.Abs(projectDirectory)
	requiredVariables := component.RequiredVariables

	for _, manifest := range component.Manifests {
		manifestName := strings.TrimSuffix(filepath.Base(manifest), filepath.Ext(manifest))
		manifestPath := filepath.Join(absoluteProjectDirectory, manifest)
		manifestDirectory := filepath.Dir(manifestPath)
		outputYamlFile := fmt.Sprintf("%v/%v-generated.yaml", manifestDirectory, manifestName)

		envs, err := generateEnvs(requiredVariables, os.Getenv)
		if err != nil {
			return err
		}

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
			Output: !dryrun,
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
func parseComponentArgs(args []string, configuredComponents []ComponentDescription) (components []ComponentDescription, err error) {
	if len(configuredComponents) < 1 {
		return components, fmt.Errorf("no components found - double check you are in a stack directory with configured components")
	}
	argMap := make(map[string]bool)
	for _, arg := range args {
		argMap[arg] = true
	}

	for _, component := range configuredComponents {
		if len(args) >= 1 {
			if argMap[component.Name] == true {
				components = append(components, component)
			}
			continue
		}
		components = append(components, component)
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
	upCmd.Flags().BoolP("dryrun", "d", true, "Generate yaml only, do not kubectl apply")
	upCmd.Flags().Lookup("wait").NoOptDefVal = "300"
}
