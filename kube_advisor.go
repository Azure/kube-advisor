package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/olekukonko/tablewriter"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metrics "k8s.io/metrics/pkg/client/clientset_generated/clientset"
)

func checkContainer(c v1.Container, p v1.Pod, pm v1beta1.PodMetrics) (PodStatusCheck, bool) {
	sc := PodStatusCheck{
		ContainerName: c.Name,
		PodName:       p.Name,
		Missing:       make(map[string]bool),
	}

	for _, c := range pm.Containers {
		sc.PodCPU = c.Usage.Cpu().String()
		sc.PodMemory = c.Usage.Memory().String()
	}

	if c.Resources.Limits.Cpu().IsZero() {
		sc.Missing["CPU Resource Limits Missing"] = true
	}
	if c.Resources.Limits.Memory().IsZero() {
		sc.Missing["Memory Resource Limits Missing"] = true
	}
	if c.Resources.Requests.Cpu().IsZero() {
		sc.Missing["CPU Request Limits Missing"] = true
	}
	if c.Resources.Requests.Memory().IsZero() {
		sc.Missing["Memory Request Limits Missing"] = true
	}
	if len(sc.Missing) == 0 {
		return PodStatusCheck{}, false
	}
	return sc, true
}

// PodStatusCheck represents a container and its resource and request limit status
type PodStatusCheck struct {
	PodName       string
	ContainerName string
	PodCPU        string
	PodMemory     string
	Missing       map[string]bool
}

type NodeStatusCheck struct {
	NodeName   string
	NodeCPU    string
	NodeMemory string
}

func main() {
	kubePtr := flag.Bool("use-kubeconfig", false, "use kubeconfig on local system")
	flag.Parse()

	var kubeconfig string

	if *kubePtr == true {
		kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	} else {
		kubeconfig = ""
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)

	if err != nil {
		log.Fatal(err)
	}

	// Create an rest client not targeting specific API version
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	metricClient, err := metrics.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	statusChecksWrapper := make(map[string][]*PodStatusCheck)

	pods, _ := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		log.Fatalln("failed to get pods:", err)
	}
	for _, p := range pods.Items {
		containers := p.Spec.Containers
		for _, c := range containers {
			podMetricsList, _ := metricClient.MetricsV1beta1().PodMetricses("").List(metav1.ListOptions{})
			for _, pm := range podMetricsList.Items {
				if p.Name == pm.Name {
					status, ok := checkContainer(c, p, pm)
					if ok {
						statusChecksWrapper[p.Namespace] = append(statusChecksWrapper[p.Namespace], &status)
					}
				}
			}
		}
	}

	var nodeStatuses []*NodeStatusCheck
	nodeMetricsList, _ := metricClient.MetricsV1beta1().NodeMetricses().List(metav1.ListOptions{})
	for _, nm := range nodeMetricsList.Items {
		ns := NodeStatusCheck{NodeName: nm.Name, NodeCPU: nm.Usage.Cpu().String(), NodeMemory: nm.Usage.Memory().String()}
		nodeStatuses = append(nodeStatuses, &ns)
	}

	nodeTable := tablewriter.NewWriter(os.Stdout)
	nodeTable.SetHeader([]string{"Node", "Node CPU Usage", "Node Memory Usage"})
	nodeTable.SetHeaderColor(tablewriter.Colors{tablewriter.Bold, tablewriter.BgBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgBlackColor})
	nodeTable.SetAutoMergeCells(true)
	nodeTable.SetRowLine(true)
	for _, ns := range nodeStatuses {
		nodeTable.Append([]string{ns.NodeName, ns.NodeCPU, ns.NodeMemory})
	}

	issuesTable := tablewriter.NewWriter(os.Stdout)
	for k, statusChecks := range statusChecksWrapper {
		issuesTable.SetHeader([]string{"Namespace", "Pod Name", "Pod CPU/Memory", "Container", "Issue"})
		issuesTable.SetHeaderColor(tablewriter.Colors{tablewriter.Bold, tablewriter.BgBlackColor},
			tablewriter.Colors{tablewriter.Bold, tablewriter.BgBlackColor},
			tablewriter.Colors{tablewriter.Bold, tablewriter.BgBlackColor},
			tablewriter.Colors{tablewriter.Bold, tablewriter.BgBlackColor},
			tablewriter.Colors{tablewriter.Bold, tablewriter.BgBlackColor})
		issuesTable.SetAutoMergeCells(true)
		issuesTable.SetRowLine(true)
		for _, s := range statusChecks {
			for key := range s.Missing {
				resourceString := fmt.Sprintf("%v / %v", s.PodCPU, s.PodMemory)
				issuesTable.Append([]string{k, s.PodName, resourceString, s.ContainerName, key})
			}
		}
	}

	remediationTable := tablewriter.NewWriter(os.Stdout)
	remediationTable.SetHeader([]string{"Issue", "Remediation"})
	remediationTable.SetHeaderColor(tablewriter.Colors{tablewriter.Bold, tablewriter.BgBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgBlackColor})
	remediationTable.SetAutoMergeCells(true)
	remediationTable.Append([]string{"CPU Resource Requests Missing", "Consider setting resource and request limits to prevent resource starvation: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/"})
	remediationTable.Append([]string{"Memory Resource Requests Missing", "Consider setting resource and request limits to prevent resource starvation: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/"})
	remediationTable.Append([]string{"CPU Resource Limits Missing", "Consider setting resource and request limits to prevent resource starvation: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/"})
	remediationTable.Append([]string{"Memory Resource Limits Missing", "Consider setting resource and request limits to prevent resource starvation: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/"})

	issuesTable.Render()
	nodeTable.Render()
	remediationTable.Render()
}
