package gb

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

type joypad uint8

func (j *joypad) write(addr uint16, v uint8) {
	switch addr {
	case ioRegs.P1:
	// TODO
	default:
		unmappedWrite("joypad", addr, v)
	}
}

func (j *joypad) read(addr uint16) uint8 {
	switch addr {
	case ioRegs.P1:
		// TODO
		return 0
	default:
		unmappedRead("joypad", addr)
		return 0
	}
}
