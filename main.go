package main

import (
	"bufio"
	"log"
	"net"
	"strings"
	"time"

	dw "remotecmds/download"
	"remotecmds/extconnection"
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
		log.Println(scanr.Text())
		commParams := strings.Split(scanr.Text(), " ")
		switch {
		case commParams[0] == "download":
			if len(commParams) < 3 {
				w.WriteString("not enougth arguments\n")
				w.Flush()
				break
			}
			dw.Download(commParams[1], commParams[2])
		case commParams[0] == "dwnldlist":
			for _, s := range dw.GetDownloadList() {
				w.WriteString(s)
			}
			w.WriteString("\n")
			w.Flush()
		default:
			w.WriteString("unknown command\n")
			w.Flush()
		}

	}
	return nil
}

func main() {
	server := &Server{
		Addr:        ":8080",
		IdleTimeout: 20 * time.Second,
		MaxBuffer:   64,
		MaxRead:     1024,
	}
	server.ListenAndServe()

	// download.Download("https://golang.org/lib/godoc/images/footer-gopher.jpg", "./")
}
