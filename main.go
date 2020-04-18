package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/flga/gb/gb"
	"github.com/veandco/go-sdl2/sdl"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "no rom provided")
		os.Exit(1)
	}

	if err := run(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(romPath string) error {
	rom, err := os.Open(romPath)
	if err != nil {
		return fmt.Errorf("could not load rom: %w", err)
	}
	defer rom.Close()

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow(
		"gb",
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		160*8, 144*8,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		panic(err)
	}

	renderer.SetIntegerScale(true)
	renderer.SetLogicalSize(160, 144)

	console := gb.New()
	if err := console.InsertCartridge(rom); err != nil {
		return fmt.Errorf("could not load rom: %w", err)
	}

	fmt.Println(console.CartridgeInfo())

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
				break
			}
		}

		renderer.SetDrawColor(255, 0, 0, 255)
		renderer.FillRect(&sdl.Rect{X: 0, Y: 0, W: 200, H: 200})
		renderer.Present()
	}

	return nil
}
