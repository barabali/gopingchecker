package main

import (
	"fmt"
	"net/http"
	"time"
	"log"
	"k8s.io/client-go/kubernetes"
)

func (currentTarget Target) Run(clientset *kubernetes.Clientset,check_period int,timeout int) {
	fmt.Println("In Run for service: " + currentTarget.Service.Name)

	var currentReadiness TargetReady
	currentReadiness.Name = currentTarget.Service.Name
	currentReadiness.Ready = false
	
	var netclient = &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
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
						currentTarget.Ready = true
					} else {
						fmt.Println("Non-OK HTTP status:", resp.StatusCode)
						currentTarget.Ready = false
					}
					resp.Body.Close()
				} else {
					log.Output(1,err.Error())
					fmt.Println("HTTP request error")
					currentTarget.Ready = false
				}

				//changed currentReadiness
				if currentReadiness.Ready != currentTarget.Ready {
					currentReadiness.Ready = currentTarget.Ready
					readiness <- currentReadiness
				}

			}
		time.Sleep(time.Duration(check_period) * time.Second)
	}
}