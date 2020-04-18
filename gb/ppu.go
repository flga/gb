package gb

type ppu struct{}

func (p *ppu) clock(b *bus) {
	//todo: frames
}

func (p *ppu) read(addr uint16) uint8 {
	return 0 // todo
}

func (p *ppu) write(addr uint16, v uint8) {
	// todo
}
