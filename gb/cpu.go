package gb

const (
	vectorVBlank uint16 = 0x40
	vectorLCDc   uint16 = 0x48
	vectorSerial uint16 = 0x50
	vectorTimer  uint16 = 0x58
	vectorHTL    uint16 = 0x60
)

type cpuFlags uint8

const (
	_ cpuFlags = 1 << iota
	_
	_
	_
	fc // carry
	fh // halfCarry
	fn // negative
	fz // zero
)

func (f cpuFlags) String() string {
	buf := make([]rune, 0, 4)
	if f&fc > 0 {
		buf = append(buf, 'C')
	}
	if f&fh > 0 {
		buf = append(buf, 'H')
	}
	if f&fn > 0 {
		buf = append(buf, 'N')
	}
	if f&fz > 0 {
		buf = append(buf, 'Z')
	}
	return string(buf)
}

func (f *cpuFlags) set(flag cpuFlags, v bool) {
	*f &^= flag

	if v {
		*f |= flag
	}
}

type cpuStatus uint8

// just guessing for now
const (
	fetch cpuStatus = 1 << iota
	execute
	read
	write
	interrupt
	halt
)

type cpu struct {
	A    uint8
	F    cpuFlags
	B, C uint8
	D, E uint8
	H, L uint8
	SP   uint16
	PC   uint16

	// address uint16
	// data    uint8

	// status  cpuStatus
	// opstack []op
}

func (c *cpu) clock(b *bus) {
	// if len(c.opstack) == 0 {
	// 	return
	// }

	// head := len(c.opstack) - 1
	// c.opstack[head](b)
	// c.opstack = c.opstack[:head]
}

func (c *cpu) read(addr uint16) uint8 {
	return 0 // todo
}

func (c *cpu) write(addr uint16, v uint8) {
	// todo
}

