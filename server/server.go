package server

import (
	"log"
	"net"
	"remotecmds/extconnection"
	"time"
)

// Server struct for server
type Server struct {
	Addr        string
	IdleTimeout time.Duration
	MaxBuffer   int64
	MaxRead     int
	inShutdown  bool
	listener    net.Listener
	cons        map[*net.Conn]struct{}
}

// Shutdown server
func (srv *Server) Shutdown() {
	srv.inShutdown = true
	log.Println("In shutdown...")
	srv.listener.Close()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.Printf("Waiting on %v connections\n", len(srv.cons))
		}
		if len(srv.cons) == 0 {
			log.Println("STOP")
		}
	}
}

// ListenAndServe start server
func (srv Server) ListenAndServe() {
	addr := srv.Addr
	if addr == "" {
		addr = ":8080"
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println(err)
	}
	defer listener.Close()

	srv.listener = listener

	for {
		if srv.inShutdown {
			break
		}
		newConn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		}
		conn := &extconnection.Conn{
			Conn:        newConn,
			IdleTimeout: srv.IdleTimeout,
			MaxBuffer:   srv.MaxBuffer,
			MaxRead:     srv.MaxRead,
		}

		if srv.cons == nil {
			srv.cons = make(map[*net.Conn]struct{})
		}

		srv.cons[&newConn] = struct{}{}
		conn.Conn.SetDeadline(time.Now().Add(srv.IdleTimeout))
		log.Printf("accepted connection from %v", conn.RemoteAddr())
		go srv.handle(conn)
	}
}
