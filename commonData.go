package main

import (
	"fmt"
	"strconv"
	v1 "k8s.io/api/core/v1"
)

type Target struct {
	Service v1.Service `json:"service"`
	Pods    *v1.PodList `json:"pods"`
	Ready   bool `json:"pingStatus"`
}

type TargetReady struct {
	Name string `json:"serviceName"`
	Ready bool `json:"status"`
}

var (
	statusMap = make(map[string]bool)
	currentTargets = make(map[string]Target)
	readiness = make(chan TargetReady)
)

func reader() {
	for{
		s := <-readiness 
		fmt.Println("Teszt channel read:" + s.Name + ", "+ strconv.FormatBool(s.Ready))
		statusMap[s.Name]=s.Ready
	}
}
