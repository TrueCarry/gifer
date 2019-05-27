package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const (
	MB = 1 << 20
)

func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	r := mux.NewRouter()

	log.Printf("Start gifer server on %s", port)
	r.SkipClean(true)
	r.HandleFunc(`/unsafe/{dimension:\d+x\d+}/{filters:filters:\w{3,}\(.*\)}/{source:.*}`, resizeFromURLHandler()).Methods("GET")
	r.HandleFunc(`/unsafe/{dimension:\d+x\d+}/{filters:filters:\w{3,}\(.*\)}`, resizeFromFileHandler()).Methods("POST")
	r.HandleFunc("/version", versionHandler())
	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:" + port,
		WriteTimeout: 5 * time.Minute, // big file
		ReadTimeout:  5 * time.Minute, // big file
	}

	log.Fatal(srv.ListenAndServe())
}

func parseParams(req *http.Request) (string, string, error) {
	var (
		dimension string
		format    string
		err       error
	)
	dimension = parseDimension(mux.Vars(req)["dimension"])
	if format, err = parseFormat(mux.Vars(req)["filters"]); err != nil {
		log.Printf("[ERROR] Bad format: %s", err)
		return "", "", err
	}
	return dimension, format, nil
}

func parseFormat(format string) (string, error) {
	switch format {
	case "filters:gifv(mp4)":
		return "mp4", nil
	case "filters:gifv(webm)":
		return "webm", nil
	default:
		return "", fmt.Errorf("bad format")
	}
}

func downloadSource(sourceURL string) (string, string, error) {
	resp, err := http.Get(sourceURL)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	size := resp.Header.Get("Content-Length")

	file, err := ioutil.TempFile("", "inp")
	if err != nil {
		return "", "", err
	}

	_, err = io.Copy(file, resp.Body)
	return file.Name(), size, err
}

func parseDimension(dim string) string {
	dimensions := strings.Split(dim, "x")
	widht, height := dimensions[0], dimensions[1]
	w, _ := strconv.ParseInt(widht, 10, 64)
	h, _ := strconv.ParseInt(height, 10, 64)
	var dimension string
	switch {
	case w > 0 && h == 0:
		dimension = fmt.Sprintf("scale=trunc(%d/2)*2:-2", w)
	case h > 0 && w == 0:
		dimension = fmt.Sprintf("scale=-2:trunc(%d/2)*2", h)
	case w > 0 && h > 0:
		dimension = fmt.Sprintf("scale=w=%v:h=%v:force_original_aspect_ratio=increase,crop=%v:%v", w, h, w, h)
	default:
		dimension = "scale=trunc(iw/2)*2:trunc(ih/2)*2"
	}
	return dimension
}

func versionHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Println("[DEBUG] Hit version")
		cmd := exec.Command("ffmpeg", "--help")

		var out bytes.Buffer
		cmd.Stdout = &out

		var errout bytes.Buffer
		cmd.Stderr = &errout

		err := cmd.Run()

		if err != nil {
			log.Printf("[ERROR] FFmpeg output: %v, %v, %v\n", err, out.String(), errout.String())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Printf("[DEBUG] FFmpeg output: %v", out.String())

		w.WriteHeader(http.StatusOK)
	})
}
