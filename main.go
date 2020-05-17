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
	debug := flag.Bool("d", false, "print debug info")
	windows := flag.Bool("w", false, "show vram and nametable windows")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "no rom provided")
		os.Exit(1)
	}

	if err := run(flag.Arg(0), *debug, *windows); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(romPath string, debug bool, windows bool) error {
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

	renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
	renderer.SetIntegerScale(true)
	renderer.SetLogicalSize(160, 144)
	renderer.SetDrawColor(0, 0, 0, 0xff)

	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, 160, 144)
	if err != nil {
		panic(err)
	}

	var nametableWindow, vramWindow *sdl.Window
	var nametableRenderer, vramRenderer *sdl.Renderer
	var nametableTex, vramTex *sdl.Texture
	if windows {
		var err error
		nametableWindow, err = sdl.CreateWindow(
			"gb",
			sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
			512*2, 256*2,
			sdl.WINDOW_SHOWN,
		)
		if err != nil {
			panic(err)
		}
		defer nametableWindow.Destroy()

		nametableRenderer, err = sdl.CreateRenderer(nametableWindow, -1, sdl.RENDERER_ACCELERATED)
		if err != nil {
			panic(err)
		}

		nametableRenderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
		nametableRenderer.SetIntegerScale(true)
		nametableRenderer.SetLogicalSize(512, 256)
		nametableRenderer.SetDrawColor(0, 0, 0, 0xff)

		nametableTex, err = nametableRenderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, 512, 256)
		if err != nil {
			panic(err)
		}

		vramWindow, err = sdl.CreateWindow(
			"gb",
			sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
			128*4, 192*4,
			sdl.WINDOW_SHOWN,
		)
		if err != nil {
			panic(err)
		}
		defer vramWindow.Destroy()

		vramRenderer, err = sdl.CreateRenderer(vramWindow, -1, sdl.RENDERER_ACCELERATED)
		if err != nil {
			panic(err)
		}

		vramRenderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
		vramRenderer.SetIntegerScale(true)
		vramRenderer.SetLogicalSize(128, 192)
		vramRenderer.SetDrawColor(0, 0, 0, 0xff)

		vramTex, err = vramRenderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, 128, 192)
		if err != nil {
			panic(err)
		}
	}

	cart, err := gb.NewCartridge(rom)
	if err != nil {
		return fmt.Errorf("could not load rom: %w", err)
	}

	console := gb.New(cart, debug)
	fmt.Println(console.CartridgeInfo())

	console.PowerOn()

	var fullscreen, grid bool
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

				case evt.Keysym.Sym == sdl.K_g && evt.State == sdl.PRESSED && evt.Repeat == 0:
					grid = !grid
					break

				case evt.Keysym.Sym == sdl.K_d && evt.State == sdl.PRESSED && evt.Repeat == 0 && evt.Keysym.Mod&sdl.KMOD_CTRL > 0:
					if err := console.DumpWram(); err != nil {
						fmt.Fprintf(os.Stderr, "unable to dump wram: %v\n", err)
					}

				case evt.Keysym.Sym == sdl.K_s && evt.State == sdl.PRESSED && evt.Repeat == 0 && evt.Keysym.Mod&sdl.KMOD_ALT > 0:
					console.ToggleSprites()
				case evt.Keysym.Sym == sdl.K_b && evt.State == sdl.PRESSED && evt.Repeat == 0 && evt.Keysym.Mod&sdl.KMOD_ALT > 0:
					console.ToggleBackground()
				case evt.Keysym.Sym == sdl.K_w && evt.State == sdl.PRESSED && evt.Repeat == 0 && evt.Keysym.Mod&sdl.KMOD_ALT > 0:
					console.ToggleWindow()

				case evt.Keysym.Sym == sdl.K_x:
					console.Press(gb.A, evt.State == sdl.PRESSED)
				case evt.Keysym.Sym == sdl.K_z:
					console.Press(gb.B, evt.State == sdl.PRESSED)
				case evt.Keysym.Sym == sdl.K_RETURN:
					console.Press(gb.Start, evt.State == sdl.PRESSED)
				case evt.Keysym.Sym == sdl.K_BACKSPACE:
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

		renderer.SetDrawColor(0x00, 0x00, 0x00, 0xff)
		renderer.Clear()
		renderer.Copy(tex, nil, nil)

		if grid {
			renderer.SetDrawColor(0xff, 0x00, 0x00, 0x33)
			for y := int32(0); y < 18; y++ {
				if y%2 == 0 {
					for x := int32(1); x < 20; x += 2 {
						renderer.FillRect(&sdl.Rect{X: x * 8, Y: y * 8, W: 8, H: 8})
					}
				}
				if y%2 == 1 {
					for x := int32(0); x < 20; x += 2 {
						renderer.FillRect(&sdl.Rect{X: x * 8, Y: y * 8, W: 8, H: 8})
					}
				}
			}
		}
		renderer.Present()

		if windows {
			nametableFrame := console.DrawNametables()
			nametableData, _, err := nametableTex.Lock(nil)
			if err != nil {
				panic(err)
			}
			copy(nametableData, nametableFrame)
			nametableTex.Unlock()
			nametableRenderer.SetDrawColor(0x00, 0x00, 0x00, 0xff)
			nametableRenderer.Clear()
			nametableRenderer.Copy(nametableTex, nil, nil)

			if grid {
				nametableRenderer.SetDrawColor(0xff, 0x00, 0x00, 0x33)
				for y := int32(0); y < 32; y++ {
					if y%2 == 0 {
						for x := int32(1); x < 64; x += 2 {
							nametableRenderer.FillRect(&sdl.Rect{X: x * 8, Y: y * 8, W: 8, H: 8})
						}
					}
					if y%2 == 1 {
						for x := int32(0); x < 64; x += 2 {
							nametableRenderer.FillRect(&sdl.Rect{X: x * 8, Y: y * 8, W: 8, H: 8})
						}
					}
				}
				nametableRenderer.SetDrawColor(0x1a, 0xcb, 0xe8, 0x99)
				nametableRenderer.DrawLine(32*8, 0, 32*8, 32*8)
			}
			nametableRenderer.Present()

			vramFrame := console.DrawVram()
			vramData, _, err := vramTex.Lock(nil)
			if err != nil {
				panic(err)
			}
			copy(vramData, vramFrame)
			vramTex.Unlock()
			vramRenderer.SetDrawColor(0x00, 0x00, 0x00, 0xff)
			vramRenderer.Clear()
			vramRenderer.Copy(vramTex, nil, nil)

			if grid {
				vramRenderer.SetDrawColor(0xff, 0x00, 0x00, 0x33)
				for y := int32(0); y < 24; y++ {
					if y%2 == 0 {
						for x := int32(1); x < 16; x += 2 {
							vramRenderer.FillRect(&sdl.Rect{X: x * 8, Y: y * 8, W: 8, H: 8})
						}
					}
					if y%2 == 1 {
						for x := int32(0); x < 16; x += 2 {
							vramRenderer.FillRect(&sdl.Rect{X: x * 8, Y: y * 8, W: 8, H: 8})
						}
					}
				}
				vramRenderer.SetDrawColor(0x1a, 0xcb, 0xe8, 0x99)
				vramRenderer.DrawLine(0, 8*8, 128, 8*8)
				vramRenderer.DrawLine(0, 16*8, 128, 16*8)
			}
			vramRenderer.Present()
		}

		frameTime := float64(time.Since(frameStart).Milliseconds())
		if frameTime < targetFrameTime {
			time.Sleep(time.Duration(targetFrameTime-frameTime) * time.Millisecond)
		}
	}

	return nil
}
