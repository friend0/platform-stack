package cmd

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// environmentCmd represents the environment command
var environmentCmd = &cobra.Command{
	Use:   "environment [target]",
	Short: "Get or set the current active environment.",
	Long: `Get or set the current active environment.
If no args are provided, the current environment is retrieved. 
If a target argument is provided, then stack will activate the configured environment with name matching target.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var environment EnvironmentDescription
		if len(args) == 0 {
			environment, err = getEnvironment()
			if err != nil {
				return err
			}
		} else {
			targetEnvironment := args[0]
			environment, err = setEnvironment(targetEnvironment)
			if err != nil {
				return err
			}
		}
		if environment == (EnvironmentDescription{}) {
			return fmt.Errorf("blank env returned")
		}
		fmt.Printf("Switched to environment \"%v\".\n", environment.Name)
		return nil
	},
}

// getEnvironment inspects the current kubectx and environment variables to determine the active environment.
// This determination is made based on the EnvironmentDescriptions provided at the top level of the project's stack configuration file.
func getEnvironment() (currentEnvironment EnvironmentDescription, err error) {
	currentContext := getContext()
	configuredEnvironments := config.Environments
	if len(configuredEnvironments) <= 0 {
		return EnvironmentDescription{}, fmt.Errorf("no environments found - double check you are in a stack directory with configured environments")
	}

	for _, env := range configuredEnvironments {
		var contextActivation, envActivation bool
		if currentContext == env.Activation.Context {
			contextActivation = true
		}
		if  len(env.Activation.Env) == 0 {
			envActivation = true
		} else {
			activationEnvs := strings.Split(env.Activation.Env, "=")
			activationEnvKey, activationEnvValue := activationEnvs[0], activationEnvs[1]
			envActivation = os.Getenv(activationEnvKey) == activationEnvValue
		}
		if contextActivation && envActivation {
			return env, nil
		}
	}
	return EnvironmentDescription{}, fmt.Errorf("no environment active under current confitions")
}

// setEnvironment sets the current kubectx and environment flags to those defined by the EnvironmentDescription with name
// matching the provided argument. EnvironmentDescriptions are defined at the top level of a stack configuration file.
func setEnvironment(targetEnvironmentName string) (targetEnvironment EnvironmentDescription, err error) {
	if len(config.Environments) <= 0 {
		return EnvironmentDescription{}, fmt.Errorf("no environments found - double check you are in a stack directory with configured environments")
	}

	for _, env := range config.Environments {
		if env.Name == targetEnvironmentName {
			targetEnvironment = env
			break
		}
	}
	if targetEnvironment == (EnvironmentDescription{}) {
		return targetEnvironment, fmt.Errorf("target environment not found")
	}

	// activate the context and environment variables described
	err = setContext(targetEnvironment.Activation.Context)
	if err != nil {
		return targetEnvironment, err
	}

	envKeyValue := strings.Split(targetEnvironment.Activation.Env, "=")
	if len(envKeyValue) != 2 {
		return targetEnvironment, fmt.Errorf("expected actiavtion env as `key=value`, got `%v` instead", targetEnvironment.Activation.Env)
	}
	viper.Set(envKeyValue[0], envKeyValue[1])
	return targetEnvironment, nil
}


func init() {
	rootCmd.AddCommand(environmentCmd)
}
