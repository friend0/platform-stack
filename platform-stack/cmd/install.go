package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs dependencies needed to run stack commands.",
	Long:  `Installs dependencies needed to run stack commands.`,
	RunE:  runInstall,
}

func runInstall(cmd *cobra.Command, args []string) (err error) {
	fmt.Println("Installing development dependencies...")
	deps := map[string]string{
		"xcode-select -v": "xcode-select --install",
		"brew --version": `usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"`,
		"kubectl": `brew install kubectl`,
		"kubetpl": `curl -sSL https://github.com/shyiko/kubetpl/releases/download/0.9.0/kubetpl-0.9.0-"$(bash -c '[[ $OSTYPE == darwin* ]] && echo darwin || echo linux')"-amd64 -o kubetpl && chmod a+x kubetpl && sudo mv kubetpl /usr/local/bin/`,
	}

	var installed []string
	for dep, installCmd := range deps {
		if !dependencyExists(dep) {
			err = installDependency(installCmd)
			if err != nil {
				return err
			}
			installed = append(installed, dep)
		}
	}

	fmt.Printf("Installed %v dependencies\n", len(installed))
	return nil
}

func dependencyExists(arg string) bool {
	err := exec.Command("sh", "-c", arg).Run()

	if exiterr, ok := err.(*exec.ExitError); ok {
		// The program has exited with an exit code != 0
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			//return false
			fmt.Println(status.ExitStatus())
			return status.ExitStatus() == 0
		}
	}

	return true
}

func installDependency(arg string) (err error) {
	cmd := exec.Command("sh", "-c", arg)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	return err
}

func init() {
	rootCmd.AddCommand(installCmd)
}
