package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// Version reported by -v parameter.
var Version = "1.2.2"

func main() {

	var (
		listen  string
		allow   string
		version bool
	)
	flag.StringVar(&listen, "l", ":5000", "host:port for listening")
	flag.StringVar(&allow, "a", "10.0.0.0/8", "allowed connection sources")
	flag.BoolVar(&version, "v", false, "show version")
	flag.Parse()

	if version {
		fmt.Printf("v%s\n", Version)
		return
	}

	server, err := newServer(listen, allow)
	if err != nil {
		log.Fatal(err)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM) // SIGINT (^C) and SIGTERM
	s := <-sigc
	log.Printf("%v received, stopping...", s)
	server.Close()
	// TODO: wait for all goroutines to finish
}
