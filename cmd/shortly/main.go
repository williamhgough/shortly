package main

import (
	"flag"
	"log"

	"github.com/williamhgough/shortly"
)

var (
	// set the default port for the service
	port uint = 8080
)

func main() {
	// Create flags for configuring ports for each service to run on.
	flag.UintVar(&port, "port", port, "HTTP port to run the service on")
	flag.Parse()

	// Create and run new shortly service
	svc := shortly.New(port)
	if err := svc.Start(); err != nil {
		log.Fatalf("could not start HTTP server: %s", err)
	}
}
