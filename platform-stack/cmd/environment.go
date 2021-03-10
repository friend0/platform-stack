package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/gookit/color"
	"github.com/pkg/errors"
	"io"
	"os"
	"strings"

	"github.com/altiscope/platform-stack/pkg/schema/latest"
	"github.com/spf13/cobra"
)

// environmentCmd represents the environment command
var environmentCmd = &cobra.Command{
	Use:   "environment [target]",
	Short: "Get or set the current active environment.",
	Long: `Get or set the current active environment.
If no args are provided, the current environment is retrieved. 
If a target argument is provided, then stack will activate the configured environment with name matching target.`,
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"env"},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validateConfiguredEnvironments(config.Environments, getContext(), os.Getenv)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var environment latest.EnvironmentDescription
		if len(args) == 0 {
			environment, err = getEnvironment()
			if err != nil {
				return err
			}

			if environment != (latest.EnvironmentDescription{}) {
				fmt.Printf("Current stack environment \"%v\". \nEnvironmentDescription:\n", environment.Name)
			} else {
				fmt.Println("No environment currently active.")
			}
			res, _ := json.MarshalIndent(environment, "", "    ")
			color.Info.Println(string(res))
		} else {
			targetEnvironment := args[0]
			environment, err = setEnvironment(targetEnvironment, os.Stdout)
			if err != nil {
				return err
			}
			if environment == (latest.EnvironmentDescription{}) {
				return fmt.Errorf("blank env returned")
			}
		}
		return nil
	},
}

// isEnvActive determines if the current environment is active under current system conditions
func isEnvActive(env latest.EnvironmentDescription, kubectx string, getEnv func(string) string) bool {
	var contextActivation, envActivation bool
	kubectxs := strings.Split(env.Activation.Context, "||")
	for i, ctx := range kubectxs {
		kubectxs[i] = strings.TrimSpace(ctx)
	}
	for _, context := range kubectxs {
		if kubectx == context {
			contextActivation = true
		}
	}

	if len(env.Activation.Env) == 0 {
		envActivation = true
	} else {
		activationEnvs := strings.Split(env.Activation.Env, "=")
		if len(activationEnvs) >= 2 {
			activationEnvKey, activationEnvValue := activationEnvs[0], activationEnvs[1]
			envActivation = getEnv(activationEnvKey) == activationEnvValue
		}
	}
	return contextActivation && envActivation
}

// validateConfiguredEnvironments checks that the environment section of the project config is consistent
// and has all required fields.
func validateConfiguredEnvironments(configuredEnvironments []latest.EnvironmentDescription, kubectx string, getEnv func(string) string) (err error) {
	var numActive int
	for i, env := range configuredEnvironments {
		if env.Name == "" {
			return fmt.Errorf("environment[%v] has no name", i)
		}
		if env.Activation == (latest.ActivationDescription{}) {
			return fmt.Errorf("environment[%v] has no ActivationDescription", i)
		} else {
			if env.Activation.Context == "" {
				return fmt.Errorf("environment[%v] has no ActivationDescription.Contexts defined: expecting string or || concatenated list", env.Name)
			}
		}
		if isEnvActive(env, kubectx, getEnv) {
			numActive++
		}
	}
	if numActive <= 1 {
		return nil
	} else {
		return fmt.Errorf("multiple configurations active")
	}
}

// getEnvironment inspects the current kubectx and environment variables to determine the active environment.
// This determination is made based on the EnvironmentDescriptions provided at the top level of the project's stack configuration file.
func getEnvironment() (currentEnvironment latest.EnvironmentDescription, err error) {
	if len(config.Environments) <= 0 {
		return latest.EnvironmentDescription{}, fmt.Errorf("no environments found - double check you are in a stack directory with configured environments")
	}
	currentContext := getContext()
	currentEnvironment, err = getCurrentEnvironment(config.Environments, currentContext, os.Getenv)
	if err != nil {
		return currentEnvironment, err
	}
	return currentEnvironment, nil
}

// getCurrentEnvironment encapsulates retrieval of the current environment into a testable unit
func getCurrentEnvironment(configuredEnvironments []latest.EnvironmentDescription, kubectx string, getEnv func(string) string) (latest.EnvironmentDescription, error) {
	err := validateConfiguredEnvironments(configuredEnvironments, kubectx, getEnv)
	if err != nil {
		return latest.EnvironmentDescription{}, errors.Wrap(err, "environment validation failed")
	}
	for _, env := range configuredEnvironments {
		envActive := isEnvActive(env, kubectx, getEnv)
		if envActive {
			return env, nil
		}
	}
	return latest.EnvironmentDescription{}, nil
}

