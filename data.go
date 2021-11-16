package main

import (
	"fmt"
	"strconv"
)

var (
	statusMap = make(map[string]bool)
)

func reader() {
	for{
		s := <-readiness 
		fmt.Println("Teszt channel read:" + s.name + ", "+ strconv.FormatBool(s.ready))
		statusMap[s.name]=s.ready
	}
}
