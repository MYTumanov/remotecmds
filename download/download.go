package download

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
	dErr
	stop
)

var dStatusToStr = map[dStatus]string{
	queue:      "queue",
	inprogress: "inprogress",
	complete:   "complete",
	dErr:       "err",
	stop:       "stop",
}

type downloads struct {
	fileName     string
	filePath     string
	url          string
	status       dStatus
	byteDownload int64
	stopLoad     chan bool
}

var downloadsList map[string]*downloads

func (d *downloads) Write(p []byte) (int, error) {
	n := len(p)
	select {
	case <-d.stopLoad:
		log.Println("Stop load signal")
		return n, NewErrStopDownloadSignal("Stop load signal")
	default:
		d.byteDownload += int64(n)
		log.Println(d.byteDownload)
		return n, nil
	}
}

// Download ownloads files from url to path
func Download(urlStr string, path string) {
	if downloadsList == nil {
		downloadsList = make(map[string]*downloads)
	}
	newDownload := &downloads{
		fileName:     getName(urlStr),
		filePath:     path,
		url:          urlStr,
		status:       queue,
		stopLoad:     make(chan bool),
		byteDownload: 0,
	}
	downloadsList[urlStr] = newDownload
	err := newDownload.download()
	if err != nil {
		log.Println(err)
		newDownload.status = dErr
	}

}

func (d *downloads) download() error {
	d.status = inprogress
	client := &http.Client{}
	req, err := http.NewRequest("GET", d.url, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	req.Header.Set("Range", "bytes="+strconv.FormatInt(d.byteDownload, 10)+"-")

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer resp.Body.Close()

	// Work with file

	fullTempPath := filepath.Join(d.filePath, d.fileName+".temp")
	fullPath := filepath.Join(d.filePath, d.fileName)
	// f, err = os.Create(fullTempPath)
	f, err := os.OpenFile(fullTempPath, os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Println(err)
		return err
	}
	defer f.Close()

	d.byteDownload, err = io.Copy(f, io.TeeReader(resp.Body, d))
	if err != nil {
		log.Println(err)
		switch err.(type) {
		case *ErrStopDownloadSignal:
			err = f.Sync()
			if err != nil {
				log.Println(err)
			}
			d.status = stop
			return err
		default:
			return err
		}
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

// StopDownload stop download by id
func StopDownload() {}

// ResumeDownload resumes given download
func ResumeDownload(d string) {
	if dwnld, ok := downloadsList[d]; ok && dwnld.status == stop {
		err := dwnld.download()
		if err != nil {
			log.Println(err)
		}
		return
	}
	log.Println("Not found download: " + d)
}
