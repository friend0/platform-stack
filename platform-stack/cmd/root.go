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
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.internal-logflights-new.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {

		dir, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)

		}

		viper.AddConfigPath(dir)
		viper.SetConfigName(".stack")
	}

	viper.SetEnvPrefix(viper.GetString("env_prefix"))
	viper.AutomaticEnv() // read in environment variables that match

	fmt.Println("config file used:", viper.ConfigFileUsed())
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

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
