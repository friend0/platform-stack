package cmd

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const kubectlLogsTemplate = `kubectl logs {{if .Stream}} -f {{end}} {{if .ContainerName}}--container {{ .ContainerName }}{{else}}--all-containers=true{{end}} {{ .PodName}}`

type KubectlLogsRequest struct {
	PodName       string
	ContainerName string
	Stream        bool
}

var (
	streamLogs bool
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs <deployment> [container]",
	Short: "Show logs for a pod of the given k8s deployment (or a container in it).",
	Long:  `Show logs for a pod of the given k8s deployment (or a container in it).`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return configPreRunnerE(cmd, args)
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return initK8s("")
	},
	Args: cobra.RangeArgs(1, 2),
	RunE: showLogs,
}

func showLogs(cmd *cobra.Command, args []string) (err error) {

	api := clientset.CoreV1()

	ns, _ := cmd.Flags().GetString("namespace")
	label, _ := cmd.Flags().GetStringSlice("label")
	field, _ := cmd.Flags().GetStringSlice("field")

	// add deployment name to labels
	label = append(label, fmt.Sprintf("app=%v", args[0]))

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
	var podContainerNames []string

	if len(args) == 2 {
		targetContainerName = args[1]
		podContainerNames = append(podContainerNames, targetContainerName)
	} else {
		podContainerNames = make([]string, len(targetPod.Spec.Containers))
		for i, container := range targetPod.Spec.Containers {
			podContainerNames[i] = container.Name
		}
		if len(podContainerNames) == 1 {
			targetContainerName = podContainerNames[0]
		} else {
			// default to showing all containers
			targetContainerName = ""
		}
	}

	fmt.Printf("Showing logs for pod %v [containers: %v]\n", targetPod.Name, strings.Join(podContainerNames, ", "))

	fetchLogsCmd, err := GenerateCommand(kubectlLogsTemplate, KubectlLogsRequest{
		PodName:       targetPod.Name,
		ContainerName: targetContainerName,
		Stream:        streamLogs,
	})

	if err != nil {
		return err
	}

	fetchLogsCmd.Stdout = os.Stdout
	fetchLogsCmd.Stderr = os.Stderr
	if err := fetchLogsCmd.Run(); err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(logsCmd)

	logsCmd.Flags().BoolVarP(&streamLogs, "follow", "f", false, "stream (follow) logs as they happen")
	logsCmd.Flags().String("namespace", "", "Namespace")
	logsCmd.Flags().StringSlice("label", []string{}, "Label selector")
	logsCmd.Flags().StringSlice("field", []string{}, "Field selector")
}
