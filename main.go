package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

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
	timeout_string := os.Getenv("TIMEOUT")

	check_period, err := strconv.Atoi(check_period_string)
	if err != nil {
		check_period = 5
		fmt.Println("ERROR converting string to int for check period!")
	}

	timeout, err := strconv.Atoi(timeout_string)
	if err != nil {
		timeout = 5
		fmt.Println("ERROR converting string to int for timeout!")
	}

	serviceNames := strings.SplitN(service_filter, ",", -1)

	createTargetGoroutines(clientset, serviceNames, check_period, timeout)
	
	go reader()
	go startMetricsServer()
	go startRESTServer()

	for {
		refreshTargets(clientset, serviceNames, currentTargets,check_period,timeout)
		time.Sleep(time.Duration(check_period) * time.Second)
	}
}

func createTargetGoroutines(clientset *kubernetes.Clientset, serviceNames []string, check_period int, timeout int) {

	createStartTargets(clientset, serviceNames, currentTargets)
	for _, target := range currentTargets {
		fmt.Println("Starting target Run: "+ target.Service.Name)
		go target.Run(clientset, check_period,timeout,target.Channel)
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