package gb

import "io"

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

type GameBoy struct{}

func New() *GameBoy {
	return &GameBoy{}
}

func (gb *GameBoy) InsertCartridge(c io.Reader) error { return nil }
func (gb *GameBoy) Reset()                            {}
func (gb *GameBoy) PowerOn()                          {}
func (gb *GameBoy) PowerOff()                         {}
func (gb *GameBoy) SetVolume(vol float64)             {}
func (gb *GameBoy) Press(btns Button)                 {}
