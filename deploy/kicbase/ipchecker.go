package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
)

func main() {
	http.HandleFunc("/", handler) // each request calls handler
	fmt.Printf("Starting server at port 8080\n")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}

// handler echoes the Path component of the requested URL.
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Receive request uri %s at port 8080\n", r.RequestURI)
	out, err := exec.Command("docker", "ps").Output()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("docker ps output:\n%s\n", string(out))
	fmt.Fprintf(w, "allow")
}
