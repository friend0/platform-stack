package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

const kubectlLogsTemplate = `kubectl logs {{if .Stream}} -f {{end}} --all-containers=true deployments/{{ .Deployment}}`

type KubectlLogsRequest struct {
	Deployment string
	Stream     bool
}

var (
	streamLogs bool
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs <component>",
	Short: "Show logs for a given running pod / deployment by short name",
	Long:  `Show logs for a given running pod / deployment by short name.`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  showLogs,
}

func showLogs(cmd *cobra.Command, args []string) (err error) {
	fmt.Printf("Showing logs for %v\n", args[0])

	fetchLogsCmd, err := GenerateCommand(kubectlLogsTemplate, KubectlLogsRequest{
		Deployment: args[0],
		Stream:     streamLogs,
	})

	if err != nil {
		return err
	}

	fetchLogsCmd.Stdout = os.Stdout
	fetchLogsCmd.Stderr = os.Stderr
	if err := fetchLogsCmd.Run(); err != nil {
		return err
	}
	return nil
}

func init() {
	logsCmd.Flags().BoolVarP(&streamLogs, "follow", "f", false, "follow (stream) logs as they happen")
	rootCmd.AddCommand(logsCmd)

}
