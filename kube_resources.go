package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/fatih/color"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func checkContainer(c v1.Container) LimitCheckStatus {
	var cpuLimitMissing, memoryLimitMissing, cpuRequestMissing, memoryRequestMissing bool

	if c.Resources.Limits.Cpu().IsZero() {
		cpuLimitMissing = true
	}
	if c.Resources.Limits.Memory().IsZero() {
		memoryLimitMissing = true
	}
	if c.Resources.Requests.Cpu().IsZero() {
		cpuRequestMissing = true
	}
	if c.Resources.Requests.Memory().IsZero() {
		memoryRequestMissing = true
	}
	return LimitCheckStatus{c.Name, cpuLimitMissing, memoryLimitMissing, cpuRequestMissing, memoryRequestMissing}
}

// LimitCheckStatus represents a container and its resource and request limit status
type LimitCheckStatus struct {
	ContainerName  string
	ResourceCPU    bool
	ResourceMemory bool
	RequestCPU     bool
	RequestMemory  bool
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

	deployments, err := clientset.AppsV1().Deployments("").List(metav1.ListOptions{})
	daemonsets, err := clientset.AppsV1().DaemonSets("").List(metav1.ListOptions{})
	statefulsets, err := clientset.AppsV1().StatefulSets("").List(metav1.ListOptions{})
	if err != nil {
		log.Fatalln("failed to get deployments:", err)
	}

	statuses := make(map[string][]*LimitCheckStatus)

	// Gather container statuses from Deployments
	for _, d := range deployments.Items {
		containers := d.Spec.Template.Spec.Containers
		for _, c := range containers {
			status := checkContainer(c)
			statuses[d.GetName()] = append(statuses[d.GetName()], &status)
		}
	}

	// Gather container statuses from StatefulSets
	for _, ss := range statefulsets.Items {
		containers := ss.Spec.Template.Spec.Containers
		for _, c := range containers {
			status := checkContainer(c)
			statuses[ss.GetName()] = append(statuses[ss.GetName()], &status)
		}
	}

	// Gather container statuses from DaemonSets
	for _, ds := range daemonsets.Items {
		containers := ds.Spec.Template.Spec.Containers
		for _, c := range containers {
			status := checkContainer(c)
			statuses[ds.GetName()] = append(statuses[ds.GetName()], &status)
		}
	}

	for k, limitStatuses := range statuses {
		c := color.New(color.FgBlue, color.Underline, color.Bold)
		c.Printf("Deployment/DaemonSet/StatefulSet name: %s\n", k)
		for _, s := range limitStatuses {
			cc := color.New(color.Bold)
			cc.Printf("Container: %s\n", s.ContainerName)
			v := reflect.Indirect(reflect.ValueOf(s))
			// Purposely skip the first struct field, Name, which is a string
			for i := 1; i < 5; i++ {
				key := v.Type().Field(i).Name
				value := v.Field(i).Bool()
				if value == true {
					color.Red("- %+v has no limit set and may cause node resource starvation", key)
				} else {
					color.Green("- %+v limit is set", key)
				}
			}
		}
	}
}
