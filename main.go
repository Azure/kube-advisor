package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var ns string
	flag.StringVar(&ns, "namespace", "", "namespace")

	// Bootstrap k8s configuration from local 	Kubernetes config file
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	log.Println("Using kubeconfig file: ", kubeconfig)
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
	if err != nil {
		log.Fatalln("failed to get deployments:", err)
	}

	type Container struct {
		Name                       string
		ResourceCPUMissingLimit    bool
		ResourceMemoryMissingLimit bool
	}

	offenders := make(map[string][]*Container)

	for _, d := range deployments.Items {
		// fmt.Printf("[%s] %s * %d replicas\n", d.GetNamespace(), d.GetName(), *d.Spec.Replicas)
		// fmt.Printf("[%s] %s * %s containers\n", d.GetNamespace(), d.GetName(), d.Spec.Template.Spec.Containers)
		containers := d.Spec.Template.Spec.Containers
		for _, c := range containers {
			cpuLimitMissing := false
			memoryLimitMissing := false
			if c.Resources.Limits.Cpu().IsZero() {
				cpuLimitMissing = true
			}
			if c.Resources.Limits.Memory().IsZero() {
				memoryLimitMissing = true
			}
			if cpuLimitMissing || memoryLimitMissing {
				container := Container{c.Name, cpuLimitMissing, memoryLimitMissing}
				offenders[d.GetName()] = append(offenders[d.GetName()], &container)
			}
		}
	}

	fmt.Println("The following containers have no resources limits and may cause node resource starvation: ")
	for k, v := range offenders {
		fmt.Println("-----")
		c := color.New(color.FgBlue, color.Underline, color.Bold)
		c.Printf("Deployment name: %s\n", k)
		for _, c := range v {
			cc := color.New(color.Bold)
			cc.Printf("- Container: %s\n", c.Name)
			if c.ResourceCPUMissingLimit {
				color.Red("cpuLimitMissing: %v", c.ResourceCPUMissingLimit)
			} else {
				color.Green("cpuLimitMissing: %v", c.ResourceCPUMissingLimit)
			}
			if c.ResourceMemoryMissingLimit {
				color.Red("memoryLimitMissing: %v", c.ResourceMemoryMissingLimit)
			} else {
				color.Green("memoryLimitMissing: %v", c.ResourceMemoryMissingLimit)
			}
		}
	}
}
