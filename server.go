package main

import (
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type server struct {
	listener net.Listener
	allowed  *net.IPNet
	wg       sync.WaitGroup
}

func newServer(listenAddress string, allowCIDR string) (s *server, err error) {
	//var lc net.ListenConfig
	// listener, err := lc.Listen(ctx, "tcp", listenAdress)
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return
	}
	_, allowed, err := net.ParseCIDR(allowCIDR)
	if err != nil {
		return
	}
	s = &server{
		listener: listener,
		allowed:  allowed,
	}
	s.wg.Add(1)
	go s.serve()
	return
}

func (s *server) Close() error {
	err := s.listener.Close()
	// wait for all goroutines to finish
	s.wg.Wait()
	return err
}

// Serve handles connections.
func (s *server) serve() error {
	defer s.wg.Done()
	log.Printf("Listening on %s", s.listener.Addr().String())
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}
		s.wg.Add(1)
		go s.handleConn(conn)
	}
}

const requestTimeout = time.Minute

func (s *server) handleConn(conn net.Conn) {
	defer s.wg.Done()
	//defer conn.Close() // NOTE: write half closed in proxy

	address, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	clientIP := net.ParseIP(address)
	if clientIP == nil || !s.allowed.Contains(clientIP) {
		conn.Close()
		log.Printf("Request from %s not allowed", clientIP)
		return
	}

	// read proxy request
	conn.SetReadDeadline(time.Now().Add(requestTimeout))
	source, destination, err := readHeader(conn)
	if err != nil {
		conn.Close()
		log.Printf("%v", err)
		// probably not a valid client, silently disconnect
		return
	}
	conn.SetReadDeadline(time.Time{})

	log.Printf("Request to proxy from %s to %s via %s", conn.RemoteAddr(), destination, source)

	// dial before returning socks response
	dstConn, err := net.DialTCP("tcp", source, destination)
	if err != nil {
		conn.Close()
		// "dial tcp 136.144.181.209:0->149.210.165.215:25: connect: connection timed out" after 2m10s
		// "refused", "network is unreachable"
		log.Printf("%v", err)
		return
	}
	//defer dstConn.Close() // NOTE: write half closed in proxy

	// wait for banner greeting?

	errCh := make(chan error, 2)

	// ? https://godoc.org/golang.org/x/sync/errgroup#pkg-index
	go proxy(dstConn, conn, errCh)
	go proxy(conn, dstConn, errCh)

	// wait for both to return
	for i := 0; i < 2; i++ {
		err := <-errCh
		if err != nil {
			// readfrom tcp 136.144.181.209:35455->108.177.127.27:25: splice: connection reset by peer
			// readfrom tcp 185.28.63.190:37080->67.195.228.94:25: splice: connection reset by peer
			// readfrom tcp 185.28.63.135:48580->67.195.204.75:25: use of closed network connection
			log.Printf("%v", err)
		}
	}
	log.Printf("Done proxying from %s to %s via %s", conn.RemoteAddr(), destination, source)
}

var pool = &sync.Pool{
	New: func() interface{} {
		return make([]byte, 32*1024)
	},
}

type halfCloser interface {
	//CloseRead() error
	CloseWrite() error
}

// Proxy copies data from src to dst until EOF.
// If dst implements CloseWrite (*net.TCPConn) it is called.
func proxy(src net.Conn, dst net.Conn, errCh chan error) {
	buf := pool.Get().([]byte)
	_, err := io.CopyBuffer(dst, src, buf)
	pool.Put(buf)
	// If source connection closed, signal destination by closing outgoing connection.
	// Otherwise destination might wait and reverse direction proxy does not finish.
	//if conn, ok := dst.(halfCloser); ok {
	//	conn.CloseWrite()
	//}
	dst.Close()
	//if conn, ok := src.(halfCloser); ok {
	//	conn.CloseRead()
	//}
	errCh <- err
}
