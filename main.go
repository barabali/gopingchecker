package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	currentTargets = make(map[string]Target)
	readiness = make(chan TargetReady)
)

func createTargets(clientset *kubernetes.Clientset, serviceNames []string, check_period int) {

	refreshTargets(clientset, serviceNames, currentTargets)
	for _, target := range currentTargets {
		fmt.Println("Starting target Run: "+ target.Service.Name)
		go target.Run(clientset, check_period)
	}
}

func main() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	//get config
	service_filter := os.Getenv("SERVICE_FILTER")
	check_period_string := os.Getenv("CHECK_PERIOD")

	check_period, err := strconv.Atoi(check_period_string)
	if err != nil {
		check_period = 5
		fmt.Println("ERROR converting string to int for check period!")
	}

	serviceNames := strings.SplitN(service_filter, ",", -1)

	createTargets(clientset, serviceNames, check_period)

	go startHTTPServer()

	for {
		refreshTargets(clientset, serviceNames, currentTargets)
		time.Sleep(time.Duration(check_period) * time.Second)
	}
}

func refreshTargets(clientset *kubernetes.Clientset, serviceNames []string, currentTargets map[string]Target) {
	//get services
	services_all, err := clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println(err.Error())
	}

	for _, name := range serviceNames {
		var found bool = false
		for _, service := range services_all.Items {
			//get service, pods based on service name
			if strings.HasPrefix(service.Name, name) {
				found = true
				var watched Target
				watched.Service = service

				//get pods for service in default namespace
				servicepods, err := getPodsForSvc(&service, "default", clientset)
				if err != nil {
					fmt.Println(err.Error())
				}
				watched.Pods = servicepods

				currentTargets[service.Name] = watched
				fmt.Println("Target added/refreshed: " + watched.Service.Name)
			}
		}

		if !found {
			fmt.Println("Service not found for configured service name: " + name)
			//Removing service in case it existed before
			delete(currentTargets, name)
		}
	}
}

/*func (services *v1.Service) getService(name string) {
	var watched_service *v1.Service
	for _, s := range services.Items {
		if strings.HasPrefix(s.Name, "docker-go-ping") {
			watched_service = s.DeepCopy()
			fmt.Printf("Service name: %s \n", watched_service.Name)
		}
	}
}*/

func getPodsForSvc(svc *v1.Service, namespace string, k8sClient *kubernetes.Clientset) (*v1.PodList, error) {
	set := labels.Set(svc.Spec.Selector)
	listOptions := metav1.ListOptions{LabelSelector: set.AsSelector().String()}
	pods, err := k8sClient.CoreV1().Pods(namespace).List(context.TODO(), listOptions)
	return pods, err
}