// isBuildEnvActive determines if the current build environment is active under current system conditions
func isBuildEnvActive(env latest.EnvironmentDescription, getEnv func(string) string) bool {
	var envActivation bool
	if len(env.Activation.Env) == 0 {
		envActivation = true
	} else {
		activationEnvs := strings.Split(env.Activation.Env, "=")
		if len(activationEnvs) >= 2 {
			activationEnvKey, activationEnvValue := activationEnvs[0], activationEnvs[1]
			envActivation = getEnv(activationEnvKey) == activationEnvValue
		}
	}
	return envActivation
}

// getBuildEnvironment inspects the current environment variables to determine the active environment.
// This determination is made based on the EnvironmentDescriptions provided at the top level of the project's stack configuration file.
func getBuildEnvironment() (currentEnvironment latest.EnvironmentDescription, err error) {
	if len(config.Environments) <= 0 {
		return latest.EnvironmentDescription{}, fmt.Errorf("no environments found - double check you are in a stack directory with configured environments")
	}
	currentEnvironment, err = getCurrentBuildEnvironment(config.Environments, os.Getenv)
	if err != nil {
		return currentEnvironment, err
	}
	return currentEnvironment, nil
}

// getBuildEnvironment will return the current build environment (no kubectx considered)
func getCurrentBuildEnvironment(configuredEnvironments []latest.EnvironmentDescription, getEnv func(string) string) (latest.EnvironmentDescription, error) {
	for _, env := range configuredEnvironments {
		envActive := isBuildEnvActive(env, getEnv)
		if envActive {
			return env, nil
		}
	}
	return latest.EnvironmentDescription{}, nil
}

// setEnvironment sets the current kubectx and environment flags to those defined by the EnvironmentDescription with name
// matching the provided argument. EnvironmentDescriptions are defined at the top level of a stack configuration file.
func setEnvironment(targetEnvironmentName string, out io.Writer) (targetEnvironment latest.EnvironmentDescription, err error) {
	if len(config.Environments) <= 0 {
		return latest.EnvironmentDescription{}, fmt.Errorf("no environments found - double check you are in a stack directory with configured environments")
	}
	for _, env := range config.Environments {
		if env.Name == targetEnvironmentName {
			targetEnvironment = env
			break
		}
	}
	if targetEnvironment == (latest.EnvironmentDescription{}) {
		return targetEnvironment, fmt.Errorf("target environment not found")
	}

	// activate the context and environment variables described
	kubectxs := strings.Split(targetEnvironment.Activation.Context, "||")
	kubectxIndex := 0
	currentContext := getContext()
	for i, ctx := range kubectxs {
		kubectxs[i] = strings.TrimSpace(ctx)
		if currentContext == kubectxs[i] {
			kubectxIndex = i
		}
	}
	err = setContext(kubectxs[kubectxIndex])
	if err != nil {
		return targetEnvironment, err
	}

	if targetEnvironment.Activation.Env != "" {
		envKeyValue := strings.Split(targetEnvironment.Activation.Env, "=")
		if len(envKeyValue) != 2 {
			return targetEnvironment, fmt.Errorf("expected actiavtion env as `key=value`, got `%v` instead", targetEnvironment.Activation.Env)
		}
		activationEnvKey, activationEnvValue := envKeyValue[0], envKeyValue[1]
		if os.Getenv(activationEnvKey) == activationEnvValue {
			_, _ = fmt.Fprintf(out, "Switched to environment \"%v\".\n", targetEnvironment.Name)

		} else {
			_, err = fmt.Fprintf(out, "Target environment requires parent process environment variables to be set. Run the following in your terminal:\n\t$ export %v=%v\n", envKeyValue[0], envKeyValue[1])
		}
		if err != nil {
			return targetEnvironment, err
		}
	} else {
		_, _ = fmt.Fprintf(out, "Switched to environment \"%v\".\n", targetEnvironment.Name)
	}
	return targetEnvironment, nil
}

func init() {
	rootCmd.AddCommand(environmentCmd)
}