func (c *cpu) ld(target *uint8, v uint8) {
	*target = v
}
func (c *cpu) executeInst(b *bus) {
	op := b.read(c.PC)
	c.PC++

	switch op {
	// ld reg/imm
	case 0x06: // LD B,n 06 8
		v := b.read(c.PC)
		c.PC++
		c.ld(&c.B, v)
	case 0x0E: // LD C,n 0E 8
		v := b.read(c.PC)
		c.PC++
		c.ld(&c.C, v)
	case 0x16: // LD D,n 16 8
		v := b.read(c.PC)
		c.PC++
		c.ld(&c.D, v)
	case 0x1E: // LD E,n 1E 8
		v := b.read(c.PC)
		c.PC++
		c.ld(&c.E, v)
	case 0x26: // LD H,n 26 8
		v := b.read(c.PC)
		c.PC++
		c.ld(&c.H, v)
	case 0x2E: // LD L,n 2E 8
		v := b.read(c.PC)
		c.PC++
		c.ld(&c.L, v)

	// ld reg/reg
	case 0x7F: // LD A,A 7F 4
		c.ld(&c.A, c.A)
	case 0x78: // LD A,B 78 4
		c.ld(&c.A, c.B)
	case 0x79: // LD A,C 79 4
		c.ld(&c.A, c.C)
	case 0x7A: // LD A,D 7A 4
		c.ld(&c.A, c.D)
	case 0x7B: // LD A,E 7B 4
		c.ld(&c.A, c.E)
	case 0x7C: // LD A,H 7C 4
		c.ld(&c.A, c.H)
	case 0x7D: // LD A,L 7D 4
		c.ld(&c.A, c.L)
	case 0x7E: // LD A,(HL) 7E 8
		addr := uint16(c.H)<<8 | uint16(c.L)
		c.ld(&c.A, b.read(addr))
	case 0x40: // LD B,B 40 4
		c.ld(&c.B, c.B)
	case 0x41: // LD B,C 41 4
		c.ld(&c.B, c.C)
	case 0x42: // LD B,D 42 4
		c.ld(&c.B, c.D)
	case 0x43: // LD B,E 43 4
		c.ld(&c.B, c.E)
	case 0x44: // LD B,H 44 4
		c.ld(&c.B, c.H)
	case 0x45: // LD B,L 45 4
		c.ld(&c.B, c.L)
	case 0x46: // LD B,(HL) 46 8
		addr := uint16(c.H)<<8 | uint16(c.L)
		c.ld(&c.B, b.read(addr))
	case 0x48: // LD C,B 48 4
		c.ld(&c.C, c.B)
	case 0x49: // LD C,C 49 4
		c.ld(&c.C, c.C)
	case 0x4A: // LD C,D 4A 4
		c.ld(&c.C, c.D)
	case 0x4B: // LD C,E 4B 4
		c.ld(&c.C, c.E)
	case 0x4C: // LD C,H 4C 4
		c.ld(&c.C, c.H)
	case 0x4D: // LD C,L 4D 4
		c.ld(&c.C, c.L)
	case 0x4E: // LD C,(HL) 4E 8
		addr := uint16(c.H)<<8 | uint16(c.L)
		c.ld(&c.C, b.read(addr))
	case 0x50: // LD D,B 50 4
		c.ld(&c.D, c.B)
	case 0x51: // LD D,C 51 4
		c.ld(&c.D, c.C)
	case 0x52: // LD D,D 52 4
		c.ld(&c.D, c.D)
	case 0x53: // LD D,E 53 4
		c.ld(&c.D, c.E)
	case 0x54: // LD D,H 54 4
		c.ld(&c.D, c.H)
	case 0x55: // LD D,L 55 4
		c.ld(&c.D, c.L)
	case 0x56: // LD D,(HL) 56 8
		addr := uint16(c.H)<<8 | uint16(c.L)
		c.ld(&c.D, b.read(addr))
	case 0x58: // LD E,B 58 4
		c.ld(&c.E, c.B)
	case 0x59: // LD E,C 59 4
		c.ld(&c.E, c.C)
	case 0x5A: // LD E,D 5A 4
		c.ld(&c.E, c.D)
	case 0x5B: // LD E,E 5B 4
		c.ld(&c.E, c.E)
	case 0x5C: // LD E,H 5C 4
		c.ld(&c.E, c.H)
	case 0x5D: // LD E,L 5D 4
		c.ld(&c.E, c.L)
	case 0x5E: // LD E,(HL) 5E 8
		addr := uint16(c.H)<<8 | uint16(c.L)
		c.ld(&c.E, b.read(addr))
	case 0x60: // LD H,B 60 4
		c.ld(&c.H, c.B)
	case 0x61: // LD H,C 61 4
		c.ld(&c.H, c.C)
	case 0x62: // LD H,D 62 4
		c.ld(&c.H, c.D)
	case 0x63: // LD H,E 63 4
		c.ld(&c.H, c.E)
	case 0x64: // LD H,H 64 4
		c.ld(&c.H, c.H)
	case 0x65: // LD H,L 65 4
		c.ld(&c.H, c.L)
	case 0x66: // LD H,(HL) 66 8
		addr := uint16(c.H)<<8 | uint16(c.L)
		c.ld(&c.H, b.read(addr))
	case 0x68: // LD L,B 68 4
		c.ld(&c.L, c.B)
	case 0x69: // LD L,C 69 4
		c.ld(&c.L, c.C)
	case 0x6A: // LD L,D 6A 4
		c.ld(&c.L, c.D)
	case 0x6B: // LD L,E 6B 4
		c.ld(&c.L, c.E)
	case 0x6C: // LD L,H 6C 4
		c.ld(&c.L, c.H)
	case 0x6D: // LD L,L 6D 4
		c.ld(&c.L, c.L)
	case 0x6E: // LD L,(HL) 6E 8
		addr := uint16(c.H)<<8 | uint16(c.L)
		c.ld(&c.L, b.read(addr))
	case 0x70: // LD (HL),B 70 8
		addr := uint16(c.H)<<8 | uint16(c.L)
		b.write(addr, c.B)
	case 0x71: // LD (HL),C 71 8
		addr := uint16(c.H)<<8 | uint16(c.L)
		b.write(addr, c.C)
	case 0x72: // LD (HL),D 72 8
		addr := uint16(c.H)<<8 | uint16(c.L)
		b.write(addr, c.D)
	case 0x73: // LD (HL),E 73 8
		addr := uint16(c.H)<<8 | uint16(c.L)
		b.write(addr, c.E)
	case 0x74: // LD (HL),H 74 8
		addr := uint16(c.H)<<8 | uint16(c.L)
		b.write(addr, c.H)
	case 0x75: // LD (HL),L 75 8
		addr := uint16(c.H)<<8 | uint16(c.L)
		b.write(addr, c.L)
	case 0x36: // LD (HL),n 36 12
		addr := uint16(c.H)<<8 | uint16(c.L)
		v := b.read(c.PC)
		c.PC++
		b.write(addr, v)

	// lda
	case 0x0A: // LD A,(BC) 0A 8
		addr := uint16(c.B)<<8 | uint16(c.C)
		c.ld(&c.A, b.read(addr))
	case 0x1A: // LD A,(DE) 1A 8
		addr := uint16(c.D)<<8 | uint16(c.E)
		c.ld(&c.A, b.read(addr))
	case 0xFA: // LD A,(nn) FA 16
		lo := uint16(b.read(c.PC))
		c.PC++
		hi := uint16(b.read(c.PC))
		c.PC++
		addr := hi<<8 | lo
		c.ld(&c.A, b.read(addr))
	case 0x3E: // LD A,# 3E 8
		v := b.read(c.PC)
		c.PC++
		c.ld(&c.A, v)

	// sta
	case 0x47: // LD B,A 47 4
		c.ld(&c.B, c.A)
	case 0x4F: // LD C,A 4F 4
		c.ld(&c.C, c.A)
	case 0x57: // LD D,A 57 4
		c.ld(&c.D, c.A)
	case 0x5F: // LD E,A 5F 4
		c.ld(&c.E, c.A)
	case 0x67: // LD H,A 67 4
		c.ld(&c.H, c.A)
	case 0x6F: // LD L,A 6F 4
		c.ld(&c.L, c.A)
	case 0x02: // LD (BC),A 02 8
		lo := uint16(c.C)
		hi := uint16(c.B)
		addr := hi<<8 | lo
		b.write(addr, c.A)
	case 0x12: // LD (DE),A 12 8
		lo := uint16(c.E)
		hi := uint16(c.D)
		addr := hi<<8 | lo
		b.write(addr, c.A)
	case 0x77: // LD (HL),A 77 8
		lo := uint16(c.L)
		hi := uint16(c.H)
		addr := hi<<8 | lo
		b.write(addr, c.A)
	case 0xEA: // LD (nn),A EA 16
		lo := uint16(b.read(c.PC))
		c.PC++
		hi := uint16(b.read(c.PC))
		c.PC++
		addr := hi<<8 | lo
		b.write(addr, c.A)

	// c loads
	case 0xF2: // LD A,(C) F2 8
		addr := 0xFF00 + uint16(c.C)
		c.ld(&c.A, b.read(addr))
	case 0xE2: // LD ($FF00+C),A E2 8
		addr := 0xFF00 + uint16(c.C)
		b.write(addr, c.A)

	// inc/dec loads
	case 0x3A: // LDD A,(HL) 3A 8
		lo := uint16(c.L)
		hi := uint16(c.H)
		addr := hi<<8 | lo
		c.ld(&c.A, b.read(addr))
		addr--
		c.L = uint8(addr & 0xFF)
		c.H = uint8(addr >> 8 & 0xFF)
	case 0x32: //LDD (HL),A 32 8
		lo := uint16(c.L)
		hi := uint16(c.H)
		addr := hi<<8 | lo
		b.write(addr, c.A)
		addr--
		c.L = uint8(addr & 0xFF)
		c.H = uint8(addr >> 8 & 0xFF)
	case 0x2A: // LDI A,(HL) 2A 8
		lo := uint16(c.L)
		hi := uint16(c.H)
		addr := hi<<8 | lo
		c.ld(&c.A, b.read(addr))
		addr++
		c.L = uint8(addr & 0xFF)
		c.H = uint8(addr >> 8 & 0xFF)
	case 0x22: // LDI (HL),A 22 8
		lo := uint16(c.L)
		hi := uint16(c.H)
		addr := hi<<8 | lo
		b.write(addr, c.A)
		addr++
		c.L = uint8(addr & 0xFF)
		c.H = uint8(addr >> 8 & 0xFF)

	// indexed loads (I think this might be zero page, need to double check)
	case 0xE0: // LD ($FF00+n),A E0 12
		n := uint16(b.read(c.PC))
		c.PC++
		addr := uint16(b.read(0xFF00 + n))
		b.write(addr, c.A)
	case 0xF0: // LD A,($FF00+n) F0 12
		n := uint16(b.read(c.PC))
		c.PC++
		addr := uint16(b.read(0xFF00 + n))
		c.ld(&c.A, b.read(addr))

	//16 bit loads
	case 0x01: // LD BC,nn 01 12
		c.C = b.read(c.PC)
		c.PC++
		c.B = b.read(c.PC)
		c.PC++
	case 0x11: // LD DE,nn 11 12
		c.E = b.read(c.PC)
		c.PC++
		c.D = b.read(c.PC)
		c.PC++
	case 0x21: // LD HL,nn 21 12
		c.L = b.read(c.PC)
		c.PC++
		c.H = b.read(c.PC)
		c.PC++
	case 0x31: // LD SP,nn 31 12
		lo := uint16(b.read(c.PC))
		c.PC++
		hi := uint16(b.read(c.PC))
		c.PC++
		c.SP = hi<<8 | lo
	case 0xF9: // LD SP,HL F9 8
		lo := uint16(c.L)
		hi := uint16(c.H)
		c.SP = hi<<8 | lo
		b.read(c.SP) // TODO: assumption
	case 0xF8: //LDHL SP,n F8 12
		n := b.read(c.PC)
		c.PC++

		lo := uint8(c.SP & 0xFF)
		hi := uint8(c.SP >> 8 & 0xFF)

		// clear carries
		c.F.set(fh, false)
		c.F.set(fc, false)
		// adc
		c.L = c.doAddc(lo, n)
		c.H = c.doAddc(hi, 0)
		// patch zero
		c.F.set(fz, false)

		b.read(c.SP) // TODO: assumption
	case 0x08: // LD (nn),SP 08 20
		lo := b.read(c.PC)
		c.PC++
		hi := b.read(c.PC)
		c.PC++
		addr := uint16(hi)<<8 | uint16(lo)
		b.write(addr, uint8(c.SP&0xFF))
		b.write(addr+1, uint8(c.SP>>8&0xFF))
	case 0xF5: // PUSH AF F5 16
		b.read(c.SP) // TODO: assumption
		c.SP--       // TODO: confirm sp dec order
		b.write(c.SP, c.A)
		c.SP-- // TODO: confirm sp dec order
		b.write(c.SP, uint8(c.F))
	case 0xC5: // PUSH BC C5 16
		b.read(c.SP) // TODO: assumption
		c.SP--       // TODO: confirm sp dec order
		b.write(c.SP, c.B)
		c.SP-- // TODO: confirm sp dec order
		b.write(c.SP, c.C)
	case 0xD5: // PUSH DE D5 16
		b.read(c.SP) // TODO: assumption
		c.SP--       // TODO: confirm sp dec order
		b.write(c.SP, c.D)
		c.SP-- // TODO: confirm sp dec order
		b.write(c.SP, c.E)
	case 0xE5: // PUSH HL E5 16
		b.read(c.SP) // TODO: assumption
		c.SP--       // TODO: confirm sp dec order
		b.write(c.SP, c.H)
		c.SP-- // TODO: confirm sp dec order
		b.write(c.SP, c.L)
	case 0xF1: // POP AF F1 12
		c.F = cpuFlags(b.read(c.SP))
		c.SP++
		c.A = b.read(c.SP)
		c.SP++
	case 0xC1: // POP BC C1 12
		c.C = b.read(c.SP)
		c.SP++
		c.B = b.read(c.SP)
		c.SP++
	case 0xD1: // POP DE D1 12
		c.E = b.read(c.SP)
		c.SP++
		c.D = b.read(c.SP)
		c.SP++
	case 0xE1: // POP HL E1 12
		c.L = b.read(c.SP)
		c.SP++
		c.H = b.read(c.SP)
		c.SP++
	case 0x87: // ADD A,A 87 4
		c.F.set(fc, false)
		c.A = c.doAddc(c.A, c.A)
	case 0x80: // ADD A,B 80 4
		c.F.set(fc, false)
		c.A = c.doAddc(c.A, c.B)
	case 0x81: // ADD A,C 81 4
		c.F.set(fc, false)
		c.A = c.doAddc(c.A, c.C)
	case 0x82: // ADD A,D 82 4
		c.F.set(fc, false)
		c.A = c.doAddc(c.A, c.D)
	case 0x83: // ADD A,E 83 4
		c.F.set(fc, false)
		c.A = c.doAddc(c.A, c.E)
	case 0x84: // ADD A,H 84 4
		c.F.set(fc, false)
		c.A = c.doAddc(c.A, c.H)
	case 0x85: // ADD A,L 85 4
		c.F.set(fc, false)
		c.A = c.doAddc(c.A, c.L)
	case 0x86: // ADD A,(HL) 86 8
		lo := uint16(c.L)
		hi := uint16(c.H)
		c.F.set(fc, false)
		c.A = c.doAddc(c.A, b.read(hi<<8|lo))
	case 0xC6: // ADD A,# C6 8
		v := b.read(c.PC)
		c.PC++
		c.F.set(fc, false)
		c.A = c.doAddc(c.A, v)
	}
}

