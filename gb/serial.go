package gb

type serial struct{}

func (s *serial) clock(gb *GameBoy) {}

func (s *serial) write(addr uint16, v uint8) {
	switch addr {
	case ioRegs.SB:
		// TODO
	case ioRegs.SC:
		// TODO
	default:
		unmappedWrite("serial", addr, v)
	}
}

func (s *serial) read(addr uint16) uint8 {
	switch addr {
	case ioRegs.SB:
		return 0 // TODO
	case ioRegs.SC:
		return 0 // TODO
	default:
		unmappedRead("serial", addr)
		return 0
	}
}
