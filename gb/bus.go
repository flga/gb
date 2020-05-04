package gb

import "fmt"

type joypad struct{}

func (j *joypad) read(addr uint16) uint8     { return 0 }
func (j *joypad) write(addr uint16, v uint8) {}

type serial struct{}

func (s *serial) clock(b bus)                {}
func (s *serial) read(addr uint16) uint8     { return 0 }
func (s *serial) write(addr uint16, v uint8) {}

type timer struct{}

func (t *timer) clock(b bus)                {}
func (t *timer) read(addr uint16) uint8     { return 0 }
func (t *timer) write(addr uint16, v uint8) {}

type memory []byte

func (r memory) read(addr uint16) uint8 {
	return r[int(addr)%cap(r)]
}
func (r memory) write(addr uint16, v uint8) {
	r[int(addr)%cap(r)] = v
}

type bus interface {
	peek(addr uint16) uint8
	read(addr uint16) uint8
	poke(addr uint16, v uint8)
	write(addr uint16, v uint8)
}

type mmu struct {
	cpu       *cpu
	ppu       *ppu
	apu       *apu
	joypad    *joypad
	serial    *serial
	timer     *timer
	wram      memory
	hram      memory
	cartridge *cartridge

	cycles uint64
}

func (b *mmu) init() {
	// io registers init
	b.poke(0xFF05, 0x00) // TIMA
	b.poke(0xFF06, 0x00) // TMA
	b.poke(0xFF07, 0x00) // TAC
	b.poke(0xFF10, 0x80) // NR10
	b.poke(0xFF11, 0xBF) // NR11
	b.poke(0xFF12, 0xF3) // NR12
	b.poke(0xFF14, 0xBF) // NR14
	b.poke(0xFF16, 0x3F) // NR21
	b.poke(0xFF17, 0x00) // NR22
	b.poke(0xFF19, 0xBF) // NR24
	b.poke(0xFF1A, 0x7F) // NR30
	b.poke(0xFF1B, 0xFF) // NR31
	b.poke(0xFF1C, 0x9F) // NR32
	b.poke(0xFF1E, 0xBF) // NR33
	b.poke(0xFF20, 0xFF) // NR41
	b.poke(0xFF21, 0x00) // NR42
	b.poke(0xFF22, 0x00) // NR43
	b.poke(0xFF23, 0xBF) // NR44
	b.poke(0xFF24, 0x77) // NR50
	b.poke(0xFF25, 0xF3) // NR51
	b.poke(0xFF26, 0xF1) // NR52
	b.poke(0xFF40, 0x91) // LCDC
	b.poke(0xFF42, 0x00) // SCY
	b.poke(0xFF43, 0x00) // SCX
	b.poke(0xFF45, 0x00) // LYC
	b.poke(0xFF47, 0xFC) // BGP
	b.poke(0xFF48, 0xFF) // OBP0
	b.poke(0xFF49, 0xFF) // OBP1
	b.poke(0xFF4A, 0x00) // WY
	b.poke(0xFF4B, 0x00) // WX
	b.poke(0xFFFF, 0x00) // IE
}

func (b *mmu) clock() {
	// todo: frame & apu samples
	// todo: frequencies

	b.cpu.clock(b)

	b.apu.clock(b)

	b.ppu.clock(b)
	b.ppu.clock(b)
	b.ppu.clock(b)
	b.ppu.clock(b)

	b.cycles++
}

