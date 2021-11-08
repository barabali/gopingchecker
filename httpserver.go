package main

import (
	"fmt"
	"net/http"
	"strconv"
)

var (
	statusMap = make(map[string]bool)
)

func handler(w http.ResponseWriter, r *http.Request) {
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

func reader() {
	for{
		s := <-readiness 
		fmt.Println("Teszt channel read:" + s.name + ", "+ strconv.FormatBool(s.ready))
		statusMap[s.name]=s.ready
	}
}

func startHTTPServer() {
	//create http server for displaying metrics
	http.HandleFunc("/metrics", handler)
	go reader()
	http.ListenAndServe(":8082", nil)
}