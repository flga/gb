package gb

import "fmt"

type memory []byte

func (r memory) read(addr uint16) uint8 {
	return r[int(addr)%cap(r)]
}
func (r memory) write(addr uint16, v uint8) {
	r[int(addr)%cap(r)] = v
}

type bus struct {
	cpu       *cpu
	ppu       *ppu
	apu       *apu
	ram       memory
	wram      memory
	hram      memory
	cartridge *cartridge
}

func newBus() *bus {
	ret := &bus{
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
	// io registers init
	// [$FF05] = $00   ; TIMA
	// [$FF06] = $00   ; TMA
	// [$FF07] = $00   ; TAC
	// [$FF10] = $80   ; NR10
	// [$FF11] = $BF   ; NR11
	// [$FF12] = $F3   ; NR12
	// [$FF14] = $BF   ; NR14
	// [$FF16] = $3F   ; NR21
	// [$FF17] = $00   ; NR22
	// [$FF19] = $BF   ; NR24
	// [$FF1A] = $7F   ; NR30
	// [$FF1B] = $FF   ; NR31
	// [$FF1C] = $9F   ; NR32
	// [$FF1E] = $BF   ; NR33
	// [$FF20] = $FF   ; NR41
	// [$FF21] = $00   ; NR42
	// [$FF22] = $00   ; NR43
	// [$FF23] = $BF   ; NR44
	// [$FF24] = $77   ; NR50
	// [$FF25] = $F3   ; NR51
	// [$FF26] = $F1-GB, $F0-SGB ; NR52
	// [$FF40] = $91   ; LCDC
	// [$FF42] = $00   ; SCY
	// [$FF43] = $00   ; SCX
	// [$FF45] = $00   ; LYC
	// [$FF47] = $FC   ; BGP
	// [$FF48] = $FF   ; OBP0
	// [$FF49] = $FF   ; OBP1
	// [$FF4A] = $00   ; WY
	// [$FF4B] = $00   ; WX
	// [$FFFF] = $00   ; IE

	return ret
}

func (b *bus) clock() {
	// todo: frame & apu samples
	// todo: frequencies

	b.cpu.clock(b)
	b.apu.clock(b)
	b.ppu.clock(b)
}

func (b *bus) read(addr uint16) uint8 {
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
		// todo: registers
		return 0
	}
	if addr < 0xFFFF {
		return b.hram.read(addr)
	}

	if addr == 0xFFFF {
		return b.cpu.read(addr)
	}

	panic(fmt.Sprintf("unmapped read at 0%X", addr))
}

func (b *bus) write(addr uint16, v uint8) {
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
		// todo: registers
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