func (b *mmu) peek(addr uint16) uint8 {
	addr = b.cartridge.translateRead(addr)

	// Start	End		Description						Notes
	// 0x0000	0x3FFF	16KB ROM bank 00				From cartridge, usually a fixed bank
	// 0x4000	0x7FFF	16KB ROM Bank 01~NN				From cartridge, switchable bank via MB (if any)
	// 0x8000	0x9FFF	8KB Video RAM (VRAM)			Only bank 0 in Non-CGB mode Switchable bank 0/1 in CGB mode
	// 0xA000	0xBFFF	8KB External RAM				In cartridge, switchable bank if any
	// 0xC000	0xCFFF	4KB Work RAM (WRAM) bank 0
	// 0xD000	0xDFFF	4KB Work RAM (WRAM) bank 1~N	Only bank 1 in Non-CGB mode Switchable bank 1~7 in CGB mode
	// 0xE000	0xFDFF	Mirror of C000~DDFF (ECHO RAM)	Typically not used
	// 0xFE00	0xFE9F	Sprite attribute table (OAM
	// 0xFEA0	0xFEFF	Not Usable
	// 0xFF00	0xFF7F	I/O Registers
	// 0xFF80	0xFFFE	High RAM (HRAM)
	// 0xFFFF	0xFFFF	Interrupts Enable Register (IE)

	// rom
	if addr >= 0x0000 && addr <= 0x7FFF {
		return b.cartridge.read(addr)
	}

	// vram
	if addr >= 0x8000 && addr <= 0x9FFF {
		return b.ppu.read(addr)
	}

	// eram
	if addr >= 0xA000 && addr <= 0xBFFF {
		return b.cartridge.read(addr)
	}

	// wram
	if addr >= 0xC000 && addr <= 0xFDFF {
		return b.wram.read(addr)
	}

	// oam
	if addr >= 0xFE00 && addr <= 0xFE9F {
		return b.ppu.read(addr)
	}

	// unusable
	if addr >= 0xFEA0 && addr <= 0xFEFF {
		return 0
	}

	// P1
	if addr == 0xFF00 {
		return b.joypad.read(addr)
	}

	// serial
	if addr >= 0xFF01 && addr <= 0xFF02 {
		return b.serial.read(addr)
	}

	// timer
	if addr >= 0xFF04 && addr <= 0xFF07 {
		return b.timer.read(addr)
	}

	// IF
	if addr == 0xFF0F {
		return b.cpu.read(addr)
	}

	// pulse1
	if addr >= 0xFF10 && addr <= 0xFF14 {
		return b.apu.read(addr)
	}

	// pulse2
	if addr >= 0xFF16 && addr <= 0xFF19 {
		return b.apu.read(addr)
	}

	// wave
	if addr >= 0xFF1A && addr <= 0xFF1E {
		return b.apu.read(addr)
	}

	// wave pattern
	if addr >= 0xFF30 && addr <= 0xFF3F {
		return b.apu.read(addr)
	}

	// noise
	if addr >= 0xFF20 && addr <= 0xFF23 {
		return b.apu.read(addr)
	}

	// apu ctrl
	if addr >= 0xFF24 && addr <= 0xFF26 {
		return b.apu.read(addr)
	}

	// lcdc, lcdstat, scroll y/x, ly, lyc
	if addr >= 0xFF40 && addr <= 0xFF45 {
		return b.ppu.read(addr)
	}

	// window y/x
	if addr >= 0xFF4A && addr <= 0xFF4B {
		return b.ppu.read(addr)
	}

	// dma
	if addr == 0xFF46 {
		return b.ppu.read(addr)
	}

	// palettes
	if addr >= 0xFF47 && addr <= 0xFF49 {
		return b.ppu.read(addr)
	}

	// hram
	if addr >= 0xFF80 && addr <= 0xFFFE {
		return b.hram.read(addr)
	}

	// IE
	if addr == 0xFFFF {
		return b.cpu.read(addr)
	}

	panic(fmt.Sprintf("unmapped read at 0%X", addr))
}

func (b *mmu) read(addr uint16) uint8 {
	v := b.peek(addr)
	b.apu.clock(b)
	b.ppu.clock(b)
	b.ppu.clock(b)
	b.ppu.clock(b)
	b.ppu.clock(b)
	b.serial.clock(b)
	b.timer.clock(b)

	b.cycles++
	return v
}

