package main

import (
	"fmt"
	"net/http"
	"time"
	"log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type Target struct {
	Service v1.Service
	Pods    *v1.PodList
	ready   bool
}

type TargetReady struct {
	name string
	ready bool
}

//var currentTarget Target

func (currentTarget Target) Run(clientset *kubernetes.Clientset,check_period int) {
	//currentTargets := make(map[string]Target)
	fmt.Println("In Run for service: " + currentTarget.Service.Name)

	var currentReadiness TargetReady
	currentReadiness.name = currentTarget.Service.Name
	currentReadiness.ready = false
	
	var netclient = &http.Client{
		Timeout: 5 * time.Second,
	}

	for {
			fmt.Println("Checking service: " + currentTarget.Service.Name + currentTarget.Service.Spec.Ports[0].String())

			//for all pods under service, TODO percentage availability...
			for _, podOfService := range currentTarget.Pods.Items {

				//get ping url from pod labels
				url := podOfService.Annotations["ping"]

				if url == "" {
					url = "ping"
				}


				resp, err := netclient.Get("http://" + podOfService.Status.PodIP + ":" + fmt.Sprint(currentTarget.Service.Spec.Ports[0].Port) + "/" + url)
				if err == nil {
					if resp.StatusCode == http.StatusOK {
						fmt.Println("Pod "+podOfService.Name+" ping http status: ", resp.StatusCode)
						currentTarget.ready = true
					} else {
						fmt.Println("Non-OK HTTP status:", resp.StatusCode)
						currentTarget.ready = false
					}
					resp.Body.Close()
				} else {
					log.Output(1,err.Error())
					fmt.Println("HTTP request error")
					currentTarget.ready = false
				}

				//changed currentReadiness
				if currentReadiness.ready != currentTarget.ready {
					currentReadiness.ready = currentTarget.ready
					readiness <- currentReadiness
				}

			}
		time.Sleep(time.Duration(check_period) * time.Second)
	}
}