package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

const kubectlDeleteSecretTemplate = `kubectl delete secrets {{if .SecretName}} {{- .SecretName -}} {{end}} -l stack={{ .StackName}}`

type KubectlDeleteSecretsRequest struct {
	SecretName string
	StackName  string
}

// environmentListCmd represents the environmentList command
var secretsDeleteCmd = &cobra.Command{
	Use:   "delete [secretType]",
	Short: "Delete the named stock secret.",
	Long:  `Delete the named stock secret.`,
	Args: cobra.MaximumNArgs(1),
	RunE:  deleteSecret,
}

func deleteSecret(cmd *cobra.Command, args []string) error {
	request := KubectlDeleteSecretsRequest{
		"",
		config.Stack.Name,
	}

	fmt.Println(len(args))
	if len(args) > 0 {
		fmt.Println(args[0])
		secretType, ok := secretTypesSecretNamesMap[args[0]]
		if ok {
			request.SecretName = secretType
		} else {
			return fmt.Errorf("invalid secret type specified")
		}

	} else {
		confirmWithUser("you are about to delete all secrets for the stack")
	}

	deleteSecretsCmd, err := GenerateCommand(kubectlDeleteSecretTemplate, request)
	if err != nil {
		return err
	}

	// todo: remove env append if able
	deleteSecretsCmd.Stdout = os.Stdout
	deleteSecretsCmd.Stderr = os.Stderr
	if err := deleteSecretsCmd.Run(); err != nil {
		return err
	}

	return nil
}

func init() {
	secretsCmd.AddCommand(secretsDeleteCmd)
}