func (b *mmu) poke(addr uint16, v uint8) {
	addr = b.cartridge.translateWrite(addr)

	// Start	End		Description						Notes
	// 0x0000	0x3FFF	16KB ROM bank 00				From cartridge, usually a fixed bank
	// 0x4000	0x7FFF	16KB ROM Bank 01~NN				From cartridge, switchable bank via MB (if any)
	// 0x8000	0x9FFF	8KB Video RAM (VRAM)			Only bank 0 in Non-CGB mode Switchable bank 0/1 in CGB mode
	// 0xA000	0xBFFF	8KB External RAM				In cartridge, switchable bank if any
	// 0xC000	0xCFFF	4KB Work RAM (WRAM) bank 0
	// 0xD000	0xDFFF	4KB Work RAM (WRAM) bank 1~N	Only bank 1 in Non-CGB mode Switchable bank 1~7 in CGB mode
	// 0xE000	0xFDFF	Mirror of C000~DDFF (ECHO RAM)	Typically not used
	// 0xFE00	0xFE9F	Sprite attribute table (OAM
	// 0xFEA0	0xFEFF	Not Usable
	// 0xFF00	0xFF7F	I/O Registers
	// 0xFF80	0xFFFE	High RAM (HRAM)
	// 0xFFFF	0xFFFF	Interrupts Enable Register (IE)

	// rom
	if addr >= 0x0000 && addr <= 0x7FFF {
		b.cartridge.write(addr, v)
		return
	}

	// vram
	if addr >= 0x8000 && addr <= 0x9FFF {
		b.ppu.write(addr, v)
		return
	}

	// eram
	if addr >= 0xA000 && addr <= 0xBFFF {
		b.cartridge.write(addr, v)
		return
	}

	// wram
	if addr >= 0xC000 && addr <= 0xFDFF {
		b.wram.write(addr, v)
		return
	}

	// oam
	if addr >= 0xFE00 && addr <= 0xFE9F {
		b.ppu.write(addr, v)
		return
	}

	// unusable
	if addr >= 0xFEA0 && addr <= 0xFEFF {
		return
	}

	// P1
	if addr == 0xFF00 {
		b.joypad.write(addr, v)
		return
	}

	// serial
	if addr >= 0xFF01 && addr <= 0xFF02 {
		b.serial.write(addr, v)
		return
	}

	// timer
	if addr >= 0xFF04 && addr <= 0xFF07 {
		b.timer.write(addr, v)
		return
	}

	// IF
	if addr == 0xFF0F {
		b.cpu.write(addr, v)
		return
	}

	// pulse1
	if addr >= 0xFF10 && addr <= 0xFF14 {
		b.apu.write(addr, v)
		return
	}

	// pulse2
	if addr >= 0xFF16 && addr <= 0xFF19 {
		b.apu.write(addr, v)
		return
	}

	// wave
	if addr >= 0xFF1A && addr <= 0xFF1E {
		b.apu.write(addr, v)
		return
	}

	// wave pattern
	if addr >= 0xFF30 && addr <= 0xFF3F {
		b.apu.write(addr, v)
		return
	}

	// noise
	if addr >= 0xFF20 && addr <= 0xFF23 {
		b.apu.write(addr, v)
		return
	}

	// apu ctrl
	if addr >= 0xFF24 && addr <= 0xFF26 {
		b.apu.write(addr, v)
		return
	}

	// lcdc, lcdstat, scroll y/x, ly, lyc
	if addr >= 0xFF40 && addr <= 0xFF45 {
		b.ppu.write(addr, v)
		return
	}

	// window y/x
	if addr >= 0xFF4A && addr <= 0xFF4B {
		b.ppu.write(addr, v)
		return
	}

	// dma
	if addr == 0xFF46 {
		b.ppu.write(addr, v)
		return
	}

	// palettes
	if addr >= 0xFF47 && addr <= 0xFF49 {
		b.ppu.write(addr, v)
		return
	}

	// hram
	if addr >= 0xFF80 && addr <= 0xFFFE {
		b.hram.write(addr, v)
		return
	}

	// IE
	if addr == 0xFFFF {
		b.cpu.write(addr, v)
		return
	}

	panic(fmt.Sprintf("unmapped write at 0%X", addr))
}

func (b *mmu) write(addr uint16, v uint8) {
	b.poke(addr, v)
	b.apu.clock(b)
	b.ppu.clock(b)
	b.ppu.clock(b)
	b.ppu.clock(b)
	b.ppu.clock(b)
	b.serial.clock(b)
	b.timer.clock(b)
	b.cycles++
}
