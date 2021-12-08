package main

import (
	"context"
	"fmt"
	"strconv"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/apimachinery/pkg/labels"
)

type Target struct {
	Service v1.Service `json:"service"`
	Pods    *v1.PodList `json:"pods"`
	Ready   bool `json:"pingStatus"`
	Channel chan string
}

type TargetReady struct {
	Name string `json:"serviceName"`
	Ready bool `json:"status"`
}

type targets map[string]Target

var (
	statusMap = make(map[string]bool)
	currentTargets = make(targets)
	readiness = make(chan TargetReady)
	messages = make(chan string,5)
)

//type targetnames []Target

func reader() {
	for{
		s := <-readiness 
		fmt.Println("Teszt channel read:" + s.Name + ", "+ strconv.FormatBool(s.Ready))
		statusMap[s.Name]=s.Ready
	}
}

func refreshTargets(clientset *kubernetes.Clientset, configServiceNames []string, currentTargets map[string]Target,check_period int, timeout int) {

	//get services
	k8s_Services, err := clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println(err.Error())
	}

	//Comparing k8s, map and config
	for _, name := range configServiceNames {
		var found_k8s bool = false
		for _, service := range k8s_Services.Items {
			if service.Name == name {
				found_k8s = true

				//Create new if not in map
				if _, ok := currentTargets[name]; !ok {
					fmt.Println("This is where new creation will be")
					var newTarget Target = createSingleTarget(clientset,service)
					go newTarget.Run(clientset, check_period,timeout,newTarget.Channel)
				}

				//TODO refresh pods, service details 

			}
		}

		//In config but not in k8s
		if !found_k8s {
			fmt.Println("Service not found in k8s: " + name)

			//Removing service from map in case it existed before
			if target, ok := currentTargets[name]; ok {
				target.Channel <- "Stop"
				delete(currentTargets, name)
			}
		}
	}

	//removed from config, should remove from map too
	deleteServicesInMapButNotInConfig(configServiceNames)

}

func deleteServicesInMapButNotInConfig(configServiceNames []string){
	for name,_ := range currentTargets {
		if !stringInSlice(name,configServiceNames) {
			delete(currentTargets, name)
			delete(statusMap,name)
		}
	}
}

func stringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

func createSingleTarget(clientset *kubernetes.Clientset, service v1.Service) Target{
	var watched Target
	watched.Service = service

	//get pods for service in default namespace
	servicepods, err := getPodsForSvc(&service, "default", clientset)
	if err != nil {
		fmt.Println(err.Error())
	}
	watched.Pods = servicepods
	watched.Channel =  make(chan string)
	currentTargets[service.Name] = watched
	fmt.Println("Target added: " + watched.Service.Name)
	return watched
}


func createStartTargets(clientset *kubernetes.Clientset, serviceNames []string, currentTargets map[string]Target) {
	//get services
	services_all, err := clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println(err.Error())
	}

	for _, name := range serviceNames {
		var found bool = false
		for _, service := range services_all.Items {
			//get service, pods based on service name
			if service.Name == name {
				found = true
				createSingleTarget(clientset,service)
			}
		}

		if !found {
			fmt.Println("Service not found for configured service name: " + name)
		}
	}
}

func getPodsForSvc(svc *v1.Service, namespace string, k8sClient *kubernetes.Clientset) (*v1.PodList, error) {
	set := labels.Set(svc.Spec.Selector)
	listOptions := metav1.ListOptions{LabelSelector: set.AsSelector().String()}
	pods, err := k8sClient.CoreV1().Pods(namespace).List(context.TODO(), listOptions)
	return pods, err
}