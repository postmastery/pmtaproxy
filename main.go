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
var Version = "1.3.0"

func main() {

	// Timestamp added by journalctl.
	log.SetFlags(0)

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

	s, err := newServer(listen, allow)
	if err != nil {
		log.Fatal(err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM) // SIGINT (^C) and SIGTERM
	select {
	case err := <-s.ErrCh:
		log.Printf("%v; stopping...", err)
	case sig := <-sigCh:
		log.Printf("%v signal; stopping...", sig)
	}
	s.Close()
}
