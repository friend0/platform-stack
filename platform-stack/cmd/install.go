package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"text/template"

	"github.com/spf13/cobra"
)

type DependencyVersionData struct {
	Version string
}

type DependencyDescription struct {
	os      []string
	version string
	test    string
	install map[string][]string
}

var dependencyVersionMap map[string]string

var StackCLIDependencies = map[string]DependencyDescription{
	"bats": {
		os:   []string{"darwin", "linux"},
		test: "bats -v",
		install: map[string][]string{
			"darwin": []string{
				"git clone --depth 1 --branch v1.5.0 https://github.com/bats-core/bats-core.git && cd bats-core && sudo ./install.sh /usr/local && cd - && rm -rf bats-core",
			},
			"linux": []string{
				"git clone --depth 1 --branch v1.5.0 https://github.com/bats-core/bats-core.git && cd bats-core && sudo ./install.sh /usr/local && cd - && rm -rf bats-core",
			},
		},
	},
	"xcode": {
		os:   []string{"darwin"},
		test: "xcode-select -v",
		install: map[string][]string{
			"darwin": []string{"xcode-select --install"},
		},
	},
	"gsm-buddy": {
		os:      []string{"darwin", "linux"},
		version: "v0.1.1",
		test:    "gsm-buddy",
		install: map[string][]string{
			"darwin": []string{
				`curl -sSL https://github.com/yamaszone/gcp-secret-manager-buddy/releases/download/{{ .Version }}/gcp-secret-manager-buddy-{{ .Version }}-darwin-amd64 -o gsm-buddy`,
				"chmod a+x gsm-buddy",
				"sudo mv gsm-buddy /usr/local/bin/",
			},
			"linux": []string{
				`curl -sSL https://github.com/yamaszone/gcp-secret-manager-buddy/releases/download/{{ .Version }}/gcp-secret-manager-buddy-{{ .Version }}-linux-amd64 -o gsm-buddy`,
				"chmod a+x gsm-buddy",
				"sudo mv gsm-buddy /usr/local/bin/",
			},
		},
	},
	"kubectl": {
		os:      []string{"darwin", "linux"},
		test:    "kubectl",
		version: getEnv("KUBECTL_VERSION", "v1.21.5"),
		install: map[string][]string{
			"darwin": []string{
				"curl -LO https://storage.googleapis.com/kubernetes-release/release/{{ .Version }}/bin/darwin/amd64/kubectl",
				"chmod +x ./kubectl",
				"sudo mv ./kubectl /usr/local/bin/kubectl"},
			"linux": []string{
				"curl -LO https://storage.googleapis.com/kubernetes-release/release/{{ .Version }}/bin/linux/amd64/kubectl && chmod +x ./kubectl && sudo mv ./kubectl /usr/local/bin/kubectl",
			},
		},
	},
	"kubetpl": {
		os:      []string{"darwin", "linux"},
		version: "0.9.0",
		test:    "kubetpl",
		install: map[string][]string{
			"darwin": []string{
				`curl -sSL https://github.com/shyiko/kubetpl/releases/download/{{ .Version }}/kubetpl-{{ .Version }}-darwin-amd64 -o kubetpl`,
				"chmod a+x kubetpl",
				"sudo mv kubetpl /usr/local/bin/",
			},
			"linux": []string{
				`curl -sSL https://github.com/shyiko/kubetpl/releases/download/{{ .Version }}/kubetpl-{{ .Version }}-linux-amd64 -o kubetpl`,
				"chmod a+x kubetpl",
				"sudo mv kubetpl /usr/local/bin/",
			},
		},
	},
	"minikube": {
		os:      []string{"darwin", "linux"},
		test:    "minikube",
		version: getEnv("MINIKUBE_VERSION", "v1.21.0"),
		install: map[string][]string{
			"darwin": []string{
				"curl -Lo minikube https://storage.googleapis.com/minikube/releases/{{ .Version }}/minikube-darwin-amd64 && chmod +x minikube && sudo cp minikube /usr/local/bin/ && rm minikube",
			},
			"linux": []string{
				"curl -Lo minikube https://storage.googleapis.com/minikube/releases/{{ .Version }}/minikube-linux-amd64 && chmod +x minikube && sudo cp minikube /usr/local/bin/ && rm minikube",
			},
		},
	},
	"tilt": {
		os:      []string{"darwin", "linux"},
		version: "v0.17.11",
		test:    "tilt version",
		install: map[string][]string{
			"darwin": []string{
				"curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash",
			},
			"linux": []string{
				"curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash",
			},
		},
	},
}

