package main

import (
	"fmt"
	"net/http"
	"time"
	"log"
	"k8s.io/client-go/kubernetes"
)

func (currentTarget Target) Run(clientset *kubernetes.Clientset,check_period int,timeout int,minpodpercentage int, channel chan string) {
	fmt.Println("Started Run for service: " + currentTarget.Service.Name)

	var currentReadiness TargetReady
	currentReadiness.Name = currentTarget.Service.Name
	currentReadiness.Ready = false

	var podpercent PodPercent
	podpercent.Name = currentTarget.Service.Name
	podpercent.Podpercent = 0

	var netclient = &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	for {
		if (currentTarget.checkChannelStop(currentReadiness)){
			return
		}

		//fmt.Println("Checking service: " + currentTarget.Service.Name + currentTarget.Service.Spec.Ports[0].String())
		//for all pods under service, TODO percentage availability...

		//get ping url from service annotations
		url := currentTarget.Service.Annotations["pingurl"]
		port := currentTarget.Service.Annotations["pingport"]
		if url == "" {
			fmt.Println("No pingurl annotation on service: "+currentTarget.Service.Name+", using default: /ping")
			url = "ping"
		}
		if port == "" {
			fmt.Println("No pingport annotation on service: "+currentTarget.Service.Name)
			port = "80"
		}

		resp, err := netclient.Get("http://" + currentTarget.Service.Spec.ClusterIP + ":" + fmt.Sprint(port) + "/" + url)
		if err == nil {
			if resp.StatusCode == http.StatusOK {
				//fmt.Println("Pod "+podOfService.Name+" ping http status: ", resp.StatusCode)
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

		//Activate only if percentage is required
		if minpodpercentage > 0 {
			numOfPods := len(currentTarget.Pods.Items)
			currentAvailablePods := 0.0

			for _, podOfService := range currentTarget.Pods.Items {
				//get ping url from pod labels

				resp, err := netclient.Get("http://" + podOfService.Status.PodIP + ":" + fmt.Sprint(port) + "/" + url)
				if err == nil {
					if resp.StatusCode == http.StatusOK {
						//fmt.Println("Pod "+podOfService.Name+" ping http status: ", resp.StatusCode)
						currentAvailablePods++
					} else {
						fmt.Println("Non-OK HTTP status:", resp.StatusCode)
					}
					resp.Body.Close()
				} else {
					log.Output(1,err.Error())
					fmt.Println("HTTP request error")
				}
			}

			//Get percentage
			currentAvailability := currentAvailablePods / float64(numOfPods)
			podpercent.Podpercent = currentAvailability*100
			podpercentchannel <- podpercent

			//Check requirement
			if podpercent.Podpercent < float64(minpodpercentage) {
				currentTarget.Ready = false
			}
		}

		//changed currentReadiness	
		if currentReadiness.Ready != currentTarget.Ready {
			currentReadiness.Ready = currentTarget.Ready
			readiness <- currentReadiness
		}

		time.Sleep(time.Duration(check_period) * time.Second)
	}
}

func (currentTarget Target) checkChannelStop(currentReadiness TargetReady) bool {
	select {
		case x, ok := <- currentTarget.Channel:
			if ok {
				fmt.Println(x+", read in "+currentTarget.Service.Name+", stopping goroutine.")
				currentReadiness.Ready = false
				readiness <- currentReadiness
				return true
			} else {
				fmt.Println("Channel closed!")
			}
		default:
			fmt.Println("Channel empty.")
			//Nothing on channel
	}
	return false
}