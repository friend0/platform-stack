package cmd

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	// See: https://github.com/kubernetes/client-go/blob/53c7adfd0294caa142d961e1f780f74081d5b15f/examples/out-of-cluster-client-configuration/main.go#L31
	// and https://github.com/kubernetes/client-go/issues/242
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"os"
	"os/exec"
	"text/template"
)

// alias for simple mocking in test. Do not remove
var execCommand = exec.Command

var stackConfigurationFile string
var stackConfigurationFileName string

var (
	clientset        *kubernetes.Clientset
	currentNamespace string
)

// config is the global configuration object made available to all root sub-commands.
// It has trivial values up until the `initConfig` function is run.
var config Config
var Version = "development"

type StackDescription struct {
	Name string
}

type ActivationDescription struct {
	ConfirmWithUser bool
	Env             string
	Context         string
}

type EnvironmentDescription struct {
	Name       string
	Activation ActivationDescription
}

type ComponentDescription struct {
	Name              string                 `json:"name"`
	RequiredVariables []string               `json:"required_variables"`
	Exposable         bool                   `json:"exposable"`
	Containers        []ContainerDescription `json:"containers"`
	Manifests         []string               `json:"manifests"`
	TemplateConfig    []string               `json:"template_config"`
}

type ContainerDescription struct {
	Dockerfile   string   `json:"dockerfile"`
	Context      string   `json:"context"`
	Image        string   `json:"image"`
	Environments []string `json:"environments"`
}

type ManifestDescription struct {
	Dockerfile string `json:"dockerfile"`
	Context    string `json:"context"`
	Image      string `json:"image"`
}

type Config struct {
	Components   []ComponentDescription
	Environments []EnvironmentDescription
	Stack        StackDescription
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "stack",
	Short: "Commands for building, deploying, and maintaining platform services.",
	Long:  `Commands for building, deploying, and maintaining platform services.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		version, err := cmd.Flags().GetBool("version")
		if err != nil {
			return err
		}
		if version {
			fmt.Printf("Stack CLI Version: %v\n", Version)
			return nil
		}
		return cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// todo do not allow config file path directly until project directory is appropriately overriden to reflect config's location
	//rootCmd.PersistentFlags().StringVar(&stackConfigurationFile, "config", "", "config file (default is $HOME/.{{name of project}}.yaml)")
	rootCmd.PersistentFlags().StringVar(&stackConfigurationFileName, "stack_configuration", ".stack-local", "set the name of the configuration file to be used")
	rootCmd.PersistentFlags().StringP("project_directory", "r", ".", "set the project directory of the stack")
	rootCmd.Flags().BoolP("version", "v", false, "Print the stack CLI version")
	viper.BindPFlag("project_directory", rootCmd.PersistentFlags().Lookup("project_directory"))
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if stackConfigurationFile != "" {
		viper.SetConfigFile(stackConfigurationFile)
	} else {
		configDirectory := viper.GetString("project_directory")
		viper.AddConfigPath(configDirectory)
		viper.SetConfigName(stackConfigurationFileName)
	}

	// todo: allow configurable env prefix
	viper.SetEnvPrefix(viper.GetString("env_prefix"))
	viper.AutomaticEnv() // read in environment variables that match

	viper.ReadInConfig()

	// Defaults
	viper.SetDefault("env", "local")
	viper.Unmarshal(&config)
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

func containsString(slice []string, element string) bool {
	for _, elem := range slice {
		if elem == element {
			return true
		}
	}
	return false
}

// initK8s initializes a global clientset object using the system KUBECONFIG, with default merging rules
func initK8s(kubectx string) (err error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	if kubectx != "" {
		configOverrides.CurrentContext = kubectx
	}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return err
	}
	currentNamespace, _, err = kubeConfig.Namespace()
	if err != nil {
		return err
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	for {
		if clientset != nil {
			break
		}
	}
	return nil
}
