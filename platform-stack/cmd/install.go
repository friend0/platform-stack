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
	"xcode": {
		os:   []string{"darwin"},
		test: "xcode-select -v",
		install: map[string][]string{
			"darwin": []string{"xcode-select --install"},
		},
	},
	"kubectl": {
		os:      []string{"darwin", "linux"},
		test:    "kubectl",
		version: "v1.17.0",
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
				`curl -sSL https://github.com/shyiko/kubetpl/releases/download/{{ .Version }}/kubetpl-{{ .Version }}-$(bash -c '[[ $OSTYPE == darwin* ]] && echo darwin || echo linux')-amd64 -o kubetpl && chmod a+x kubetpl && sudo mv kubetpl /usr/local/bin/`,
			},
		},
	},
}

func parseDependencyVersionOverrides(dependencyVersions []string) (map[string]string, error) {
	dependencyVersionMap = make(map[string]string)
	if len(dependencyVersions) == 0 {
		return dependencyVersionMap, fmt.Errorf("no dependency version assignments provided")
	}
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
					installed = append(installed, fmt.Sprintf("will not install %v", dep))
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
							installed = append(installed, "installed %v")
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

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().BoolP("dryrun", "d", false, "Select deployment environment")
	healthCmd.Flags().StringSliceP("dependency_versions", "d", []string{}, "Comma separated list of dependency version assignments as `dependency_name=version`")
}
