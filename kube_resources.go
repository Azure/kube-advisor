package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/fatih/color"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

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
	if err != nil {
		log.Fatalln("failed to get deployments:", err)
	}

	type Container struct {
		Name           string
		ResourceCPU    bool
		ResourceMemory bool
		RequestCPU     bool
		RequestMemory  bool
	}

	offenders := make(map[string][]*Container)

	for _, d := range deployments.Items {
		containers := d.Spec.Template.Spec.Containers
		for _, c := range containers {
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
			if cpuLimitMissing || memoryLimitMissing {
				container := Container{c.Name, cpuLimitMissing, memoryLimitMissing, cpuRequestMissing, memoryRequestMissing}
				offenders[d.GetName()] = append(offenders[d.GetName()], &container)
			}
		}
	}

	fmt.Println("The following containers have no resources limits and may cause node resource starvation: ")
	for k, containers := range offenders {
		c := color.New(color.FgBlue, color.Underline, color.Bold)
		c.Printf("Deployment name: %s\n", k)
		for _, c := range containers {
			cc := color.New(color.Bold)
			cc.Printf("- Container: %s\n", c.Name)
			v := reflect.Indirect(reflect.ValueOf(c))
			// Purposely skip the first struct field, Name, which is a string
			for i := 1; i < 5; i++ {
				key := v.Type().Field(i).Name
				value := v.Field(i).Bool()
				if value == true {
					color.Red("%+v has no limit set", key)
				} else {
					color.Green("%+v limit is set", key)
				}
			}
		}
	}
}
