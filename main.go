package main

import (
	"bytes"
	"fmt"
	"image/png"
	"io"
	"log"
	"net/http"
	"sync"
)

func main() {
	images, errors := Listen()
	go func() {
		err := <-errors
		log.Fatal(err)
	}()
	var currentFrame BitvisImage
	var currentFrameLock sync.RWMutex
	currentFrameUpdate := sync.NewCond(currentFrameLock.RLocker())

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/html")
		res.Write([]byte(`
			<html>
			<head>
				<title>Bitvis</title>
			</head>
			<body style="margin:0; background:#000">
					<table width="100%" height="100%">
							<tr>
									<td valign=middle align=center>
											<a href="frame.png" target="_blank">
												<img style="width:100%; image-rendering:-moz-crisp-edges; image-rendering:pixelated" src="stream.mpng" />
											</a>
									</td>
							</tr>
					</table>
			</body>
			</html>
		`))
	})
	mux.HandleFunc("/stream.mpng", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=--pngboundary")
		w.WriteHeader(http.StatusOK)
		for {
			currentFrameLock.RLock()
			currentFrameUpdate.Wait()
			img := currentFrame
			currentFrameLock.RUnlock()
			buf := encodeImage(&img)
			w.Write([]byte("--pngboundary"))
			w.Write([]byte("Content-Type: image/png\n"))
			w.Write([]byte(fmt.Sprintf("Content-Length: %d\n\n", len(buf))))
			if _, err := io.Copy(w, bytes.NewReader(buf)); err != nil {
				return
			}
		}
	})
	mux.HandleFunc("/frame.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		currentFrameLock.RLock()
		img := currentFrame
		currentFrameLock.RUnlock()
		w.Write(encodeImage(&img))
	})
	go func() {
		log.Fatal(http.ListenAndServe(":13378", mux))
	}()

	for img := range images {
		currentFrameLock.Lock()
		currentFrame = img
		currentFrameUpdate.Broadcast()
		currentFrameLock.Unlock()
	}
}

func encodeImage(img *BitvisImage) []byte {
	var buf bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.BestSpeed}
	enc.Encode(&buf, img)
	return buf.Bytes()
}
