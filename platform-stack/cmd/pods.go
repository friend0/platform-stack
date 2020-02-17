package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
	"os"
	"text/template"
)

var (
	podSuccessConditions = []metav1.TableRowCondition{{Type: metav1.RowCompleted, Status: metav1.ConditionTrue, Reason: string(v1.PodSucceeded), Message: "The pod has completed successfully."}}
	podFailedConditions  = []metav1.TableRowCondition{{Type: metav1.RowCompleted, Status: metav1.ConditionTrue, Reason: string(v1.PodFailed), Message: "The pod failed."}}
)

// PodColumns defines a set of columns to print pod details row-wise
type PodColumns struct {
	Name           string
	Ready          string
	Status         string
	Restarts       int64
	Age            string
	IP             string
	Node           string
	Nominated      string
	ReadinessGates string
	Healthy        bool
	Namespace      string
	Images         []string
}

// podsCmd represents the pods command
var podsCmd = &cobra.Command{
	Use:   "pods",
	Short: "List running pods.",
	Long:  `List running pods.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return initK8s()
	},
	RunE: pods,
}

func pods(cmd *cobra.Command, args []string) (err error) {

	ns, _ := cmd.Flags().GetString("namespace")
	label, _ := cmd.Flags().GetStringSlice("label")
	field, _ := cmd.Flags().GetStringSlice("field")

	api := clientset.CoreV1()

	podList, err := getPodsList(api, ns, label, field)
	_, err = printPodList(podList, os.Stdout)
	if err != nil {
		return err
	}
	return nil
}

// getPodsList retrieves a PodList from the given namespace, labels, and fields
func getPodsList(api v12.CoreV1Interface, ns string, label, field []string) (list *v1.PodList, err error) {

	defaultLabel := viper.GetString("stack")

	labelSelect := ""
	if defaultLabel != "" {
		labelSelect = fmt.Sprintf("stack=%v", defaultLabel)
	}
	for _, elem := range label {
		labelSelect += elem
	}

	fieldSelect := ""
	for _, elem := range field {
		fieldSelect += elem
	}

	listOptions := metav1.ListOptions{
		LabelSelector: labelSelect,
		FieldSelector: fieldSelect,
	}

	pods, err := api.Pods(ns).List(listOptions)
	if err != nil {
		return pods, err
	}
	return pods, nil
}

// printPodList returns a set of rows for printing a PodList
// borrows heavily from https://github.com/kubernetes/kubernetes/blob/master/pkg/printers/internalversion/printers.go
func printPodList(podList *v1.PodList, out io.Writer) ([]PodColumns, error) {
	template2 := "{{ printf \"%-48v\" .Name }}{{ printf \"%-8v\" .Ready }}{{ printf \"%-16v\" .Status }}{{ printf \"%-16v\" .Restarts }}{{ printf \"%-16v\" .Age }}{{ printf \"%-16v\" .IP }}{{ printf \"%-24v\" .Node }}{{ printf \"%-16v\" .Nominated }}{{ printf \"%-16v\" .ReadinessGates }}{{ printf \"%v\" .Images }}\n"

	columnsTemplate := "%-48v%-8v%-16v%-16v%-16v%-16v%-24v%-16v%-16v%v\n"
	header := fmt.Sprintf(columnsTemplate, "NAME", "READY", "STATUS", "RESTARTS", "AGE", "IP", "NODE", "NOMINATED", "READINESS", "IMAGES")
	_, _ = fmt.Fprintf(out, header)

	rows := make([]PodColumns, 0, len(podList.Items))
	for i := range podList.Items {
		r, err := printPod(&podList.Items[i])
		if err != nil {
			return nil, err
		}

		t := template.Must(template.New("podColumns").Parse(template2))

		err = t.Execute(out, r)

		rows = append(rows, r)
	}

	return rows, nil
}

// printPod returns pod details as a PodColumn
// borrows heavily from https://github.com/kubernetes/kubernetes/blob/master/pkg/printers/internalversion/printers.go
func printPod(pod *v1.Pod) (podDetail PodColumns, err error) {

	healthy := true
	restarts := 0
	totalContainers := len(pod.Spec.Containers)
	readyContainers := 0

	reason := string(pod.Status.Phase)
	if pod.Status.Reason != "" {
		reason = pod.Status.Reason
	}

	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: pod},
	}

	switch pod.Status.Phase {
	case v1.PodSucceeded:
		row.Conditions = podSuccessConditions
	case v1.PodFailed:
		row.Conditions = podFailedConditions
	}

	initializing := false
	for i := range pod.Status.InitContainerStatuses {
		container := pod.Status.InitContainerStatuses[i]
		restarts += int(container.RestartCount)
		switch {
		case container.State.Terminated != nil && container.State.Terminated.ExitCode == 0:
			continue
		case container.State.Terminated != nil:
			// initialization is failed
			if len(container.State.Terminated.Reason) == 0 {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Init:Signal:%d", container.State.Terminated.Signal)
					healthy = false
				} else {
					reason = fmt.Sprintf("Init:ExitCode:%d", container.State.Terminated.ExitCode)
					if container.State.Terminated.ExitCode != 0 {
						healthy = false
					}
				}
			} else {
				reason = "Init:" + container.State.Terminated.Reason
			}
			initializing = true
		case container.State.Waiting != nil && len(container.State.Waiting.Reason) > 0 && container.State.Waiting.Reason != "PodInitializing":
			reason = "Init:" + container.State.Waiting.Reason
			initializing = true
		default:
			reason = fmt.Sprintf("Init:%d/%d", i, len(pod.Spec.InitContainers))
			initializing = true
		}
		break
	}

	if !initializing {
		restarts = 0
		hasRunning := false
		for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
			container := pod.Status.ContainerStatuses[i]

			restarts += int(container.RestartCount)
			if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
				reason = container.State.Waiting.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason != "" {
				reason = container.State.Terminated.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason == "" {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else if container.Ready && container.State.Running != nil {
				hasRunning = true
				readyContainers++
			}
		}

		// change pod status back to "Running" if there is at least one container still reporting as "Running" status
		if reason == "Completed" && hasRunning {
			reason = "Running"
		}
	} else {
		healthy = false
	}

	if readyContainers != totalContainers {
		healthy = false
	}

	if pod.DeletionTimestamp != nil && pod.Status.Reason == v1.TaintNodeUnreachable {
		reason = "Unknown"
		healthy = false
	} else if pod.DeletionTimestamp != nil {
		reason = "Terminating"
		healthy = false
	}

	images := make([]string, len(pod.Spec.Containers))
	for i, container := range pod.Spec.Containers {
		images[i] = container.Image
	}

	podDetail = PodColumns{
		Name:      pod.Name,
		Ready:     fmt.Sprintf("%d/%d", readyContainers, totalContainers),
		Status:    reason,
		Restarts:  int64(restarts),
		Age:       translateTimestampSince(pod.CreationTimestamp),
		Healthy:   healthy,
		Namespace: pod.Namespace,
		Images:    images,
	}

	if true {
		nodeName := pod.Spec.NodeName
		nominatedNodeName := pod.Status.NominatedNodeName
		podIP := ""
		if len(pod.Status.PodIPs) > 0 {
			podIP = pod.Status.PodIPs[0].IP
		}

		if podIP == "" {
			podIP = "<none>"
		}
		if nodeName == "" {
			nodeName = "<none>"
		}
		if nominatedNodeName == "" {
			nominatedNodeName = "<none>"
		}

		readinessGates := "<none>"
		if len(pod.Spec.ReadinessGates) > 0 {
			trueConditions := 0
			for _, readinessGate := range pod.Spec.ReadinessGates {
				conditionType := readinessGate.ConditionType
				for _, condition := range pod.Status.Conditions {
					if condition.Type == conditionType {
						if condition.Status == v1.ConditionTrue {
							trueConditions++
						}
						break
					}
				}
			}
			readinessGates = fmt.Sprintf("%d/%d", trueConditions, len(pod.Spec.ReadinessGates))
		}

		podDetail.IP = podIP
		podDetail.Node = nodeName
		podDetail.Nominated = nominatedNodeName
		podDetail.ReadinessGates = readinessGates

	}

	return podDetail, nil
}

func init() {
	rootCmd.AddCommand(podsCmd)

	podsCmd.Flags().StringP("namespace", "n", "", "Namespace")
	podsCmd.Flags().StringSliceP("label", "l", []string{}, "Label selector")
	podsCmd.Flags().StringSliceP("field", "f", []string{}, "Field selector")

}
