package gb

type bus struct {
	cpu       *cpu
	ppu       *ppu
	apu       *apu
	cartridge *cartridge
}
