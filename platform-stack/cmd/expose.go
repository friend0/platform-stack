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
	LocalPort string
	RemotePort string
}


var forwardCmds = map[string]*exec.Cmd {
	"frontend": nil,
	"logflights": nil,
	"redis": nil,
}

// exposeCmd represents the expose command
var exposeCmd = &cobra.Command{
	Use:   "expose",
	Short: "Exposes a kubernetes deployment to your local machine.",
	Long: `Exposes a kubernetes deployment to your local machine.`,
	Args: func(cmd *cobra.Command, args []string) error {

		if len(args) != 3 {
			return fmt.Errorf("Expecting exactly three args")
		}

		// todo: filter only exposable components
		components, _ := parseComponentArgs(args)
		for idx, component := range components {

			if component.Name == args[0] {
				break
			}
			if idx >= len(components) - 1 {
				return fmt.Errorf("component is not exposable")
			}
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		fmt.Printf("Exposing %v", args[0])

		// todo: can make cleanup async by tracking PIDs in leveldb
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		err = exposeDeployment(args[0], args[1], args[2])
		if err != nil {
			return err
		}

		<-c
		forwardCmd, ok := forwardCmds[args[0]]
		if ok && forwardCmd != nil {
			fmt.Println()
			return forwardCmd.Process.Kill()
		}

		return nil
	},
}

func exposeDeployment(deployment, localPort, remotePort string) (err error) {

	exposeCmd, err := GenerateCommand(kubectlExposeTemplate, KubectlExposeRequest{
		Deployment: deployment,
		LocalPort: localPort,
		RemotePort: remotePort,
	})

	if err != nil {
		return err
	}

	exposeCmd.Stdout = os.Stdout
	exposeCmd.Stderr = os.Stderr
	if err := exposeCmd.Start(); err != nil {
		return err
	}
	forwardCmds[deployment] = exposeCmd
	return nil
}


func init() {
	rootCmd.AddCommand(exposeCmd)
}
