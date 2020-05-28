package cmd

import (
	"fmt"
	"github.com/spf13/viper"
	"os"

	//"os"

	"github.com/spf13/cobra"
)

var secretTypesSecretNamesMap = map[string]string{
	"registry": "acr-service-principal",
}

const kubectlCreateRegistrySecretTemplate = `kubectl create secret docker-registry {{ .SecretName }} --docker-server=https://{{ .ContainerRegistry }}.azurecr.io --docker-username={{ .ServicePrincipleID }} --docker-password={{ .ServicePrinciplePassword }}/
											 kubectl label secret acr-service-principal stack={{ .StackName }}`

const kubectlGetSecretTemplate = `kubectl get secrets -l stack={{ .StackName}}`

type KubectlCreateRegistrySecretsRequest struct {
	SecretName 				 string
	ContainerRegistry        string
	ServicePrincipleID       string
	ServicePrinciplePassword string
	StackName                string
}

type KubectlListRegistrySecretsRequest struct {
	StackName                string
}


// secretsCmd represents the secrets command
var secretsCmd = &cobra.Command{
	Use:   "secrets [secretType]",
	Short: "Utility command for distributing credentials with Kubernetes secrets.",
	Long: `Utility command for distributing credentials with Kubernetes secrets.

Makes securely distributed credentials available to the stack.

Available SecretTypes:
- registry: specify the registry type for creating the 'acr-service-principal' for manifest imagePullSecrets
"SERVICE_PRINCIPLE_ID" and "SERVICE_PRINCIPLE_PASSWORD" variables to be set in the host environment.
`,
	Args: cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := viper.BindPFlag("registry", cmd.Flags().Lookup("registry"))
		if err != nil {
			return err
		}
		return nil
	},
	RunE: createSecret,
}

func createSecret(cmd *cobra.Command, args []string) error {
	secretType := args[0]
	if len(args) > 0 {
		switch secretType {
		case "registry":
			return createRegistrySecret(cmd, args)
		default:
			return fmt.Errorf("error")
		}
	} else {
		return listRegistrySecret(cmd, args)
	}
}

func labelSecret(cmd *cobra.Command, args []string) error {
	return nil
}

func createRegistrySecret(cmd *cobra.Command, args []string) error {
	secretName, _ := secretTypesSecretNamesMap[args[0]]

	spid := viper.GetString("SERVICE_PRINCIPLE_ID")
	sppwd := viper.GetString("SERVICE_PRINCIPLE_PASSWORD")

	if spid == "" || sppwd == "" {
		return fmt.Errorf("SERVICE_PRINCIPLE_ID or SERVICE_PRINCIPLE_PASSWORD must be set in order to create a Registry secret")
	}

	createRegistrySecretCmd, err := GenerateCommand(kubectlCreateRegistrySecretTemplate, KubectlCreateRegistrySecretsRequest{
		secretName,
		viper.GetString("registry"),
		spid,
		sppwd,
		config.Stack.Name,
	})
	if err != nil {
		return err
	}

	// todo: remove env append if able
	createRegistrySecretCmd.Stdout = os.Stdout
	createRegistrySecretCmd.Stderr = os.Stderr
	if err := createRegistrySecretCmd.Run(); err != nil {
		return err
	}

	return nil
}

func listRegistrySecret(cmd *cobra.Command, args []string) error {

	getSecretsCmd, err := GenerateCommand(kubectlGetSecretTemplate, KubectlListRegistrySecretsRequest{
		config.Stack.Name,
	})
	if err != nil {
		return err
	}

	// todo: remove env append if able
	getSecretsCmd.Stdout = os.Stdout
	getSecretsCmd.Stderr = os.Stderr
	if err := getSecretsCmd.Run(); err != nil {
		return err
	}

	return nil
}


func init() {
	rootCmd.AddCommand(secretsCmd)
	secretsCmd.Flags().StringP("registry", "c", "airbusutm", "Name of registry referenced by secret")
}
