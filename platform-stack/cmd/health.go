package cmd

import (
	"context"
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/spf13/cobra"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

// podsCmd represents the pods command
var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Get the health of the stack.",
	Long:  `List running pods.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return initK8s()
	},
	RunE: health,
}

func health(cmd *cobra.Command, args []string) (err error) {

	api := clientset.CoreV1()

	ns, _ := cmd.Flags().GetString("namespace")
	label, _ := cmd.Flags().GetStringSlice("label")
	field, _ := cmd.Flags().GetStringSlice("field")

	podList, err := getPodsList(api, ns, label, field)
	if err != nil {
		return err
	}
	if len(podList.Items) == 0 {
		return fmt.Errorf("no pods found")
	}

	_, err = printPodListHealth(podList, os.Stdout)
	return err
}

func printPodListHealth(pods *v1.PodList, out io.Writer) (podsHealthy bool, err error) {

	podsHealthy = true
	var podsMeta []PodColumns
	unhealthyPodsMap := make(map[string]*v1.Pod)
	for i := range pods.Items {
		podMeta, _ := printPod(&pods.Items[i])
		podsMeta = append(podsMeta, podMeta)
		if !podMeta.Healthy {
			unhealthyPodsMap[(&pods.Items[i]).Name] = &pods.Items[i]
			podsHealthy = false
		}
	}

	if podsHealthy {
		_, _ = fmt.Fprintf(out, "All pods are healthy\n")
	} else {
		_, _ = fmt.Fprintf(out, "Not all pods are healthy\n")
	}
	for _, podDetail := range podsMeta {
		if podDetail.Healthy {
			_, _ = fmt.Fprintf(out, "✔️  %v in namespace `%v` is healthy\n", podDetail.Name, podDetail.Namespace)
		} else {
			podsHealthy = false
			_, _ = fmt.Fprintf(out, "✖️  %v in namespace `%v` is not healthy\n", podDetail.Name, podDetail.Namespace)
			if !podDetail.Healthy {
				podDetailHeader := fmt.Sprintf("\n\tPod Details `%v`\n", podDetail.Name)
				_, _ = fmt.Fprintf(out, podDetailHeader)
				_, _ = fmt.Fprintf(out, "\t%v\n", strings.Repeat("=", utf8.RuneCountInString(podDetailHeader)))

				for _, condition := range unhealthyPodsMap[podDetail.Name].Status.Conditions {
					_, _ = fmt.Fprintf(out, "\t%v: %v\n", condition.Type, condition.Status)
				}

				for _, container := range unhealthyPodsMap[podDetail.Name].Status.ContainerStatuses {
					containerDetailHeader := fmt.Sprintf("\n\tContainer Details `%v`\n", container.Name)
					_, _ = fmt.Fprintf(out, containerDetailHeader)
					if container.State.Waiting != nil || container.State.Terminated != nil {
						_, _ = fmt.Fprintf(out, "\t%v\n", strings.Repeat("=", utf8.RuneCountInString(containerDetailHeader)))
					}
					if container.State.Waiting != nil {
						_, _ = fmt.Fprintf(out, "\tContainer Waiting: %v\n", container.State.Waiting.Message)
					}
					if container.State.Terminated != nil {
						_, _ = fmt.Fprintf(out, "\tContainer Terminated with non-zero ExitCode: %v: %v\n", container.State.Terminated.ExitCode, container.State.Terminated.Message)
					}
				}
			}
		}
	}
	return podsHealthy, nil
}

func waitForStackWithTimeout(api v12.CoreV1Interface, cmd *cobra.Command, timeoutMs time.Duration) (err error, ctx context.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), timeoutMs*time.Millisecond)
	defer cancel() // releases resources if slowOperation completes before timeout elapses
	err = waitForStack(api, cmd, ctx)
	return err, ctx
}

func waitForStack(api v12.CoreV1Interface, cmd *cobra.Command, ctx context.Context) (err error) {

	ns, _ := cmd.Flags().GetString("namespace")
	label, _ := cmd.Flags().GetStringSlice("label")
	field, _ := cmd.Flags().GetStringSlice("field")

	backoffConfig := backoff.NewExponentialBackOff()

	backoffConfig.Multiplier = 2
	backoffConfig.MaxInterval = 10 * time.Second
	ticker := backoff.NewTicker(backoffConfig)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			podList, err := getPodsList(api, ns, label, field)
			if err != nil {
				return err
			}
			null, err := os.Open(os.DevNull) // For read access.
			if err != nil {
				return err
			}
			healthy, err := printPodListHealth(podList, null)
			if err != nil {
				return err
			}
			if healthy {
				return nil
			}
		}
	}

}

func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}

func init() {
	rootCmd.AddCommand(healthCmd)

	healthCmd.Flags().BoolP("wide", "w", true, "Wide cell")

	healthCmd.Flags().String("namespace", "", "Namespace")
	healthCmd.Flags().StringSlice("label", []string{}, "Label selectors")
	healthCmd.Flags().StringSlice("field", []string{}, "Field selectors")
}
