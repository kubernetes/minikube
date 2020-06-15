package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Yes!\n")
	fmt.Fprint(w, r.URL.EscapedPath())
	log.Printf("----REQUEST: %+v----\n", r)
	fmt.Printf("----REQUEST: %+v----\n", r)
}

func main() {
	log.Print("Metadata server started!")

	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
