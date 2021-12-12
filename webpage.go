package main

import (
	"fmt"
	"net/http"
	"html/template"
//	"encoding/json"
//	v1 "k8s.io/api/core/v1"
)

func webpage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Web testpage called!!!")
	//Parsing HTML
	t, err := template.ParseFiles("page.html")
	if err != nil {
		fmt.Println(err)
	}

	items := struct {
		Name string
		City string
	}{
		Name: "MyName",
		City: "MyCity",
	}

	t.Execute(w, items)
}


func startWebPage() {
	/*r := gin.Default()
	r.Static("/css", "./static/css")
	r.Static("/img", "./static/img")
	r.Static("/scss", "./static/scss")
	r.Static("/vendor", "./static/vendor")
	r.Static("/js", "./static/js")
	r.StaticFile("/favicon.ico", "./img/favicon.ico")

	r.LoadHTMLGlob("")

	r.Run(":8084") */

	
	//create http server
	http.Handle("/", http.FileServer(http.Dir("css/")))
	http.HandleFunc("/testPage", webpage)
	http.ListenAndServe(":8084", nil)
}