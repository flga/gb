package gb

type timerControl uint8

const (
	timerEnable      = 1 << 2
	timerClockSelect = 0x3
)

func (t timerControl) enabled() bool {
	return t&timerEnable > 0
}

func (t timerControl) freq() (cpu int, machine int) {
	switch t & timerClockSelect {
	case 0:
		return cpuFreq / 1024, machineFreq / 1024
	case 1:
		return cpuFreq / 16, machineFreq / 16
	case 2:
		return cpuFreq / 64, machineFreq / 64
	case 3:
		return cpuFreq / 256, machineFreq / 256
	default:
		return 0, 0
	}
}

type timer struct {
	TIMA uint8
	TMA  uint8
	TAC  uint8
}

func (t *timer) clock(gb *GameBoy) {}

func (t *timer) write(addr uint16, v uint8) {
	switch addr {
	case ioRegs.TIMA:
		t.TIMA = v
	case ioRegs.TMA:
		t.TMA = v
	case ioRegs.TAC:
		t.TAC = v
	default:
		unmappedWrite("timer", addr, v)
	}
}

func (t *timer) read(addr uint16) uint8 {
	switch addr {
	case ioRegs.TIMA:
		return t.TIMA
	case ioRegs.TMA:
		return t.TMA
	case ioRegs.TAC:
		return 0xF8 | t.TAC
	default:
		unmappedRead("timer", addr)
		return 0
	}
}

type divider uint8

func (d *divider) clock(gb *GameBoy) {}

func (d *divider) write(addr uint16, v uint8) {
	switch addr {
	case ioRegs.DIV:
		*d = 0
	default:
		unmappedWrite("divider", addr, v)
	}
}

func (d *divider) read(addr uint16) uint8 {
	switch addr {
	case ioRegs.DIV:
		return uint8(*d)
	default:
		unmappedRead("divider", addr)
		return 0
	}
}
