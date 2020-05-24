package gb

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type state uint8

const (
	run state = 1 << iota
	stop
	halt
	interruptDispatch
	dma
)

func (s state) String() string {
	out := make([]rune, 0, 26)
	if s&run > 0 {
		out = append(out, []rune("run")...)

	}
	if s&stop > 0 {
		if len(out) > 0 {
			out = append(out, '|')
		}
		out = append(out, []rune("stop")...)

	}
	if s&halt > 0 {
		if len(out) > 0 {
			out = append(out, '|')
		}
		out = append(out, []rune("halt")...)

	}
	if s&interruptDispatch > 0 {
		if len(out) > 0 {
			out = append(out, '|')
		}
		out = append(out, []rune("dispatch")...)

	}
	if s&dma > 0 {
		if len(out) > 0 {
			out = append(out, '|')
		}
		out = append(out, []rune("dma")...)

	}
	return string(out)
}

type hram [127]byte

func (r *hram) read(addr uint16) uint8 {
	return r[int(addr)%cap(r)]
}
func (r *hram) write(addr uint16, v uint8) {
	r[int(addr)%cap(r)] = v
}

type wram [8 * KiB]byte

func (r *wram) read(addr uint16) uint8 {
	return r[int(addr)%cap(r)]
}
func (r *wram) write(addr uint16, v uint8) {
	r[int(addr)%cap(r)] = v
}

type GameBoy struct {
	state state

	cpu           *cpu
	timer         *timer
	interruptCtrl *interruptCtrl
	dmaCtrl       *dmaCtrl
	apu           *apu
	ppu           *ppu
	serial        busDevice
	cartridge     *Cartridge
	joypad        *joypad

	hram hram
	wram wram

	machineCycles uint64
	Debug         bool
}

func (gb *GameBoy) PowerOn() {
	if gb == nil {
		return
	}

	gb.cpu = &cpu{}
	gb.timer = &timer{}
	gb.interruptCtrl = &interruptCtrl{}
	gb.dmaCtrl = &dmaCtrl{}
	gb.apu = &apu{p1: pulse{isPulse1: true}}
	gb.ppu = &ppu{}
	gb.serial = &serial{}
	gb.joypad = &joypad{}
	// gb.cartridge =     cartridge{}

	gb.cpu.init(0x0100)

	// io registers init
	gb.write(ioRegs.SC, 0x7E)
	gb.write(ioRegs.TIMA, 0x0)
	gb.write(ioRegs.TMA, 0x00)
	gb.write(ioRegs.TAC, 0x00)
	gb.write(ioRegs.NR10, 0x80)
	gb.write(ioRegs.NR11, 0xBF)
	gb.write(ioRegs.NR12, 0xF3)
	gb.write(ioRegs.NR14, 0xBF)
	gb.write(ioRegs.NR21, 0x3F)
	gb.write(ioRegs.NR22, 0x00)
	gb.write(ioRegs.NR24, 0xBF)
	gb.write(ioRegs.NR30, 0x7F)
	gb.write(ioRegs.NR31, 0xFF)
	gb.write(ioRegs.NR32, 0x9F)
	gb.write(ioRegs.NR33, 0xBF)
	gb.write(ioRegs.NR41, 0xFF)
	gb.write(ioRegs.NR42, 0x00)
	gb.write(ioRegs.NR43, 0x00)
	gb.write(ioRegs.NR44, 0xBF)
	gb.write(ioRegs.NR50, 0x77)
	gb.write(ioRegs.NR51, 0xF3)
	gb.write(ioRegs.NR52, 0xF1)
	gb.write(ioRegs.LCDC, 0x91)
	gb.write(ioRegs.SCY, 0x00)
	gb.write(ioRegs.SCX, 0x00)
	gb.write(ioRegs.LYC, 0x00)
	gb.write(ioRegs.BGP, 0xFC)
	gb.write(ioRegs.OBP0, 0xFF)
	gb.write(ioRegs.OBP1, 0xFF)
	gb.write(ioRegs.WY, 0x00)
	gb.write(ioRegs.WX, 0x00)
	gb.write(ioRegs.IE, 0x00)

	gb.timer.DIV = 0xABCC
	gb.joypad.p1 = 0xCF

	gb.state = run
}

