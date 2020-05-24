package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"image/color"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/flga/gb/gb"
	"github.com/veandco/go-sdl2/sdl"
)

const targetFrameTime = 1000 / float64(60)

var (
	black        = color.RGBA{0x00, 0x00, 0x00, 0xFF}
	gridColor    = color.RGBA{0xff, 0x00, 0x00, 0x33}
	dividerColor = color.RGBA{0x1a, 0xcb, 0xe8, 0x99}
)

func init() {
	runtime.LockOSThread()
}

func main() {
	debug := flag.Bool("d", false, "print debug info")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	go func() {
		<-sigChan
		cancel()
	}()

	if err := run(ctx, flag.Arg(0), *debug); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, romPath string, debug bool) error {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	mainWindow := &window{
		Title:         "gb",
		W:             160,
		H:             144,
		Scale:         4,
		WindowFlags:   sdl.WINDOW_SHOWN | sdl.WINDOW_RESIZABLE,
		RendererFlags: sdl.RENDERER_ACCELERATED | sdl.RENDERER_PRESENTVSYNC,
		BlendMode:     sdl.BLENDMODE_BLEND,
		PixelFormat:   sdl.PIXELFORMAT_ABGR8888,
	}
	defer mainWindow.Destroy()

	nametableWindow := &window{
		// Hidden:        true,
		Title:         "nametables",
		W:             512,
		H:             256,
		Scale:         2,
		WindowFlags:   sdl.WINDOW_HIDDEN,
		RendererFlags: sdl.RENDERER_ACCELERATED | sdl.RENDERER_PRESENTVSYNC,
		BlendMode:     sdl.BLENDMODE_BLEND,
		PixelFormat:   sdl.PIXELFORMAT_ABGR8888,
	}
	defer nametableWindow.Destroy()

	vramWindow := &window{
		// Hidden:        true,
		Title:         "vram",
		W:             128,
		H:             192,
		Scale:         4,
		WindowFlags:   sdl.WINDOW_HIDDEN,
		RendererFlags: sdl.RENDERER_ACCELERATED | sdl.RENDERER_PRESENTVSYNC,
		BlendMode:     sdl.BLENDMODE_BLEND,
		PixelFormat:   sdl.PIXELFORMAT_ABGR8888,
	}
	defer vramWindow.Destroy()

	console := &gb.GameBoy{
		Debug: debug,
	}
	defer console.Save()
	if romPath != "" {
		if err := loadRom(romPath, console); err != nil {
			return err
		}
	}

	running := true
	turbo := false
Loop:
	for running {
		frameStart := time.Now()
		select {
		case <-ctx.Done():
			running = false
		default:
		}

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch evt := event.(type) {
			case *sdl.QuitEvent:
				running = false
				break Loop
			case *sdl.WindowEvent:
				switch evt.Event {
				case sdl.WINDOWEVENT_CLOSE:
					switch evt.WindowID {
					case mainWindow.ID:
						running = false
						break Loop
					case nametableWindow.ID:
						nametableWindow.Hide()
					case vramWindow.ID:
						vramWindow.Hide()
					}
				}

			case *sdl.DropEvent:
				if evt.Type != sdl.DROPFILE {
					continue
				}
				if err := loadRom(evt.File, console); err != nil {
					return err
				}

			case *sdl.KeyboardEvent:
				switch {
				case evt.Keysym.Sym == sdl.K_w && evt.State == sdl.PRESSED && evt.Repeat == 0 && evt.Keysym.Mod&sdl.KMOD_CTRL > 0:
					switch evt.WindowID {
					case mainWindow.ID:
						running = false
					case nametableWindow.ID:
						nametableWindow.Hide()
					case vramWindow.ID:
						vramWindow.Hide()
					}

				case evt.Keysym.Sym == sdl.K_f && evt.State == sdl.PRESSED && evt.Repeat == 0 && evt.WindowID == mainWindow.ID:
					mainWindow.ToggleFullscreen()

				case evt.Keysym.Sym == sdl.K_SPACE && evt.State == sdl.PRESSED && evt.Repeat == 0:
					turbo = true
				case evt.Keysym.Sym == sdl.K_SPACE && evt.State == sdl.RELEASED && evt.Repeat == 0:
					turbo = false

				case evt.Keysym.Sym == sdl.K_d && evt.State == sdl.PRESSED && evt.Repeat == 0:
					console.Debug = !console.Debug

				case evt.Keysym.Sym == sdl.K_g && evt.State == sdl.PRESSED && evt.Repeat == 0:
					if evt.Keysym.Mod&sdl.KMOD_ALT > 0 {
						mainWindow.ToggleGrid()
						nametableWindow.ToggleGrid()
						vramWindow.ToggleGrid()
					} else {
						switch evt.WindowID {
						case mainWindow.ID:
							mainWindow.ToggleGrid()
						case nametableWindow.ID:
							nametableWindow.ToggleGrid()
						case vramWindow.ID:
							vramWindow.ToggleGrid()
						}
					}

				case evt.Keysym.Sym == sdl.K_d && evt.State == sdl.PRESSED && evt.Repeat == 0 && evt.Keysym.Mod&sdl.KMOD_CTRL > 0:
					if err := console.DumpWram(); err != nil {
						fmt.Fprintf(os.Stderr, "unable to dump wram: %v\n", err)
					}

				case evt.Keysym.Sym == sdl.K_F1 && evt.State == sdl.PRESSED && evt.Repeat == 0:
					nametableWindow.Focus()
				case evt.Keysym.Sym == sdl.K_F2 && evt.State == sdl.PRESSED && evt.Repeat == 0:
					vramWindow.Focus()

				case evt.Keysym.Sym == sdl.K_s && evt.State == sdl.PRESSED && evt.Repeat == 0 && evt.Keysym.Mod&sdl.KMOD_ALT > 0:
					console.ToggleSprites()
				case evt.Keysym.Sym == sdl.K_b && evt.State == sdl.PRESSED && evt.Repeat == 0 && evt.Keysym.Mod&sdl.KMOD_ALT > 0:
					console.ToggleBackground()
				case evt.Keysym.Sym == sdl.K_w && evt.State == sdl.PRESSED && evt.Repeat == 0 && evt.Keysym.Mod&sdl.KMOD_ALT > 0:
					console.ToggleWindow()

				// case evt.Keysym.Sym == sdl.K_s && evt.State == sdl.PRESSED && evt.Repeat == 0 && evt.Keysym.Mod&sdl.KMOD_CTRL > 0:
				// 	console.Save()

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

		if turbo {
			for i := 0; i < 9; i++ {
				console.ClockFrame()
			}
		}

		frame := console.ClockFrame()
		mainWindow.Clear(black)
		mainWindow.Update(frame)
		mainWindow.DrawGrid(gridColor)

		nametableFrame := console.DrawNametables()
		nametableWindow.Clear(black)
		nametableWindow.Update(nametableFrame)
		nametableWindow.DrawGrid(gridColor)
		nametableWindow.SetDrawColor(dividerColor)
		nametableWindow.DrawDivider(dividerColor, 2, true)

		vramFrame := console.DrawVram()
		vramWindow.Clear(black)
		vramWindow.Update(vramFrame)
		vramWindow.DrawGrid(gridColor)
		vramWindow.DrawDivider(dividerColor, 3, false)

		frameTime := float64(time.Since(frameStart).Milliseconds())
		if frameTime < targetFrameTime {
			time.Sleep(time.Duration(targetFrameTime-frameTime) * time.Millisecond)
		}

		mainWindow.Present()
		nametableWindow.Present()
		vramWindow.Present()
	}

	return nil
}

func loadRom(path string, console *gb.GameBoy) error {
	rom, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not load rom: %w", err)
	}
	defer rom.Close()

	cart, err := gb.NewCartridge(rom)
	if err != nil {
		return fmt.Errorf("could not load rom: %w", err)
	}

	if !cart.Saveable() {
		if err := console.InsertCartridge(cart, nil, nil); err != nil {
			return err
		}
		console.PowerOn()
		return nil
	}

	savr, savw, err := openSavFile(strings.TrimSuffix(path, filepath.Ext(path)) + ".sav")
	if err != nil {
		return fmt.Errorf("could not load sav: %w", err)
	}

	if err := console.InsertCartridge(cart, savr, savw); err != nil {
		return err
	}

	fmt.Println(console.CartridgeInfo())
	console.PowerOn()

	return nil
}

func openSavFile(path string) (r io.Reader, w io.WriteCloser, err error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, nil, err
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		f.Close()
		return nil, nil, err
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		f.Close()
		return nil, nil, err
	}

	return bytes.NewReader(data), f, nil
}

