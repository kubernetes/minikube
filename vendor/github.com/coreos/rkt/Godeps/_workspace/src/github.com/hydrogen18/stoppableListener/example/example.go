package main

import (
	"fmt"
	"github.com/hydrogen18/stoppableListener"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func helloHttp(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	fmt.Fprintf(rw, "Hello HTTP!\n")
}

func main() {
	originalListener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	sl, err := stoppableListener.New(originalListener)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", helloHttp)
	server := http.Server{}

	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT)
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		defer wg.Done()
		server.Serve(sl)
	}()

	fmt.Printf("Serving HTTP\n")
	select {
	case signal := <-stop:
		fmt.Printf("Got signal:%v\n", signal)
	}
	fmt.Printf("Stopping listener\n")
	sl.Stop()
	fmt.Printf("Waiting on server\n")
	wg.Wait()

}
