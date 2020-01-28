package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

const kubectlLogsTemplate = `kubectl logs --all-containers=true deployment/{{ .Deployment}}`

type KubectlLogsRequest struct {
	Deployment string
}

// pod, container, namespace?

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs <component>",
	Short: "Show logs for a given running pod by short name",
	Long: `Show logs for a given running pod by short name.`,
	Args: func(cmd *cobra.Command, args []string) error {

		if len(args) != 1 {
			return fmt.Errorf("expecting argument <component>: see `stack logs help`")
		}

		err := viper.Unmarshal(&config)
		if err != nil {
			return err
		}

		if len(config.Components) < 1 {
			return fmt.Errorf("no configured components")
		}

		for idx, component := range config.Components {
			if component.Name == args[0] {
				if !component.Exposable {
					return fmt.Errorf("component not exposable")
				}
				break
			}
			if idx >= len(config.Components)-1 {
				return fmt.Errorf("component not found")
			}
		}
		return nil
	},
	RunE: runShowLogs,
}

func runShowLogs(cmd *cobra.Command, args []string) (err error) {
	fmt.Printf("Showing logs for %v", args[0])

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	forwardCmd, err := showLogs(args[0])
	if err != nil {
		return err
	}

	<-c
	if forwardCmd != nil {
		return forwardCmd.Process.Kill()
	}
	return nil
}

func showLogs(deployment string) (cmd *exec.Cmd, err error) {

	fetchLogsCmd, err := GenerateCommand(kubectlLogsTemplate, KubectlLogsRequest{
		Deployment: deployment,
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
	rootCmd.AddCommand(logsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// logsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// logsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
