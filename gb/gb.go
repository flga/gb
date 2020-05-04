package gb

import (
	"bytes"
	"io"
	"io/ioutil"
)

type Button uint8

const (
	A Button = 1 << iota
	B
	Up
	Down
	Left
	Right
	Start
	Select
)

type GameBoy struct {
	cpu       *cpu
	ppu       *ppu
	apu       *apu
	joypad    *joypad
	serial    *serial
	timer     *timer
	cartridge *cartridge
	bus       *mmu
}

func New(r io.Reader, disasm bool) (*GameBoy, error) {

	cart, err := newCartridge(r)
	if err != nil {
		return nil, err
	}

	var cpu cpu
	var ppu ppu
	var apu apu
	var joypad joypad
	var serial serial
	var timer timer

	cpu.init(0x0100)
	cpu.disasm = disasm
	// ppu.init()
	// apu.init()
	// joypad.init()
	// serial.init()
	// timer.init()

	bus := mmu{
		cpu:       &cpu,
		ppu:       &ppu,
		apu:       &apu,
		joypad:    &joypad,
		serial:    &serial,
		timer:     &timer,
		wram:      make(memory, 8*KiB),
		hram:      make(memory, 127),
		cartridge: cart,
	}

	bus.init()

	return &GameBoy{
		cpu:       &cpu,
		ppu:       &ppu,
		apu:       &apu,
		joypad:    &joypad,
		serial:    &serial,
		timer:     &timer,
		cartridge: cart,
		bus:       &bus,
	}, nil
}

func (gb *GameBoy) Clock() {
	gb.cpu.clock(gb.bus)
}

func (gb *GameBoy) ClockFrame() []uint8 {
	start := gb.bus.cycles
	for gb.bus.cycles < start+17556 {
		gb.bus.clock()
	}
	return gb.ppu.frame[:]
}

func (gb *GameBoy) InsertCartridge(r io.Reader) error {
	c, err := newCartridge(r)
	if err != nil {
		return err
	}

	gb.bus.cartridge = c
	return nil
}

func (gb *GameBoy) CartridgeInfo() CartridgeInfo {
	if gb == nil || gb.bus.cartridge == nil {
		return CartridgeInfo{}
	}

	return gb.bus.cartridge.info
}

func (gb *GameBoy) Reset()                {}
func (gb *GameBoy) PowerOn()              {}
func (gb *GameBoy) PowerOff()             {}
func (gb *GameBoy) SetVolume(vol float64) {}
func (gb *GameBoy) Press(btns Button)     {}

func (gb *GameBoy) DumpWram() error {
	f, err := ioutil.TempFile(".", "wram-*.bin")
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, bytes.NewReader(gb.bus.wram))
	if err != nil {
		return err
	}

	return nil
}
