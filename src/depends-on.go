package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"strings"
	"time"
)

type arrayFlags []string
func (i *arrayFlags) String() string {
	return strings.Join(*i, ",")
}
func (i *arrayFlags) Set(value string) error {
	*i = append(*i, strings.TrimSpace(value))
	return nil
}

var jobs arrayFlags
var services arrayFlags
var checkInterval int


func main() {
	flag.Var(&jobs, "job", "Job, which successfull completion to wait for. Can be specified multiple times")
	flag.Var(&services, "service", "Service, which pods to wait for. Can be specified multiple times")
	flag.IntVar(&checkInterval, "check_interval", 5, "Seconds to wait between check attempts")
	flag.Parse()

	if len(jobs) == 0 && len(services) == 0 {
		fmt.Println("No jobs or services provided. Exiting...")
		os.Exit(0)
	}

	nsb, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		panic(err.Error())
	}
	ns := string(nsb)
	fmt.Printf("Determined namespace: %s\n", ns)

	fmt.Println("Creating client...")
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	//TODO: import revision
	//TODO: rewrite with watch?

	//jobs
	for _, job_name := range jobs {
		fmt.Printf("Getting '%s' job object...\n", job_name)
		job, err := clientset.BatchV1().Jobs(ns).Get(job_name, metav1.GetOptions{})
		for {
			if errors.IsNotFound(err) {
				fmt.Printf("%s job not found. Retrying in %d seconds...\n", job_name, checkInterval)
				time.Sleep(time.Duration(checkInterval) * time.Second)
			} else if err != nil {
				fmt.Printf("Error getting %s job object: %v Retrying in %d seconds...\n", job.Name, err.Error(), checkInterval)
				time.Sleep(time.Duration(checkInterval) * time.Second)
			} else {
				if job.Status.Active >= 1 {
					fmt.Printf("%s job is not completed yet. Retrying in %d seconds...\n", job.Name, checkInterval)
					time.Sleep(time.Duration(checkInterval) * time.Second)
				} else if job.Status.Succeeded >= 1 {
					fmt.Printf("%s job succeeded\n", job.Name)
					break
				} else if job.Status.Failed >= 1 {
					fmt.Printf("%s job failed\n", job.Name)
					panic("Job failed\n")
				}
			}
			time.Sleep(time.Duration(1) * time.Second)
			job, err = clientset.BatchV1().Jobs(ns).Get(job_name, metav1.GetOptions{})
		}
	}

	//services
	for _, service_name := range services {
		fmt.Printf("Getting '%s' service object...\n", service_name)
		service, err := clientset.CoreV1().Services(ns).Get(service_name, metav1.GetOptions{})
		for {
			if errors.IsNotFound(err) {
				fmt.Printf("%s service not found. Retrying in %d seconds...\n", service_name, checkInterval)
				time.Sleep(time.Duration(checkInterval) * time.Second)
			} else if err != nil {
				fmt.Printf("Error getting %s service object: %v Retrying in %d seconds...\n", service.Name, err.Error(), checkInterval)
				time.Sleep(time.Duration(checkInterval) * time.Second)
			} else {
				break
			}
			service, err = clientset.CoreV1().Services(ns).Get(service_name, metav1.GetOptions{})
		}

		set := labels.Set(service.Spec.Selector)

		for {
			fmt.Printf("Getting pods for the '%s' service...\n", service.Name)
			pods, err := clientset.CoreV1().Pods(ns).List(metav1.ListOptions{LabelSelector: set.AsSelector().String()})
			if err != nil {
				panic(err.Error())
			}

			if len(pods.Items) < 1 {
				fmt.Printf("No pods found for the '%s' service. Retrying...\n", service.Name)
				time.Sleep(1 * time.Second)
				continue
			}

			fmt.Printf("Checking readiness of the '%s' service pods...\n", service.Name)

			ready_pod_found := false

			for _, pod := range pods.Items {
				for _, cond := range pod.Status.Conditions {
					if cond.Type == "Ready" && cond.Status == "True" {
						fmt.Printf("%s is ready.\n", pod.GetName())
						ready_pod_found = true
						break
					}
				}
				if ready_pod_found == true {
					break
				}
				fmt.Printf("%s is not ready yet. Retrying in %d seconds...\n", pod.GetName(), checkInterval)
			}
			if ready_pod_found == true {
				break
			}
			time.Sleep(time.Duration(checkInterval) * time.Second)
		}
	}
}
