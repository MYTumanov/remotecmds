package server

import (
	"bufio"
	"log"
	dw "remotecmds/download"
	ex "remotecmds/extconnection"
	"strings"
)

func (srv Server) handle(conn *ex.Conn) error {
	defer func() {
		log.Printf("closing connection from %v", conn.RemoteAddr())
		delete(srv.cons, &conn.Conn)
		conn.Close()
	}()

	// r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn.Conn)
	scanr := bufio.NewScanner(conn.Conn)

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

		if commParams[0] != "loggin" && (conn.UserName == "" || conn.UserName == " ") {
			w.WriteString("use command 'loggin'\n")
			w.Flush()
			continue
		}

		switch {
		case commParams[0] == "loggin":
			if len(commParams) < 2 {
				w.WriteString("not enougth arguments\n")
				w.Flush()
				break
			}
			w.WriteString(conn.Loggin(commParams[1]) + "\n")
			w.Flush()
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