type window struct {
	Title                      string
	W, H                       int
	Scale                      int
	WindowFlags, RendererFlags uint32
	BlendMode                  sdl.BlendMode
	PixelFormat                uint32

	ID       uint32
	window   *sdl.Window
	renderer *sdl.Renderer
	tex      *sdl.Texture

	once       sync.Once
	fullscreen bool
	grid       bool
}

func (w *window) init() error {
	var err error
	w.once.Do(func() {
		w.window, err = sdl.CreateWindow(
			w.Title,
			sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
			int32(w.W*w.Scale), int32(w.H*w.Scale),
			w.WindowFlags,
		)
		if err != nil {
			return
		}
		w.ID, err = w.window.GetID()
		if err != nil {
			return
		}

		w.renderer, err = sdl.CreateRenderer(w.window, -1, w.RendererFlags)
		if err != nil {
			return
		}
		err = w.renderer.SetDrawBlendMode(w.BlendMode)
		if err != nil {
			return
		}
		err = w.renderer.SetIntegerScale(true)
		if err != nil {
			return
		}
		err = w.renderer.SetLogicalSize(int32(w.W), int32(w.H))
		if err != nil {
			return
		}
		err = w.renderer.SetDrawColor(0x00, 0x00, 0x00, 0xFF)
		if err != nil {
			return
		}

		w.tex, err = w.renderer.CreateTexture(w.PixelFormat, sdl.TEXTUREACCESS_STREAMING, int32(w.W), int32(w.H))
		if err != nil {
			return
		}
	})

	if err != nil {
		w.Destroy()
	}

	return err
}

