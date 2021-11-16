package cmd

import (
	"encoding/base64"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

const gsmFetchSecretTemplate = `GOOGLE_APPLICATION_CREDENTIALS={{ .ServiceAccountFile }} gsm-buddy get -i {{ .SecretInputFile }} -p {{ .GcpProjectId }} > {{ .SecretOutputFile }} 2>error.log`
const gsmAccessSecretTemplate = `gcloud secrets versions access latest --secret="gsm-secret-reader-global-blue" --project={{ .GcpProjectId }} > {{ .GsmReaderJsonFile }}`

type GsmAccessSecretsRequest struct {
	GcpProjectId      string
	GsmReaderJsonFile string
}

type GsmFetchSecretsRequest struct {
	GcpProjectId       string
	ServiceAccountFile string
	SecretInputFile    string
	SecretOutputFile   string
}

var secretsFetchCmd = &cobra.Command{
	Use:   "fetch [-e <env>] [-p <gcp-project-id>] [-i <input-file-directory>] [-o <output-file-directory>]",
	Short: "Fetch secrets for the secret IDs in the input file.",
	Long: `Fetch secrets for the secret IDs in the input file.
Example:
	Input: cat deployments/secret-ids-ci.json:
	{
		"GCR_ACCESS_KEY_PUSH": "gcr-utmgcr-push-circleci-blue",
		"IBM_WEATHER_API_KEY": "platform-weather-local-default-IBM_WEATHER_API_KEY"
	}

	$ stack secrets fetch -e ci -p utmgsmdev -i deployments -o deployments
	Output: cat deployments/secrets-ci.json:
	{"GCR_ACCESS_KEY_PUSH": "***", "IBM_WEATHER_API_KEY":"***"}
	`,
	Args: cobra.MaximumNArgs(4),
	RunE: fetchSecrets,
}

func fetchSecrets(cmd *cobra.Command, args []string) error {
	env, _ := cmd.Flags().GetString("env")
	project, _ := cmd.Flags().GetString("project")
	input, _ := cmd.Flags().GetString("input")
	output, _ := cmd.Flags().GetString("output")
	sa, _ := cmd.Flags().GetString("service-account")
	saFlavor, _ := cmd.Flags().GetString("sa-version")

	if env == "local" || os.Getenv("USE_GSM_IAM_ROLE") == "yes" {
		accessSecretsCmd, err := GenerateCommand(gsmAccessSecretTemplate, GsmAccessSecretsRequest{
			GcpProjectId:      project,
			GsmReaderJsonFile: sa,
		})
		if err != nil {
			return err
		}
		// For debugging only
		//fmt.Println(accessSecretsCmd)

		accessSecretsCmd.Stdout = os.Stdout
		accessSecretsCmd.Stderr = os.Stderr
		if err := accessSecretsCmd.Run(); err != nil {
			return err
		}
	} else {
		secretReaderEnvVar := ""
		if env == "local" || env == "ci" {
			secretReaderEnvVar = "GSM_SECRET_READER_DEV_" + strings.ToUpper(saFlavor)
		} else if env == "prev" || env == "preview" {
			secretReaderEnvVar = "GSM_SECRET_READER_PREV" + strings.ToUpper(saFlavor)
		} else if env == "stg" || env == "staging" {
			secretReaderEnvVar = "GSM_SECRET_READER_STG" + strings.ToUpper(saFlavor)
		} else if env == "prod" || env == "production" {
			secretReaderEnvVar = "GSM_SECRET_READER_PROD" + strings.ToUpper(saFlavor)
		} else {
			return fmt.Errorf("no GCP Secret Manager Project configured for the target environment.")
		}

		if os.Getenv(secretReaderEnvVar) == "" {
			return fmt.Errorf("GSM_SECRET_READER_{DEV|PREV|STG|PROD}_{BLUE|GREEN} not set.")
		}

		saKey, err := base64.StdEncoding.DecodeString(os.Getenv(secretReaderEnvVar))
		if err != nil {
			return fmt.Errorf("failed to decode GSM reader service account key:", err)
		}

		if _, err := os.Stat(sa); os.IsNotExist(err) {
			_, err := os.Create(sa)
			if err != nil {
				return fmt.Errorf("failed to create a file to save GSM reader service account key:", err)
			}
		}

		err = os.WriteFile(sa, saKey, 0644)
		if err != nil {
			return fmt.Errorf("failed to save GSM reader service account key:", err)
		}
	}

	request := GsmFetchSecretsRequest{
		GcpProjectId:       project,
		ServiceAccountFile: sa,
		SecretInputFile:    fmt.Sprintf("%s/secret-ids-%s.json", input, env),
		SecretOutputFile:   fmt.Sprintf("%s/secrets-%s.json", output, env),
	}

	fetchSecretsCmd, err := GenerateCommand(gsmFetchSecretTemplate, request)
	if err != nil {
		return err
	}
	// For debugging only
	//fmt.Println(fetchSecretsCmd)

	fetchSecretsCmd.Stdout = os.Stdout
	fetchSecretsCmd.Stderr = os.Stderr
	if err := fetchSecretsCmd.Run(); err != nil {
		return err
	}

	return nil
}

func init() {
	secretsCmd.AddCommand(secretsFetchCmd)
	secretsFetchCmd.Flags().StringP("env", "e", "local", "Deployment target (e.g. local, ci, prod, etc.)")
	secretsFetchCmd.Flags().StringP("project", "p", "utmgsmdev", "GCP Project ID for Secret Manager (e.g. utmgsmdev, utmgsmstg, utmgsm, etc.)")
	secretsFetchCmd.Flags().StringP("input", "i", "deployments", "Directory for the secret ID manifest file (manifest file needs to be named as: 'secret-ids-<env>.json')")
	secretsFetchCmd.Flags().StringP("output", "o", "deployments", "Directory for the output file (to be stored as 'secrets-<env>.json')")
	secretsFetchCmd.Flags().StringP("service-account", "s", "/tmp/gsm-secret-reader.json", "Path to the GSM Reader service account")
	secretsFetchCmd.Flags().StringP("sa-version", "v", "blue", "Service account version flavor (blue|green) used as a postfix for environment variable 'GSM_SECRET_READER_<e>_<v>' to allow rotation")
}
