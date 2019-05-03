package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/williamhgough/shortly/pkg/adding"
	"github.com/williamhgough/shortly/pkg/hashing"
	"github.com/williamhgough/shortly/pkg/http/rest"
	"github.com/williamhgough/shortly/pkg/redirect"
	"github.com/williamhgough/shortly/pkg/storage/memory"
)

var (
	// set the default port for the service
	port        uint = 8080
	storageType      = "memory"
)

func main() {
	// Create flags for configuring ports for each service to run on.
	flag.UintVar(&port, "port", port, "HTTP port to run the service on")
	flag.StringVar(&storageType, "storage", storageType, "Storage type to use [memory]")
	flag.Parse()

	var adder adding.Service
	var redirector redirect.Service

	switch storageType {
	case "memory":
		mem := memory.NewMapRepository()
		adder = adding.NewService(mem, hashing.NewSimpleHasher())
		redirector = redirect.NewService(mem)
	default:
		mem := memory.NewMapRepository()
		adder = adding.NewService(mem, hashing.NewSimpleHasher())
		redirector = redirect.NewService(mem)
	}

	router := rest.Handler(adder, redirector)
	fmt.Println("Shortly server is on tap now: http://localhost:8080")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}