func (w *window) Show() error {
	if err := w.init(); err != nil {
		return err
	}

	w.WindowFlags |= sdl.WINDOW_SHOWN
	w.window.Show()

	return nil
}

func (w *window) Focus() error {
	if err := w.init(); err != nil {
		return err
	}

	if w.WindowFlags&sdl.WINDOW_SHOWN == 0 {
		w.WindowFlags |= sdl.WINDOW_SHOWN
		w.window.Show()
	}

	w.window.Raise()

	return nil
}

func (w *window) Hide() error {
	if err := w.init(); err != nil {
		return err
	}

	w.WindowFlags &^= sdl.WINDOW_SHOWN
	w.window.Hide()

	return nil
}

func (w *window) Toggle() error {
	if err := w.init(); err != nil {
		return err
	}

	if w.WindowFlags&sdl.WINDOW_SHOWN > 0 {
		w.WindowFlags &^= sdl.WINDOW_SHOWN
		w.window.Hide()
	} else {
		w.WindowFlags |= sdl.WINDOW_SHOWN
		w.window.Show()
	}

	return nil
}

func (w *window) ToggleFullscreen(f ...uint32) error {
	if w.WindowFlags&sdl.WINDOW_SHOWN == 0 {
		return nil
	}

	if err := w.init(); err != nil {
		return err
	}

	if w.fullscreen {
		w.fullscreen = false
		return w.window.SetFullscreen(0)
	}

	var flag uint32
	if len(f) == 0 {
		flag = sdl.WINDOW_FULLSCREEN
	} else {
		flag = f[0]
	}

	w.fullscreen = true
	return w.window.SetFullscreen(flag)
}

