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

type InstallDescription struct {
	os      []string
	test    string
	install map[string][]string
}

var StackCLIDependencies = map[string]InstallDescription{
	"xcode": {
		os:   []string{"darwin"},
		test: "xcode-select -v",
		install: map[string][]string{
			"darwin": []string{"xcode-select --install"},
		},
	},
	"kubectl": {
		os:   []string{"darwin", "linux"},
		test: "kubectl",
		install: map[string][]string{
			"darwin": []string{
				"curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.17.0/bin/darwin/amd64/kubectl",
				"chmod +x ./kubectl",
				"sudo mv ./kubectl /usr/local/bin/kubectl"},
			"linux": []string{
				"curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.17.0/bin/linux/amd64/kubectl && chmod +x ./kubectl && sudo mv ./kubectl /usr/local/bin/kubectl",
			},
		},
	},
	"kubetpl": {
		os:   []string{"darwin", "linux"},
		test: "kubetpl",
		install: map[string][]string{
			"darwin": []string{
				`curl -sSL https://github.com/shyiko/kubetpl/releases/download/0.9.0/kubetpl-0.9.0-darwin-amd64 -o kubetpl`,
				"chmod a+x kubetpl",
				"sudo mv kubetpl /usr/local/bin/",
			},
			"linux": []string{
				`curl -sSL https://github.com/shyiko/kubetpl/releases/download/0.9.0/kubetpl-0.9.0-$(bash -c '[[ $OSTYPE == darwin* ]] && echo darwin || echo linux')-amd64 -o kubetpl && chmod a+x kubetpl && sudo mv kubetpl /usr/local/bin/`,
			},
		},
	},
}


func runInstall(cmd *cobra.Command, args []string) (err error) {
	fmt.Println("Installing development dependencies...")

	var installed []string
	goos := runtime.GOOS
	for dep, install := range StackCLIDependencies {
		for _, os := range install.os {
			if os == goos {
				if !dependencyExists(install.test) {
					// todo:
					fmt.Printf("Installing %v...\n", dep)
					installCmds, ok := install.install[os]
					if ok {
						err = installDependency(installCmds)
						if err != nil {
							return err
						}
						installed = append(installed, dep)
					}
				}
			}
		}
	}

	fmt.Printf("Installed dependencies: %v\n", installed)
	return nil
}

func dependencyExists(arg string) bool {
	err := exec.Command("sh", "-c", arg).Run()
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
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
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return fmt.Errorf("install exited with code %v", status.ExitStatus())
			}
		}
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(installCmd)
}
