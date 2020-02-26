package cmd

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
)

const kubectlContextTemplate = `kubectl config {{ .ConfigCommand }}`

type kubectlContextRequest struct {
	ConfigCommand 	string
}

// contextCmd represents the context command
var contextCmd = &cobra.Command{
	Use:   "context [target]",
	Short: "Get or set the current active kubectx.",
	Long: `Get or set the current active kubectx
If no args are provided, the current context is retrieved. 
If a target argument is provided, then stack will activate the configured context for the environment with name matching target.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return initK8s("")
	},
	RunE:  func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			currentContext := getContext()
			if currentContext == "" {
				return fmt.Errorf("no context was returned")
			}
			fmt.Println(currentContext)
		} else {
			return setContext(args[0])
		}
		return nil
	},
}

// runContextCommandFunction returns a cobra command handler for running kubectl-config commands
func runContextCommandFunction(configCommand string) func (cmd *cobra.Command, args []string) (err error){
	return func(cmd *cobra.Command, args []string) (err error) {
		return runContextFunction(configCommand, os.Stdout)()
	}
}
// runContextFunction returns a function handler for running kubectl-config commands
func runContextFunction(configCommand string, out io.Writer) func ()(error){
	return func()error{
		contextsCmd, err := GenerateCommand(kubectlContextTemplate, kubectlContextRequest{
			ConfigCommand: configCommand,
		})
		if err != nil {
			return err
		}
		contextsCmd.Stdout = out
		contextsCmd.Stderr = os.Stderr
		if err := contextsCmd.Run(); err != nil {
			return err
		}
		return nil
	}
}

// getContext uses kubectl to retrieve the current kubectx
func getContext() string {
	var currentContext bytes.Buffer
	err := runContextFunction(fmt.Sprintf("current-context"), &currentContext)()
	if err != nil {
		return ""
	}
	return strings.TrimRight(currentContext.String(), "\n")
}

// setContext uses kubectl to activate the provided targetContext
func setContext(targetContext string) error {
	return runContextFunction(fmt.Sprintf("use-context %v", targetContext), os.Stdout)()
}

func init() {
	rootCmd.AddCommand(contextCmd)
}
