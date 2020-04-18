package gb

type bus struct {
	cpu       *cpu
	ppu       *ppu
	apu       *apu
	cartridge *cartridge
}

func (b *bus) clock() {
	// todo: frame & apu samples
	// todo: frequencies

	b.cpu.clock(b)
	b.apu.clock(b)
	b.ppu.clock(b)
}

func (b *bus) read(addr uint16, v uint8) uint8 {
	addr = b.cartridge.translate(addr)

	// todo: memory map
	return 0
}

func (b *bus) write(addr uint16, v uint8) {
	addr = b.cartridge.translate(addr)

	// todo: memory map
}
