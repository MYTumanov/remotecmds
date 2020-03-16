package download

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// https://golang.org/lib/godoc/images/footer-gopher.jpg

func init() {
	log.SetPrefix("Download:")
}

type dStatus int

const (
	queue dStatus = iota
	inprogress
	complete
	err
)

var dStatusToStr = map[dStatus]string{
	queue:      "queue",
	inprogress: "inprogress",
	complete:   "complete",
	err:        "err",
}

type downloads struct {
	fileName     string
	filePath     string
	url          string
	status       dStatus
	byteDownload int64
}

var downloadsList map[string]*downloads

func (d *downloads) Write(p []byte) (int, error) {
	n := len(p)
	d.byteDownload += int64(n)
	fmt.Println(d.byteDownload)
	return n, nil
}

// Download ownloads files from url to path
func Download(urlStr string, path string) {
	if downloadsList == nil {
		downloadsList = make(map[string]*downloads)
	}
	newDownload := &downloads{
		fileName: getName(urlStr),
		filePath: path,
		url:      urlStr,
		status:   queue,
	}
	downloadsList[urlStr] = newDownload
	err := newDownload.download()
	if err != nil {
		log.Println(err)
	}

}

func (d *downloads) download() error {
	d.status = inprogress
	resp, err := http.Get(d.url)
	if err != nil {
		log.Println(err)
		return err
	}
	defer resp.Body.Close()

	fullTempPath := filepath.Join(d.filePath, d.fileName+".temp")
	fullPath := filepath.Join(d.filePath, d.fileName)
	f, err := os.Create(fullTempPath)
	if err != nil {
		log.Println(err)
		return err
	}
	defer f.Close()

	d.byteDownload, err = io.Copy(f, io.TeeReader(resp.Body, d))
	if err != nil {
		log.Println(err)
		return err
	}

	err = f.Close()
	if err != nil {
		log.Println(err)
		return err
	}

	if err := os.Rename(fullTempPath, fullPath); err != nil {
		log.Println(err)
		return err
	}
	d.status = complete

	return nil
}

func getName(url string) string {
	name := url[strings.LastIndex(url, "/")+1:]
	log.Println(name)
	return name
}

// GetDownloadList return list of downloads
func GetDownloadList() []string {
	if downloadsList == nil {
		return nil
	}
	var list []string
	for _, d := range downloadsList {
		list = append(list, d.fileName+" "+dStatusToStr[d.status])
	}
	return list
}
