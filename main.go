package main

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	r := mux.NewRouter()

	log.Printf("Start gifer server on %s", port)
	r.SkipClean(true)
	r.HandleFunc(`/unsafe/{dimension:\d+x\d+}/{filters:filters:\w{3,}\(.*\)}/{source:.*}`, resizeHandler())
	r.HandleFunc("/version", versionHandler())
	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:" + port,
		WriteTimeout: 5 * time.Minute, // big file
		ReadTimeout:  5 * time.Minute, // big file
	}

	log.Fatal(srv.ListenAndServe())
}

// If we'd like to keep the aspect ratio,
// we need to specify only one component, either width or height, and set the other component to -1.
// ffmpeg -i big.gif -vf scale=320:-1 small.gif
func resizeHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Println("[DEBUG] Hit convert")
		dimension := parseDimension(mux.Vars(req)["dimension"])
		sourcePath, err := downloadSource(mux.Vars(req)["source"])
		if err != nil {
			log.Printf("[ERROR] Download source error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("[DEBUG]\nSource: %s [DEBUG]\nDimension: %s", sourcePath, dimension)
		cmd := exec.Command("ffmpeg", "-i", sourcePath, "-vf", dimension, "small.gif")

		var out bytes.Buffer
		cmd.Stdout = &out

		var errout bytes.Buffer
		cmd.Stderr = &errout

		err = cmd.Run()
		if err != nil {
			log.Printf("[ERROR] FFmpeg command : %v, %v, %v\n", err, out.String(), errout.String())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

func downloadSource(sourceUrl string) (string, error) {
	resp, err := http.Get(sourceUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	filepath := "/tmp/toconvert.gif"
	out, err := os.Create(filepath) // TODO Delete file after convertion
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return filepath, nil
}

func parseDimension(dim string) string {
	dimensions := strings.Split(dim, "x")
	widht, height := dimensions[0], dimensions[1]
	w, _ := strconv.ParseInt(widht, 10, 64)
	h, _ := strconv.ParseInt(height, 10, 64)
	var dimension string
	if w > 0 && h == 0 {
		dimension = fmt.Sprintf("scale=%d:-1", w)
	}
	if h > 0 && w == 0 {
		dimension = fmt.Sprintf("scale=-1:%d", h)
	}
	if w > 0 && h > 0 {
		dimension = fmt.Sprintf("scale=%d:%d", w, h)
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