func (gb *GameBoy) InsertCartridge(cart *Cartridge, savReader io.Reader, savWriter io.WriteCloser) error {
	if gb == nil {
		return nil
	}

	if cart == nil {
		return errors.New("gb: invalid cart")
	}

	if cart.Saveable() && (savReader == nil || savWriter == nil) {
		return errors.New("gb: cart is saveable but reader or writer not provided")
	}

	if gb.cartridge != nil {
		if err := gb.cartridge.save(); err != nil {
			return err
		}
	}

	gb.cartridge = cart
	if !cart.Saveable() {
		return nil
	}

	data, err := ioutil.ReadAll(savReader)
	if err != nil {
		return fmt.Errorf("gb: unable to read sav data: %w", err)
	}
	cart.loadSave(data)

	cart.savWriter = savWriter

	return nil
}

func (gb *GameBoy) Save() error {
	return gb.cartridge.save()
}

func (gb *GameBoy) CartridgeInfo() CartridgeInfo {
	if gb == nil || gb.cartridge == nil {
		return CartridgeInfo{}
	}

	return gb.cartridge.CartridgeInfo
}

func (gb *GameBoy) SetVolume(vol float64) {}

func (gb *GameBoy) Press(btns Button, pressed bool) {
	if gb == nil {
		return
	}

	gb.joypad.press(gb, btns, pressed)
}

func (gb *GameBoy) ExecuteInst() {
	if gb == nil {
		return
	}

	if gb.Debug {
		disassemble(gb, os.Stdout)
	}
	gb.cpu.clock(gb)
}

func (gb *GameBoy) clockCompensate() {
	if gb == nil {
		return
	}

	gb.timer.clock(gb)
	gb.timer.clock(gb)
	gb.timer.clock(gb)
	gb.timer.clock(gb)
	gb.interruptCtrl.clock(gb)
	gb.dmaCtrl.clock(gb)
	gb.apu.clock(gb)
	gb.ppu.clock(gb)
	gb.ppu.clock(gb)
	gb.ppu.clock(gb)
	gb.ppu.clock(gb)
	gb.serial.clock(gb)
	gb.machineCycles++
}

func (gb *GameBoy) ClockFrame() []uint8 {
	if gb == nil {
		return []uint8{}
	}

	start := gb.machineCycles
	for gb.machineCycles < start+17556 {
		gb.ExecuteInst()
	}
	return gb.ppu.frame[:]
}

func (gb *GameBoy) DrawNametables() []uint8 {
	if gb == nil {
		return []uint8{}
	}

	return gb.ppu.nametableFrame()
}
func (gb *GameBoy) DrawVram() []uint8 {
	if gb == nil {
		return []uint8{}
	}

	return gb.ppu.vramFrame()
}

func (gb *GameBoy) ToggleSprites() {
	if gb == nil {
		return
	}

	gb.ppu.hideSprites = !gb.ppu.hideSprites
}
func (gb *GameBoy) ToggleBackground() {
	if gb == nil {
		return
	}

	gb.ppu.hideBackground = !gb.ppu.hideBackground
}
func (gb *GameBoy) ToggleWindow() {
	if gb == nil {
		return
	}

	gb.ppu.hideWindow = !gb.ppu.hideWindow
}

func (gb *GameBoy) DumpWram() error {
	if gb == nil {
		return nil
	}

	f, err := ioutil.TempFile(".", "wram-*.bin")
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, bytes.NewReader(gb.wram[:]))
	if err != nil {
		return err
	}

	return nil
}