func parseDependencyVersionOverrides(dependencyVersions []string) (map[string]string, error) {
	dependencyVersionMap = make(map[string]string)
	for _, dependencyVersion := range dependencyVersions {
		split := strings.Split(dependencyVersion, "=")
		if len(split) != 2 || split[1] == "" {
			return dependencyVersionMap, fmt.Errorf("expecting dependency version as `dependency_name=dependency_version`: got `%v` instead", dependencyVersion)
		}
		dependencyVersionMap[split[0]] = split[1]
	}
	return dependencyVersionMap, nil
}

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs dependencies needed to run stack commands.",
	Long:  `Installs dependencies needed to run stack commands.`,
	Args: func(cmd *cobra.Command, args []string) error {
		dependencyVersions, _ := cmd.Flags().GetStringSlice("dependency_versions")
		dvm, err := parseDependencyVersionOverrides(dependencyVersions)
		if err != nil {
			return err
		}
		dependencyVersionMap = dvm
		return nil
	},
	RunE: runInstall,
}

func runInstall(cmd *cobra.Command, args []string) (err error) {

	dependencyVersions, _ := cmd.Flags().GetStringSlice("dependency_versions")
	for _, dependencyVersion := range dependencyVersions {
		strings.Split(dependencyVersion, "=")
	}

	dryRun, _ := cmd.Flags().GetBool("dryrun")
	if !dryRun {
		fmt.Println("Installing development dependencies...")
	} else {
		fmt.Println("[dry run] Installing development dependencies...")
	}

	installed, err := installDependencies(StackCLIDependencies, dryRun)
	if err != nil {
		return err
	}
	fmt.Printf("Install summary:\n%v", strings.Join(installed, ",\n"))
	return nil
}

func installDependencies(dependencies map[string]DependencyDescription, dryRun bool) (installed []string, err error) {

	goos := runtime.GOOS
	for dep, install := range dependencies {
		for _, osName := range install.os {
			if osName == goos {
				exists := dependencyExists(install.test)
				if exists {
					installed = append(installed, fmt.Sprintf("`%v` already exists - skipping install", dep))
				} else {
					installed = append(installed, fmt.Sprintf("will install %v", dep))
				}
				if !dryRun {
					if !exists {
						fmt.Printf("Installing %v...\n", dep)
						installDependencyCmds, ok := install.install[osName]
						if ok {
							dependencyVersion := install.version
							overrideVersion, dependencyVersionOverride := dependencyVersionMap[dep]
							if dependencyVersionOverride {
								dependencyVersion = overrideVersion
							}
							err = installDependency(installDependencyCmds, dependencyVersion)
							if err != nil {
								installed = append(installed, "failed installing %v")
								return installed, err
							}
							installed = append(installed, fmt.Sprintf("installed %v", dep))
						}
					}
				}
			}
		}
	}

	return installed, nil
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

func installDependency(args []string, version string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("no install args were provided")
	}
	for _, arg := range args {

		var installString bytes.Buffer
		tmpl, err := template.New("installCommand").Parse(arg)
		if err != nil {
			return err
		}
		err = tmpl.Execute(&installString, DependencyVersionData{version})
		if err != nil {
			return err
		}

		cmd := exec.Command("sh", "-c", installString.String())

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return fmt.Errorf("install exited with code %v", status.ExitStatus())
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().BoolP("dryrun", "d", false, "Select deployment environment")
	healthCmd.Flags().StringSliceP("dependency_versions", "d", []string{}, "Comma separated list of dependency version assignments as `dependency_name=version`")
}