func (w *window) Clear(c color.RGBA) error {
	if w.WindowFlags&sdl.WINDOW_SHOWN == 0 {
		return nil
	}

	if err := w.init(); err != nil {
		return err
	}

	if err := w.renderer.SetDrawColor(c.R, c.G, c.B, c.A); err != nil {
		return err
	}

	return w.renderer.Clear()
}

func (w *window) SetDrawColor(c color.RGBA) error {
	if w.WindowFlags&sdl.WINDOW_SHOWN == 0 {
		return nil
	}

	if err := w.init(); err != nil {
		return err
	}

	return w.renderer.SetDrawColor(c.R, c.G, c.B, c.A)
}

func (ww *window) FillRect(x, y, w, h int) error {
	if ww.WindowFlags&sdl.WINDOW_SHOWN == 0 {
		return nil
	}

	if err := ww.init(); err != nil {
		return err
	}

	return ww.renderer.FillRect(&sdl.Rect{X: int32(x), Y: int32(y), W: int32(w), H: int32(h)})
}

func (w *window) DrawLine(x1, y1, x2, y2 int) error {
	if w.WindowFlags&sdl.WINDOW_SHOWN == 0 {
		return nil
	}

	if err := w.init(); err != nil {
		return err
	}

	return w.renderer.DrawLine(int32(x1), int32(y1), int32(x2), int32(y2))
}

func (w *window) ToggleGrid() {
	w.grid = !w.grid
}

func (w *window) DrawGrid(c color.RGBA) error {
	if w.WindowFlags&sdl.WINDOW_SHOWN == 0 {
		return nil
	}

	if !w.grid {
		return nil
	}

	if err := w.SetDrawColor(c); err != nil {
		return err
	}
	for y := int32(0); y < int32(w.H); y++ {
		if y%2 == 0 {
			for x := int32(1); x < int32(w.W/8); x += 2 {
				if err := w.renderer.FillRect(&sdl.Rect{X: x * 8, Y: y * 8, W: 8, H: 8}); err != nil {
					return err
				}
			}
		}
		if y%2 == 1 {
			for x := int32(0); x < int32(w.W/8); x += 2 {
				if err := w.renderer.FillRect(&sdl.Rect{X: x * 8, Y: y * 8, W: 8, H: 8}); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (w *window) DrawDivider(c color.RGBA, count int32, horizontal bool) error {
	if w.WindowFlags&sdl.WINDOW_SHOWN == 0 {
		return nil
	}

	if !w.grid {
		return nil
	}

	if err := w.SetDrawColor(c); err != nil {
		return err
	}

	width := int32(w.W)
	height := int32(w.H)
	if horizontal {
		for i := int32(1); i < int32(count); i++ {
			if err := w.renderer.DrawLine(width/count*i, 0, width/count*i, height); err != nil {
				return err
			}
		}
	} else {
		for i := int32(1); i < int32(count); i++ {
			if err := w.renderer.DrawLine(0, height/count*i, width, height/count*i); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *window) Update(pix []byte) error {
	if w.WindowFlags&sdl.WINDOW_SHOWN == 0 {
		return nil
	}

	if err := w.init(); err != nil {
		return err
	}

	buf, _, err := w.tex.Lock(nil)
	if err != nil {
		return err
	}
	copy(buf, pix)
	w.tex.Unlock()

	return w.renderer.Copy(w.tex, nil, nil)
}

func (w *window) Present() error {
	if w.WindowFlags&sdl.WINDOW_SHOWN == 0 {
		return nil
	}

	if err := w.init(); err != nil {
		return err
	}

	w.renderer.Present()
	return nil
}

func (w *window) Destroy() {
	if w.tex != nil {
		w.tex.Destroy()
		w.tex = nil
	}
	if w.renderer != nil {
		w.renderer.Destroy()
		w.renderer = nil
	}
	if w.window != nil {
		w.window.Destroy()
		w.window = nil
		w.ID = 0
		w.fullscreen = false
	}

	w.once = sync.Once{}
}
