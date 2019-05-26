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

	"github.com/cloudfoundry/bytefmt"
	"github.com/gorilla/mux"
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

func resizeHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Println("[DEBUG] Hit convert")
		var (
			dimension string
			size      string
			format    string
			err       error
		)
		dimension = parseDimension(mux.Vars(req)["dimension"])
		if format, err = parseFormat(mux.Vars(req)["filters"]); err != nil {
			log.Printf("[ERROR] Bad format: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		sourcePath, size, err := downloadSource(mux.Vars(req)["source"])
		defer os.Remove(sourcePath)
		if err != nil {
			log.Printf("[ERROR] Download source error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Printf("[DEBUG] Process file -> Extension: x / Dimension: %s", dimension)

		sourceSize, _ := strconv.ParseInt(size, 10, 64)

		log.Printf("[DEBUG] File resized length before: %s", bytefmt.ByteSize(uint64(sourceSize)))

		outfile, err := ioutil.TempFile("", "res")
		if err != nil {
			log.Printf("[ERROR] Create Outfile error %v", err)
			return
		}
		// outfile.Close()
		defer os.Remove(outfile.Name())

		cmd := exec.Command("ffmpeg",
			"-an", // disable audio
			"-y",  // overwrite
			// "-trans_color", "ffffff", // TODO read from input
			"-i", sourcePath, // set input
			"-vf", dimension,
			"-pix_fmt", "yuv420p",
			// "-movflags", "frag_keyframe",
			"-movflags", "faststart",
			// "-qmin", "10", // the minimum quantizer (default 4, range 0–63), lower - better quality --- VP9 only
			// "-qmax", "42", // the maximum quantizer (default 63, range qmin–63) higher - lower quality --- VP9 only
			"-crf", "23", // enable constant bitrate(0-51) lower - better
			"-preset", "medium", // quality preset
			"-maxrate", "500k", // max bitrate. higher - better
			"-profile:v", "baseline", // https://trac.ffmpeg.org/wiki/Encode/H.264 - compatibility level
			"-level", "4.0", // ^^^
			"-f", format,
			outfile.Name(),
		)

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

		output, _ := ioutil.ReadAll(outfile)
		imageLen := len(output)

		log.Printf("[DEBUG] File resized length after: %s", bytefmt.ByteSize(uint64(imageLen)))

		w.Header().Set("X-Filename", "video."+format)
		w.Header().Set("Content-Type", "video/"+format)
		w.Header().Set("Content-Length", strconv.Itoa(imageLen))
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(output)
		if err != nil {
			log.Printf("[ERROR] Output write error %v", err)
		}
	})
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
