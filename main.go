package main

import (
	"flag"
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
	disasm := flag.Bool("d", false, "print disassembly when executing an instr")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "no rom provided")
		os.Exit(1)
	}

	if err := run(flag.Arg(0), *disasm); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(romPath string, disasm bool) error {
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

	console, err := gb.New(rom, disasm)
	if err != nil {
		return fmt.Errorf("could not load rom: %w", err)
	}
	fmt.Println(console.CartridgeInfo())

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch evt := event.(type) {
			case *sdl.QuitEvent:
				running = false
				break

			case *sdl.KeyboardEvent:
				if evt.State == sdl.PRESSED && evt.Repeat == 0 && evt.Keysym.Sym == sdl.K_d && evt.Keysym.Mod&sdl.KMOD_CTRL > 0 {
					if err := console.DumpWram(); err != nil {
						fmt.Fprintf(os.Stderr, "unable to dump wram: %v\n", err)
					}
				}
			}
		}

		for i := 0; i < 69905; i++ {
			console.Clock()
		}

		renderer.SetDrawColor(255, 0, 0, 255)
		renderer.FillRect(&sdl.Rect{X: 0, Y: 0, W: 200, H: 200})
		renderer.Present()
	}

	return nil
}