func (gb *GameBoy) read(addr uint16) uint8 {
	if gb == nil {
		return 0
	}

	// Start	End		Description						Notes
	// 0x0000	0x3FFF	16KB ROM bank 00				From cartridge, usually a fixed bank
	// 0x4000	0x7FFF	16KB ROM Bank 01~NN				From cartridge, switchable bank via MB (if any)
	// 0x8000	0x9FFF	8KB Video RAM (VRAM)			Only bank 0 in Non-CGB mode Switchable bank 0/1 in CGB mode
	// 0xA000	0xBFFF	8KB External RAM				In cartridge, switchable bank if any
	// 0xC000	0xCFFF	4KB Work RAM (WRAM) bank 0
	// 0xD000	0xDFFF	4KB Work RAM (WRAM) bank 1~N	Only bank 1 in Non-CGB mode Switchable bank 1~7 in CGB mode
	// 0xE000	0xFDFF	Mirror of C000~DDFF (ECHO RAM)	Typically not used
	// 0xFE00	0xFE9F	Sprite attribute table (OAM
	// 0xFEA0	0xFEFF	Not Usable
	// 0xFF00	0xFF7F	I/O Registers
	// 0xFF80	0xFFFE	High RAM (HRAM)
	// 0xFFFF	0xFFFF	Interrupts Enable Register (IE)

	// rom
	if addr >= 0x0000 && addr <= 0x7FFF {
		return gb.cartridge.read(addr)
	}

	// vram
	if addr >= 0x8000 && addr <= 0x9FFF {
		return gb.ppu.read(addr)
	}

	// eram
	if addr >= 0xA000 && addr <= 0xBFFF {
		return gb.cartridge.read(addr)
	}

	// wram
	if addr >= 0xC000 && addr <= 0xFDFF {
		return gb.wram.read(addr)
	}

	// oam
	if addr >= 0xFE00 && addr <= 0xFE9F {
		return gb.ppu.read(addr)
	}

	// unusable
	if addr >= 0xFEA0 && addr <= 0xFEFF {
		return 0
	}

	// P1
	if addr == 0xFF00 {
		return gb.joypad.read(addr)
	}

	// serial
	if addr >= 0xFF01 && addr <= 0xFF02 {
		return gb.serial.read(addr)
	}

	// timer
	if addr >= 0xFF04 && addr <= 0xFF07 {
		return gb.timer.read(addr)
	}

	// IF
	if addr == 0xFF0F {
		return gb.interruptCtrl.read(addr)
	}

	// pulse1
	if addr >= 0xFF10 && addr <= 0xFF14 {
		return gb.apu.read(addr)
	}

	// pulse2
	if addr >= 0xFF16 && addr <= 0xFF19 {
		return gb.apu.read(addr)
	}

	// wave
	if addr >= 0xFF1A && addr <= 0xFF1E {
		return gb.apu.read(addr)
	}

	// wave pattern
	if addr >= 0xFF30 && addr <= 0xFF3F {
		return gb.apu.read(addr)
	}

	// noise
	if addr >= 0xFF20 && addr <= 0xFF23 {
		return gb.apu.read(addr)
	}

	// apu ctrl
	if addr >= 0xFF24 && addr <= 0xFF26 {
		return gb.apu.read(addr)
	}

	// lcdc, lcdstat, scroll y/x, ly, lyc
	if addr >= 0xFF40 && addr <= 0xFF45 {
		return gb.ppu.read(addr)
	}

	// window y/x
	if addr >= 0xFF4A && addr <= 0xFF4B {
		return gb.ppu.read(addr)
	}

	// dma
	if addr == 0xFF46 {
		return gb.ppu.read(addr)
	}

	// palettes
	if addr >= 0xFF47 && addr <= 0xFF49 {
		return gb.ppu.read(addr)
	}

	// hram
	if addr >= 0xFF80 && addr <= 0xFFFE {
		return gb.hram.read(addr)
	}

	// IE
	if addr == 0xFFFF {
		return gb.interruptCtrl.read(addr)
	}

	// panic(fmt.Sprintf("unmapped read at 0%X", addr))
	return 0xFF
}

