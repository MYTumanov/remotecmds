package main

import (
	sr "remotecmds/server"
	"time"
)

func main() {
	server := &sr.Server{
		Addr:        ":8080",
		IdleTimeout: 200 * time.Second,
		MaxBuffer:   64,
		MaxRead:     1024,
	}
	server.ListenAndServe()

	// download.Download("https://golang.org/lib/godoc/images/footer-gopher.jpg", "./")
}
