package gb

import "fmt"

type joypad struct{}

func (j *joypad) read(addr uint16) uint8     { return 0 }
func (j *joypad) write(addr uint16, v uint8) {}

type serial struct{}

func (s *serial) read(addr uint16) uint8     { return 0 }
func (s *serial) write(addr uint16, v uint8) {}

type timer struct{}

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
	read(addr uint16) uint8
	write(addr uint16, v uint8)
}

type mmu struct {
	cpu       *cpu
	ppu       *ppu
	apu       *apu
	joypad    *joypad
	serial    *serial
	timer     *timer
	ram       memory
	wram      memory
	hram      memory
	cartridge *cartridge
}

func newMMU() *mmu {
	ret := &mmu{
		cpu: &cpu{
			A:  0x01,
			F:  0xB0,
			B:  0x00,
			C:  0x13,
			D:  0x00,
			E:  0xD8,
			H:  0x01,
			L:  0x4D,
			SP: 0xFFFE,
			PC: 0x0100,
		},
		ppu:  &ppu{},
		apu:  &apu{},
		ram:  nil, //todo
		wram: nil, //todo
		hram: nil, //todo
	}

	ret.cpu.init()

	// io registers init
	ret.write(0xFF05, 0x00) // TIMA
	ret.write(0xFF06, 0x00) // TMA
	ret.write(0xFF07, 0x00) // TAC
	ret.write(0xFF10, 0x80) // NR10
	ret.write(0xFF11, 0xBF) // NR11
	ret.write(0xFF12, 0xF3) // NR12
	ret.write(0xFF14, 0xBF) // NR14
	ret.write(0xFF16, 0x3F) // NR21
	ret.write(0xFF17, 0x00) // NR22
	ret.write(0xFF19, 0xBF) // NR24
	ret.write(0xFF1A, 0x7F) // NR30
	ret.write(0xFF1B, 0xFF) // NR31
	ret.write(0xFF1C, 0x9F) // NR32
	ret.write(0xFF1E, 0xBF) // NR33
	ret.write(0xFF20, 0xFF) // NR41
	ret.write(0xFF21, 0x00) // NR42
	ret.write(0xFF22, 0x00) // NR43
	ret.write(0xFF23, 0xBF) // NR44
	ret.write(0xFF24, 0x77) // NR50
	ret.write(0xFF25, 0xF3) // NR51
	ret.write(0xFF26, 0xF1) // NR52
	ret.write(0xFF40, 0x91) // LCDC
	ret.write(0xFF42, 0x00) // SCY
	ret.write(0xFF43, 0x00) // SCX
	ret.write(0xFF45, 0x00) // LYC
	ret.write(0xFF47, 0xFC) // BGP
	ret.write(0xFF48, 0xFF) // OBP0
	ret.write(0xFF49, 0xFF) // OBP1
	ret.write(0xFF4A, 0x00) // WY
	ret.write(0xFF4B, 0x00) // WX
	ret.write(0xFFFF, 0x00) // IE

	return ret
}

func (b *mmu) clock() {
	// todo: frame & apu samples
	// todo: frequencies

	b.cpu.clock(b)
	b.apu.clock(b)
	b.ppu.clock(b)
}

func (b *mmu) read(addr uint16) uint8 {
	addr = b.cartridge.translateRead(addr)

	// Start	End		Description						Notes
	// 0000		3FFF	16KB ROM bank 00				From cartridge, usually a fixed bank
	// 4000		7FFF	16KB ROM Bank 01~NN				From cartridge, switchable bank via MB (if any)
	// 8000		9FFF	8KB Video RAM (VRAM)			Only bank 0 in Non-CGB mode Switchable bank 0/1 in CGB mode
	// A000		BFFF	8KB External RAM				In cartridge, switchable bank if any
	// C000		CFFF	4KB Work RAM (WRAM) bank 0
	// D000		DFFF	4KB Work RAM (WRAM) bank 1~N	Only bank 1 in Non-CGB mode Switchable bank 1~7 in CGB mode
	// E000		FDFF	Mirror of C000~DDFF (ECHO RAM)	Typically not used
	// FE00		FE9F	Sprite attribute table (OAM
	// FEA0		FEFF	Not Usable
	// FF00		FF7F	I/O Registers
	// FF80		FFFE	High RAM (HRAM)
	// FFFF		FFFF	Interrupts Enable Register (IE)

	switch addr {
	case 0xFF0F, 0xFFFF:
		return b.cpu.read(addr)
	}

	if addr < 0x4000 {
		return b.cartridge.read(addr)
	}
	if addr < 0xA000 {
		return b.ppu.read(addr)
	}
	if addr < 0xC000 {
		return b.cartridge.read(addr)
	}
	if addr < 0xFE00 {
		return b.wram.read(addr)
	}
	if addr < 0xFEA0 {
		return b.ppu.read(addr)
	}
	if addr < 0xFF00 {
		return 0
	}
	if addr < 0xFF80 {
		switch addr {
		case 0xFF00: // P1
			return b.joypad.read(addr)
		case 0xFF01, 0xFF02:
			return b.serial.read(addr)
		case 0xFF04, 0xFF05, 0xFF06, 0xFF07:
			return b.timer.read(addr)
		case 0xFF0F: // IF
			return b.cpu.read(addr)
		case 0xFF10, 0xFF11, 0xFF12, 0xFF13, 0xFF14, 0xFF16, 0xFF17, 0xFF18, 0xFF19, 0xFF1A, 0xFF1B, 0xFF1C, 0xFF1D, 0xFF1E, 0xFF20, 0xFF21, 0xFF22, 0xFF23, 0xFF24, 0xFF25, 0xFF26:
			return b.apu.read(addr)
		case 0xFF30, 0xFF31, 0xFF32, 0xFF33, 0xFF34, 0xFF35, 0xFF36, 0xFF37, 0xFF38, 0xFF39, 0xFF3A, 0xFF3B, 0xFF3C, 0xFF3D, 0xFF3E, 0xFF3F:
			return b.ppu.read(addr)
		case 0xFF40, 0xFF41:
			return b.ppu.read(addr)
		case 0xFF42, 0xFF43:
			return b.ppu.read(addr)
		case 0xFF44, 0xFF45:
			return b.ppu.read(addr)
		case 0xFF46:
			return b.ppu.read(addr)
		case 0xFF47:
			return b.ppu.read(addr)
		case 0xFF48, 0xFF49:
			return b.ppu.read(addr)
		case 0xFF4A, 0xFF4B:
			return b.ppu.read(addr)
		}

		return 0
	}
	if addr < 0xFFFF {
		return b.hram.read(addr)
	}

	panic(fmt.Sprintf("unmapped read at 0%X", addr))
	return 0
}

