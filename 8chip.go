package main

import (
	"flag"
	"fmt"
	"image"
	"math/rand"
	"os"
	"time"

	_ "image/png"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

var (
	scale float64
	rom   string
	debug bool
)

// noise is a beep.Streamer for playing a sound
func noise() beep.Streamer {
	return beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		for i := range samples {
			samples[i][0] = rand.Float64()*2 - 1
			samples[i][1] = rand.Float64()*2 - 1
		}
		return len(samples), true
	})
}

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

	sr := beep.SampleRate(44100)
	speaker.Init(sr, sr.N(time.Second/10))

	batch := pixel.NewBatch(&pixel.TrianglesData{}, pic)
	vm := newChipVM(rom)
	var (
		frames     = 0
		second     = time.Tick(time.Second)      // for fps
		delayticks = time.Tick(time.Second / 60) // for timers
		pixelPos   [32][64]pixel.Rect
		done       = make(chan bool)
		play       = true
	)
	// Note down Rect for all the pixels
	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			xf, yf := float64(x)*scale, float64(y)*scale
			pixelPos[y][x] = pixel.R(xf, (32*scale)-yf, xf+scale, (32*scale)-yf-scale)
		}
	}
	for !win.Closed() {
		win.Clear(colornames.Black)

		vm.fetchNextOpcode(win)

		// Draw over the screen
		batch.Clear()
		for y := 0; y < 32; y++ {
			for x := 0; x < 64; x++ {
				if vm.display[y][x] == 1 {
					square := pixel.NewSprite(pic, pic.Bounds())
					square.Draw(batch, pixel.IM.ScaledXY(pixel.ZV, pixel.V(scale, scale)).Moved(pixelPos[y][x].Center()))
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
		case <-delayticks:
			if vm.delay > 0 {
				vm.delay--
			}
			if vm.sound > 2 {
				vm.sound--
				if play {
					speaker.Play(beep.Seq(beep.Take(sr.N((time.Second/1000)*20), noise()), beep.Callback(func() {
						done <- true
					})))
					play = false
				}
			}
		case <-done:
			play = true
		default:
		}
	}
}

func main() {
	flag.Float64Var(&scale, "scale", 10, "resolution scaling factor")
	flag.Float64Var(&scale, "s", 10, "resolution scaling factor (shorthand)")
	flag.BoolVar(&debug, "debug", false, "print various things")
	flag.StringVar(&rom, "rom", "", "chip8 rom")
	flag.Parse()
	pixelgl.Run(run)
}
