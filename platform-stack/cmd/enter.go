package cmd

import (
	"bytes"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

const kubectlExecTemplate = `kubectl exec -it {{ .PodName }}{{if .ContainerName}} --container {{ .ContainerName }}{{end}} {{ .Command}}`

type KubectlExecRequest struct {
	PodName       string
	ContainerName string
	Command       string
}

// enterCmd represents the enter command
var enterCmd = &cobra.Command{
	Use:   "enter <deployment> [container]",
	Args:  cobra.MinimumNArgs(1),
	Short: "Initiates a terminal session to a container in a pod of the given k8s deployment",
	Long:  `Initiates a terminal session to a container in a pod of the given k8s deployment`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return initK8s("")
	},
	RunE: enter,
}

func enter(cmd *cobra.Command, args []string) (err error) {

	api := clientset.CoreV1()

	ns, _ := cmd.Flags().GetString("namespace")
	label, _ := cmd.Flags().GetStringSlice("label")
	field, _ := cmd.Flags().GetStringSlice("field")

	//defaultLabel := viper.GetString("stack")
	//if defaultLabel != "" {
	//	label = append(label, fmt.Sprintf("stack=%v", defaultLabel))
	//}

	pods, err := getPodsList(api, ns, label, field)
	if err != nil {
		return err
	}

	// locate target pod
	var targetPod *v1.Pod
	if len(pods.Items) == 1 {
		targetPod = &pods.Items[0]
	} else {
		if len(pods.Items) == 0 {
			return fmt.Errorf("no pods matching labels %v", label)
		}
		matchingPods := make([]string, len(pods.Items))
		for i, pod := range pods.Items {
			matchingPods[i] = pod.Name
		}
		return fmt.Errorf("multiple pods matching given app label: %v", strings.Join(matchingPods, ", "))
	}

	var targetContainerName string
	targetShell, _ := cmd.Flags().GetString("shell")
	if len(args) >= 2 {
		targetContainerName = args[1]
	} else {
		targetContainerName = targetPod.Spec.Containers[0].Name
	}

	enterCmd, err := enterContainerCommand(targetPod, targetContainerName, targetShell)
	if err != nil {
		return err
	}
	if err := enterCmd.Start(); err != nil {
		return err
	}
	return enterCmd.Wait()

}

// enterContainer will attempt to use the given shell session on the given container in the given pod
func enterContainerCommand(pod *v1.Pod, containerName, shell string) (cmd *exec.Cmd, err error) {

	var targetContainer *v1.Container

	// locate target container
	containers := make(map[string]*v1.Container, len(pod.Spec.Containers))
	containerList := make([]string, len(pod.Spec.Containers))
	for i, container := range pod.Spec.Containers {
		if i == 0 {
			targetContainer = &container
		}
		containerList[i] = container.Name
		containers[container.Name] = &container
	}

	targetContainer = containers[containerName]

	if len(containers) != 1 {

		if len(containers) == 0 {
			return cmd, fmt.Errorf("no containers found in the given pod")
		}

		if containerName == "" {
			return cmd, fmt.Errorf("multiple containers for the given pod: %v: please provide a container name as an additional argument", strings.Join(containerList, ", "))
		}

		container, ok := containers[containerName]
		if ok {
			targetContainer = container
		} else {
			return cmd, fmt.Errorf("no container matching `%v`: containers found for the given pod: %v", containerName, strings.Join(containerList, ", "))
		}
	}

	// determine target shell if one was not provided
	if shell == "" {
		availableShells, err := getAvailableShells(pod, targetContainer)
		if err != nil {
			return cmd, err
		}
		shell = availableShells[0]
		fmt.Printf("available shells: %v: using first available: %v\n", strings.Join(availableShells, ", "), shell)
	}

	// generate and run kubectl exec command
	generateExecCmd, err := GenerateCommand(kubectlExecTemplate, KubectlExecRequest{
		PodName:       pod.Name,
		ContainerName: targetContainer.Name,
		Command:       shell,
	})
	if err != nil {
		return cmd, err
	}

	generateExecCmd.Stdin = os.Stdin
	generateExecCmd.Stdout = os.Stdout
	generateExecCmd.Stderr = os.Stderr
	return generateExecCmd, nil
}

// getAvailableShells takes an input pod and container and returns a list of shells that are available
// relies on /etc/shells
func getAvailableShells(targetPod *v1.Pod, targetContainer *v1.Container) (shells []string, err error) {
	getShells, err := GenerateCommand(kubectlExecTemplate, KubectlExecRequest{
		PodName:       targetPod.Name,
		ContainerName: targetContainer.Name,
		Command:       "cat /etc/shells | grep /",
	})
	if err != nil {
		return shells, err
	}

	var shellsBuf bytes.Buffer
	getShells.Stdout = &shellsBuf

	if err := getShells.Run(); err != nil {
		return shells, err
	}
	shells = strings.Fields(shellsBuf.String())
	if len(shells) < 1 {
		return shells, fmt.Errorf("could not locate any available shells")
	}
	return shells, nil
}

func init() {
	rootCmd.AddCommand(enterCmd)

	enterCmd.Flags().StringP("shell", "s", "", "Provide a target shell.")

	enterCmd.Flags().String("namespace", "", "Namespace")
	enterCmd.Flags().StringSlice("label", []string{}, "Label selector")
	enterCmd.Flags().StringSlice("field", []string{}, "Field selector")
}
