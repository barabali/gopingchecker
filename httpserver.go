package main

import (
	"fmt"
	"net/http"
)

func prometheusmetrics(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "# HELP services statuses\n")
	for key,value := range statusMap {
		//fmt.Fprint(w, "status{serviceName="+service.Service.Name+"} "+strconv.FormatBool(service.ready)+"\n")
		if value {
			fmt.Fprint(w, "status{serviceName="+key+"} "+"1\n")
		} else {
			fmt.Fprint(w, "status{serviceName="+key+"} "+"0\n")
		}
	}
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w,"{\"status\":\"ok\"}")
	return
}

func startMetricsServer() {
	//create http server for displaying metrics
	http.HandleFunc("/metrics", prometheusmetrics)
	http.HandleFunc("/ping", ping)
	http.ListenAndServe(":8082", nil)
}