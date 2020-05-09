package gb

type timerControl uint8

const (
	timerEnable      = 1 << 2
	timerClockSelect = 0x3
)

func (t timerControl) enabled() bool {
	return t&timerEnable > 0
}

func (t timerControl) freq() (machineCycles uint64) {
	switch t & timerClockSelect {
	case 0:
		return 1024
	case 1:
		return 16
	case 2:
		return 64
	case 3:
		return 256
	default:
		return 0
	}
}

type timer struct {
	TIMA   uint8
	TMA    uint8
	TAC    timerControl
	clocks uint64 // todo: dont keep count here
	reload bool
}

func (t *timer) clock(gb *GameBoy) {
	if !t.TAC.enabled() {
		return
	}

	if t.reload {
		t.reload = false
		t.TIMA = t.TMA
		gb.interruptCtrl.raise(timerInterrupt)
	}

	if t.clocks%t.TAC.freq() == 0 {
		prev := t.TIMA
		t.TIMA++
		if t.TIMA < prev {
			t.reload = true
		}
	}
	t.clocks++
}

func (t *timer) write(addr uint16, v uint8) {
	switch addr {
	case ioRegs.TIMA:
		t.TIMA = v
	case ioRegs.TMA:
		t.TMA = v
	case ioRegs.TAC:
		t.TAC = timerControl(v & 0x7)
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
		return uint8(0xF8 | t.TAC)
	default:
		unmappedRead("timer", addr)
		return 0
	}
}

type divider uint16

func (d *divider) clock(gb *GameBoy) {
	*d++
}

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
		return uint8(*d >> 8)
	default:
		unmappedRead("divider", addr)
		return 0
	}
}
