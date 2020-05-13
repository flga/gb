package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/flga/gb/gb"
	"github.com/veandco/go-sdl2/sdl"
)

const targetFrameTime = 1000 / float64(60)

func init() {
	runtime.LockOSThread()
}

func main() {
	disasm := flag.Bool("d", false, "print debug info")
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
		160*4, 144*4,
		sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE,
	)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}

	// nametableWindow, err := sdl.CreateWindow(
	// 	"gb",
	// 	sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
	// 	512*2, 256*2,
	// 	sdl.WINDOW_SHOWN,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// defer nametableWindow.Destroy()

	// nametableRenderer, err := sdl.CreateRenderer(nametableWindow, -1, sdl.RENDERER_ACCELERATED)
	// if err != nil {
	// 	panic(err)
	// }

	// vramWindow, err := sdl.CreateWindow(
	// 	"gb",
	// 	sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
	// 	128*4, 192*4,
	// 	sdl.WINDOW_SHOWN,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// defer vramWindow.Destroy()

	// vramRenderer, err := sdl.CreateRenderer(vramWindow, -1, sdl.RENDERER_ACCELERATED)
	// if err != nil {
	// 	panic(err)
	// }

	renderer.SetIntegerScale(true)
	renderer.SetLogicalSize(160, 144)
	renderer.SetDrawColor(0, 0, 0, 0xff)

	// nametableRenderer.SetIntegerScale(true)
	// nametableRenderer.SetLogicalSize(512, 256)
	// nametableRenderer.SetDrawColor(0, 0, 0, 0xff)

	// vramRenderer.SetIntegerScale(true)
	// vramRenderer.SetLogicalSize(128, 192)
	// vramRenderer.SetDrawColor(0, 0, 0, 0xff)

	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, 160, 144)
	if err != nil {
		panic(err)
	}

	// nametableTex, err := nametableRenderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, 512, 256)
	// if err != nil {
	// 	panic(err)
	// }

	// vramTex, err := vramRenderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, 128, 192)
	// if err != nil {
	// 	panic(err)
	// }

	// _ = nametableTex
	// _ = vramTex

	cart, err := gb.NewCartridge(rom)
	if err != nil {
		return fmt.Errorf("could not load rom: %w", err)
	}

	console := gb.New(cart, disasm)
	fmt.Println(console.CartridgeInfo())

	console.PowerOn()

	var fullscreen bool
	running := true
	for running {
		frameStart := time.Now()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch evt := event.(type) {
			case *sdl.QuitEvent:
				running = false
				break

			case *sdl.KeyboardEvent:
				switch {
				case evt.Keysym.Sym == sdl.K_w && evt.State == sdl.PRESSED && evt.Repeat == 0 && evt.Keysym.Mod&sdl.KMOD_CTRL > 0:
					running = false
					break

				case evt.Keysym.Sym == sdl.K_f && evt.State == sdl.PRESSED && evt.Repeat == 0:
					if fullscreen {
						window.SetFullscreen(0)
					} else {
						window.SetFullscreen(sdl.WINDOW_FULLSCREEN)
					}
					fullscreen = !fullscreen
					break

				case evt.Keysym.Sym == sdl.K_d && evt.State == sdl.PRESSED && evt.Repeat == 0:
					console.ToggleDebugInfo()
					break

				case evt.Keysym.Sym == sdl.K_d && evt.State == sdl.PRESSED && evt.Repeat == 0 && evt.Keysym.Mod&sdl.KMOD_CTRL > 0:
					if err := console.DumpWram(); err != nil {
						fmt.Fprintf(os.Stderr, "unable to dump wram: %v\n", err)
					}

				case evt.Keysym.Sym == sdl.K_a:
					console.Press(gb.A, evt.State == sdl.PRESSED)
				case evt.Keysym.Sym == sdl.K_b:
					console.Press(gb.B, evt.State == sdl.PRESSED)
				case evt.Keysym.Sym == sdl.K_z:
					console.Press(gb.Start, evt.State == sdl.PRESSED)
				case evt.Keysym.Sym == sdl.K_x:
					console.Press(gb.Select, evt.State == sdl.PRESSED)
				case evt.Keysym.Sym == sdl.K_UP:
					console.Press(gb.Up, evt.State == sdl.PRESSED)
				case evt.Keysym.Sym == sdl.K_DOWN:
					console.Press(gb.Down, evt.State == sdl.PRESSED)
				case evt.Keysym.Sym == sdl.K_LEFT:
					console.Press(gb.Left, evt.State == sdl.PRESSED)
				case evt.Keysym.Sym == sdl.K_RIGHT:
					console.Press(gb.Right, evt.State == sdl.PRESSED)
				}
			}
		}

		frame := console.ClockFrame()
		data, _, err := tex.Lock(nil)
		if err != nil {
			panic(err)
		}
		copy(data, frame)
		tex.Unlock()

		renderer.Clear()
		renderer.Copy(tex, nil, nil)
		renderer.Present()

		// nametableFrame := console.DrawNametables()
		// nametableData, _, err := nametableTex.Lock(nil)
		// if err != nil {
		// 	panic(err)
		// }
		// copy(nametableData, nametableFrame)
		// nametableTex.Unlock()
		// nametableRenderer.Clear()
		// nametableRenderer.Copy(nametableTex, nil, nil)
		// nametableRenderer.Present()

		// vramFrame := console.DrawVram()
		// vramData, _, err := vramTex.Lock(nil)
		// if err != nil {
		// 	panic(err)
		// }
		// copy(vramData, vramFrame)
		// vramTex.Unlock()
		// vramRenderer.Clear()
		// vramRenderer.Copy(vramTex, nil, nil)
		// vramRenderer.SetDrawColor(0xff, 0x00, 0x00, 0xff)
		// for x := int32(1); x < 16; x++ {
		// 	vramRenderer.DrawLine(x*8, 0, x*8, 192)
		// }
		// for y := int32(1); y < 24; y++ {
		// 	vramRenderer.DrawLine(0, y*8, 128, y*8)
		// }
		// vramRenderer.Present()

		frameTime := float64(time.Since(frameStart).Milliseconds())
		if frameTime < targetFrameTime {
			time.Sleep(time.Duration(targetFrameTime-frameTime) * time.Millisecond)
		}
	}

	return nil
}
