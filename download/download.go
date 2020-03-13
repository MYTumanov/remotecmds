package download

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// https://golang.org/lib/godoc/images/footer-gopher.jpg

type dStatus int

const (
	queue dStatus = iota
	inprogress
	complete
	err
)

type downloads struct {
	fileName     string
	filePath     string
	url          string
	status       dStatus
	byteDownload int
}

// Download ownloads files from url to path
func Download(urlStr string, path string) {
	log.SetPrefix("Download:")
	resp, err := http.Get(urlStr)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	fName := getName(urlStr)
	fullTempPath := filepath.Join(path, fName+".temp")
	fullPath := filepath.Join(path, fName)
	f, err := os.Create(fullTempPath)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		log.Println(err)
	}

	err = f.Close()
	if err != nil {
		log.Println(err)
	}

	if err := os.Rename(fullTempPath, fullPath); err != nil {
		log.Println(err)
	}
}

func getName(url string) string {
	name := url[strings.LastIndex(url, "/")+1:]
	log.Println(name)
	return name
}
