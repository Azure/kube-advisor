package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func checkContainer(c v1.Container) (StatusCheck, bool) {
	m := make(map[string]bool)
	sc := StatusCheck{ContainerName: c.Name, Missing: m}

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
		return StatusCheck{}, false
	}
	return sc, true
}

// StatusCheck represents a container and its resource and request limit status
type StatusCheck struct {
	ContainerName string
	Missing       map[string]bool
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

	statusChecksWrapper := make(map[string][]*StatusCheck)

	// Gather container statusChecksWrapper from Deployments
	for _, d := range deployments.Items {
		containers := d.Spec.Template.Spec.Containers
		for _, c := range containers {
			status, ok := checkContainer(c)
			if ok {
				statusChecksWrapper[d.GetName()] = append(statusChecksWrapper[d.GetName()], &status)
			}
		}
	}

	// Gather container statusChecksWrapper from StatefulSets
	for _, ss := range statefulsets.Items {
		containers := ss.Spec.Template.Spec.Containers
		for _, c := range containers {
			status, ok := checkContainer(c)
			if ok {
				statusChecksWrapper[ss.GetName()] = append(statusChecksWrapper[ss.GetName()], &status)
			}
		}
	}

	// Gather container statusChecksWrapper from DaemonSets
	for _, ds := range daemonsets.Items {
		containers := ds.Spec.Template.Spec.Containers
		for _, c := range containers {
			status, ok := checkContainer(c)
			if ok {
				statusChecksWrapper[ds.GetName()] = append(statusChecksWrapper[ds.GetName()], &status)
			}
		}
	}

	for k, statusChecks := range statusChecksWrapper {
		c := color.New(color.FgBlue, color.Underline, color.Bold)
		c.Printf("Deployment/DaemonSet/StatefulSet name: %s\n", k)
		for _, s := range statusChecks {
			cc := color.New(color.Bold)
			cc.Printf("Container: %s\n", s.ContainerName)
			for key := range s.Missing {
				color.Red("- %+v", key)
			}
		}
	}
}
