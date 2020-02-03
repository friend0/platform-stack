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
	fmt.Println("install called")
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
