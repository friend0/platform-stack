package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
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

	// todo:
	deps := map[string]struct{
		os []string
		test      string
		install map[string][]string
	} {
		"xcode": {
			os: []string{"darwin"},
			test: "xcode-select -v",
			install: map[string][]string{
				"darwin": []string{"xcode-select --install"},
			},
		},
		"kubectl": {
			os: []string{"darwin", "linux"},
			test: "kubectl",
			install: map[string][]string{
				"darwin":[]string{
					"curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.17.0/bin/darwin/amd64/kubectl",
					"chmod +x ./kubectl",
					"sudo mv ./kubectl /usr/local/bin/kubectl"},
			},
		},
		"kubetpl": {
			os: []string{"darwin", "linux"},
			test: "kubetpl",
			install: map[string][]string{
				"darwin":[]string{
					`curl -sSL https://github.com/shyiko/kubetpl/releases/download/0.9.0/kubetpl-0.9.0-darwin-amd64 -o kubetpl`,
					"chmod a+x kubetpl",
					"sudo mv kubetpl /usr/local/bin/",
				},
				"linux":[]string{
					`curl -sSL https://github.com/shyiko/kubetpl/releases/download/0.9.0/kubetpl-0.9.0-linux-amd64 -o kubetpl`,
					"chmod a+x kubetpl",
					"sudo mv kubetpl /usr/local/bin/",
				},
			},
		},
	}

	var installed []string
	goos := runtime.GOOS
	for dep, install := range deps {
		for _, os := range install.os {
			if os == goos {
				if !dependencyExists(install.test) {
					err = installDependency(install.install[os])
					if err != nil {
						return err
					}
					installed = append(installed, dep)
				}
			}
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

func installDependency(args []string) (err error) {
	for _, arg := range args {
		cmd := exec.Command("sh", "-c", arg)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(installCmd)
}
