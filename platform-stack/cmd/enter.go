package cmd

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	v1 "k8s.io/api/core/v1"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const kubectlExecTemplate = `kubectl exec -it {{ .PodName }}{{if .ContainerName}} --container {{ .ContainerName }}{{end}} {{ .Command}}`

type KubectlExecRequest struct {
	PodName string
	ContainerName string
	Command   string
}

// enterCmd represents the enter command
var enterCmd = &cobra.Command{
	Use:   "enter <pod> [container]",
	Args:  cobra.MinimumNArgs(1),
	Short: "Initiates a terminal session for a pod in the given Deployment.",
	Long:  `Initiates a terminal session for a pod in the given Deployment.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		initK8s()
	},
	RunE: enterPod,
}

func enterPod(cmd *cobra.Command, args []string) (err error) {

	// todo: wrap this common functionality into a helper in pods
	ns, _ := cmd.Flags().GetString("namespace")
	label, _ := cmd.Flags().GetStringSlice("label")
	field, _ := cmd.Flags().GetStringSlice("field")

	api := clientset.CoreV1()

	// todo: warn on missing default label
	defaultLabel := viper.GetString("stack")

	labelSelect := ""
	if defaultLabel != "" {
		labelSelect = fmt.Sprintf("stack=%v", defaultLabel)
	}
	if args[0] != "" {
		labelSelect = fmt.Sprintf("app=%v", args[0])
	}
	for _, elem := range label {
		labelSelect += elem
	}

	fieldSelect := ""
	for _, elem := range field {
		fieldSelect += elem
	}

	podList, err := getPodsList(api, ns, labelSelect, fieldSelect)
	if err != nil {
		return err
	}

	// error if no pods, or multiple matching pods are found
	if len(podList.Items) == 0 {
		return fmt.Errorf("no pods matching labels %v", labelSelect)
	}
	if len(podList.Items) > 1 {
		var matchingPods []string
		for _, pod := range podList.Items {
			matchingPods = append(matchingPods, pod.Name)
		}
		return fmt.Errorf("multiple pods matching given app label: %v", strings.Join(matchingPods, ", "))
	}


	targetPod := podList.Items[0]

	var targetContainer v1.Container

	// handle pod with many containers
	containers := targetPod.Spec.Containers
	if len(containers) > 1 {
		var containerList []string
		var containerPresent bool

		for _, container := range containers {
			containerList = append(containerList, container.Name)
			if len(args) > 1 {
				if args[1] == container.Name {
					targetContainer = container
					containerPresent = true
				}
			}
		}

		if len(args) <= 1 {
			return fmt.Errorf("multiple containers for the given pod: %v: please provide a container name as an additional argument", strings.Join(containerList, ", "))
		} else {
			if !containerPresent {
				return fmt.Errorf("provided container is not present in pod")
			}

		}
	} else if len(containers) == 1 {
		targetContainer = containers[0]
	}


	// determine target shell if one was not provided
	targetShell, _ := cmd.Flags().GetString("shell")
	if targetShell == "" {
		availableShells, err := getAvailableShells(&targetPod, &targetContainer)
		if err !=  nil {
			return err
		}
		if len(availableShells) < 1 {
			return fmt.Errorf("could not locate any available shells")
		}
		targetShell = availableShells[0]
		fmt.Printf("available shells: %v: using first available: %v\n", strings.Join(availableShells, ", "), targetShell)
	}

	generateExecCmd, err := GenerateCommand(kubectlExecTemplate, KubectlExecRequest{
		PodName: targetPod.Name,
		ContainerName: targetContainer.Name,
		Command:   targetShell,
	})

	if err != nil {
		return err
	}

	generateExecCmd.Stdin = os.Stdin
	generateExecCmd.Stdout = os.Stdout
	generateExecCmd.Stderr = os.Stderr

	fmt.Println(generateExecCmd)
	if err := generateExecCmd.Start(); err != nil {
		return err
	}

	return generateExecCmd.Wait()
}


func getAvailableShells(targetPod *v1.Pod, targetContainer *v1.Container) (shells []string, err error) {
	// cat /etc/shells
	getShells, err := GenerateCommand(kubectlExecTemplate, KubectlExecRequest{
		PodName: targetPod.Name,
		ContainerName: targetContainer.Name,
		Command:   "cat /etc/shells | grep /",
	})
	if err != nil {
		return shells, err
	}

	var shellsBuf bytes.Buffer
	getShells.Stdout = &shellsBuf

	if err := getShells.Run(); err != nil {
		return shells, err
	}
	fmt.Println("SHELBUF")
	fmt.Println(strings.Fields(shellsBuf.String()), len(shellsBuf.String()))
	return strings.Fields(shellsBuf.String()), nil
}

func init() {
	rootCmd.AddCommand(enterCmd)

	enterCmd.Flags().StringP("shell", "s", "", "Provide a target shell.")
	enterCmd.Flags().StringP("pod", "p", "", "Provide a specific pod to enter.")

	enterCmd.Flags().StringP("namespace", "n", "", "Namespace")
	enterCmd.Flags().StringSliceP("label", "l", []string{}, "Label selector")
	enterCmd.Flags().StringSliceP("field", "f", []string{}, "Field selector")
}
