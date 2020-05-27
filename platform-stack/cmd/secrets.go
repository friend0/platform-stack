package cmd

import (
	"fmt"
	"github.com/spf13/viper"
	"os"

	//"os"

	"github.com/spf13/cobra"
)

const kubectlCreateRegistrySecretTemplate = `kubectl create secret docker-registry acr-service-principal --docker-server=https://{{ .ContainerRegistry }}.azurecr.io --docker-username={{ .ServicePrincipleID }} --docker-password={{ .ServicePrinciplePassword }}`

type KubectlCreateRegistrySecretRequest struct {
	ContainerRegistry string
	ServicePrincipleID string
	ServicePrinciplePassword string
}

// secretsCmd represents the secrets command
var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Utility command for distributing credentials with Kubernetes secrets.",
	Long: `Utility command for distributing credentials with Kubernetes secrets.

Makes securely distributed credentials available to the stack.
Available stock secrets:
- "registry": Creates "acr-service-principal" secret to be used as an imagePullSecret in Kubernetes manifests. Requires
"SERVICE_PRINCIPLE_ID" and "SERVICE_PRINCIPLE_PASSWORD" variables to be set in the host environment.
`,
	Args: cobra.ExactArgs(1),
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
	switch secretType {
	case "registry":
		return createRegistrySecret(cmd, args)
	default:
		return fmt.Errorf("error")
	}
}

func createRegistrySecret(cmd *cobra.Command, args []string) error {

	spid := viper.GetString("SERVICE_PRINCIPLE_ID")
	sppwd := viper.GetString("SERVICE_PRINCIPLE_PASSWORD")

	if spid == "" || sppwd == "" {
		return fmt.Errorf("SERVICE_PRINCIPLE_ID or SERVICE_PRINCIPLE_PASSWORD must be set in order to create a Registry secret")
	}

	createRegistrySecretCmd, err := GenerateCommand(kubectlCreateRegistrySecretTemplate, KubectlCreateRegistrySecretRequest{
		viper.GetString("registry"),
		spid,
		sppwd,
	})
	if err != nil {
		return err
	}

	// todo: remove env append if able
	createRegistrySecretCmd.Stdout = os.Stdout
	createRegistrySecretCmd.Stderr = os.Stderr
	fmt.Println("+%v", createRegistrySecretCmd)
	if err := createRegistrySecretCmd.Run(); err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(secretsCmd)
	secretsCmd.Flags().StringP("registry", "c", "airbusutm", "Name of registry referenced by secret")
}
