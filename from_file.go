package main

import (
	"bytes"
	"github.com/cloudfoundry/bytefmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"os/exec"
	"strconv"

	"log"
	"net/http"
)

func resizeFromFileHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Println("[DEBUG] Hit resize from FILE ...")
		dimension, format, err := parseParams(req)
		if err != nil {
			log.Printf("[ERROR] Download source error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := req.ParseMultipartForm(5 * MB); nil != err {
			log.Printf("[ERROR] while parse: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for _, fheaders := range req.MultipartForm.File {
			for _, hdr := range fheaders {
				log.Printf("Income file len: %d", hdr.Size)

				var err error
				var infile multipart.File

				if infile, err = hdr.Open(); err != nil {
					log.Printf("[ERROR] Handle open error: %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				inputfile, err := ioutil.TempFile("", "*")
				if err != nil {
					log.Printf("[ERROR] Create Input error %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				io.Copy(inputfile, infile)

				log.Printf("[DEBUG] Process file -> Dimension: %s", dimension)

				sourceSize := hdr.Size

				log.Printf("[DEBUG] File resized length before: %s", bytefmt.ByteSize(uint64(sourceSize)))

				outfile, err := ioutil.TempFile("", "res")
				if err != nil {
					log.Printf("[ERROR] Create Outfile error %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				defer func() {
					os.Remove(outfile.Name())
					os.Remove(inputfile.Name())
				}()

				cmd := exec.Command("ffmpeg",
					"-an", // disable audio
					"-y",  // overwrite
					// "-trans_color", "ffffff", // TODO read from input
					"-i", inputfile.Name(), // set input
					"-vf", dimension,
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

			}
		}
	})
}
