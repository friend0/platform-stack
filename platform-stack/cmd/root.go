package cmd

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"os"
	"os/exec"
	"text/template"
)

// alias for simple mocking in test. Do not remove
var execCommand = exec.Command

var cfgFile string
var clientset *kubernetes.Clientset

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "stack",
	Short: "Commands for building, deploying, and maintaining platform services.",
	Long:  `Commands for building, deploying, and maintaining platform services.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.{{name of project}}.yaml)")
	rootCmd.PersistentFlags().StringP("project_directory", "r", ".", "set the project directory for stack command")
	viper.BindPFlag("project_directory", rootCmd.PersistentFlags().Lookup("project_directory"))

	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {

		configDirectory := viper.GetString("project_directory")
		if configDirectory != "." {
			viper.AddConfigPath(configDirectory)
		} else {
			dir, err := os.Getwd()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)

			}
			viper.AddConfigPath(dir)
		}
		viper.SetConfigName(".stack")
	}

	viper.SetEnvPrefix(viper.GetString("env_prefix"))
	viper.AutomaticEnv() // read in environment variables that match

	viper.ReadInConfig()

	// Defaults
	viper.SetDefault("deployment_directory", "./deployments")
	viper.SetDefault("build_directory", "./build")

}

// GenerateCommandString builds a non-executable command string
func GenerateCommandString(tmpl string, data interface{}) (cmd string, err error) {
	var templateBytes bytes.Buffer

	funcMap := template.FuncMap{
		"minus_one": func(i int) int {
			return i - 1
		},
	}

	parsedTemplate, err := template.New("").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", err
	}

	err = parsedTemplate.Execute(&templateBytes, data)
	if err != nil {
		return "", err
	}

	return templateBytes.String(), nil
}

// GenerateCommand returns an executable command from an input template, and corresponding data interface.
func GenerateCommand(tmpl string, data interface{}) (cmd *exec.Cmd, err error) {

	result, err := GenerateCommandString(tmpl, data)
	if err != nil {
		return nil, err
	}
	cmd = execCommand("sh", "-c", result)
	return cmd, err
}

// confirmWithUser ensures an action with confirmation from user input
func confirmWithUser(confirmationText string) (confirmation bool) {

	var response string

	affirmative := []string{"y", "Y", "yes", "Yes", "YES"}
	negative := []string{"n", "N", "no", "No", "NO"}

	if confirmationText != "" {
		fmt.Printf("%v - are you sure you want to proceed?", confirmationText)
	}

	_, err := fmt.Scanln(&response)
	if err != nil {
		return false
	}

	if containsString(affirmative, response) {
		return true
	} else if containsString(negative, response) {
		return false
	} else {
		fmt.Println("Please type yes or no and then press enter:")
		return confirmWithUser(confirmationText)
	}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}


func containsString(slice []string, element string) bool {
	for _, elem := range slice {
		if elem == element {
			return true
		}
	}
	return false
}
