package main

import (
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"net"
	"time"
)

const (
	LedpanelWidth  = 120
	LedpanelHeight = 48
)

func Listen() (chan image.Image, chan error) {
	errs := make(chan error, 1)
	out := make(chan image.Image, 1)
	out <- bitvisImage{}
	go func() {
		defer close(errs)
		defer close(out)
		service, err := net.Listen("tcp", ":1338")
		if err != nil {
			errs <- err
			return
		}
		for {
			conn, err := service.Accept()
			if err != nil {
				errs <- err
				return
			}
			go func() {
				if err := handleConnection(conn, out); err != nil {
					log.Printf("Error handling connection: %v", err)
				}
			}()
		}
	}()
	return out, errs
}

func handleConnection(conn net.Conn, out chan<- image.Image) error {
	for {
		if err := conn.SetReadDeadline(time.Now().Add(time.Second * 4)); err != nil {
			return err
		}

		var boundary [1]byte
		if _, err := io.ReadFull(conn, boundary[:]); err != nil {
			return err
		}
		if boundary[0] != ':' {
			continue
		}
		io.CopyN(ioutil.Discard, conn, 2)

		var img bitvisImage
		if _, err := io.ReadFull(conn, img[:]); err != nil {
			return err
		}
		select {
		case out <- img:
		default:
		}
	}
}

type bitvisImage [LedpanelWidth * LedpanelHeight / 4]uint8

func (img bitvisImage) ColorModel() color.Model {
	return color.RGBAModel
}

func (img bitvisImage) Bounds() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: LedpanelWidth, Y: LedpanelHeight},
	}
}

func (img bitvisImage) At(x, y int) color.Color {
	return bitvisColor(img[y*LedpanelWidth/4+x/4] >> ((3 - uint(x)%4) * 2))
}

type bitvisColor uint8

func (c bitvisColor) RGBA() (r, g, b, a uint32) {
	return uint32(c&2>>1) * 0xffff, uint32(c&1) * 0xffff, 0, 0xffff
}