func (gb *GameBoy) write(addr uint16, v uint8) {
	if gb == nil {
		return
	}

	// Start	End		Description						Notes
	// 0x0000	0x3FFF	16KB ROM bank 00				From cartridge, usually a fixed bank
	// 0x4000	0x7FFF	16KB ROM Bank 01~NN				From cartridge, switchable bank via MB (if any)
	// 0x8000	0x9FFF	8KB Video RAM (VRAM)			Only bank 0 in Non-CGB mode Switchable bank 0/1 in CGB mode
	// 0xA000	0xBFFF	8KB External RAM				In cartridge, switchable bank if any
	// 0xC000	0xCFFF	4KB Work RAM (WRAM) bank 0
	// 0xD000	0xDFFF	4KB Work RAM (WRAM) bank 1~N	Only bank 1 in Non-CGB mode Switchable bank 1~7 in CGB mode
	// 0xE000	0xFDFF	Mirror of C000~DDFF (ECHO RAM)	Typically not used
	// 0xFE00	0xFE9F	Sprite attribute table (OAM
	// 0xFEA0	0xFEFF	Not Usable
	// 0xFF00	0xFF7F	I/O Registers
	// 0xFF80	0xFFFE	High RAM (HRAM)
	// 0xFFFF	0xFFFF	Interrupts Enable Register (IE)

	// rom
	if addr >= 0x0000 && addr <= 0x7FFF {
		gb.cartridge.write(addr, v)
		return
	}

	// vram
	if addr >= 0x8000 && addr <= 0x9FFF {
		gb.ppu.write(addr, v)
		return
	}

	// eram
	if addr >= 0xA000 && addr <= 0xBFFF {
		gb.cartridge.write(addr, v)
		return
	}

	// wram
	if addr >= 0xC000 && addr <= 0xFDFF {
		gb.wram.write(addr, v)
		return
	}

	// oam
	if addr >= 0xFE00 && addr <= 0xFE9F {
		gb.ppu.write(addr, v)
		return
	}

	// unusable
	if addr >= 0xFEA0 && addr <= 0xFEFF {
		return
	}

	// P1
	if addr == 0xFF00 {
		gb.joypad.write(addr, v)
		return
	}

	// serial
	if addr >= 0xFF01 && addr <= 0xFF02 {
		gb.serial.write(addr, v)
		return
	}

	// timer
	if addr >= 0xFF04 && addr <= 0xFF07 {
		gb.timer.write(addr, v)
		return
	}

	// IF
	if addr == 0xFF0F {
		gb.interruptCtrl.write(addr, v)
		return
	}

	// pulse1
	if addr >= 0xFF10 && addr <= 0xFF14 {
		gb.apu.write(addr, v)
		return
	}

	// pulse2
	if addr >= 0xFF16 && addr <= 0xFF19 {
		gb.apu.write(addr, v)
		return
	}

	// wave
	if addr >= 0xFF1A && addr <= 0xFF1E {
		gb.apu.write(addr, v)
		return
	}

	// wave pattern
	if addr >= 0xFF30 && addr <= 0xFF3F {
		gb.apu.write(addr, v)
		return
	}

	// noise
	if addr >= 0xFF20 && addr <= 0xFF23 {
		gb.apu.write(addr, v)
		return
	}

	// apu ctrl
	if addr >= 0xFF24 && addr <= 0xFF26 {
		gb.apu.write(addr, v)
		return
	}

	// lcdc, lcdstat, scroll y/x, ly, lyc
	if addr >= 0xFF40 && addr <= 0xFF45 {
		gb.ppu.write(addr, v)
		return
	}

	// window y/x
	if addr >= 0xFF4A && addr <= 0xFF4B {
		gb.ppu.write(addr, v)
		return
	}

	// dma
	if addr == 0xFF46 {
		gb.dmaCtrl.write(addr, v)
		gb.state |= dma
		return
	}

	// palettes
	if addr >= 0xFF47 && addr <= 0xFF49 {
		gb.ppu.write(addr, v)
		return
	}

	// hram
	if addr >= 0xFF80 && addr <= 0xFFFE {
		gb.hram.write(addr, v)
		return
	}

	// IE
	if addr == 0xFFFF {
		gb.interruptCtrl.write(addr, v)
		return
	}

	// panic(fmt.Sprintf("unmapped write at 0%X", addr))
}

type busDevice interface {
	clock(gb *GameBoy)
	read(addr uint16) uint8
	write(addr uint16, v uint8)
}
