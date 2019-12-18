package cmd

import (
	"bytes"
	"fmt"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"text/template"
)

// alias for simple mocking in test. Do not remove
var execCommand = exec.Command

var cfgFile string

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
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.internal-logflights-new.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".internal-logflights-new" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".internal-logflights-new")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
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

	fmt.Println("confirming with user")
	var response string

	affirmative := []string{"y", "Y", "yes", "Yes", "YES"}
	negative := []string{"n", "N", "no", "No", "NO"}

	if confirmationText != "" {
		fmt.Printf("%v - ", confirmationText)
	}
	fmt.Println("are you sure you want to proceed?")

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

	return confirmation
}