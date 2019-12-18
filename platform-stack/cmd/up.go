package cmd

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
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

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Brings up the Logflights stack.",
	Long: `Brings up the Logflights stack.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		fmt.Println("Deploying Logflights objects")
		return upAllComponents()
	},
}

// todo: share this map with downAllComponents
func upAllComponents() (err error) {

	componentBuildMap := map[string]func()error {
		"config": upConfigComponents,
		"redis": upRedisComponents,
		"database": upDatabaseComponents,
		//"datadog": buildDatadogComponents,
		"frontend": upFrontendComponents,
		"logflights": upLogflightsComponents,
		"celery": upCeleryComponents,
	}

	for component, componentBuildFunction := range componentBuildMap {
		fmt.Println("Building component ", component)
		err := componentBuildFunction()
		if err != nil {
			fmt.Printf("Build %v component failed", component)
			return err
		}
	}
	return nil
}

func upConfigComponents() (err error) {

	outputYamlFile := "deployments/config-generated.yaml"
	generateYamlCmd, err := GenerateCommand(kubetplRenderTemplate, KubetplRenderRequest{
		Manifest: "deployments/config.yaml",
		Env: []string{fmt.Sprintf("ENV=%v", env)},
	})

	applyYamlCmd, err := GenerateCommand(kubectlApplyTemplate, KubectlApplyRequest{
		YamlFile: outputYamlFile,
	})

	if err != nil {
		return err
	}

	var generatedYaml bytes.Buffer
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


func upDatabaseComponents() (err error) {

	outputYamlFile := "deployments/database-generated.yaml"

	//requiredVariables := []string{"POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB"}
	requiredVariables := []string{"LF_POSTGRES_PASSWORD"}
	envs, err := generateEnvs(requiredVariables)
	if err != nil {
		return err
	}

	generateYamlCmd, err := GenerateCommand(kubetplRenderTemplate, KubetplRenderRequest{
		Manifest: "deployments/database.yaml",
		EnvFrom: []string{"deployments/config-local.env"},
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


func upFrontendComponents() (err error) {

	outputYamlFile := "deployments/frontend-generated.yaml"

	requiredVariables := []string{}
	envs, err := generateEnvs(requiredVariables)
	if err != nil {
		return err
	}

	generateYamlCmd, err := GenerateCommand(kubetplRenderTemplate, KubetplRenderRequest{
		Manifest: "deployments/frontend.yaml",
		EnvFrom: []string{"deployments/config-local.env"},
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

func upLogflightsComponents() (err error) {

	outputYamlFile := "deployments/logflights-generated.yaml"

	requiredVariables := []string{"GOOGLE_MAPS_API_KEY", "LF_DB_PASS"}
	envs, err := generateEnvs(requiredVariables)
	if err != nil {
		return err
	}

	generateYamlCmd, err := GenerateCommand(kubetplRenderTemplate, KubetplRenderRequest{
		Manifest: "deployments/logflights.yaml",
		EnvFrom: []string{"deployments/config-local.env"},
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

func upRedisComponents() (err error) {


	outputYamlFile := "./deployments/redis-generated.yaml"

	requiredVariables := []string{}
	envs, err := generateEnvs(requiredVariables)
	if err != nil {
		return err
	}

	cmd, err := GenerateCommand(kubetplRenderTemplate, KubetplRenderRequest{
		Manifest: "./deployments/redis.yaml",
		Env: envs,
	})

	applyYamlCmd, err := GenerateCommand(kubectlApplyTemplate, KubectlApplyRequest{
		YamlFile: outputYamlFile,
	})

	if err != nil {
		return err
	}

	var generatedYaml bytes.Buffer
	cmd.Stdout = &generatedYaml
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
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

func upCeleryComponents() (err error) {


	outputYamlFile := "./deployments/celery-generated.yaml"

	requiredVariables := []string{"GOOGLE_MAPS_API_KEY", "LF_DB_PASS"}
	envs, err := generateEnvs(requiredVariables)
	if err != nil {
		return err
	}

	generateYamlCmd, err := GenerateCommand(kubetplRenderTemplate, KubetplRenderRequest{
		Manifest: "./deployments/celery.yaml",
		EnvFrom: []string{"deployments/config-local.env"},
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
