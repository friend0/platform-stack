package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// TODO: set namespace? (`kubectl -n FOO`)
const kubectlLogsTemplate = `kubectl logs {{if .Stream}} -f {{end}} --all-containers=true deployments/{{ .Deployment}}`

type KubectlLogsRequest struct {
	Deployment string
	Stream bool
}

var (
	streamLogs bool
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs <component>",
	Short: "Show logs for a given running pod by short name",
	Long: `Show logs for a given running pod by short name.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runShowLogs,
}

func runShowLogs(cmd *cobra.Command, args []string) (err error) {
	fmt.Printf("Showing logs for %v\n", args[0])

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	forwardCmd, err := showLogs(args[0], streamLogs)
	if err != nil {
		return err
	}

	<-c
	if forwardCmd != nil {
		return forwardCmd.Process.Kill()
	}
	return nil
}

func showLogs(deployment string, streamLogs bool) (cmd *exec.Cmd, err error) {

	fetchLogsCmd, err := GenerateCommand(kubectlLogsTemplate, KubectlLogsRequest{
		Deployment: deployment,
		Stream: streamLogs,
	})

	if err != nil {
		return fetchLogsCmd, err
	}

	fetchLogsCmd.Stdout = os.Stdout
	fetchLogsCmd.Stderr = os.Stderr
	if err := fetchLogsCmd.Start(); err != nil {
		return fetchLogsCmd, err
	}
	forwardCmds[deployment] = fetchLogsCmd
	return fetchLogsCmd, nil
}



func init() {
	logsCmd.Flags().BoolVarP(&streamLogs, "follow", "f", false, "follow (stream) logs as they happen")
	rootCmd.AddCommand(logsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// logsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// logsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
