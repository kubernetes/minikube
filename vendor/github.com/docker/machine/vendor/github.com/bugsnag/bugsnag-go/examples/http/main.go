package main

import (
	"github.com/bugsnag/bugsnag-go"
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("/", Get)

	bugsnag.Configure(bugsnag.Configuration{
		APIKey: "066f5ad3590596f9aa8d601ea89af845",
	})

	log.Println("Serving on 9001")
	http.ListenAndServe(":9001", bugsnag.Handler(nil))
}

func Get(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("OK\n"))

	var a struct{}
	crash(a)
}

func crash(a interface{}) string {
	return a.(string)
}
