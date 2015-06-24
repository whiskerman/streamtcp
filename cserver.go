package main

import (
	. "./classes"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <port>\n", os.Args[0])
		os.Exit(-1)
	}

	go func() {
		log.Println(http.ListenAndServe("192.168.1.5:6060", nil))

	}()

	server := CreateServer(nil)

	fmt.Printf("Running on %s\n", os.Args[1])
	server.Start(os.Args[1])

}
