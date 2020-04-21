package gb

type op func(opcode uint8, b bus)

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
	if f&fz > 0 {
		buf = append(buf, 'Z')
	} else {
		buf = append(buf, '-')
	}

	if f&fn > 0 {
		buf = append(buf, 'N')
	} else {
		buf = append(buf, '-')
	}

	if f&fh > 0 {
		buf = append(buf, 'H')
	} else {
		buf = append(buf, '-')
	}

	if f&fc > 0 {
		buf = append(buf, 'C')
	} else {
		buf = append(buf, '-')
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

	table [256]op

	// address uint16
	// data    uint8

	// status  cpuStatus
	// opstack []op
}

func (c *cpu) init() {
	c.genTable()
}

func (c *cpu) clock(b bus) {
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

func (c *cpu) executeInst(b bus) {
	op := b.read(c.PC)
	c.PC++

	c.table[op](op, b)
}

func (c *cpu) genTable() {
	c.table = [256]op{
		c.nop, c.ld_rr_d16, c.ld_irr_r, c.inc_rr, c.inc_r, c.dec_r, c.ld_r_d8, c.rlca, c.ld_ia16_sp, c.add_rr_rr, c.ld_r_irr, c.dec_rr, c.inc_r, c.dec_r, c.ld_r_d8, c.rrca, c.stop_,
		c.ld_rr_d16, c.ld_irr_r, c.inc_rr, c.inc_r, c.dec_r, c.ld_r_d8, c.rla, c.jr_r8, c.add_rr_rr, c.ld_r_irr, c.dec_rr, c.inc_r, c.dec_r, c.ld_r_d8, c.rra, c.jr_NZ_r8,
		c.ld_rr_d16, c.ld__r, c.inc_rr, c.inc_r, c.dec_r, c.ld_r_d8, c.daa, c.jr_Z_r8, c.add_rr_rr, c.ld_r_hlid, c.dec_rr, c.inc_r, c.dec_r, c.ld_r_d8, c.cpl, c.jr_NC_r8,
		c.ld_sp_d16, c.ld__r, c.inc_sp, c.inc_irr, c.dec_irr, c.ld_irr_d8, c.scf, c.jr_r_r8, c.add_rr_sp, c.ld_r_hlid, c.dec_sp, c.inc_r, c.dec_r, c.ld_r_d8, c.ccf, c.ld_r_r,
		c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_irr, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_irr, c.ld_r_r, c.ld_r_r,
		c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_irr, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_irr, c.ld_r_r, c.ld_r_r,
		c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_irr, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_irr, c.ld_r_r, c.ld_irr_r,
		c.ld_irr_r, c.ld_irr_r, c.ld_irr_r, c.ld_irr_r, c.ld_irr_r, c.halt, c.ld_irr_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_irr, c.ld_r_r, c.add_r_r,
		c.add_r_r, c.add_r_r, c.add_r_r, c.add_r_r, c.add_r_r, c.add_r_irr, c.add_r_r, c.adc_r_r, c.adc_r_r, c.adc_r_r, c.adc_r_r, c.adc_r_r, c.adc_r_r, c.adc_r_irr, c.adc_r_r, c.sub_r,
		c.sub_r, c.sub_r, c.sub_r, c.sub_r, c.sub_r, c.sub_irr, c.sub_r, c.sbc_r_r, c.sbc_r_r, c.sbc_r_r, c.sbc_r_r, c.sbc_r_r, c.sbc_r_r, c.sbc_r_irr, c.sbc_r_r, c.and_r,
		c.and_r, c.and_r, c.and_r, c.and_r, c.and_r, c.and_irr, c.and_r, c.xor_r, c.xor_r, c.xor_r, c.xor_r, c.xor_r, c.xor_r, c.xor_irr, c.xor_r, c.or_r,
		c.or_r, c.or_r, c.or_r, c.or_r, c.or_r, c.or_irr, c.or_r, c.cp_r, c.cp_r, c.cp_r, c.cp_r, c.cp_r, c.cp_r, c.cp_irr, c.cp_r, c.ret_NZ,
		c.pop_rr, c.jp_NZ_a16, c.jp_a16, c.call_NZ_a16, c.push_rr, c.add_r_d8, c.rst_, c.ret_Z, c.ret, c.jp_Z_a16, c.prefix_, c.call_Z_a16, c.call_a16, c.adc_r_d8, c.rst_, c.ret_NC,
		c.pop_rr, c.jp_NC_a16, c.call_NC_a16, c.push_rr, c.sub_d8, c.rst_, c.ret_r, c.reti, c.jp_r_a16, c.call_r_a16, c.sbc_r_d8, c.rst_, c.ldh_ia8_r, c.pop_rr, c.ld_ir_r, c.push_rr,
		c.and_d8, c.rst_, c.add_sp_r8, c.jp_irr, c.ld_ia16_r, c.xor_d8, c.rst_, c.ldh_r_ia8, c.pop_rr, c.ld_r_ir, c.di, c.push_rr, c.or_d8, c.rst_, c.ld_rr_SP_r8, c.ld_sp_rr,
		c.ld_r_ia16, c.ei, c.cp_d8, c.rst_,
	}
}

func (c *cpu) addc8(a, b uint8) uint8 {
	var carry uint8
	if c.F&fc > 0 {
		carry = 1
	}

	a += b + carry

	c.F.set(fz, a == 0)
	c.F.set(fn, false)
	c.F.set(fh, a&0xF+b&0xF+carry > 0xF)
	c.F.set(fc, uint16(a)+uint16(b)+uint16(carry) > 0xFF)

	return a
}

func (c *cpu) add8(a, b uint8) uint8 {
	a += b

	c.F.set(fz, a == 0)
	c.F.set(fn, false)
	c.F.set(fh, a&0xF+b&0xF > 0xF)
	c.F.set(fc, uint16(a)+uint16(b) > 0xFF)

	return a
}

// 0xCE ADC A,d8        2 8 0 Z 0 H C
func (c *cpu) adc_r_d8(opcode uint8, b bus) {
	v := b.read(c.PC)
	c.PC++

	c.A = c.addc8(c.A, v)
}

// 0x8E ADC A,(HL)      1 8 0 Z 0 H C
func (c *cpu) adc_r_irr(opcode uint8, b bus) {
	lo := uint16(c.L)
	hi := uint16(c.H)
	addr := hi<<8 | lo
	v := b.read(addr)

	c.A = c.addc8(c.A, v)
}

// 0x88 ADC A,B 1 4 0 Z 0 H C
// 0x89 ADC A,C 1 4 0 Z 0 H C
// 0x8A ADC A,D 1 4 0 Z 0 H C
// 0x8B ADC A,E 1 4 0 Z 0 H C
// 0x8C ADC A,H 1 4 0 Z 0 H C
// 0x8D ADC A,L 1 4 0 Z 0 H C
// 0x8F ADC A,A 1 4 0 Z 0 H C
func (c *cpu) adc_r_r(opcode uint8, b bus) {
	var v uint8

	switch opcode {
	case 0x88:
		v = c.B
	case 0x89:
		v = c.C
	case 0x8A:
		v = c.D
	case 0x8B:
		v = c.E
	case 0x8C:
		v = c.H
	case 0x8D:
		v = c.L
	case 0x8F:
		v = c.A
	}

	c.A = c.addc8(c.A, v)
}

// 0xC6 ADD A,d8        2 8 0 Z 0 H C
func (c *cpu) add_r_d8(opcode uint8, b bus) {
	v := b.read(c.PC)
	c.PC++

	c.A = c.add8(c.A, v)
}

// 0x86 ADD A,(HL)      1 8 0 Z 0 H C
func (c *cpu) add_r_irr(opcode uint8, b bus) {
	lo := uint16(c.L)
	hi := uint16(c.H)

	addr := hi<<8 | lo
	v := b.read(addr)

	c.A = c.add8(c.A, v)
}

// 0x80 ADD A,B 1 4 0 Z 0 H C
// 0x81 ADD A,C 1 4 0 Z 0 H C
// 0x82 ADD A,D 1 4 0 Z 0 H C
// 0x83 ADD A,E 1 4 0 Z 0 H C
// 0x84 ADD A,H 1 4 0 Z 0 H C
// 0x85 ADD A,L 1 4 0 Z 0 H C
// 0x87 ADD A,A 1 4 0 Z 0 H C
func (c *cpu) add_r_r(opcode uint8, b bus) {
	var v uint8

	switch opcode {
	case 0x80:
		v = c.B
	case 0x81:
		v = c.C
	case 0x82:
		v = c.D
	case 0x83:
		v = c.E
	case 0x84:
		v = c.H
	case 0x85:
		v = c.L
	case 0x87:
		v = c.A
	}

	c.A = c.add8(c.A, v)
}

// 0x09 ADD HL,BC       1 8 0 - 0 H C
// 0x19 ADD HL,DE       1 8 0 - 0 H C
// 0x29 ADD HL,HL       1 8 0 - 0 H C
func (c *cpu) add_rr_rr(opcode uint8, b bus) {
	hl := uint16(c.H)<<8 | uint16(c.L)

	var v uint16
	switch opcode {
	case 0x09:
		v = uint16(c.B)<<8 | uint16(c.C)
	case 0x19:
		v = uint16(c.D)<<8 | uint16(c.E)
	case 0x29:
		v = uint16(c.H)<<8 | uint16(c.L)
	}

	hl += v

	c.F.set(fn, false)
	c.F.set(fh, hl&0xFFF+v&0xFFF > 0xFFF)
	c.F.set(fc, uint32(hl)+uint32(v) > 0xFFFF)

	c.L = uint8(hl & 0xFF)
	c.H = uint8(hl >> 8)

	b.read(c.PC) // TODO: what actually gets read (or written)?
}

// 0x39 ADD HL,SP       1 8 0 - 0 H C
func (c *cpu) add_rr_sp(opcode uint8, b bus) {
	hl := uint16(c.H)<<8 | uint16(c.L)
	v := c.SP

	hl += v

	c.F.set(fn, false)
	c.F.set(fh, hl&0xFFF+v&0xFFF > 0xFFF)
	c.F.set(fc, uint32(hl)+uint32(v) > 0xFFFF)

	c.L = uint8(hl & 0xFF)
	c.H = uint8(hl >> 8)

	b.read(c.PC) // TODO: what actually gets read (or written)?
}

// 0xE8 ADD SP,r8       2 16 0 0 0 H C
func (c *cpu) add_sp_r8(opcode uint8, b bus) {
	r8 := b.read(c.PC)
	c.PC++

	v := uint16(r8)
	c.SP += v

	c.F.set(fz, false)
	c.F.set(fn, false)
	c.F.set(fh, c.SP&0xFFF+v&0xFFF > 0xFFF)
	c.F.set(fc, uint32(c.SP)+uint32(v) > 0xFFFF)

	b.read(c.PC) // TODO: what actually gets read (or written)?
	b.read(c.PC) // TODO: what actually gets read (or written)?
}

// 0xE6 AND d8  2 8 0 Z 0 1 0
func (c *cpu) and_d8(opcode uint8, b bus) {
	v := b.read(c.PC)
	c.PC++

	c.A &= v
	c.F.set(fz, c.A == 0)
	c.F.set(fn, false)
	c.F.set(fh, true)
	c.F.set(fc, false)
}

// 0xA6 AND (HL)        1 8 0 Z 0 1 0
func (c *cpu) and_irr(opcode uint8, b bus) {
	hi := uint16(c.H)
	lo := uint16(c.L)

	addr := hi<<8 | lo
	v := b.read(addr)

	c.A &= v
	c.F.set(fz, c.A == 0)
	c.F.set(fn, false)
	c.F.set(fh, true)
	c.F.set(fc, false)
}

// 0xA0 AND B   1 4 0 Z 0 1 0
// 0xA1 AND C   1 4 0 Z 0 1 0
// 0xA2 AND D   1 4 0 Z 0 1 0
// 0xA3 AND E   1 4 0 Z 0 1 0
// 0xA4 AND H   1 4 0 Z 0 1 0
// 0xA5 AND L   1 4 0 Z 0 1 0
// 0xA7 AND A   1 4 0 Z 0 1 0
func (c *cpu) and_r(opcode uint8, b bus) {
	var v uint8
	switch opcode {
	case 0xA0:
		v = c.B
	case 0xA1:
		v = c.C
	case 0xA2:
		v = c.D
	case 0xA3:
		v = c.E
	case 0xA4:
		v = c.H
	case 0xA5:
		v = c.L
	case 0xA7:
		v = c.A
	}

	c.A &= v
	c.F.set(fz, c.A == 0)
	c.F.set(fn, false)
	c.F.set(fh, true)
	c.F.set(fc, false)
}

// 0xD4 CALL NC,a16     3 24 12 - - - -
func (c *cpu) call_NC_a16(opcode uint8, b bus) {
	lo := uint16(b.read(c.PC))
	c.PC++
	hi := uint16(b.read(c.PC))
	c.PC++

	if c.F&fc > 0 {
		return
	}

	c.SP--
	b.write(c.SP, uint8(c.PC>>8))
	c.SP--
	b.write(c.SP, uint8(c.PC&0xFF))

	c.PC = hi<<8 | lo
	b.read(c.PC) // TODO: what actually gets read (or written)?
}

// 0xC4 CALL NZ,a16     3 24 12 - - - -
func (c *cpu) call_NZ_a16(opcode uint8, b bus) {
	lo := uint16(b.read(c.PC))
	c.PC++
	hi := uint16(b.read(c.PC))
	c.PC++

	if c.F&fz > 0 {
		return
	}

	c.SP--
	b.write(c.SP, uint8(c.PC>>8))
	c.SP--
	b.write(c.SP, uint8(c.PC&0xFF))

	c.PC = hi<<8 | lo
	b.read(c.PC) // TODO: what actually gets read (or written)?
}

// 0xCC CALL Z,a16      3 24 12 - - - -
func (c *cpu) call_Z_a16(opcode uint8, b bus) {
	lo := uint16(b.read(c.PC))
	c.PC++
	hi := uint16(b.read(c.PC))
	c.PC++

	if c.F&fz == 0 {
		return
	}

	c.SP--
	b.write(c.SP, uint8(c.PC>>8))
	c.SP--
	b.write(c.SP, uint8(c.PC&0xFF))

	c.PC = hi<<8 | lo
	b.read(c.PC) // TODO: what actually gets read (or written)?
}

// 0xCD CALL a16        3 24 0 - - - -
func (c *cpu) call_a16(opcode uint8, b bus) {
	lo := uint16(b.read(c.PC))
	c.PC++
	hi := uint16(b.read(c.PC))
	c.PC++

	c.SP--
	b.write(c.SP, uint8(c.PC>>8))
	c.SP--
	b.write(c.SP, uint8(c.PC&0xFF))

	c.PC = hi<<8 | lo
	b.read(c.PC) // TODO: what actually gets read (or written)?
}

// 0xDC CALL C,a16      3 24 12 - - - -
func (c *cpu) call_r_a16(opcode uint8, b bus) {
	lo := uint16(b.read(c.PC))
	c.PC++
	hi := uint16(b.read(c.PC))
	c.PC++

	if c.F&fc == 0 {
		return
	}

	c.SP--
	b.write(c.SP, uint8(c.PC>>8))
	c.SP--
	b.write(c.SP, uint8(c.PC&0xFF))

	c.PC = hi<<8 | lo
	b.read(c.PC) // TODO: what actually gets read (or written)?
}

// 0x3F CCF     1 4 0 - 0 0 C
func (c *cpu) ccf(opcode uint8, b bus) {
	c.F.set(fn, false)
	c.F.set(fh, false)

	if c.F&fc > 0 {
		c.F.set(fc, false)
	} else {
		c.F.set(fc, true)
	}
}

// 0xFE CP d8   2 8 0 Z 1 H C
func (c *cpu) cp_d8(opcode uint8, b bus) {
	v := b.read(c.PC)
	c.PC++

	c.F.set(fz, c.A == v)
	c.F.set(fn, true)
	c.F.set(fh, v < c.A)
	c.F.set(fc, v > c.A)
}

// 0xBE CP (HL) 1 8 0 Z 1 H C
func (c *cpu) cp_irr(opcode uint8, b bus) {
	lo := uint16(c.L)
	hi := uint16(c.H)
	addr := hi<<8 | lo
	v := b.read(addr)

	c.F.set(fz, c.A == v)
	c.F.set(fn, true)
	c.F.set(fh, v < c.A)
	c.F.set(fc, v > c.A)
}

// 0xB8 CP B    1 4 0 Z 1 H C
// 0xB9 CP C    1 4 0 Z 1 H C
// 0xBA CP D    1 4 0 Z 1 H C
// 0xBB CP E    1 4 0 Z 1 H C
// 0xBC CP H    1 4 0 Z 1 H C
// 0xBD CP L    1 4 0 Z 1 H C
// 0xBF CP A    1 4 0 Z 1 H C
func (c *cpu) cp_r(opcode uint8, b bus) {
	var v uint8
	switch opcode {
	case 0xB8:
		v = c.B
	case 0xB9:
		v = c.C
	case 0xBA:
		v = c.D
	case 0xBB:
		v = c.E
	case 0xBC:
		v = c.H
	case 0xBD:
		v = c.L
	case 0xBF:
		v = c.A
	}

	c.F.set(fz, c.A == v)
	c.F.set(fn, true)
	c.F.set(fh, v < c.A)
	c.F.set(fc, v > c.A)
}

// 0x2F CPL     1 4 0 - 1 1 -
func (c *cpu) cpl(opcode uint8, b bus) {
	c.A = c.A ^ 0xFF
	c.F.set(fn, true)
	c.F.set(fh, true)
}

// 0x27 DAA     1 4 0 Z - 0 C
func (c *cpu) daa(opcode uint8, b bus) { panic("not implemented") }

// 0x35 DEC (HL)        1 12 0 Z 1 H -
func (c *cpu) dec_irr(opcode uint8, b bus) {
	lo := uint16(c.L)
	hi := uint16(c.H)
	addr := hi<<8 | lo

	v := b.read(addr)

	c.F.set(fz, v-1 == 0)
	c.F.set(fn, true)
	c.F.set(fh, v&0xF == 0)

	b.write(addr, v-1)
}

// 0x05 DEC B   1 4 0 Z 1 H -
// 0x0D DEC C   1 4 0 Z 1 H -
// 0x15 DEC D   1 4 0 Z 1 H -
// 0x1D DEC E   1 4 0 Z 1 H -
// 0x25 DEC H   1 4 0 Z 1 H -
// 0x2D DEC L   1 4 0 Z 1 H -
// 0x3D DEC A   1 4 0 Z 1 H -
func (c *cpu) dec_r(opcode uint8, b bus) {
	var r *uint8

	switch opcode {
	case 0x05:
		r = &c.B
	case 0x0D:
		r = &c.C
	case 0x15:
		r = &c.D
	case 0x1D:
		r = &c.E
	case 0x25:
		r = &c.H
	case 0x2D:
		r = &c.L
	case 0x3D:
		r = &c.A
	}

	v := *r
	c.F.set(fz, v-1 == 0)
	c.F.set(fn, true)
	c.F.set(fh, v&0xF == 0)
	*r = v - 1
}

// 0x0B DEC BC  1 8 0 - - - -
// 0x1B DEC DE  1 8 0 - - - -
// 0x2B DEC HL  1 8 0 - - - -
func (c *cpu) dec_rr(opcode uint8, b bus) {
	var rrhi, rrlo *uint8

	switch opcode {
	case 0x0B:
		rrhi = &c.B
		rrlo = &c.C
	case 0x1B:
		rrhi = &c.D
		rrlo = &c.E
	case 0x2B:
		rrhi = &c.H
		rrlo = &c.L
	}

	v := uint16(*rrhi)<<8 | uint16(*rrlo)
	v--
	*rrhi = uint8(v >> 8)
	*rrlo = uint8(v & 0xFF)

	b.read(c.PC) // TODO: what actually gets read (or written)?
}

// 0x3B DEC SP  1 8 0 - - - -
func (c *cpu) dec_sp(opcode uint8, b bus) {
	c.SP--
	b.read(c.PC) // TODO: what actually gets read (or written)?
}

// 0xF3 DI      1 4 0 - - - -
func (c *cpu) di(opcode uint8, b bus) { panic("not implemented") }

// 0xFB EI      1 4 0 - - - -
func (c *cpu) ei(opcode uint8, b bus) { panic("not implemented") }

// 0x76 HALT    1 4 0 - - - -
func (c *cpu) halt(opcode uint8, b bus) { panic("not implemented") }

// 0x34 INC (HL)        1 12 0 Z 0 H -
func (c *cpu) inc_irr(opcode uint8, b bus) {
	lo := uint16(c.L)
	hi := uint16(c.H)
	addr := hi<<8 | lo
	v := b.read(addr)

	c.F.set(fz, v+1 == 0)
	c.F.set(fn, false)
	c.F.set(fh, v&0xF == 0xF)

	b.write(addr, v+1)
}

// 0x04 INC B   1 4 0 Z 0 H -
// 0x0C INC C   1 4 0 Z 0 H -
// 0x14 INC D   1 4 0 Z 0 H -
// 0x1C INC E   1 4 0 Z 0 H -
// 0x24 INC H   1 4 0 Z 0 H -
// 0x2C INC L   1 4 0 Z 0 H -
// 0x3C INC A   1 4 0 Z 0 H -
func (c *cpu) inc_r(opcode uint8, b bus) {
	var r *uint8

	switch opcode {
	case 0x04:
		r = &c.B
	case 0x0C:
		r = &c.C
	case 0x14:
		r = &c.D
	case 0x1C:
		r = &c.E
	case 0x24:
		r = &c.H
	case 0x2C:
		r = &c.L
	case 0x3C:
		r = &c.A
	}

	v := *r
	c.F.set(fz, v+1 == 0)
	c.F.set(fn, false)
	c.F.set(fh, v&0xF == 0xF)
	*r = v + 1
}

// 0x03 INC BC  1 8 0 - - - -
// 0x13 INC DE  1 8 0 - - - -
// 0x23 INC HL  1 8 0 - - - -
func (c *cpu) inc_rr(opcode uint8, b bus) {
	var rrhi, rrlo *uint8

	switch opcode {
	case 0x03:
		rrhi = &c.B
		rrlo = &c.C
	case 0x13:
		rrhi = &c.D
		rrlo = &c.E
	case 0x23:
		rrhi = &c.H
		rrlo = &c.L
	}

	v := uint16(*rrhi)<<8 | uint16(*rrlo)
	v++
	*rrhi = uint8(v >> 8)
	*rrlo = uint8(v & 0xFF)

	b.read(c.PC) // TODO: what actually gets read (or written)?
}

// 0x33 INC SP  1 8 0 - - - -
func (c *cpu) inc_sp(opcode uint8, b bus) {
	c.SP++
	b.read(c.PC) // TODO: what actually gets read (or written)?
}

// 0xD2 JP NC,a16       3 16 12 - - - -
func (c *cpu) jp_NC_a16(opcode uint8, b bus) {
	lo := uint16(b.read(c.PC))
	c.PC++
	hi := uint16(b.read(c.PC))
	c.PC++

	if c.F&fc > 0 {
		return
	}

	c.PC = hi<<8 | lo
	b.read(c.PC) // TODO: what actually gets read (or written)?
}

// 0xC2 JP NZ,a16       3 16 12 - - - -
func (c *cpu) jp_NZ_a16(opcode uint8, b bus) {
	lo := uint16(b.read(c.PC))
	c.PC++
	hi := uint16(b.read(c.PC))
	c.PC++

	if c.F&fz > 0 {
		return
	}

	c.PC = hi<<8 | lo
	b.read(c.PC) // TODO: what actually gets read (or written)?
}

// 0xCA JP Z,a16        3 16 12 - - - -
func (c *cpu) jp_Z_a16(opcode uint8, b bus) {
	lo := uint16(b.read(c.PC))
	c.PC++
	hi := uint16(b.read(c.PC))
	c.PC++

	if c.F&fz == 0 {
		return
	}

	c.PC = hi<<8 | lo
	b.read(c.PC) // TODO: what actually gets read (or written)? }
}

// 0xC3 JP a16  3 16 0 - - - -
func (c *cpu) jp_a16(opcode uint8, b bus) {
	lo := uint16(b.read(c.PC))
	c.PC++
	hi := uint16(b.read(c.PC))
	c.PC++

	c.PC = hi<<8 | lo
	b.read(c.PC) // TODO: what actually gets read (or written)?
}

// 0xE9 JP (HL) 1 4 0 - - - -
func (c *cpu) jp_irr(opcode uint8, b bus) {
	lo := uint16(c.L)
	hi := uint16(c.H)

	c.PC = hi<<8 | lo
}

// 0xDA JP C,a16        3 16 12 - - - -
func (c *cpu) jp_r_a16(opcode uint8, b bus) {
	lo := uint16(b.read(c.PC))
	c.PC++
	hi := uint16(b.read(c.PC))
	c.PC++

	if c.F&fc == 0 {
		return
	}

	c.PC = hi<<8 | lo
	b.read(c.PC) // TODO: what actually gets read (or written)? }
}

// 0x30 JR NC,r8        2 12 8 - - - -
func (c *cpu) jr_NC_r8(opcode uint8, b bus) {
	r8 := uint16(int8(b.read(c.PC)))
	c.PC++

	if c.F&fc > 0 {
		return
	}

	c.PC += r8
	b.read(c.PC) // TODO: what actually gets read (or written)? }
}

// 0x20 JR NZ,r8        2 12 8 - - - -
func (c *cpu) jr_NZ_r8(opcode uint8, b bus) {
	r8 := uint16(int8(b.read(c.PC)))
	c.PC++

	if c.F&fz > 0 {
		return
	}

	c.PC += r8
	b.read(c.PC) // TODO: what actually gets read (or written)? }
}

// 0x28 JR Z,r8 2 12 8 - - - -
func (c *cpu) jr_Z_r8(opcode uint8, b bus) {
	r8 := uint16(int8(b.read(c.PC)))
	c.PC++

	if c.F&fz == 0 {
		return
	}

	c.PC += uint16(r8)
	b.read(c.PC) // TODO: what actually gets read (or written)? } }
}

// 0x18 JR r8   2 12 0 - - - -
func (c *cpu) jr_r8(opcode uint8, b bus) {
	r8 := uint16(int8(b.read(c.PC)))
	c.PC++

	c.PC += r8
	b.read(c.PC) // TODO: what actually gets read (or written)? }
}

// 0x38 JR C,r8 2 12 8 - - - -
func (c *cpu) jr_r_r8(opcode uint8, b bus) {
	r8 := uint16(int8(b.read(c.PC)))
	c.PC++

	if c.F&fc == 0 {
		return
	}

	c.PC += uint16(r8)
	b.read(c.PC) // TODO: what actually gets read (or written)? } }
}

// 0x22 LD (HL+),A      1 8 0 - - - -
// 0x32 LD (HL-),A      1 8 0 - - - -
func (c *cpu) ld__r(opcode uint8, b bus) {
	addr := uint16(c.H)<<8 | uint16(c.L)
	b.write(addr, c.A)

	switch opcode {
	case 0x22:
		addr++
	case 0x32:
		addr--
	}
	c.H = uint8(addr >> 8)
	c.L = uint8(addr & 0xFF)
}

// 0xEA LD (a16),A      3 16 0 - - - -
func (c *cpu) ld_ia16_r(opcode uint8, b bus) {
	lo := uint16(b.read(c.PC))
	c.PC++
	hi := uint16(b.read(c.PC))
	c.PC++

	addr := hi<<8 | lo
	b.write(addr, c.A)
}

// 0x08 LD (a16),SP     3 20 0 - - - -
func (c *cpu) ld_ia16_sp(opcode uint8, b bus) {
	lo := uint16(b.read(c.PC))
	c.PC++
	hi := uint16(b.read(c.PC))
	c.PC++

	addr := hi<<8 | lo
	b.write(addr, uint8(c.SP&0xFF))
	b.write(addr+1, uint8(c.SP>>8))
}

// 0xE2 LD (C),A        2 8 0 - - - -
func (c *cpu) ld_ir_r(opcode uint8, b bus) {
	b.write(0xFF00+uint16(c.C), c.A)
}

// 0x36 LD (HL),d8      2 12 0 - - - -
func (c *cpu) ld_irr_d8(opcode uint8, b bus) {
	d8 := b.read(c.PC)
	c.PC++

	addr := uint16(c.H)<<8 | uint16(c.L)
	b.write(addr, d8)
}

// 0x02 LD (BC),A       1 8 0 - - - -
// 0x12 LD (DE),A       1 8 0 - - - -
// 0x70 LD (HL),B       1 8 0 - - - -
// 0x71 LD (HL),C       1 8 0 - - - -
// 0x72 LD (HL),D       1 8 0 - - - -
// 0x73 LD (HL),E       1 8 0 - - - -
// 0x74 LD (HL),H       1 8 0 - - - -
// 0x75 LD (HL),L       1 8 0 - - - -
// 0x77 LD (HL),A       1 8 0 - - - -
func (c *cpu) ld_irr_r(opcode uint8, b bus) {
	var irrhi, irrlo, r *uint8

	switch opcode {
	case 0x02:
		irrhi = &c.B
		irrlo = &c.C
		r = &c.A
	case 0x12:
		irrhi = &c.D
		irrlo = &c.E
		r = &c.A
	case 0x70:
		irrhi = &c.H
		irrlo = &c.L
		r = &c.B
	case 0x71:
		irrhi = &c.H
		irrlo = &c.L
		r = &c.C
	case 0x72:
		irrhi = &c.H
		irrlo = &c.L
		r = &c.D
	case 0x73:
		irrhi = &c.H
		irrlo = &c.L
		r = &c.E
	case 0x74:
		irrhi = &c.H
		irrlo = &c.L
		r = &c.H
	case 0x75:
		irrhi = &c.H
		irrlo = &c.L
		r = &c.L
	case 0x77:
		irrhi = &c.H
		irrlo = &c.L
		r = &c.A
	}

	lo := uint16(*irrlo)
	hi := uint16(*irrhi)
	v := *r
	addr := hi<<8 | lo
	b.write(addr, v)
}

// 0x2A LD A,(HL+)      1 8 0 - - - -
// 0x3A LD A,(HL-)      1 8 0 - - - -
func (c *cpu) ld_r_hlid(opcode uint8, b bus) {
	addr := uint16(c.H)<<8 | uint16(c.L)
	c.A = b.read(addr)

	switch opcode {
	case 0x2A:
		addr++
	case 0x3A:
		addr--
	}

	c.H = uint8(addr >> 8)
	c.L = uint8(addr & 0xFF)
}

// 0x06 LD B,d8 2 8 0 - - - -
// 0x0E LD C,d8 2 8 0 - - - -
// 0x16 LD D,d8 2 8 0 - - - -
// 0x1E LD E,d8 2 8 0 - - - -
// 0x26 LD H,d8 2 8 0 - - - -
// 0x2E LD L,d8 2 8 0 - - - -
// 0x3E LD A,d8 2 8 0 - - - -
func (c *cpu) ld_r_d8(opcode uint8, b bus) {
	d8 := b.read(c.PC)
	c.PC++

	var r *uint8
	switch opcode {
	case 0x06:
		r = &c.B
	case 0x0E:
		r = &c.C
	case 0x16:
		r = &c.D
	case 0x1E:
		r = &c.E
	case 0x26:
		r = &c.H
	case 0x2E:
		r = &c.L
	case 0x3E:
		r = &c.A
	}

	*r = d8
}

// 0xFA LD A,(a16)      3 16 0 - - - -
func (c *cpu) ld_r_ia16(opcode uint8, b bus) {
	lo := uint16(b.read(c.PC))
	c.PC++
	hi := uint16(b.read(c.PC))
	c.PC++

	addr := hi<<8 | lo
	c.A = b.read(addr)
}

// 0xF2 LD A,(C)        2 8 0 - - - -
func (c *cpu) ld_r_ir(opcode uint8, b bus) {
	c.A = b.read(0xFF00 + uint16(c.C))
}

// 0x0A LD A,(BC)       1 8 0 - - - -
// 0x1A LD A,(DE)       1 8 0 - - - -
// 0x46 LD B,(HL)       1 8 0 - - - -
// 0x4E LD C,(HL)       1 8 0 - - - -
// 0x56 LD D,(HL)       1 8 0 - - - -
// 0x5E LD E,(HL)       1 8 0 - - - -
// 0x66 LD H,(HL)       1 8 0 - - - -
// 0x6E LD L,(HL)       1 8 0 - - - -
// 0x7E LD A,(HL)       1 8 0 - - - -
func (c *cpu) ld_r_irr(opcode uint8, b bus) {
	var r, irrhi, irrlo *uint8

	switch opcode {
	case 0x0A:
		r = &c.A
		irrhi = &c.B
		irrlo = &c.C
	case 0x1A:
		r = &c.A
		irrhi = &c.D
		irrlo = &c.E
	case 0x46:
		r = &c.B
		irrhi = &c.H
		irrlo = &c.L
	case 0x4E:
		r = &c.C
		irrhi = &c.H
		irrlo = &c.L
	case 0x56:
		r = &c.D
		irrhi = &c.H
		irrlo = &c.L
	case 0x5E:
		r = &c.E
		irrhi = &c.H
		irrlo = &c.L
	case 0x66:
		r = &c.H
		irrhi = &c.H
		irrlo = &c.L
	case 0x6E:
		r = &c.L
		irrhi = &c.H
		irrlo = &c.L
	case 0x7E:
		r = &c.A
		irrhi = &c.H
		irrlo = &c.L
	}

	lo := uint16(*irrlo)
	hi := uint16(*irrhi)
	addr := hi<<8 | lo
	*r = b.read(addr)
}

// 0x40 LD B,B  1 4 0 - - - -
// 0x41 LD B,C  1 4 0 - - - -
// 0x42 LD B,D  1 4 0 - - - -
// 0x43 LD B,E  1 4 0 - - - -
// 0x44 LD B,H  1 4 0 - - - -
// 0x45 LD B,L  1 4 0 - - - -
// 0x47 LD B,A  1 4 0 - - - -
// 0x48 LD C,B  1 4 0 - - - -
// 0x49 LD C,C  1 4 0 - - - -
// 0x4A LD C,D  1 4 0 - - - -
// 0x4B LD C,E  1 4 0 - - - -
// 0x4C LD C,H  1 4 0 - - - -
// 0x4D LD C,L  1 4 0 - - - -
// 0x4F LD C,A  1 4 0 - - - -
// 0x50 LD D,B  1 4 0 - - - -
// 0x51 LD D,C  1 4 0 - - - -
// 0x52 LD D,D  1 4 0 - - - -
// 0x53 LD D,E  1 4 0 - - - -
// 0x54 LD D,H  1 4 0 - - - -
// 0x55 LD D,L  1 4 0 - - - -
// 0x57 LD D,A  1 4 0 - - - -
// 0x58 LD E,B  1 4 0 - - - -
// 0x59 LD E,C  1 4 0 - - - -
// 0x5A LD E,D  1 4 0 - - - -
// 0x5B LD E,E  1 4 0 - - - -
// 0x5C LD E,H  1 4 0 - - - -
// 0x5D LD E,L  1 4 0 - - - -
// 0x5F LD E,A  1 4 0 - - - -
// 0x60 LD H,B  1 4 0 - - - -
// 0x61 LD H,C  1 4 0 - - - -
// 0x62 LD H,D  1 4 0 - - - -
// 0x63 LD H,E  1 4 0 - - - -
// 0x64 LD H,H  1 4 0 - - - -
// 0x65 LD H,L  1 4 0 - - - -
// 0x67 LD H,A  1 4 0 - - - -
// 0x68 LD L,B  1 4 0 - - - -
// 0x69 LD L,C  1 4 0 - - - -
// 0x6A LD L,D  1 4 0 - - - -
// 0x6B LD L,E  1 4 0 - - - -
// 0x6C LD L,H  1 4 0 - - - -
// 0x6D LD L,L  1 4 0 - - - -
// 0x6F LD L,A  1 4 0 - - - -
// 0x78 LD A,B  1 4 0 - - - -
// 0x79 LD A,C  1 4 0 - - - -
// 0x7A LD A,D  1 4 0 - - - -
// 0x7B LD A,E  1 4 0 - - - -
// 0x7C LD A,H  1 4 0 - - - -
// 0x7D LD A,L  1 4 0 - - - -
// 0x7F LD A,A  1 4 0 - - - -
func (c *cpu) ld_r_r(opcode uint8, b bus) {
	var r1, r2 *uint8

	switch opcode {
	case 0x40:
		r1 = &c.B
		r2 = &c.B
	case 0x41:
		r1 = &c.B
		r2 = &c.C
	case 0x42:
		r1 = &c.B
		r2 = &c.D
	case 0x43:
		r1 = &c.B
		r2 = &c.E
	case 0x44:
		r1 = &c.B
		r2 = &c.H
	case 0x45:
		r1 = &c.B
		r2 = &c.L
	case 0x47:
		r1 = &c.B
		r2 = &c.A
	case 0x48:
		r1 = &c.C
		r2 = &c.B
	case 0x49:
		r1 = &c.C
		r2 = &c.C
	case 0x4A:
		r1 = &c.C
		r2 = &c.D
	case 0x4B:
		r1 = &c.C
		r2 = &c.E
	case 0x4C:
		r1 = &c.C
		r2 = &c.H
	case 0x4D:
		r1 = &c.C
		r2 = &c.L
	case 0x4F:
		r1 = &c.C
		r2 = &c.A
	case 0x50:
		r1 = &c.D
		r2 = &c.B
	case 0x51:
		r1 = &c.D
		r2 = &c.C
	case 0x52:
		r1 = &c.D
		r2 = &c.D
	case 0x53:
		r1 = &c.D
		r2 = &c.E
	case 0x54:
		r1 = &c.D
		r2 = &c.H
	case 0x55:
		r1 = &c.D
		r2 = &c.L
	case 0x57:
		r1 = &c.D
		r2 = &c.A
	case 0x58:
		r1 = &c.E
		r2 = &c.B
	case 0x59:
		r1 = &c.E
		r2 = &c.C
	case 0x5A:
		r1 = &c.E
		r2 = &c.D
	case 0x5B:
		r1 = &c.E
		r2 = &c.E
	case 0x5C:
		r1 = &c.E
		r2 = &c.H
	case 0x5D:
		r1 = &c.E
		r2 = &c.L
	case 0x5F:
		r1 = &c.E
		r2 = &c.A
	case 0x60:
		r1 = &c.H
		r2 = &c.B
	case 0x61:
		r1 = &c.H
		r2 = &c.C
	case 0x62:
		r1 = &c.H
		r2 = &c.D
	case 0x63:
		r1 = &c.H
		r2 = &c.E
	case 0x64:
		r1 = &c.H
		r2 = &c.H
	case 0x65:
		r1 = &c.H
		r2 = &c.L
	case 0x67:
		r1 = &c.H
		r2 = &c.A
	case 0x68:
		r1 = &c.L
		r2 = &c.B
	case 0x69:
		r1 = &c.L
		r2 = &c.C
	case 0x6A:
		r1 = &c.L
		r2 = &c.D
	case 0x6B:
		r1 = &c.L
		r2 = &c.E
	case 0x6C:
		r1 = &c.L
		r2 = &c.H
	case 0x6D:
		r1 = &c.L
		r2 = &c.L
	case 0x6F:
		r1 = &c.L
		r2 = &c.A
	case 0x78:
		r1 = &c.A
		r2 = &c.B
	case 0x79:
		r1 = &c.A
		r2 = &c.C
	case 0x7A:
		r1 = &c.A
		r2 = &c.D
	case 0x7B:
		r1 = &c.A
		r2 = &c.E
	case 0x7C:
		r1 = &c.A
		r2 = &c.H
	case 0x7D:
		r1 = &c.A
		r2 = &c.L
	case 0x7F:
		r1 = &c.A
		r2 = &c.A
	}

	*r1 = *r2
}

// 0xF8 LD HL,SP+r8     2 12 0 0 0 H C
func (c *cpu) ld_rr_SP_r8(opcode uint8, b bus) {
	v := uint16(b.read(c.SP))
	c.SP++

	result := c.SP + v
	c.L = uint8(result & 0xFF)
	c.H = uint8(result >> 8)

	c.F.set(fz, false)
	c.F.set(fn, false)
	c.F.set(fh, c.SP&0xF+v&0xF > 0xF)
	c.F.set(fc, uint32(c.SP)+uint32(v) > 0xFFFF)
}

// 0x01 LD BC,d16       3 12 0 - - - -
// 0x11 LD DE,d16       3 12 0 - - - -
// 0x21 LD HL,d16       3 12 0 - - - -
func (c *cpu) ld_rr_d16(opcode uint8, b bus) {
	var rrhi, rrlo *uint8

	switch opcode {
	case 0x01:
		rrhi = &c.B
		rrlo = &c.C
	case 0x11:
		rrhi = &c.D
		rrlo = &c.E
	case 0x21:
		rrhi = &c.H
		rrlo = &c.L
	}

	*rrhi = b.read(c.PC)
	c.PC++
	*rrlo = b.read(c.PC)
	c.PC++
}

// 0x31 LD SP,d16       3 12 0 - - - -
func (c *cpu) ld_sp_d16(opcode uint8, b bus) {
	lo := uint16(b.read(c.PC))
	c.PC++
	hi := uint16(b.read(c.PC))
	c.PC++

	c.SP = hi<<8 | lo
}

// 0xF9 LD SP,HL        1 8 0 - - - -
func (c *cpu) ld_sp_rr(opcode uint8, b bus) {
	c.SP = uint16(c.H)<<8 | uint16(c.L)
	c.read(c.PC) // TODO: what actually gets read (or written)?
}

// 0xE0 LDH (a8),A      2 12 0 - - - -
func (c *cpu) ldh_ia8_r(opcode uint8, b bus) {
	a8 := b.read(c.PC)
	c.PC++
	b.write(uint16(a8), c.A)
}

// 0xF0 LDH A,(a8)      2 12 0 - - - -
func (c *cpu) ldh_r_ia8(opcode uint8, b bus) {
	a8 := b.read(c.PC)
	c.PC++
	c.A = b.read(uint16(a8))
}

// 0x00 NOP     1 4 0 - - - -
func (c *cpu) nop(opcode uint8, b bus) {}

// 0xF6 OR d8   2 8 0 Z 0 0 0
func (c *cpu) or_d8(opcode uint8, b bus) {
	v := b.read(c.PC)
	c.PC++

	c.A |= v
	c.F.set(fz, c.A == 0)
}

// 0xB6 OR (HL) 1 8 0 Z 0 0 0
func (c *cpu) or_irr(opcode uint8, b bus) {
	lo := uint16(c.L)
	hi := uint16(c.H)
	addr := hi<<8 | lo
	v := b.read(addr)

	c.A |= v
	c.F.set(fz, c.A == 0)
}

// 0xB0 OR B    1 4 0 Z 0 0 0
// 0xB1 OR C    1 4 0 Z 0 0 0
// 0xB2 OR D    1 4 0 Z 0 0 0
// 0xB3 OR E    1 4 0 Z 0 0 0
// 0xB4 OR H    1 4 0 Z 0 0 0
// 0xB5 OR L    1 4 0 Z 0 0 0
// 0xB7 OR A    1 4 0 Z 0 0 0
func (c *cpu) or_r(opcode uint8, b bus) {
	var v uint8
	switch opcode {
	case 0xB0:
		v = c.B
	case 0xB1:
		v = c.C
	case 0xB2:
		v = c.D
	case 0xB3:
		v = c.E
	case 0xB4:
		v = c.H
	case 0xB5:
		v = c.L
	case 0xB7:
		v = c.A
	}

	c.A |= v
	c.F.set(fz, c.A == 0)
}

// 0xC1 POP BC  1 12 0 - - - -
// 0xD1 POP DE  1 12 0 - - - -
// 0xE1 POP HL  1 12 0 - - - -
// 0xF1 POP AF  1 12 0 Z N H C
func (c *cpu) pop_rr(opcode uint8, b bus) {
	var rrhi, rrlo *uint8

	switch opcode {
	case 0xC1:
		rrhi = &c.B
		rrlo = &c.C
	case 0xD1:
		rrhi = &c.D
		rrlo = &c.E
	case 0xE1:
		rrhi = &c.H
		rrlo = &c.L
	case 0xF1:
		rrhi = &c.A
		rrlo = (*uint8)(&c.F)
	}

	*rrlo = b.read(c.SP)
	c.SP++
	*rrhi = b.read(c.SP)
	c.SP++
}

// 0xCB PREFIX CB       1 4 0 - - - -
func (c *cpu) prefix_(opcode uint8, b bus) { panic("not implemented") }

// 0xC5 PUSH BC 1 16 0 - - - -
// 0xD5 PUSH DE 1 16 0 - - - -
// 0xE5 PUSH HL 1 16 0 - - - -
// 0xF5 PUSH AF 1 16 0 - - - -
func (c *cpu) push_rr(opcode uint8, b bus) {
	var rrhi, rrlo *uint8

	switch opcode {
	case 0xC5:
		rrhi = &c.B
		rrlo = &c.C
	case 0xD5:
		rrhi = &c.D
		rrlo = &c.E
	case 0xE5:
		rrhi = &c.H
		rrlo = &c.L
	case 0xF5:
		rrhi = &c.A
		rrlo = (*uint8)(&c.F)
	}

	c.SP--
	b.write(c.SP, *rrhi)
	c.SP--
	b.write(c.SP, *rrlo)
}

// 0xC9 RET     1 16 0 - - - -
func (c *cpu) ret(opcode uint8, b bus) {
	lo := uint16(b.read(c.SP))
	c.SP++
	hi := uint16(b.read(c.SP))
	c.SP++

	c.PC = hi<<8 | lo

	c.read(c.PC) // TODO: what actually gets read (or written)?
}

// 0xD0 RET NC  1 20 8 - - - -
func (c *cpu) ret_NC(opcode uint8, b bus) {
	c.read(c.PC) // TODO: what actually gets read (or written)?

	lo := uint16(b.read(c.SP))
	c.SP++
	hi := uint16(b.read(c.SP))
	c.SP++

	if c.F&fc > 0 {
		return
	}

	c.PC = hi<<8 | lo

	c.read(c.PC) // TODO: what actually gets read (or written)?
}

// 0xC0 RET NZ  1 20 8 - - - -
func (c *cpu) ret_NZ(opcode uint8, b bus) {
	c.read(c.PC) // TODO: what actually gets read (or written)?

	lo := uint16(b.read(c.SP))
	c.SP++
	hi := uint16(b.read(c.SP))
	c.SP++

	if c.F&fz > 0 {
		return
	}

	c.PC = hi<<8 | lo

	c.read(c.PC) // TODO: what actually gets read (or written)?
}

// 0xC8 RET Z   1 20 8 - - - -
func (c *cpu) ret_Z(opcode uint8, b bus) {
	c.read(c.PC) // TODO: what actually gets read (or written)?

	lo := uint16(b.read(c.SP))
	c.SP++
	hi := uint16(b.read(c.SP))
	c.SP++

	if c.F&fz == 0 {
		return
	}

	c.PC = hi<<8 | lo

	c.read(c.PC) // TODO: what actually gets read (or written)? }
}

// 0xD8 RET C   1 20 8 - - - -
func (c *cpu) ret_r(opcode uint8, b bus) {
	c.read(c.PC) // TODO: what actually gets read (or written)?

	lo := uint16(b.read(c.SP))
	c.SP++
	hi := uint16(b.read(c.SP))
	c.SP++

	if c.F&fc == 0 {
		return
	}

	c.PC = hi<<8 | lo

	c.read(c.PC) // TODO: what actually gets read (or written)?
}

// 0xD9 RETI    1 16 0 - - - -
func (c *cpu) reti(opcode uint8, b bus) {
	lo := uint16(b.read(c.SP))
	c.SP++
	hi := uint16(b.read(c.SP))
	c.SP++

	c.PC = hi<<8 | lo

	c.read(c.PC) // TODO: what actually gets read (or written)?
}

// 0x17 RLA     1 4 0 0 0 0 C
func (c *cpu) rla(opcode uint8, b bus) {
	var carryIn uint8
	if c.F&fc > 0 {
		carryIn = 1
	}
	carryOut := c.A & 0x80
	c.A = c.A<<1 | carryIn

	c.F.set(fz, false)
	c.F.set(fn, false)
	c.F.set(fh, false)
	c.F.set(fc, carryOut > 0)
}

// 0x07 RLCA    1 4 0 0 0 0 C
func (c *cpu) rlca(opcode uint8, b bus) {
	carryOut := c.A & 0x80
	c.A = c.A<<1 | carryOut>>7

	c.F.set(fz, false)
	c.F.set(fn, false)
	c.F.set(fh, false)
	c.F.set(fc, carryOut > 0)
}

// 0x1F RRA     1 4 0 0 0 0 C
func (c *cpu) rra(opcode uint8, b bus) {
	var carryIn uint8
	if c.F&fc > 0 {
		carryIn = 1 << 7
	}

	carryOut := c.A & 0x1
	c.A = c.A>>1 | carryIn

	c.F.set(fz, false)
	c.F.set(fn, false)
	c.F.set(fh, false)
	c.F.set(fc, carryOut > 0)
}

// 0x0F RRCA    1 4 0 0 0 0 C
func (c *cpu) rrca(opcode uint8, b bus) {
	carryOut := c.A & 0x1
	c.A = c.A>>1 | carryOut<<7

	c.F.set(fz, false)
	c.F.set(fn, false)
	c.F.set(fh, false)
	c.F.set(fc, carryOut > 0)
}

// 0xC7 RST 00H 1 16 0 - - - -
// 0xCF RST 08H 1 16 0 - - - -
// 0xD7 RST 10H 1 16 0 - - - -
// 0xDF RST 18H 1 16 0 - - - -
// 0xE7 RST 20H 1 16 0 - - - -
// 0xEF RST 28H 1 16 0 - - - -
// 0xF7 RST 30H 1 16 0 - - - -
// 0xFF RST 38H 1 16 0 - - - -
func (c *cpu) rst_(opcode uint8, b bus) { panic("not implemented") }

// 0xDE SBC A,d8        2 8 0 Z 1 H C
func (c *cpu) sbc_r_d8(opcode uint8, b bus) {
	v := b.read(c.PC)
	c.PC++

	v = v ^ 0xFF
	c.addc8(c.A, v)
	c.F.set(fn, true)
}

// 0x9E SBC A,(HL)      1 8 0 Z 1 H C
func (c *cpu) sbc_r_irr(opcode uint8, b bus) {
	lo := uint16(c.L)
	hi := uint16(c.H)
	addr := hi<<8 | lo
	v := b.read(addr)

	v = v ^ 0xFF
	c.addc8(c.A, v)
	c.F.set(fn, true)
}

// 0x98 SBC A,B 1 4 0 Z 1 H C
// 0x99 SBC A,C 1 4 0 Z 1 H C
// 0x9A SBC A,D 1 4 0 Z 1 H C
// 0x9B SBC A,E 1 4 0 Z 1 H C
// 0x9C SBC A,H 1 4 0 Z 1 H C
// 0x9D SBC A,L 1 4 0 Z 1 H C
// 0x9F SBC A,A 1 4 0 Z 1 H C
func (c *cpu) sbc_r_r(opcode uint8, b bus) {
	var v uint8
	switch opcode {
	case 0x98:
		v = c.B
	case 0x99:
		v = c.C
	case 0x9A:
		v = c.D
	case 0x9B:
		v = c.E
	case 0x9C:
		v = c.H
	case 0x9D:
		v = c.L
	case 0x9F:
		v = c.A
	}

	v = v ^ 0xFF
	c.addc8(c.A, v)
	c.F.set(fn, true)
}

// 0x37 SCF     1 4 0 - 0 0 1
func (c *cpu) scf(opcode uint8, b bus) {
	c.F.set(fc, true)
}

// 0x10 STOP 0  2 4 0 - - - -
func (c *cpu) stop_(opcode uint8, b bus) { panic("not implemented") }

// 0xD6 SUB d8  2 8 0 Z 1 H C
func (c *cpu) sub_d8(opcode uint8, b bus) {
	v := b.read(c.PC)
	c.PC++

	v = v ^ 0xFF + 1
	c.add8(c.A, v)
	c.F.set(fn, true)
}

// 0x96 SUB (HL)        1 8 0 Z 1 H C
func (c *cpu) sub_irr(opcode uint8, b bus) {
	lo := uint16(c.L)
	hi := uint16(c.H)
	addr := hi<<8 | lo
	v := b.read(addr)

	v = v ^ 0xFF + 1
	c.add8(c.A, v)
	c.F.set(fn, true)
}

// 0x90 SUB B   1 4 0 Z 1 H C
// 0x91 SUB C   1 4 0 Z 1 H C
// 0x92 SUB D   1 4 0 Z 1 H C
// 0x93 SUB E   1 4 0 Z 1 H C
// 0x94 SUB H   1 4 0 Z 1 H C
// 0x95 SUB L   1 4 0 Z 1 H C
// 0x97 SUB A   1 4 0 Z 1 H C
func (c *cpu) sub_r(opcode uint8, b bus) {
	var v uint8
	switch opcode {
	case 0x90:
		v = c.B
	case 0x91:
		v = c.C
	case 0x92:
		v = c.D
	case 0x93:
		v = c.E
	case 0x94:
		v = c.H
	case 0x95:
		v = c.L
	case 0x97:
		v = c.A
	}

	v = v ^ 0xFF + 1
	c.add8(c.A, v)
	c.F.set(fn, true)
}

// 0xEE XOR d8  2 8 0 Z 0 0 0
func (c *cpu) xor_d8(opcode uint8, b bus) {
	v := b.read(c.PC)
	c.PC++

	c.A = c.A ^ v
	c.F.set(fz, c.A == 0)
}

// 0xAE XOR (HL)        1 8 0 Z 0 0 0
func (c *cpu) xor_irr(opcode uint8, b bus) {
	lo := uint16(c.L)
	hi := uint16(c.H)
	addr := hi<<8 | lo
	v := b.read(addr)

	c.A = c.A ^ v
	c.F.set(fz, c.A == 0)
}

// 0xA8 XOR B   1 4 0 Z 0 0 0
// 0xA9 XOR C   1 4 0 Z 0 0 0
// 0xAA XOR D   1 4 0 Z 0 0 0
// 0xAB XOR E   1 4 0 Z 0 0 0
// 0xAC XOR H   1 4 0 Z 0 0 0
// 0xAD XOR L   1 4 0 Z 0 0 0
// 0xAF XOR A   1 4 0 Z 0 0 0
func (c *cpu) xor_r(opcode uint8, b bus) {
	var v uint8
	switch opcode {
	case 0xA8:
		v = c.B
	case 0xA9:
		v = c.C
	case 0xAA:
		v = c.D
	case 0xAB:
		v = c.E
	case 0xAC:
		v = c.H
	case 0xAD:
		v = c.L
	case 0xAF:
		v = c.A
	}

	c.A = c.A ^ v
	c.F.set(fz, c.A == 0)
}
