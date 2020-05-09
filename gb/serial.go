package gb

type serial struct {
	SB, SC uint8
}

func (s *serial) clock(gb *GameBoy) {}

func (s *serial) write(addr uint16, v uint8) {
	switch addr {
	case ioRegs.SB:
		s.SB = v
	case ioRegs.SC:
		s.SC = v
	default:
		unmappedWrite("serial", addr, v)
	}
}

func (s *serial) read(addr uint16) uint8 {
	switch addr {
	case ioRegs.SB:
		return s.SB
	case ioRegs.SC:
		return s.SC
	default:
		unmappedRead("serial", addr)
		return 0
	}
}
