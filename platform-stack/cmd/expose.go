package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

const kubectlExposeTemplate = `kubectl port-forward deployments/{{ .Deployment}} {{ .LocalPort }}:{{ .RemotePort }}`

type KubectlExposeRequest struct {
	Deployment string
	LocalPort  string
	RemotePort string
}

var forwardCmds = map[string]*exec.Cmd{}

// exposeCmd represents the expose command
var exposeCmd = &cobra.Command{
	Use:   "expose <component> <local port> <remote port>",
	Short: "Exposes a kubernetes deployment to your local machine.",
	Long:  `Exposes a kubernetes deployment to your local machine.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("expecting exactly three arguments: see `stack expose help`")
		}
		return nil
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return configPreRunnerE(cmd, args)
	},
	RunE: runExpose,
}

func runExpose(cmd *cobra.Command, args []string) (err error) {

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

	fmt.Printf("Exposing %v", args[0])

	// FIXME: do we need a larger channel?
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	forwardCmd, err := exposeDeployment(args[0], args[1], args[2])
	if err != nil {
		return err
	}

	<-c
	if forwardCmd != nil {
		return forwardCmd.Process.Kill()
	}
	return nil
}

func exposeDeployment(deployment, localPort, remotePort string) (cmd *exec.Cmd, err error) {

	exposeCmd, err := GenerateCommand(kubectlExposeTemplate, KubectlExposeRequest{
		Deployment: deployment,
		LocalPort:  localPort,
		RemotePort: remotePort,
	})

	if err != nil {
		return exposeCmd, err
	}

	exposeCmd.Stdout = os.Stdout
	exposeCmd.Stderr = os.Stderr
	if err := exposeCmd.Start(); err != nil {
		return exposeCmd, err
	}
	forwardCmds[deployment] = exposeCmd
	return exposeCmd, nil
}

func init() {
	rootCmd.AddCommand(exposeCmd)
}
