package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
	"strings"
	"time"
	"unicode/utf8"
)

// podsCmd represents the pods command
var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Get the health of the stack.",
	Long:  `List running pods.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		initK8s()
	},
	RunE: health,
}

func health(cmd *cobra.Command, args []string) (err error) {

	api := clientset.CoreV1()

	ns, _ := cmd.Flags().GetString("namespace")
	label, _ := cmd.Flags().GetStringSlice("label")
	field, _ := cmd.Flags().GetStringSlice("field")

	podList, err := getPodsList(api, ns, label, field)
	fmt.Println(podHealth(podList))
	return nil
}

// podHealth generates a report string given an input PodList
func podHealth(pods *v1.PodList) (output []string) {

	for _, pod := range pods.Items {
		healthy := true
		numContainersHealthy := 0

		podDetailHeader := fmt.Sprintf("\n\tPod Details `%v`\n", pod.Name)
		podDetailOutput := podDetailHeader
		podDetailOutput += fmt.Sprintf("\t%v\n", strings.Repeat("=", utf8.RuneCountInString(podDetailHeader)))

		for _, condition := range pod.Status.Conditions {
			podDetailOutput += fmt.Sprintf("\t%v: %v\n", condition.Type, condition.Status)
		}

		// check container numbers
		for _, container := range pod.Status.ContainerStatuses {
			if container.Ready {
				numContainersHealthy++
			}

			if container.State.Waiting != nil || container.State.Terminated != nil {
				containerDetailHeader := fmt.Sprintf("\n\tContainer Details `%v`\n", container.Name)
				podDetailOutput += containerDetailHeader
				podDetailOutput += fmt.Sprintf("\t%v\n", strings.Repeat("=", utf8.RuneCountInString(containerDetailHeader)))
			}
			if container.State.Waiting != nil {
				podDetailOutput += fmt.Sprintf("\tContainer Waiting: %v\n", container.State.Waiting.Message)
				healthy = false
			}
			if container.State.Terminated != nil {
				if container.State.Terminated.ExitCode == 0 {
					numContainersHealthy++
				} else {
					podDetailOutput += fmt.Sprintf("\tContainer Terminated with non-zero ExitCode: %v: %v\n", container.State.Terminated.ExitCode, container.State.Terminated.Message)
					healthy = false
				}
			}
		}

		if numContainersHealthy != len(pod.Spec.Containers) {
			healthy = false
		}

		if healthy {
			output = append(output, fmt.Sprintf("✔️  %v in namespace `%v` is healthy\n", pod.Name, pod.Namespace))
		} else {
			output = append(output, fmt.Sprintf("✖️  %v in namespace `%v` is not healthy\n", pod.Name, pod.Namespace)+podDetailOutput)
		}
	}
	return output
}

func waitForStackWithTimeout(api v12.CoreV1Interface, cmd *cobra.Command, timeoutMs time.Duration) (results []string, err error, ctx context.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), timeoutMs*time.Millisecond)
	defer cancel() // releases resources if slowOperation completes before timeout elapses
	results, err = waitForStack(api, cmd, ctx)
	return results, err, ctx
}

func waitForStack(api v12.CoreV1Interface, cmd *cobra.Command, ctx context.Context) (results []string, err error) {

	ns, _ := cmd.Flags().GetString("namespace")
	label, _ := cmd.Flags().GetStringSlice("label")
	field, _ := cmd.Flags().GetStringSlice("field")

	podList, err := getPodsList(api, ns, label, field)
	if err != nil {
		return results, err
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Println("cancelled from sig")
			return results, ctx.Err()
		default:
			fmt.Println("looping...")
			results = podHealth(podList)
			allReady := true
			for _, item := range results {
				notHealthy := strings.Contains(item, "not healthy")
				if notHealthy {
					allReady = false
				}
			}
			if allReady {
				return results, nil
			}
			time.Sleep(10 * time.Second)
		}
	}

}

func init() {
	rootCmd.AddCommand(healthCmd)

	healthCmd.Flags().StringP("namespace", "n", "", "Namespace")
	healthCmd.Flags().StringSliceP("label", "l", []string{}, "Label selectors")
	healthCmd.Flags().StringSliceP("field", "f", []string{}, "Field selectors")
}
