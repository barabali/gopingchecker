/*
Copyright 2016 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Note: the example only works with the code within the same release/branch.
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
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

type Watched struct {
	Service v1.Service
	Pods    *v1.PodList
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
	fmt.Println(service_filter)

	for {
		// get pods in all the namespaces by omitting namespace
		// Or specify namespace to get pods in particular namespace
		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

		//get services
		services, err := clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d services in the cluster\n", len(services.Items))

		var watched_service *v1.Service
		for _, s := range services.Items {
			if strings.HasPrefix(s.Name, "docker-go-ping") {
				watched_service = s.DeepCopy()
				fmt.Printf("Service name: %s \n", watched_service.Name)
			}
		}

		for _, port := range watched_service.Spec.Ports {
			fmt.Printf("Service port and nodeport: %d, %d \n", port.Port, port.NodePort)
		}

		servicepods, err := getPodsForSvc(watched_service, "default", clientset)
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("Pod connected to service and ip: %s %s \n", servicepods.Items[0].Name, servicepods.Items[0].Status.PodIP)

		resp, err := http.Get("http://" + servicepods.Items[0].Status.PodIP + ":8080/ping")
		if err != nil {
			panic(err.Error())
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err.Error())
		}

		fmt.Printf("Response: %s \n", body)

		time.Sleep(5 * time.Second)
	}
}

func getPodsForSvc(svc *v1.Service, namespace string, k8sClient *kubernetes.Clientset) (*v1.PodList, error) {
	set := labels.Set(svc.Spec.Selector)
	listOptions := metav1.ListOptions{LabelSelector: set.AsSelector().String()}
	pods, err := k8sClient.CoreV1().Pods(namespace).List(context.TODO(), listOptions)
	for _, pod := range pods.Items {
		fmt.Fprintf(os.Stdout, "pod name: %v\n", pod.Name)
	}
	return pods, err
}
