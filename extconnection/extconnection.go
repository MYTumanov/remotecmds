package extconnection

import (
	"io"
	"log"
	"net"
	"time"
)

// Conn extension of net.Conn
type Conn struct {
	net.Conn
	IdleTimeout time.Duration
	MaxBuffer   int64
	MaxRead     int
	CurrRead    int
}

func (c *Conn) Write(p []byte) (int, error) {
	c.updateDeadline()
	return c.Conn.Write(p)
}

func (c *Conn) Read(b []byte) (n int, err error) {
	c.updateDeadline()
	// log.Println("CurrRead:", c.CurrRead)
	r := io.LimitReader(c.Conn, c.MaxBuffer)
	n, err = r.Read(b)
	c.CurrRead += n
	if c.CurrRead > c.MaxRead {
		log.Println("Too much read:", c.CurrRead)
		c.Close()
	}
	return
}

func (c *Conn) updateDeadline() {
	idleTimeout := time.Now().Add(c.IdleTimeout)
	c.Conn.SetDeadline(idleTimeout)
}