func (c *cpu) doAddc(a, b uint8) uint8 {
	var carry uint16
	if c.F&fc > 0 {
		carry = 1
	}
	v := uint16(a) + uint16(b) + carry
	c.F.set(fz, v == 0)
	c.F.set(fn, false)
	c.F.set(fh, a&0xF+b&0xF > 0xF)
	c.F.set(fc, v > 0xFF)
	return uint8(v)
}

// instruction base
func (c *cpu) adc_r_d8(b *bus)    {}
func (c *cpu) adc_r_irr(b *bus)   {}
func (c *cpu) adc_r_r(b *bus)     {}
func (c *cpu) add_r_d8(b *bus)    {}
func (c *cpu) add_r_irr(b *bus)   {}
func (c *cpu) add_r_r(b *bus)     {}
func (c *cpu) add_rr_rr(b *bus)   {}
func (c *cpu) add_rr_sp(b *bus)   {}
func (c *cpu) add_sp_r8(b *bus)   {}
func (c *cpu) and_d8(b *bus)      {}
func (c *cpu) and_irr(b *bus)     {}
func (c *cpu) and_r(b *bus)       {}
func (c *cpu) call_a16(b *bus)    {}
func (c *cpu) call_NC_a16(b *bus) {}
func (c *cpu) call_NZ_a16(b *bus) {}
func (c *cpu) call_r_a16(b *bus)  {}
func (c *cpu) call_Z_a16(b *bus)  {}
func (c *cpu) ccf(b *bus)         {}
func (c *cpu) cp_d8(b *bus)       {}
func (c *cpu) cp_irr(b *bus)      {}
func (c *cpu) cpl(b *bus)         {}
func (c *cpu) cp_r(b *bus)        {}
func (c *cpu) daa(b *bus)         {}
func (c *cpu) dec_irr(b *bus)     {}
func (c *cpu) dec_r(b *bus)       {}
func (c *cpu) dec_rr(b *bus)      {}
func (c *cpu) dec_sp(b *bus)      {}
func (c *cpu) di(b *bus)          {}
func (c *cpu) ei(b *bus)          {}
func (c *cpu) halt(b *bus)        {}
func (c *cpu) inc_irr(b *bus)     {}
func (c *cpu) inc_r(b *bus)       {}
func (c *cpu) inc_rr(b *bus)      {}
func (c *cpu) inc_sp(b *bus)      {}
func (c *cpu) jp_a16(b *bus)      {}
func (c *cpu) jp_irr(b *bus)      {}
func (c *cpu) jp_NC_a16(b *bus)   {}
func (c *cpu) jp_NZ_a16(b *bus)   {}
func (c *cpu) jp_r_a16(b *bus)    {}
func (c *cpu) jp_Z_a16(b *bus)    {}
func (c *cpu) jr_NC_r8(b *bus)    {}
func (c *cpu) jr_NZ_r8(b *bus)    {}
func (c *cpu) jr_r8(b *bus)       {}
func (c *cpu) jr_r_r8(b *bus)     {}
func (c *cpu) jr_Z_r8(b *bus)     {}
func (c *cpu) ldh_ia8_r(b *bus)   {}
func (c *cpu) ldh_r_ia8(b *bus)   {}
func (c *cpu) ld_ia16_r(b *bus)   {}
func (c *cpu) ld_ia16_sp(b *bus)  {}
func (c *cpu) ld_ir_r(b *bus)     {}
func (c *cpu) ld_irr_d8(b *bus)   {}
func (c *cpu) ld_irr_r(b *bus)    {}
func (c *cpu) ld__r(b *bus)       {}
func (c *cpu) ld_r_(b *bus)       {}
func (c *cpu) ld_r_d8(b *bus)     {}
func (c *cpu) ld_r_ia16(b *bus)   {}
func (c *cpu) ld_r_ir(b *bus)     {}
func (c *cpu) ld_r_irr(b *bus)    {}
func (c *cpu) ld_r_r(b *bus)      {}
func (c *cpu) ld_rr_d16(b *bus)   {}
func (c *cpu) ld_rr_SP_r8(b *bus) {}
func (c *cpu) ld_sp_d16(b *bus)   {}
func (c *cpu) ld_sp_rr(b *bus)    {}
func (c *cpu) nop(b *bus)         {}
func (c *cpu) or_d8(b *bus)       {}
func (c *cpu) or_irr(b *bus)      {}
func (c *cpu) or_r(b *bus)        {}
func (c *cpu) pop_rr(b *bus)      {}
func (c *cpu) prefix_(b *bus)     {}
func (c *cpu) push_rr(b *bus)     {}
func (c *cpu) ret(b *bus)         {}
func (c *cpu) reti(b *bus)        {}
func (c *cpu) ret_NC(b *bus)      {}
func (c *cpu) ret_NZ(b *bus)      {}
func (c *cpu) ret_r(b *bus)       {}
func (c *cpu) ret_Z(b *bus)       {}
func (c *cpu) rla(b *bus)         {}
func (c *cpu) rlca(b *bus)        {}
func (c *cpu) rra(b *bus)         {}
func (c *cpu) rrca(b *bus)        {}
func (c *cpu) rst_(b *bus)        {}
func (c *cpu) sbc_r_d8(b *bus)    {}
func (c *cpu) sbc_r_irr(b *bus)   {}
func (c *cpu) sbc_r_r(b *bus)     {}
func (c *cpu) scf(b *bus)         {}
func (c *cpu) stop_(b *bus)       {}
func (c *cpu) sub_d8(b *bus)      {}
func (c *cpu) sub_irr(b *bus)     {}
func (c *cpu) sub_r(b *bus)       {}
func (c *cpu) xor_d8(b *bus)      {}
func (c *cpu) xor_irr(b *bus)     {}
func (c *cpu) xor_r(b *bus)       {}
