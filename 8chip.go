package main

import (
	"flag"
	"fmt"
	"image"
	"math/rand"
	"os"
	"time"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

var scale float64
var rom string

// for the pixel
func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

func run() {
	rand.Seed(time.Now().UTC().UnixNano())
	cfg := pixelgl.WindowConfig{
		Title:  "8ChipGo",
		Bounds: pixel.R(0, 0, 64*scale, 32*scale),
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	pic, err := loadPicture("pixelc.png")
	if err != nil {
		panic(err)
	}
	batch := pixel.NewBatch(&pixel.TrianglesData{}, pic)
	var (
		frames = 0
		second = time.Tick(time.Second)
	)
	var (
		pixelPos [32][64]pixel.Rect
	)
	for y, yscale := 0, 0.; y < 32; y++ {
		for x, xscale := 0, 0.; x < 64; x++ {
			xf, yf := float64(x), float64(y)
			pixelPos[y][x] = pixel.R(xf+xscale, 32.*scale-yf-yscale, xf*scale+xscale, 32.*scale-yf*scale-yscale)
			xscale += scale
		}
		yscale += scale
	}
	for !win.Closed() {
		win.Clear(colornames.Black)

		batch.Clear()

		for y := 0; y < 32; y++ {
			for x := 0; x < 64; x++ {
				if rand.Float64() > 0.5 {
					square := pixel.NewSprite(pic, pic.Bounds())
					square.Draw(batch, pixel.IM.Scaled(pixel.ZV, scale).Moved(pixelPos[y][x].Center()))
				}
			}
		}
		batch.Draw(win)
		win.Update()

		frames++
		select {
		case <-second:
			win.SetTitle(fmt.Sprintf("%s | FPS: %d", cfg.Title, frames))
			frames = 0
		default:
		}
	}
}

func main() {
	flag.Float64Var(&scale, "scale", 1, "resolution scaling factor")
	flag.Float64Var(&scale, "s", 1, "resolution scaling factor (shorthand)")
	flag.StringVar(&rom, "rom", "", "chip8 rom")
	flag.Parse()
	pixelgl.Run(run)
}
