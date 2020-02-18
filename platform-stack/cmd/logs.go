package cmd

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const kubectlLogsTemplate = `kubectl logs {{if .Stream}} -f {{end}} {{if .ContainerName}}--container {{ .ContainerName }}{{else}}--all-containers=true{{end}} deployments/{{ .Deployment}}`

type KubectlLogsRequest struct {
	PodName string
	ContainerName string
	Stream     bool
}

var (
	streamLogs bool
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs <pod> [container]",
	Short: "Show logs for a given running pod / deployment by short name",
	Long:  `Show logs for a given running pod / deployment by short name.`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  showLogs,
}

func showLogs(cmd *cobra.Command, args []string) (err error) {

	api := clientset.CoreV1()

	ns, _ := cmd.Flags().GetString("namespace")
	label, _ := cmd.Flags().GetStringSlice("label")
	field, _ := cmd.Flags().GetStringSlice("field")

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
	if len(args) >= 2 {
		targetContainerName = args[1]
	} else {
		targetContainerName = targetPod.Spec.Containers[0].Name
	}

	fmt.Printf("Showing logs for %v\n", args[0])

	fetchLogsCmd, err := GenerateCommand(kubectlLogsTemplate, KubectlLogsRequest{
		PodName: args[0],
		ContainerName: targetContainerName,
		Stream: streamLogs,
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
	logsCmd.Flags().StringP("namespace", "", "", "Namespace")
	logsCmd.Flags().StringSliceP("label", "", []string{}, "Label selector")
	logsCmd.Flags().StringSliceP("field", "", []string{}, "Field selector")

}
