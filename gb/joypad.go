package gb

type Button uint8

const (
	A Button = 1 << iota
	B
	Select
	Start
	Right
	Left
	Up
	Down
)

const (
	joypadSelectButtons   uint8 = 1 << 5
	joypadSelectDirection uint8 = 1 << 4
)

type joypad struct {
	state Button
	p1    uint8
	raise bool
}

func (j *joypad) press(gb *GameBoy, buttons Button, pressed bool) {
	if pressed {
		if j.state&buttons == 0 {
			if buttons&0x0F > 0 && j.p1&joypadSelectButtons == 0 {
				gb.interruptCtrl.raise(joypadInterrupt)
			} else if buttons&0xF0 > 0 && j.p1&joypadSelectDirection == 0 {
				gb.interruptCtrl.raise(joypadInterrupt)
			}
		}
		j.state |= buttons
	} else {
		j.state &^= buttons
	}
}

func (j *joypad) write(addr uint16, v uint8) {
	switch addr {
	case ioRegs.P1:
		j.p1 = v & 0xF0
	default:
		unmappedWrite("joypad", addr, v)
	}
}

func (j *joypad) read(addr uint16) uint8 {
	switch addr {
	case ioRegs.P1:
		var v uint8
		if j.p1&joypadSelectButtons == 0 {
			v = uint8(j.state) & 0x0F
		} else if j.p1&joypadSelectDirection == 0 {
			v = uint8(j.state) >> 4
		}

		return 0xC0 | uint8(j.p1&0xF0) | ^v&0x0F
	default:
		unmappedRead("joypad", addr)
		return 0
	}
}
