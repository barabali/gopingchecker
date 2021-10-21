package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

type Target struct {
	Service v1.Service
	Pods    *v1.PodList
	ready   bool
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
		//panic(err.Error())
	}

	serviceNames := strings.SplitN(service_filter, ",", -1)

	currentTargets := make(map[string]Target)

	for {
		refreshTargets(clientset, serviceNames, currentTargets)

		for _, service := range currentTargets {
			fmt.Println("Checking service: " + service.Service.Name + service.Service.Spec.Ports[0].String())

			//for all pods under service, TODO percentage availability...
			for _, podOfService := range service.Pods.Items {

				//get ping url from pod labels
				url := podOfService.Annotations["ping"]

				if url == "" {
					url = "ping"
				}

				resp, err := http.Get("http://" + podOfService.Status.PodIP + ":" + fmt.Sprint(service.Service.Spec.Ports[0].Port) + "/" + url)
				if err != nil {
					panic(err.Error())
				}

				if resp.StatusCode == http.StatusOK {
					fmt.Println("Pod "+podOfService.Name+" ping http status: ", resp.StatusCode)
					service.ready = true
				} else {
					fmt.Println("Non-OK HTTP status:", resp.StatusCode)
					service.ready = false
				}
				resp.Body.Close()
			}

			/*body, err := io.ReadAll(resp.Body)
			if err != nil {
				panic(err.Error())
			}

			fmt.Printf("Response: %s \n", body)*/
		}

		time.Sleep(time.Duration(check_period) * time.Second)
	}
}

func refreshTargets(clientset *kubernetes.Clientset, serviceNames []string, currentTargets map[string]Target) {
	//get services
	services_all, err := clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
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
					panic(err.Error())
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