func (b *mmu) write(addr uint16, v uint8) {
	addr = b.cartridge.translateWrite(addr)

	// Start	End		Description						Notes
	// 0000		3FFF	16KB ROM bank 00				From cartridge, usually a fixed bank
	// 4000		7FFF	16KB ROM Bank 01~NN				From cartridge, switchable bank via MB (if any)
	// 8000		9FFF	8KB Video RAM (VRAM)			Only bank 0 in Non-CGB mode Switchable bank 0/1 in CGB mode
	// A000		BFFF	8KB External RAM				In cartridge, switchable bank if any
	// C000		CFFF	4KB Work RAM (WRAM) bank 0
	// D000		DFFF	4KB Work RAM (WRAM) bank 1~N	Only bank 1 in Non-CGB mode Switchable bank 1~7 in CGB mode
	// E000		FDFF	Mirror of C000~DDFF (ECHO RAM)	Typically not used
	// FE00		FE9F	Sprite attribute table (OAM
	// FEA0		FEFF	Not Usable
	// FF00		FF7F	I/O Registers
	// FF80		FFFE	High RAM (HRAM)
	// FFFF		FFFF	Interrupts Enable Register (IE)

	switch addr {
	case 0xFF0F, 0xFFFF:
		b.cpu.write(addr, v)
		return
	}

	if addr < 0x4000 {
		b.cartridge.write(addr, v)
		return
	}
	if addr < 0xA000 {
		b.ppu.write(addr, v)
		return
	}
	if addr < 0xC000 {
		b.cartridge.write(addr, v)
		return
	}
	if addr < 0xFE00 {
		b.wram.write(addr, v)
		return
	}
	if addr < 0xFEA0 {
		b.ppu.write(addr, v)
		return
	}
	if addr < 0xFF00 {
		return
	}
	if addr < 0xFF80 {
		switch addr {
		case 0xFF00: // P1
			b.joypad.write(addr, v)
			return
		case 0xFF01, 0xFF02:
			b.serial.write(addr, v)
			return
		case 0xFF04, 0xFF05, 0xFF06, 0xFF07:
			b.timer.write(addr, v)
			return
		case 0xFF0F: // IF
			b.cpu.write(addr, v)
			return
		case 0xFF10, 0xFF11, 0xFF12, 0xFF13, 0xFF14, 0xFF16, 0xFF17, 0xFF18, 0xFF19, 0xFF1A, 0xFF1B, 0xFF1C, 0xFF1D, 0xFF1E, 0xFF20, 0xFF21, 0xFF22, 0xFF23, 0xFF24, 0xFF25, 0xFF26:
			b.apu.write(addr, v)
			return
		case 0xFF30, 0xFF31, 0xFF32, 0xFF33, 0xFF34, 0xFF35, 0xFF36, 0xFF37, 0xFF38, 0xFF39, 0xFF3A, 0xFF3B, 0xFF3C, 0xFF3D, 0xFF3E, 0xFF3F:
			b.ppu.write(addr, v)
			return
		case 0xFF40, 0xFF41:
			b.ppu.write(addr, v)
			return
		case 0xFF42, 0xFF43:
			b.ppu.write(addr, v)
			return
		case 0xFF44, 0xFF45:
			b.ppu.write(addr, v)
			return
		case 0xFF46:
			b.ppu.write(addr, v)
			return
		case 0xFF47:
			b.ppu.write(addr, v)
			return
		case 0xFF48, 0xFF49:
			b.ppu.write(addr, v)
			return
		case 0xFF4A, 0xFF4B:
			b.ppu.write(addr, v)
			return
		}
		return
	}
	if addr < 0xFFFF {
		b.hram.write(addr, v)
		return
	}

	if addr == 0xFFFF {
		b.cpu.write(addr, v)
		return
	}

	panic(fmt.Sprintf("unmapped write at 0%X", addr))
}
