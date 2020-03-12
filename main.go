package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"remotecmds/extconnection"

	"github.com/shirou/gopsutil/cpu"
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

func (srv Server) handle(conn net.Conn) error {
	defer func() {
		log.Printf("closing connection from %v", conn.RemoteAddr())
		delete(srv.cons, &conn)
		conn.Close()
	}()

	// r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	scanr := bufio.NewScanner(conn)

	for {
		if scanned := scanr.Scan(); !scanned {
			if err := scanr.Err(); err != nil {
				log.Printf("%v(%v)", err, conn.RemoteAddr())
				return err
			}
			break
		}
		w.WriteString(strings.ToUpper(scanr.Text()) + "\n")
		w.Flush()
	}
	return nil
}

func main() {
	// server := &Server{
	// 	Addr:        ":8080",
	// 	IdleTimeout: 20 * time.Second,
	// 	MaxBuffer:   64,
	// 	MaxRead:     1024,
	// }
	// server.ListenAndServe()

	for {
		usage, err := cpu.Percent(0, false)
		if err != nil {
			log.Panic(err)
		}
		fmt.Println(usage)
		time.Sleep(1 * time.Second)
	}
}
