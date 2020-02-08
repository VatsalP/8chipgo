package main

import (
	"flag"
	"fmt"
	"image"
	"os"
	"time"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

var (
	scale float64
	rom   string
	debug bool
)

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
	batch := pixel.NewBatch(&pixel.TrianglesData{}, pic)
	vm := newChipVM(rom)
	var (
		frames     = 0
		second     = time.Tick(time.Second)      // for fps
		delayticks = time.Tick(time.Second / 60) // for timers
		pixelPos   [32][64]pixel.Rect
	)
	// Note down Rect for all the pixels
	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			xf, yf := float64(x)*scale, float64(y)*scale
			pixelPos[y][x] = pixel.R(xf, (32*scale)-yf, xf+scale-1, (32*scale)-yf-scale+1)
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
			if debug {
				fmt.Printf("V: %v\n", vm.v)
				fmt.Printf("I: %x\n", vm.i)
				fmt.Printf("pc: %d\n", vm.pc)
				fmt.Printf("OPCODE: %x\n", vm.memory[vm.pc])
			}
		case <-delayticks:
			if vm.delay > 0 {
				vm.delay--
			}
			if vm.sound > 0 {
				vm.sound--
				// vm.playTone()
			}
		default:
		}
	}
}

func main() {
	flag.Float64Var(&scale, "scale", 1, "resolution scaling factor")
	flag.Float64Var(&scale, "s", 1, "resolution scaling factor (shorthand)")
	flag.StringVar(&rom, "rom", "", "chip8 rom")
	flag.BoolVar(&debug, "debug", false, "print various things")
	flag.Parse()
	pixelgl.Run(run)
}
