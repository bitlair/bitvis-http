package main

import (
	"bytes"
	"fmt"
	"image/png"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {
	images, errors := Listen()
	go func() {
		err := <-errors
		log.Fatal(err)
	}()
	var currentFrame []byte
	var currentFrameLock sync.RWMutex

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
	mux.HandleFunc("/stream.mpng", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=--pngboundary")
		res.WriteHeader(http.StatusOK)
		for {
			currentFrameLock.RLock()
			buf := currentFrame
			currentFrameLock.RUnlock()
			res.Write([]byte("--pngboundary"))
			res.Write([]byte("Content-Type: image/png\n"))
			res.Write([]byte(fmt.Sprintf("Content-Length: %d\n\n", len(buf))))
			if _, err := io.Copy(res, bytes.NewReader(buf)); err != nil {
				return
			}
			time.Sleep(time.Millisecond * 2)
		}
	})
	mux.HandleFunc("/frame.png", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "image/png")
		currentFrameLock.RLock()
		buf := currentFrame
		currentFrameLock.RUnlock()
		io.Copy(res, bytes.NewReader(buf))
	})
	go func() {
		log.Fatal(http.ListenAndServe(":13378", mux))
	}()

	for img := range images {
		var buf bytes.Buffer
		enc := png.Encoder{CompressionLevel: png.BestSpeed}
		enc.Encode(&buf, img)
		currentFrameLock.Lock()
		currentFrame = buf.Bytes()
		currentFrameLock.Unlock()
	}
}
