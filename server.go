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
	allowed  []*net.IPNet
	wg       sync.WaitGroup
	quit     chan struct{}
	// ErrCh will pass an error from accept
	ErrCh chan error
}

// newServer starts a listener in a goroutine and returns.
func newServer(listenAddress string, allowed []*net.IPNet) (s *server, err error) {
	//var lc net.ListenConfig
	// listener, err := lc.Listen(ctx, "tcp", listenAdress)
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return
	}
	s = &server{
		listener: listener,
		allowed:  allowed,
		quit:     make(chan struct{}),
		ErrCh:    make(chan error, 1),
	}
	s.wg.Add(1)
	go s.serve()
	return
}

func (s *server) Close() error {
	close(s.quit) // signal to serve() that listener is closed
	err := s.listener.Close()
	s.wg.Wait() // wait for all goroutines to finish
	return err
}

// Serve handles connections.
func (s *server) serve() {
	defer s.wg.Done()
	log.Printf("Listening on %s", s.listener.Addr().String())
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
			}
			if e, ok := err.(net.Error); ok && e.Temporary() {
				// accept tcp [::]:5000: accept4: too many open files
				log.Printf("%v; retrying...", err)
				time.Sleep(50 * time.Millisecond)
				continue
			}
			s.ErrCh <- err
			return
		}
		s.wg.Add(1)
		go s.handleConn(conn)
	}
}

const requestTimeout = time.Minute

func (s *server) handleConn(conn net.Conn) {
	defer s.wg.Done()

	address, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	clientIP := net.ParseIP(address)
	if clientIP == nil || !s.isAllowed(clientIP) {
		conn.Close()
		log.Printf("Request from %s not allowed", clientIP)
		return
	}

	// read proxy request
	_ = conn.SetReadDeadline(time.Now().Add(requestTimeout))
	source, destination, err := readHeader(conn)
	if err != nil {
		conn.Close()
		log.Printf("%v", err)
		// probably not a valid client, silently disconnect
		return
	}
	_ = conn.SetReadDeadline(time.Time{})

	log.Printf("Request to proxy from %s to %s via %s", conn.RemoteAddr(), destination, source)

	// dial before returning socks response
	dstConn, err := net.DialTCP("tcp", source, destination)
	if err != nil {
		conn.Close()
		log.Printf("%v", err)
		return
	}

	errCh := make(chan error, 2)

	// ? https://godoc.org/golang.org/x/sync/errgroup#pkg-index
	// ? use waitgroup and print error from goroutine
	// ? https://godoc.org/github.com/oklog/run
	go proxy(dstConn, conn.(*net.TCPConn), errCh)
	go proxy(conn.(*net.TCPConn), dstConn, errCh)

	// wait for both to return
	for i := 0; i < 2; i++ {
		err := <-errCh
		if err != nil {
			log.Printf("%v", err)
		}
	}
	conn.Close()
	dstConn.Close()
	log.Printf("Done proxying from %s to %s via %s", conn.RemoteAddr(), destination, source)
}

func (s *server) isAllowed(ip net.IP) bool {
	for _, cidr := range s.allowed {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

var pool = &sync.Pool{
	New: func() interface{} {
		return make([]byte, 32*1024)
	},
}

// Proxy copies data from src to dst until EOF.
// When done it closes the destination connection.
func proxy(src, dst *net.TCPConn, errCh chan error) {
	buf := pool.Get().([]byte)
	_, err := io.CopyBuffer(dst, src, buf)
	pool.Put(buf)
	// If source connection closed, signal destination by closing outgoing connection.
	// Otherwise destination might wait and reverse direction proxy does not finish.
	// Servers can choose to send data after receiving FIN packet.
	// https://github.com/kubernetes/kubernetes/blob/master/pkg/proxy/userspace/proxysocket.go#L164
	_ = dst.CloseWrite()
	_ = src.CloseRead() // ? needed
	errCh <- err
}
