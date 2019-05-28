package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/cloudfoundry/bytefmt"
	"github.com/gorilla/mux"
)

func resizeFromURLHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Println("[DEBUG] Hit resize from URL ...")
		dimension, format, err := parseParams(req)
		if err != nil {
			log.Printf("[ERROR] Download source error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var size string
		sourcePath, size, err := downloadSource(mux.Vars(req)["source"])
		defer os.Remove(sourcePath)
		if err != nil {
			log.Printf("[ERROR] Download source error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("[DEBUG] Process file -> Dimension: %s", dimension)

		sourceSize, _ := strconv.ParseInt(size, 10, 64)

		log.Printf("[DEBUG] File resized length before: %s", bytefmt.ByteSize(uint64(sourceSize)))

		outfile, err := ioutil.TempFile("", "res")
		if err != nil {
			log.Printf("[ERROR] Create Outfile error %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer func() {
			os.Remove(outfile.Name())
		}()

		cmd := exec.Command("ffmpeg",
			"-an", // disable audio
			"-y",  // overwrite
			// "-trans_color", "ffffff", // TODO read from input
			"-i", sourcePath, // set input
			"-vf", dimension,
			"-c:v", "libvpx-vp9", // https://trac.ffmpeg.org/wiki/Encode/VP9
			"-b:v", "0",
			"-crf", "30", // enable constant bitrate(0-51) lower - better
			// "-pix_fmt", "yuv420p",
			// "-movflags", "frag_keyframe",
			// "-movflags", "faststart",
			// "-qmin", "10", // the minimum quantizer (default 4, range 0–63), lower - better quality --- VP9 only
			// "-qmax", "42", // the maximum quantizer (default 63, range qmin–63) higher - lower quality --- VP9 only
			// "-crf", "23", // enable constant bitrate(0-51) lower - better
			// "-preset", "medium", // quality preset
			// "-maxrate", "500k", // max bitrate. higher - better
			// "-profile:v", "baseline", // https://trac.ffmpeg.org/wiki/Encode/H.264 - compatibility level
			// "-level", "4.0", // ^^^
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
