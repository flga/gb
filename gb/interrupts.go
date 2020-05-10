package gb

type interrupt uint8

const (
	vblankInterrupt interrupt = 1 << iota
	lcdStatInterrupt
	timerInterrupt
	serialInterrupt
	joypadInterrupt

	anyInterrupt = 0xff
)

func (i interrupt) String() string {
	buf := []byte{'-', '-', '-', '-', '-', '-', '-', '-'}

	if i&vblankInterrupt > 0 {
		buf[0] = 'V'
	}
	if i&lcdStatInterrupt > 0 {
		buf[1] = 'L'
	}
	if i&timerInterrupt > 0 {
		buf[2] = 'T'
	}
	if i&serialInterrupt > 0 {
		buf[3] = 'S'
	}
	if i&joypadInterrupt > 0 {
		buf[4] = 'J'
	}

	return string(buf)
}

type interruptCtrl struct {
	IME    bool
	IF, IE interrupt
}

func (ic *interruptCtrl) clock(gb *GameBoy) {}

func (ic *interruptCtrl) write(addr uint16, v uint8) {
	switch addr {
	case ioRegs.IF:
		ic.IF = interrupt(v)
	case ioRegs.IE:
		ic.IE = interrupt(v)
	default:
		unmappedWrite("interrupt controller", addr, v)
	}
}

func (ic *interruptCtrl) read(addr uint16) uint8 {
	switch addr {
	case ioRegs.IF:
		return 0xe0 | uint8(ic.IF)
	case ioRegs.IE:
		return uint8(ic.IE)
	default:
		unmappedRead("interrupt controller", addr)
		return 0
	}
}

func (ic *interruptCtrl) raised(interrupt interrupt) interrupt {
	return ic.IE & ic.IF & interrupt
}

func (ic *interruptCtrl) raise(interrupt interrupt) {
	ic.IF |= interrupt
}

func (ic *interruptCtrl) enabled(interrupt interrupt) interrupt {
	return ic.IE & interrupt
}

func (ic *interruptCtrl) enable(interrupt interrupt) {
	ic.IE |= interrupt
}

func (ic *interruptCtrl) ack(interrupt interrupt) {
	ic.IF &^= interrupt
}
