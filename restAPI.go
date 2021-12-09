package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	v1 "k8s.io/api/core/v1"
)

type ServiceDTO struct {
	Name string `json:"serviceName"`
	Spec v1.ServiceSpec `json:"spec"`
}

/*type PodDTO struct {
	Name string `json:"name"`
}*/

type TargetDTO struct {
	Service ServiceDTO `json:"service"`
	ServiceAnnotations map[string]string `json:"serviceAnnotations"`
	Pods    []string `json:"pods"`
	Ready   bool `json:"pingStatus"`
}

func homePage(w http.ResponseWriter, r *http.Request){
    fmt.Fprintf(w, "Welcome to the HomePage!")
    //fmt.Println("Endpoint Hit: homePage")
}

func returnAllTargetStatuses(w http.ResponseWriter, r *http.Request){
    fmt.Println("Endpoint Hit: returnAllTargetStatuses")
	w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(statusMap)
}

func returnAllTargets(w http.ResponseWriter, r *http.Request){
    fmt.Println("Endpoint Hit: returnAllTargets")

	var TargetDTOs []TargetDTO

	for _, target := range currentTargets {
		fmt.Println("Converting to DTO: "+ target.Service.Name)
		var sdto ServiceDTO
		sdto.Spec = target.Service.Spec
		sdto.Name = target.Service.Name

		var dto TargetDTO 
		dto.Service = sdto
		dto.ServiceAnnotations = target.Service.Annotations
		dto.Ready = statusMap[target.Service.Name]

		for _, podOfService := range target.Pods.Items {
			dto.Pods = append(dto.Pods,podOfService.Name)
		}

		TargetDTOs = append(TargetDTOs,dto)
		//TargetDTO dto = TargetDTO{Service: target.Service.spec,Pods target.Pods,ready: True}
	}

	w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(TargetDTOs)
}

func startRESTServer() {
	//create http server for displaying metrics
	//http.HandleFunc("/", homePage)
	http.HandleFunc("/getStatuses", returnAllTargetStatuses)
	http.HandleFunc("/getAll", returnAllTargets)
	http.ListenAndServe(":8083", nil)
}