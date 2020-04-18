package gb

import (
	"io"
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
	bus *bus
}

func New() *GameBoy {
	return &GameBoy{
		bus: newBus(),
	}
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
