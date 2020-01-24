package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

// https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt

const maxPacketSize = 104

func readHeader(r io.Reader) (source *net.TCPAddr, destination *net.TCPAddr, err error) {

	br := bufio.NewReaderSize(r, maxPacketSize)

	line, _, err := br.ReadLine()
	if err != nil {
		return
	}
	header := string(line)

	parts := strings.Split(header, " ")
	if len(parts) < 6 {
		err = fmt.Errorf("expected proxy header: %q", header)
		return
	}

	if parts[0] != "PROXY" {
		err = fmt.Errorf("expected proxy protocol: %q", parts[0])
		return
	}

	if parts[1] != "TCP4" && parts[1] != "TCP6" {
		err = fmt.Errorf("invalid address type: %q", parts[1])
		return
	}

	sourceIP := net.ParseIP(parts[2])
	if sourceIP == nil {
		err = fmt.Errorf("invalid source address: %q", parts[2])
		return
	}

	destinationIP := net.ParseIP(parts[3])
	if destinationIP == nil {
		err = fmt.Errorf("invalid destination address: %q", parts[3])
		return
	}

	sourcePort, err := strconv.Atoi(parts[4])
	if err != nil {
		err = fmt.Errorf("invalid source port: %q", parts[4])
		return
	}

	destinationPort, err := strconv.Atoi(parts[5])
	if err != nil {
		err = fmt.Errorf("invalid destination port: %q", parts[5])
		return
	}

	source = &net.TCPAddr{
		IP:   sourceIP,
		Port: sourcePort,
	}
	destination = &net.TCPAddr{
		IP:   destinationIP,
		Port: destinationPort,
	}
	return
}
