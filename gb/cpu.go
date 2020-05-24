package gb

type op func(opcode uint8, gb *GameBoy)

const (
	vectorVBlank uint16 = 0x40
	vectorLCDc   uint16 = 0x48
	vectorTimer  uint16 = 0x50
	vectorSerial uint16 = 0x58
	vectorHTL    uint16 = 0x60
)

type cpuFlags uint8

const (
	_ cpuFlags = 1 << iota
	_
	_
	_
	CY // carry
	H  // halfCarry
	N  // negative
	Z  // zero
)

func (f cpuFlags) String() string {
	buf := []rune{'-', '-', '-', '-'}
	if f.has(Z) {
		buf[0] = 'Z'
	}
	if f.has(N) {
		buf[1] = 'N'
	}
	if f.has(H) {
		buf[2] = 'H'
	}
	if f.has(CY) {
		buf[3] = 'C'
	}

	return string(buf)
}

func (f *cpuFlags) set(flag cpuFlags, v bool) {
	*f &^= flag

	if v {
		*f |= flag
	}
}

func (f *cpuFlags) has(flag cpuFlags) bool {
	return *f&flag > 0
}

func (f *cpuFlags) flip(flag cpuFlags) {
	*f ^= flag
}

type cpu struct {
	A    uint8
	F    cpuFlags
	B, C uint8
	D, E uint8
	H, L uint8
	SP   uint16
	PC   uint16

	skipPCIncBug bool
	scheduleIME  bool
	IME          bool

	table   [256]op
	cbTable [256]op
}

func (c *cpu) init(pc uint16) {
	c.A = 0x01
	c.F = 0x00
	c.B = 0xFF
	c.C = 0x13
	c.D = 0x00
	c.E = 0xC1
	c.H = 0x84
	c.L = 0x03
	c.SP = 0xFFFE
	c.PC = pc

	c.table = [256]op{
		c.nop, c.ld_rr_d16, c.ld_irr_r, c.inc_rr, c.inc_r, c.dec_r, c.ld_r_d8, c.rlca, c.ld_ia16_sp, c.add_rr_rr, c.ld_r_irr, c.dec_rr, c.inc_r, c.dec_r, c.ld_r_d8, c.rrca,
		c.stop, c.ld_rr_d16, c.ld_irr_r, c.inc_rr, c.inc_r, c.dec_r, c.ld_r_d8, c.rla, c.jr_r8, c.add_rr_rr, c.ld_r_irr, c.dec_rr, c.inc_r, c.dec_r, c.ld_r_d8, c.rra,
		c.jr_NZ_r8, c.ld_rr_d16, c.ld_hlid_r, c.inc_rr, c.inc_r, c.dec_r, c.ld_r_d8, c.daa, c.jr_Z_r8, c.add_rr_rr, c.ld_r_hlid, c.dec_rr, c.inc_r, c.dec_r, c.ld_r_d8, c.cpl,
		c.jr_NC_r8, c.ld_sp_d16, c.ld_hlid_r, c.inc_sp, c.inc_irr, c.dec_irr, c.ld_irr_d8, c.scf, c.jr_r_r8, c.add_rr_sp, c.ld_r_hlid, c.dec_sp, c.inc_r, c.dec_r, c.ld_r_d8, c.ccf,
		c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_irr, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_irr, c.ld_r_r,
		c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_irr, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_irr, c.ld_r_r,
		c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_irr, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_irr, c.ld_r_r,
		c.ld_irr_r, c.ld_irr_r, c.ld_irr_r, c.ld_irr_r, c.ld_irr_r, c.ld_irr_r, c.halt, c.ld_irr_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_r, c.ld_r_irr, c.ld_r_r,
		c.add_r_r, c.add_r_r, c.add_r_r, c.add_r_r, c.add_r_r, c.add_r_r, c.add_r_irr, c.add_r_r, c.adc_r_r, c.adc_r_r, c.adc_r_r, c.adc_r_r, c.adc_r_r, c.adc_r_r, c.adc_r_irr, c.adc_r_r,
		c.sub_r, c.sub_r, c.sub_r, c.sub_r, c.sub_r, c.sub_r, c.sub_irr, c.sub_r, c.sbc_r_r, c.sbc_r_r, c.sbc_r_r, c.sbc_r_r, c.sbc_r_r, c.sbc_r_r, c.sbc_r_irr, c.sbc_r_r,
		c.and_r, c.and_r, c.and_r, c.and_r, c.and_r, c.and_r, c.and_irr, c.and_r, c.xor_r, c.xor_r, c.xor_r, c.xor_r, c.xor_r, c.xor_r, c.xor_irr, c.xor_r,
		c.or_r, c.or_r, c.or_r, c.or_r, c.or_r, c.or_r, c.or_irr, c.or_r, c.cp_r, c.cp_r, c.cp_r, c.cp_r, c.cp_r, c.cp_r, c.cp_irr, c.cp_r,
		c.ret_NZ, c.pop_rr, c.jp_NZ_a16, c.jp_a16, c.call_NZ_a16, c.push_rr, c.add_r_d8, c.rst, c.ret_Z, c.ret, c.jp_Z_a16, c.prefix, c.call_Z_a16, c.call_a16, c.adc_r_d8, c.rst,
		c.ret_NC, c.pop_rr, c.jp_NC_a16, c.illegal, c.call_NC_a16, c.push_rr, c.sub_d8, c.rst, c.ret_r, c.reti, c.jp_r_a16, c.illegal, c.call_r_a16, c.illegal, c.sbc_r_d8, c.rst,
		c.ldh_ia8_r, c.pop_rr, c.ld_ir_r, c.illegal, c.illegal, c.push_rr, c.and_d8, c.rst, c.add_sp_r8, c.jp_irr, c.ld_ia16_r, c.illegal, c.illegal, c.illegal, c.xor_d8, c.rst,
		c.ldh_r_ia8, c.pop_rr, c.ld_r_ir, c.di, c.illegal, c.push_rr, c.or_d8, c.rst, c.ld_rr_SP_r8, c.ld_sp_rr, c.ld_r_ia16, c.ei, c.illegal, c.illegal, c.cp_d8, c.rst,
	}

	c.cbTable = [256]op{
		c.rlc_r, c.rlc_r, c.rlc_r, c.rlc_r, c.rlc_r, c.rlc_r, c.rlc_irr, c.rlc_r, c.rrc_r, c.rrc_r, c.rrc_r, c.rrc_r, c.rrc_r, c.rrc_r, c.rrc_irr, c.rrc_r,
		c.rl_r, c.rl_r, c.rl_r, c.rl_r, c.rl_r, c.rl_r, c.rl_irr, c.rl_r, c.rr_r, c.rr_r, c.rr_r, c.rr_r, c.rr_r, c.rr_r, c.rr_irr, c.rr_r,
		c.sla_r, c.sla_r, c.sla_r, c.sla_r, c.sla_r, c.sla_r, c.sla_irr, c.sla_r, c.sra_r, c.sra_r, c.sra_r, c.sra_r, c.sra_r, c.sra_r, c.sra_irr, c.sra_r,
		c.swap_r, c.swap_r, c.swap_r, c.swap_r, c.swap_r, c.swap_r, c.swap_irr, c.swap_r, c.srl_r, c.srl_r, c.srl_r, c.srl_r, c.srl_r, c.srl_r, c.srl_irr, c.srl_r,
		c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_irr, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_irr, c.bit_r,
		c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_irr, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_irr, c.bit_r,
		c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_irr, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_irr, c.bit_r,
		c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_irr, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_r, c.bit_irr, c.bit_r,
		c.res_r, c.res_r, c.res_r, c.res_r, c.res_r, c.res_r, c.res_irr, c.res_r, c.res_r, c.res_r, c.res_r, c.res_r, c.res_r, c.res_r, c.res_irr, c.res_r,
		c.res_r, c.res_r, c.res_r, c.res_r, c.res_r, c.res_r, c.res_irr, c.res_r, c.res_r, c.res_r, c.res_r, c.res_r, c.res_r, c.res_r, c.res_irr, c.res_r,
		c.res_r, c.res_r, c.res_r, c.res_r, c.res_r, c.res_r, c.res_irr, c.res_r, c.res_r, c.res_r, c.res_r, c.res_r, c.res_r, c.res_r, c.res_irr, c.res_r,
		c.res_r, c.res_r, c.res_r, c.res_r, c.res_r, c.res_r, c.res_irr, c.res_r, c.res_r, c.res_r, c.res_r, c.res_r, c.res_r, c.res_r, c.res_irr, c.res_r,
		c.set_r, c.set_r, c.set_r, c.set_r, c.set_r, c.set_r, c.set_irr, c.set_r, c.set_r, c.set_r, c.set_r, c.set_r, c.set_r, c.set_r, c.set_irr, c.set_r,
		c.set_r, c.set_r, c.set_r, c.set_r, c.set_r, c.set_r, c.set_irr, c.set_r, c.set_r, c.set_r, c.set_r, c.set_r, c.set_r, c.set_r, c.set_irr, c.set_r,
		c.set_r, c.set_r, c.set_r, c.set_r, c.set_r, c.set_r, c.set_irr, c.set_r, c.set_r, c.set_r, c.set_r, c.set_r, c.set_r, c.set_r, c.set_irr, c.set_r,
		c.set_r, c.set_r, c.set_r, c.set_r, c.set_r, c.set_r, c.set_irr, c.set_r, c.set_r, c.set_r, c.set_r, c.set_r, c.set_r, c.set_r, c.set_irr, c.set_r,
	}
}

func (c *cpu) readFrom(gb *GameBoy, addr uint16) uint8 {
	v := gb.read(addr)
	gb.clockCompensate()
	return v
}

func (c *cpu) writeTo(gb *GameBoy, addr uint16, v uint8) {
	gb.write(addr, v)
	gb.clockCompensate()
}

func (c *cpu) clock(gb *GameBoy) {
	switch {
	// case gb.state&dma > 0:
	// 	if gb.interruptCtrl.raised(anyInterrupt) > 0 && c.IME {
	// 		gb.state |= interruptDispatch
	// 	}
	// 	gb.clockCompensate()
	case gb.state&stop > 0:
		if gb.interruptCtrl.raised(anyInterrupt) > 0 {
			gb.state = interruptDispatch
		}
	// TODO: missing some component clocks here I think

	case gb.state&run > 0:
		op := c.readFrom(gb, c.PC)

		if c.scheduleIME {
			c.scheduleIME = false
			c.IME = true
		}
		if c.skipPCIncBug {
			c.skipPCIncBug = false
		} else {
			c.PC++
		}
		c.table[op](op, gb)

		if gb.interruptCtrl.raised(anyInterrupt) > 0 && c.IME {
			gb.state = interruptDispatch
		}
	case gb.state&interruptDispatch > 0:
		// TODO: we might be missing a cycle here (if the 1st cycle is not the PC fetch)
		var vector uint16

		c.IME = false

		c.SP--
		c.writeTo(gb, c.SP, uint8(c.PC>>8))
		c.SP--
		c.writeTo(gb, c.SP, uint8(c.PC&0xFF))

		gb.clockCompensate()
		gb.clockCompensate()

		intType := gb.interruptCtrl.raised(anyInterrupt)
		if intType == 0 {
			panic("no ints available")
		}
		switch {
		case intType&vblankInterrupt > 0:
			vector = vectorVBlank
			gb.interruptCtrl.ack(vblankInterrupt)
		case intType&lcdStatInterrupt > 0:
			vector = vectorLCDc
			gb.interruptCtrl.ack(lcdStatInterrupt)
		case intType&timerInterrupt > 0:
			vector = vectorTimer
			gb.interruptCtrl.ack(timerInterrupt)
		case intType&serialInterrupt > 0:
			vector = vectorSerial
			gb.interruptCtrl.ack(serialInterrupt)
		case intType&joypadInterrupt > 0:
			vector = vectorHTL
			gb.interruptCtrl.ack(joypadInterrupt)
		}

		c.PC = vector
		// if gb.state&dma == 0 {
		gb.state = run
		// }

	case gb.state&halt > 0:
		switch c.IME {
		case true:
			if gb.interruptCtrl.raised(anyInterrupt) > 0 {
				gb.state = interruptDispatch
				c.clock(gb)
			} else {
				gb.clockCompensate()
			}

		case false:
			if gb.interruptCtrl.raised(anyInterrupt) > 0 {
				c.IME = false

				c.SP--
				c.writeTo(gb, c.SP, uint8(c.PC>>8))
				c.SP--
				c.writeTo(gb, c.SP, uint8(c.PC&0xFF))

				gb.clockCompensate()
				gb.clockCompensate()

				gb.state = run
			} else {
				gb.clockCompensate()
			}
		}
	}
}

func (c *cpu) cancelInterruptEffects() {
	// fmt.Println("??")
}

func (c *cpu) adc8(a, b uint8) uint8 {
	var carry uint8
	if c.F.has(CY) {
		carry = 1
	}

	c.F.set(Z, a+b+carry == 0)
	c.F.set(N, false)
	c.F.set(H, a&0xF+b&0xF+carry > 0xF)
	c.F.set(CY, uint16(a)+uint16(b)+uint16(carry) > 0xFF)

	return a + b + carry
}

func (c *cpu) sbc8(a, b uint8) uint8 {
	c.F.flip(CY)

	v := c.adc8(a, b^0xFF)

	c.F.set(N, true)
	c.F.flip(H)
	c.F.flip(CY)

	return v
}

func (c *cpu) add8(a, b uint8) uint8 {
	c.F.set(CY, false)
	return c.adc8(a, b)
}

func (c *cpu) sub8(a, b uint8) uint8 {
	c.F.set(CY, false)
	return c.sbc8(a, b)
}

// 0xCE ADC A,d8        2 8 0 Z 0 H C
func (c *cpu) adc_r_d8(opcode uint8, gb *GameBoy) {
	v := c.readFrom(gb, c.PC)
	c.PC++

	c.A = c.adc8(c.A, v)
}

// 0x8E ADC A,(HL)      1 8 0 Z 0 H C
func (c *cpu) adc_r_irr(opcode uint8, gb *GameBoy) {
	lo := uint16(c.L)
	hi := uint16(c.H)
	addr := hi<<8 | lo
	v := c.readFrom(gb, addr)

	c.A = c.adc8(c.A, v)
}

// 0x88 ADC A,B 1 4 0 Z 0 H C
// 0x89 ADC A,C 1 4 0 Z 0 H C
// 0x8A ADC A,D 1 4 0 Z 0 H C
// 0x8B ADC A,E 1 4 0 Z 0 H C
// 0x8C ADC A,H 1 4 0 Z 0 H C
// 0x8D ADC A,L 1 4 0 Z 0 H C
// 0x8F ADC A,A 1 4 0 Z 0 H C
func (c *cpu) adc_r_r(opcode uint8, gb *GameBoy) {
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

	c.A = c.adc8(c.A, v)
}

// 0xC6 ADD A,d8        2 8 0 Z 0 H C
func (c *cpu) add_r_d8(opcode uint8, gb *GameBoy) {
	v := c.readFrom(gb, c.PC)
	c.PC++

	c.A = c.add8(c.A, v)
}

// 0x86 ADD A,(HL)      1 8 0 Z 0 H C
func (c *cpu) add_r_irr(opcode uint8, gb *GameBoy) {
	lo := uint16(c.L)
	hi := uint16(c.H)

	addr := hi<<8 | lo
	v := c.readFrom(gb, addr)

	c.A = c.add8(c.A, v)
}

// 0x80 ADD A,B 1 4 0 Z 0 H C
// 0x81 ADD A,C 1 4 0 Z 0 H C
// 0x82 ADD A,D 1 4 0 Z 0 H C
// 0x83 ADD A,E 1 4 0 Z 0 H C
// 0x84 ADD A,H 1 4 0 Z 0 H C
// 0x85 ADD A,L 1 4 0 Z 0 H C
// 0x87 ADD A,A 1 4 0 Z 0 H C
func (c *cpu) add_r_r(opcode uint8, gb *GameBoy) {
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
func (c *cpu) add_rr_rr(opcode uint8, gb *GameBoy) {
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

	c.F.set(N, false)
	c.F.set(H, hl&0xFFF+v&0xFFF > 0xFFF)
	c.F.set(CY, uint32(hl)+uint32(v) > 0xFFFF)

	hl += v

	c.L = uint8(hl & 0xFF)
	c.H = uint8(hl >> 8)

	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?
}

// 0x39 ADD HL,SP       1 8 0 - 0 H C
func (c *cpu) add_rr_sp(opcode uint8, gb *GameBoy) {
	hl := uint16(c.H)<<8 | uint16(c.L)
	v := c.SP

	c.F.set(N, false)
	c.F.set(H, hl&0xFFF+v&0xFFF > 0xFFF)
	c.F.set(CY, uint32(hl)+uint32(v) > 0xFFFF)

	hl += v

	c.L = uint8(hl & 0xFF)
	c.H = uint8(hl >> 8)

	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?
}

// 0xE8 ADD SP,r8       2 16 0 0 0 H C
func (c *cpu) add_sp_r8(opcode uint8, gb *GameBoy) {
	r8 := c.readFrom(gb, c.PC)
	c.PC++

	spl := uint8(c.SP & 0xFF)
	sph := uint8(c.SP >> 8)

	var f cpuFlags // H and CY are only affected by lsb
	if r8&0x80 > 0 {
		r8 := r8 ^ 0xFF + 1
		spl = c.sub8(spl, r8)
		f = c.F
		f.flip(H | CY) // apparently we only want carries
		sph = c.sbc8(sph, 0)
	} else {
		spl = c.add8(spl, r8)
		f = c.F
		sph = c.adc8(sph, 0)
	}

	c.SP = uint16(sph)<<8 | uint16(spl)
	c.F = f // restore lsb flags
	c.F.set(Z, false)
	c.F.set(N, false)

	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?
	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?
}

// 0xE6 AND d8  2 8 0 Z 0 1 0
func (c *cpu) and_d8(opcode uint8, gb *GameBoy) {
	v := c.readFrom(gb, c.PC)
	c.PC++

	c.A &= v
	c.F.set(Z, c.A == 0)
	c.F.set(N, false)
	c.F.set(H, true)
	c.F.set(CY, false)
}

// 0xA6 AND (HL)        1 8 0 Z 0 1 0
func (c *cpu) and_irr(opcode uint8, gb *GameBoy) {
	hi := uint16(c.H)
	lo := uint16(c.L)

	addr := hi<<8 | lo
	v := c.readFrom(gb, addr)

	c.A &= v
	c.F.set(Z, c.A == 0)
	c.F.set(N, false)
	c.F.set(H, true)
	c.F.set(CY, false)
}

// 0xA0 AND B   1 4 0 Z 0 1 0
// 0xA1 AND C   1 4 0 Z 0 1 0
// 0xA2 AND D   1 4 0 Z 0 1 0
// 0xA3 AND E   1 4 0 Z 0 1 0
// 0xA4 AND H   1 4 0 Z 0 1 0
// 0xA5 AND L   1 4 0 Z 0 1 0
// 0xA7 AND A   1 4 0 Z 0 1 0
func (c *cpu) and_r(opcode uint8, gb *GameBoy) {
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
	c.F.set(Z, c.A == 0)
	c.F.set(N, false)
	c.F.set(H, true)
	c.F.set(CY, false)
}

// 0xD4 CALL NC,a16     3 24 12 - - - -
func (c *cpu) call_NC_a16(opcode uint8, gb *GameBoy) {
	lo := uint16(c.readFrom(gb, c.PC))
	c.PC++
	hi := uint16(c.readFrom(gb, c.PC))
	c.PC++

	if c.F.has(CY) {
		return
	}

	c.SP--
	c.writeTo(gb, c.SP, uint8(c.PC>>8))
	c.SP--
	c.writeTo(gb, c.SP, uint8(c.PC&0xFF))

	c.PC = hi<<8 | lo
	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?
}

// 0xC4 CALL NZ,a16     3 24 12 - - - -
func (c *cpu) call_NZ_a16(opcode uint8, gb *GameBoy) {
	lo := uint16(c.readFrom(gb, c.PC))
	c.PC++
	hi := uint16(c.readFrom(gb, c.PC))
	c.PC++

	if c.F.has(Z) {
		return
	}

	c.SP--
	c.writeTo(gb, c.SP, uint8(c.PC>>8))
	c.SP--
	c.writeTo(gb, c.SP, uint8(c.PC&0xFF))

	c.PC = hi<<8 | lo
	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?
}

// 0xCC CALL Z,a16      3 24 12 - - - -
func (c *cpu) call_Z_a16(opcode uint8, gb *GameBoy) {
	lo := uint16(c.readFrom(gb, c.PC))
	c.PC++
	hi := uint16(c.readFrom(gb, c.PC))
	c.PC++

	if !c.F.has(Z) {
		return
	}

	c.SP--
	c.writeTo(gb, c.SP, uint8(c.PC>>8))
	c.SP--
	c.writeTo(gb, c.SP, uint8(c.PC&0xFF))

	c.PC = hi<<8 | lo
	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?
}

// 0xCD CALL a16        3 24 0 - - - -
func (c *cpu) call_a16(opcode uint8, gb *GameBoy) {
	lo := uint16(c.readFrom(gb, c.PC))
	c.PC++
	hi := uint16(c.readFrom(gb, c.PC))
	c.PC++

	c.SP--
	c.writeTo(gb, c.SP, uint8(c.PC>>8))
	c.SP--
	c.writeTo(gb, c.SP, uint8(c.PC&0xFF))

	c.PC = hi<<8 | lo
	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?
}

// 0xDC CALL C,a16      3 24 12 - - - -
func (c *cpu) call_r_a16(opcode uint8, gb *GameBoy) {
	lo := uint16(c.readFrom(gb, c.PC))
	c.PC++
	hi := uint16(c.readFrom(gb, c.PC))
	c.PC++

	if !c.F.has(CY) {
		return
	}

	c.SP--
	c.writeTo(gb, c.SP, uint8(c.PC>>8))
	c.SP--
	c.writeTo(gb, c.SP, uint8(c.PC&0xFF))

	c.PC = hi<<8 | lo
	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?
}

// 0x3F CCF     1 4 0 - 0 0 C
func (c *cpu) ccf(opcode uint8, gb *GameBoy) {
	c.F.set(N, false)
	c.F.set(H, false)

	if c.F.has(CY) {
		c.F.set(CY, false)
	} else {
		c.F.set(CY, true)
	}
}

// 0xFE CP d8   2 8 0 Z 1 H C
func (c *cpu) cp_d8(opcode uint8, gb *GameBoy) {
	v := c.readFrom(gb, c.PC)
	c.PC++

	a := c.A
	c.sub8(c.A, v)
	c.A = a
}

// 0xBE CP (HL) 1 8 0 Z 1 H C
func (c *cpu) cp_irr(opcode uint8, gb *GameBoy) {
	lo := uint16(c.L)
	hi := uint16(c.H)
	addr := hi<<8 | lo
	v := c.readFrom(gb, addr)

	a := c.A
	c.sub8(c.A, v)
	c.A = a
}

// 0xB8 CP B    1 4 0 Z 1 H C
// 0xB9 CP C    1 4 0 Z 1 H C
// 0xBA CP D    1 4 0 Z 1 H C
// 0xBB CP E    1 4 0 Z 1 H C
// 0xBC CP H    1 4 0 Z 1 H C
// 0xBD CP L    1 4 0 Z 1 H C
// 0xBF CP A    1 4 0 Z 1 H C
func (c *cpu) cp_r(opcode uint8, gb *GameBoy) {
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

	a := c.A
	c.sub8(c.A, v)
	c.A = a
}

// 0x2F CPL     1 4 0 - 1 1 -
func (c *cpu) cpl(opcode uint8, gb *GameBoy) {
	c.A = c.A ^ 0xFF
	c.F.set(N, true)
	c.F.set(H, true)
}

// 0x27 DAA     1 4 0 Z - 0 C
func (c *cpu) daa(opcode uint8, gb *GameBoy) {
	if c.F.has(N) {
		if c.F.has(CY) {
			c.A -= 0x60
		}
		if c.F.has(H) {
			c.A -= 0x6
		}
	} else {
		if c.F.has(CY) || c.A > 0x99 {
			c.A += 0x60
			c.F.set(CY, true)
		}
		if c.F.has(H) || (c.A&0x0f) > 0x09 {
			c.A += 0x6
		}
	}
	c.F.set(Z, c.A == 0)
	c.F.set(H, false)
}

// 0x35 DEC (HL)        1 12 0 Z 1 H -
func (c *cpu) dec_irr(opcode uint8, gb *GameBoy) {
	lo := uint16(c.L)
	hi := uint16(c.H)
	addr := hi<<8 | lo

	v := c.readFrom(gb, addr)

	c.F.set(Z, v-1 == 0)
	c.F.set(N, true)
	c.F.set(H, v&0xF == 0)

	c.writeTo(gb, addr, v-1)
}

// 0x05 DEC B   1 4 0 Z 1 H -
// 0x0D DEC C   1 4 0 Z 1 H -
// 0x15 DEC D   1 4 0 Z 1 H -
// 0x1D DEC E   1 4 0 Z 1 H -
// 0x25 DEC H   1 4 0 Z 1 H -
// 0x2D DEC L   1 4 0 Z 1 H -
// 0x3D DEC A   1 4 0 Z 1 H -
func (c *cpu) dec_r(opcode uint8, gb *GameBoy) {
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
	c.F.set(Z, v-1 == 0)
	c.F.set(N, true)
	c.F.set(H, v&0xF == 0)
	*r = v - 1
}

// 0x0B DEC BC  1 8 0 - - - -
// 0x1B DEC DE  1 8 0 - - - -
// 0x2B DEC HL  1 8 0 - - - -
func (c *cpu) dec_rr(opcode uint8, gb *GameBoy) {
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

	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?
}

// 0x3B DEC SP  1 8 0 - - - -
func (c *cpu) dec_sp(opcode uint8, gb *GameBoy) {
	c.SP--
	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?
}

// 0xF3 DI      1 4 0 - - - -
func (c *cpu) di(opcode uint8, gb *GameBoy) {
	c.IME = false
	c.cancelInterruptEffects()
}

// 0xFB EI      1 4 0 - - - -
func (c *cpu) ei(opcode uint8, gb *GameBoy) {
	c.scheduleIME = true
}

// 0x76 HALT    1 4 0 - - - -
func (c *cpu) halt(opcode uint8, gb *GameBoy) {
	// IME set
	if c.IME {
		gb.state = halt
		return
	}

	// Some pending
	if gb.interruptCtrl.raised(anyInterrupt) > 0 {
		c.skipPCIncBug = true
		gb.state = run
		return
	} else {
		gb.state = halt
		return
	}
}

// 0x34 INC (HL)        1 12 0 Z 0 H -
func (c *cpu) inc_irr(opcode uint8, gb *GameBoy) {
	lo := uint16(c.L)
	hi := uint16(c.H)
	addr := hi<<8 | lo
	v := c.readFrom(gb, addr)

	c.F.set(Z, v+1 == 0)
	c.F.set(N, false)
	c.F.set(H, v&0xF == 0xF)

	c.writeTo(gb, addr, v+1)
}

// 0x04 INC B   1 4 0 Z 0 H -
// 0x0C INC C   1 4 0 Z 0 H -
// 0x14 INC D   1 4 0 Z 0 H -
// 0x1C INC E   1 4 0 Z 0 H -
// 0x24 INC H   1 4 0 Z 0 H -
// 0x2C INC L   1 4 0 Z 0 H -
// 0x3C INC A   1 4 0 Z 0 H -
func (c *cpu) inc_r(opcode uint8, gb *GameBoy) {
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
	c.F.set(Z, v+1 == 0)
	c.F.set(N, false)
	c.F.set(H, v&0xF == 0xF)
	*r = v + 1
}

// 0x03 INC BC  1 8 0 - - - -
// 0x13 INC DE  1 8 0 - - - -
// 0x23 INC HL  1 8 0 - - - -
func (c *cpu) inc_rr(opcode uint8, gb *GameBoy) {
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

	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?
}

// 0x33 INC SP  1 8 0 - - - -
func (c *cpu) inc_sp(opcode uint8, gb *GameBoy) {
	c.SP++
	c.readFrom(gb, c.SP) // TODO: what actually gets read (or written)?
}

// 0xD2 JP NC,a16       3 16 12 - - - -
func (c *cpu) jp_NC_a16(opcode uint8, gb *GameBoy) {
	lo := uint16(c.readFrom(gb, c.PC))
	c.PC++
	hi := uint16(c.readFrom(gb, c.PC))
	c.PC++

	if c.F.has(CY) {
		return
	}

	c.PC = hi<<8 | lo
	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?
}

// 0xC2 JP NZ,a16       3 16 12 - - - -
func (c *cpu) jp_NZ_a16(opcode uint8, gb *GameBoy) {
	lo := uint16(c.readFrom(gb, c.PC))
	c.PC++
	hi := uint16(c.readFrom(gb, c.PC))
	c.PC++

	if c.F.has(Z) {
		return
	}

	c.PC = hi<<8 | lo
	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?
}

// 0xCA JP Z,a16        3 16 12 - - - -
func (c *cpu) jp_Z_a16(opcode uint8, gb *GameBoy) {
	lo := uint16(c.readFrom(gb, c.PC))
	c.PC++
	hi := uint16(c.readFrom(gb, c.PC))
	c.PC++

	if !c.F.has(Z) {
		return
	}

	c.PC = hi<<8 | lo
	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)? }
}

// 0xC3 JP a16  3 16 0 - - - -
func (c *cpu) jp_a16(opcode uint8, gb *GameBoy) {
	lo := uint16(c.readFrom(gb, c.PC))
	c.PC++
	hi := uint16(c.readFrom(gb, c.PC))
	c.PC++

	c.PC = hi<<8 | lo
	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?
}

// 0xE9 JP (HL) 1 4 0 - - - -
func (c *cpu) jp_irr(opcode uint8, gb *GameBoy) {
	lo := uint16(c.L)
	hi := uint16(c.H)

	c.PC = hi<<8 | lo
}

// 0xDA JP C,a16        3 16 12 - - - -
func (c *cpu) jp_r_a16(opcode uint8, gb *GameBoy) {
	lo := uint16(c.readFrom(gb, c.PC))
	c.PC++
	hi := uint16(c.readFrom(gb, c.PC))
	c.PC++

	if !c.F.has(CY) {
		return
	}

	c.PC = hi<<8 | lo
	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)? }
}

// 0x30 JR NC,r8        2 12 8 - - - -
func (c *cpu) jr_NC_r8(opcode uint8, gb *GameBoy) {
	r8 := uint16(int8(c.readFrom(gb, c.PC)))
	c.PC++

	if c.F.has(CY) {
		return
	}

	c.PC += r8
	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)? }
}

// 0x20 JR NZ,r8        2 12 8 - - - -
func (c *cpu) jr_NZ_r8(opcode uint8, gb *GameBoy) {
	r8 := uint16(int8(c.readFrom(gb, c.PC)))
	c.PC++

	if c.F.has(Z) {
		return
	}

	c.PC += r8
	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)? }
}

// 0x28 JR Z,r8 2 12 8 - - - -
func (c *cpu) jr_Z_r8(opcode uint8, gb *GameBoy) {
	r8 := uint16(int8(c.readFrom(gb, c.PC)))
	c.PC++

	if !c.F.has(Z) {
		return
	}

	c.PC += uint16(r8)
	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)? } }
}

// 0x18 JR r8   2 12 0 - - - -
func (c *cpu) jr_r8(opcode uint8, gb *GameBoy) {
	r8 := uint16(int8(c.readFrom(gb, c.PC)))
	c.PC++

	c.PC += r8
	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)? }
}

// 0x38 JR C,r8 2 12 8 - - - -
func (c *cpu) jr_r_r8(opcode uint8, gb *GameBoy) {
	r8 := uint16(int8(c.readFrom(gb, c.PC)))
	c.PC++

	if !c.F.has(CY) {
		return
	}

	c.PC += uint16(r8)
	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)? } }
}

// 0x22 LD (HL+),A      1 8 0 - - - -
// 0x32 LD (HL-),A      1 8 0 - - - -
func (c *cpu) ld_hlid_r(opcode uint8, gb *GameBoy) {
	addr := uint16(c.H)<<8 | uint16(c.L)
	c.writeTo(gb, addr, c.A)

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
func (c *cpu) ld_ia16_r(opcode uint8, gb *GameBoy) {
	lo := uint16(c.readFrom(gb, c.PC))
	c.PC++
	hi := uint16(c.readFrom(gb, c.PC))
	c.PC++

	addr := hi<<8 | lo
	c.writeTo(gb, addr, c.A)
}

// 0x08 LD (a16),SP     3 20 0 - - - -
func (c *cpu) ld_ia16_sp(opcode uint8, gb *GameBoy) {
	lo := uint16(c.readFrom(gb, c.PC))
	c.PC++
	hi := uint16(c.readFrom(gb, c.PC))
	c.PC++

	addr := hi<<8 | lo
	c.writeTo(gb, addr, uint8(c.SP&0xFF))
	c.writeTo(gb, addr+1, uint8(c.SP>>8))
}

// 0xE2 LD (C),A        2 8 0 - - - -
func (c *cpu) ld_ir_r(opcode uint8, gb *GameBoy) {
	c.writeTo(gb, 0xFF00+uint16(c.C), c.A)
}

// 0x36 LD (HL),d8      2 12 0 - - - -
func (c *cpu) ld_irr_d8(opcode uint8, gb *GameBoy) {
	d8 := c.readFrom(gb, c.PC)
	c.PC++

	addr := uint16(c.H)<<8 | uint16(c.L)
	c.writeTo(gb, addr, d8)
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
func (c *cpu) ld_irr_r(opcode uint8, gb *GameBoy) {
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
	c.writeTo(gb, addr, v)
}

// 0x2A LD A,(HL+)      1 8 0 - - - -
// 0x3A LD A,(HL-)      1 8 0 - - - -
func (c *cpu) ld_r_hlid(opcode uint8, gb *GameBoy) {
	addr := uint16(c.H)<<8 | uint16(c.L)
	c.A = c.readFrom(gb, addr)

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
func (c *cpu) ld_r_d8(opcode uint8, gb *GameBoy) {
	d8 := c.readFrom(gb, c.PC)
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
func (c *cpu) ld_r_ia16(opcode uint8, gb *GameBoy) {
	lo := uint16(c.readFrom(gb, c.PC))
	c.PC++
	hi := uint16(c.readFrom(gb, c.PC))
	c.PC++

	addr := hi<<8 | lo
	c.A = c.readFrom(gb, addr)
}

// 0xF2 LD A,(C)        2 8 0 - - - -
func (c *cpu) ld_r_ir(opcode uint8, gb *GameBoy) {
	c.A = c.readFrom(gb, 0xFF00+uint16(c.C))
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
func (c *cpu) ld_r_irr(opcode uint8, gb *GameBoy) {
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
	*r = c.readFrom(gb, addr)
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
func (c *cpu) ld_r_r(opcode uint8, gb *GameBoy) {
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
func (c *cpu) ld_rr_SP_r8(opcode uint8, gb *GameBoy) {
	r8 := c.readFrom(gb, c.PC)
	c.PC++

	spl := uint8(c.SP & 0xFF)
	sph := uint8(c.SP >> 8)

	var f cpuFlags // H and CY are only affected by lsb
	if r8&0x80 > 0 {
		r8 := r8 ^ 0xFF + 1
		spl = c.sub8(spl, r8)
		f = c.F
		f.flip(H | CY) // apparently we only want carries
		sph = c.sbc8(sph, 0)
	} else {
		spl = c.add8(spl, r8)
		f = c.F
		sph = c.adc8(sph, 0)
	}

	c.F = f // restore lsb flags
	c.F.set(Z, false)
	c.F.set(N, false)

	c.L = spl
	c.H = sph

	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?
}

// 0x01 LD BC,d16       3 12 0 - - - -
// 0x11 LD DE,d16       3 12 0 - - - -
// 0x21 LD HL,d16       3 12 0 - - - -
func (c *cpu) ld_rr_d16(opcode uint8, gb *GameBoy) {
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

	*rrlo = c.readFrom(gb, c.PC)
	c.PC++
	*rrhi = c.readFrom(gb, c.PC)
	c.PC++
}

// 0x31 LD SP,d16       3 12 0 - - - -
func (c *cpu) ld_sp_d16(opcode uint8, gb *GameBoy) {
	lo := uint16(c.readFrom(gb, c.PC))
	c.PC++
	hi := uint16(c.readFrom(gb, c.PC))
	c.PC++

	c.SP = hi<<8 | lo
}

// 0xF9 LD SP,HL        1 8 0 - - - -
func (c *cpu) ld_sp_rr(opcode uint8, gb *GameBoy) {
	c.SP = uint16(c.H)<<8 | uint16(c.L)
	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?
}

// 0xE0 LDH (a8),A      2 12 0 - - - -
func (c *cpu) ldh_ia8_r(opcode uint8, gb *GameBoy) {
	a8 := c.readFrom(gb, c.PC)
	c.PC++
	c.writeTo(gb, 0xFF00|uint16(a8), c.A)
}

// 0xF0 LDH A,(a8)      2 12 0 - - - -
func (c *cpu) ldh_r_ia8(opcode uint8, gb *GameBoy) {
	a8 := c.readFrom(gb, c.PC)
	c.PC++
	c.A = c.readFrom(gb, 0xFF00|uint16(a8))
}

// 0x00 NOP     1 4 0 - - - -
func (c *cpu) nop(opcode uint8, gb *GameBoy) {}

// 0xF6 OR d8   2 8 0 Z 0 0 0
func (c *cpu) or_d8(opcode uint8, gb *GameBoy) {
	v := c.readFrom(gb, c.PC)
	c.PC++

	c.A |= v
	c.F.set(Z, c.A == 0)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, false)
}

// 0xB6 OR (HL) 1 8 0 Z 0 0 0
func (c *cpu) or_irr(opcode uint8, gb *GameBoy) {
	lo := uint16(c.L)
	hi := uint16(c.H)
	addr := hi<<8 | lo
	v := c.readFrom(gb, addr)

	c.A |= v
	c.F.set(Z, c.A == 0)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, false)
}

// 0xB0 OR B    1 4 0 Z 0 0 0
// 0xB1 OR C    1 4 0 Z 0 0 0
// 0xB2 OR D    1 4 0 Z 0 0 0
// 0xB3 OR E    1 4 0 Z 0 0 0
// 0xB4 OR H    1 4 0 Z 0 0 0
// 0xB5 OR L    1 4 0 Z 0 0 0
// 0xB7 OR A    1 4 0 Z 0 0 0
func (c *cpu) or_r(opcode uint8, gb *GameBoy) {
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
	c.F.set(Z, c.A == 0)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, false)
}

// 0xC1 POP BC  1 12 0 - - - -
// 0xD1 POP DE  1 12 0 - - - -
// 0xE1 POP HL  1 12 0 - - - -
// 0xF1 POP AF  1 12 0 Z N H C
func (c *cpu) pop_rr(opcode uint8, gb *GameBoy) {
	var rrhi, rrlo *uint8
	var isf bool
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
		isf = true
	}

	*rrlo = c.readFrom(gb, c.SP)
	if isf {
		*rrlo &= 0xF0
	}
	c.SP++
	*rrhi = c.readFrom(gb, c.SP)
	c.SP++
}

// 0xCB PREFIX CB       1 4 0 - - - -
func (c *cpu) prefix(opcode uint8, gb *GameBoy) {
	op := c.readFrom(gb, c.PC)
	c.PC++
	c.cbTable[op](op, gb)
}

// 0xC5 PUSH BC 1 16 0 - - - -
// 0xD5 PUSH DE 1 16 0 - - - -
// 0xE5 PUSH HL 1 16 0 - - - -
// 0xF5 PUSH AF 1 16 0 - - - -
func (c *cpu) push_rr(opcode uint8, gb *GameBoy) {
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

	c.readFrom(gb, c.SP) // TODO: what actually gets read (or written)?

	c.SP--
	c.writeTo(gb, c.SP, *rrhi)
	c.SP--
	c.writeTo(gb, c.SP, *rrlo)
}

// 0xC9 RET     1 16 0 - - - -
func (c *cpu) ret(opcode uint8, gb *GameBoy) {
	lo := uint16(c.readFrom(gb, c.SP))
	c.SP++
	hi := uint16(c.readFrom(gb, c.SP))
	c.SP++

	c.PC = hi<<8 | lo

	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?
}

// 0xD0 RET NC  1 20 8 - - - -
func (c *cpu) ret_NC(opcode uint8, gb *GameBoy) {
	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?

	if c.F.has(CY) {
		return
	}

	lo := uint16(c.readFrom(gb, c.SP))
	c.SP++
	hi := uint16(c.readFrom(gb, c.SP))
	c.SP++

	c.PC = hi<<8 | lo

	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?
}

// 0xC0 RET NZ  1 20 8 - - - -
func (c *cpu) ret_NZ(opcode uint8, gb *GameBoy) {
	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?

	if c.F.has(Z) {
		return
	}

	lo := uint16(c.readFrom(gb, c.SP))
	c.SP++
	hi := uint16(c.readFrom(gb, c.SP))
	c.SP++

	c.PC = hi<<8 | lo

	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?
}

// 0xC8 RET Z   1 20 8 - - - -
func (c *cpu) ret_Z(opcode uint8, gb *GameBoy) {
	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?

	if !c.F.has(Z) {
		return
	}

	lo := uint16(c.readFrom(gb, c.SP))
	c.SP++
	hi := uint16(c.readFrom(gb, c.SP))
	c.SP++

	c.PC = hi<<8 | lo

	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)? }
}

// 0xD8 RET C   1 20 8 - - - -
func (c *cpu) ret_r(opcode uint8, gb *GameBoy) {
	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?

	if !c.F.has(CY) {
		return
	}

	lo := uint16(c.readFrom(gb, c.SP))
	c.SP++
	hi := uint16(c.readFrom(gb, c.SP))
	c.SP++

	c.PC = hi<<8 | lo

	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?
}

// 0xD9 RETI    1 16 0 - - - -
func (c *cpu) reti(opcode uint8, gb *GameBoy) {
	lo := uint16(c.readFrom(gb, c.SP))
	c.SP++
	hi := uint16(c.readFrom(gb, c.SP))
	c.SP++

	c.PC = hi<<8 | lo
	c.IME = true

	c.readFrom(gb, c.PC) // TODO: what actually gets read (or written)?
}

// 0x17 RLA     1 4 0 0 0 0 C
func (c *cpu) rla(opcode uint8, gb *GameBoy) {
	var carryIn uint8
	if c.F.has(CY) {
		carryIn = 1
	}
	carryOut := c.A & 0x80
	c.A = c.A<<1 | carryIn

	c.F.set(Z, false)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, carryOut > 0)
}

// 0x07 RLCA    1 4 0 0 0 0 C
func (c *cpu) rlca(opcode uint8, gb *GameBoy) {
	carryOut := c.A & 0x80 >> 7
	c.A = c.A<<1 | carryOut

	c.F.set(Z, false)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, carryOut > 0)
}

// 0x1F RRA     1 4 0 0 0 0 C
func (c *cpu) rra(opcode uint8, gb *GameBoy) {
	var carryIn uint8
	if c.F.has(CY) {
		carryIn = 1 << 7
	}

	carryOut := c.A & 0x1
	c.A = c.A>>1 | carryIn

	c.F.set(Z, false)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, carryOut > 0)
}

// 0x0F RRCA    1 4 0 0 0 0 C
func (c *cpu) rrca(opcode uint8, gb *GameBoy) {
	carryOut := c.A & 0x1 << 7
	c.A = c.A>>1 | carryOut

	c.F.set(Z, false)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, carryOut > 0)
}

// 0xC7 RST 00H 1 16 0 - - - -
// 0xCF RST 08H 1 16 0 - - - -
// 0xD7 RST 10H 1 16 0 - - - -
// 0xDF RST 18H 1 16 0 - - - -
// 0xE7 RST 20H 1 16 0 - - - -
// 0xEF RST 28H 1 16 0 - - - -
// 0xF7 RST 30H 1 16 0 - - - -
// 0xFF RST 38H 1 16 0 - - - -
func (c *cpu) rst(opcode uint8, gb *GameBoy) {
	var addr uint16

	switch opcode {
	case 0xC7:
		addr = 0x00
	case 0xCF:
		addr = 0x08
	case 0xD7:
		addr = 0x10
	case 0xDF:
		addr = 0x18
	case 0xE7:
		addr = 0x20
	case 0xEF:
		addr = 0x28
	case 0xF7:
		addr = 0x30
	case 0xFF:
		addr = 0x38
	}

	_ = c.readFrom(gb, c.SP)

	c.SP--
	c.writeTo(gb, c.SP, uint8(c.PC>>8))
	c.SP--
	c.writeTo(gb, c.SP, uint8(c.PC&0xFF))

	c.PC = uint16(addr)
}

func (c *cpu) illegal(opcode uint8, gb *GameBoy) { panic("illegal") }

// 0xDE SBC A,d8        2 8 0 Z 1 H C
func (c *cpu) sbc_r_d8(opcode uint8, gb *GameBoy) {
	v := c.readFrom(gb, c.PC)
	c.PC++

	c.A = c.sbc8(c.A, v)
}

// 0x9E SBC A,(HL)      1 8 0 Z 1 H C
func (c *cpu) sbc_r_irr(opcode uint8, gb *GameBoy) {
	lo := uint16(c.L)
	hi := uint16(c.H)
	addr := hi<<8 | lo
	v := c.readFrom(gb, addr)

	c.A = c.sbc8(c.A, v)
}

// 0x98 SBC A,B 1 4 0 Z 1 H C
// 0x99 SBC A,C 1 4 0 Z 1 H C
// 0x9A SBC A,D 1 4 0 Z 1 H C
// 0x9B SBC A,E 1 4 0 Z 1 H C
// 0x9C SBC A,H 1 4 0 Z 1 H C
// 0x9D SBC A,L 1 4 0 Z 1 H C
// 0x9F SBC A,A 1 4 0 Z 1 H C
func (c *cpu) sbc_r_r(opcode uint8, gb *GameBoy) {
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

	c.A = c.sbc8(c.A, v)
}

// 0x37 SCF     1 4 0 - 0 0 1
func (c *cpu) scf(opcode uint8, gb *GameBoy) {
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, true)
}

// 0x10 STOP 0  2 4 0 - - - -
func (c *cpu) stop(opcode uint8, gb *GameBoy) {
	gb.read(c.PC)
	c.PC++
	gb.state = stop
}

// 0xD6 SUB d8  2 8 0 Z 1 H C
func (c *cpu) sub_d8(opcode uint8, gb *GameBoy) {
	v := c.readFrom(gb, c.PC)
	c.PC++

	c.A = c.sub8(c.A, v)
}

// 0x96 SUB (HL)        1 8 0 Z 1 H C
func (c *cpu) sub_irr(opcode uint8, gb *GameBoy) {
	lo := uint16(c.L)
	hi := uint16(c.H)
	addr := hi<<8 | lo
	v := c.readFrom(gb, addr)

	c.A = c.sub8(c.A, v)
}

// 0x90 SUB B   1 4 0 Z 1 H C
// 0x91 SUB C   1 4 0 Z 1 H C
// 0x92 SUB D   1 4 0 Z 1 H C
// 0x93 SUB E   1 4 0 Z 1 H C
// 0x94 SUB H   1 4 0 Z 1 H C
// 0x95 SUB L   1 4 0 Z 1 H C
// 0x97 SUB A   1 4 0 Z 1 H C
func (c *cpu) sub_r(opcode uint8, gb *GameBoy) {
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

	c.A = c.sub8(c.A, v)
}

// 0xEE XOR d8  2 8 0 Z 0 0 0
func (c *cpu) xor_d8(opcode uint8, gb *GameBoy) {
	v := c.readFrom(gb, c.PC)
	c.PC++

	c.A = c.A ^ v
	c.F.set(Z, c.A == 0)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, false)
}

// 0xAE XOR (HL)        1 8 0 Z 0 0 0
func (c *cpu) xor_irr(opcode uint8, gb *GameBoy) {
	lo := uint16(c.L)
	hi := uint16(c.H)
	addr := hi<<8 | lo
	v := c.readFrom(gb, addr)

	c.A = c.A ^ v
	c.F.set(Z, c.A == 0)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, false)
}

// 0xA8 XOR B   1 4 0 Z 0 0 0
// 0xA9 XOR C   1 4 0 Z 0 0 0
// 0xAA XOR D   1 4 0 Z 0 0 0
// 0xAB XOR E   1 4 0 Z 0 0 0
// 0xAC XOR H   1 4 0 Z 0 0 0
// 0xAD XOR L   1 4 0 Z 0 0 0
// 0xAF XOR A   1 4 0 Z 0 0 0
func (c *cpu) xor_r(opcode uint8, gb *GameBoy) {
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
	c.F.set(Z, c.A == 0)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, false)
}

// 0x46 BIT 0,(HL)      2 12 0 Z 0 1 -
// 0x4E BIT 1,(HL)      2 12 0 Z 0 1 -
// 0x56 BIT 2,(HL)      2 12 0 Z 0 1 -
// 0x5E BIT 3,(HL)      2 12 0 Z 0 1 -
// 0x66 BIT 4,(HL)      2 12 0 Z 0 1 -
// 0x6E BIT 5,(HL)      2 12 0 Z 0 1 -
// 0x76 BIT 6,(HL)      2 12 0 Z 0 1 -
// 0x7E BIT 7,(HL)      2 12 0 Z 0 1 -
func (c *cpu) bit_irr(opcode uint8, gb *GameBoy) {
	addr := uint16(c.H)<<8 | uint16(c.L)
	v := c.readFrom(gb, addr)
	var mask uint8

	switch opcode {
	case 0x46:
		mask = 1 << 0
	case 0x4E:
		mask = 1 << 1
	case 0x56:
		mask = 1 << 2
	case 0x5E:
		mask = 1 << 3
	case 0x66:
		mask = 1 << 4
	case 0x6E:
		mask = 1 << 5
	case 0x76:
		mask = 1 << 6
	case 0x7E:
		mask = 1 << 7
	}

	c.F.set(Z, v&mask == 0)
	c.F.set(N, false)
	c.F.set(H, true)
}

// 0x40 BIT 0,B 2 8 0 Z 0 1 -
// 0x41 BIT 0,C 2 8 0 Z 0 1 -
// 0x42 BIT 0,D 2 8 0 Z 0 1 -
// 0x43 BIT 0,E 2 8 0 Z 0 1 -
// 0x44 BIT 0,H 2 8 0 Z 0 1 -
// 0x45 BIT 0,L 2 8 0 Z 0 1 -
// 0x47 BIT 0,A 2 8 0 Z 0 1 -
// 0x48 BIT 1,B 2 8 0 Z 0 1 -
// 0x49 BIT 1,C 2 8 0 Z 0 1 -
// 0x4A BIT 1,D 2 8 0 Z 0 1 -
// 0x4B BIT 1,E 2 8 0 Z 0 1 -
// 0x4C BIT 1,H 2 8 0 Z 0 1 -
// 0x4D BIT 1,L 2 8 0 Z 0 1 -
// 0x4F BIT 1,A 2 8 0 Z 0 1 -
// 0x50 BIT 2,B 2 8 0 Z 0 1 -
// 0x51 BIT 2,C 2 8 0 Z 0 1 -
// 0x52 BIT 2,D 2 8 0 Z 0 1 -
// 0x53 BIT 2,E 2 8 0 Z 0 1 -
// 0x54 BIT 2,H 2 8 0 Z 0 1 -
// 0x55 BIT 2,L 2 8 0 Z 0 1 -
// 0x57 BIT 2,A 2 8 0 Z 0 1 -
// 0x58 BIT 3,B 2 8 0 Z 0 1 -
// 0x59 BIT 3,C 2 8 0 Z 0 1 -
// 0x5A BIT 3,D 2 8 0 Z 0 1 -
// 0x5B BIT 3,E 2 8 0 Z 0 1 -
// 0x5C BIT 3,H 2 8 0 Z 0 1 -
// 0x5D BIT 3,L 2 8 0 Z 0 1 -
// 0x5F BIT 3,A 2 8 0 Z 0 1 -
// 0x60 BIT 4,B 2 8 0 Z 0 1 -
// 0x61 BIT 4,C 2 8 0 Z 0 1 -
// 0x62 BIT 4,D 2 8 0 Z 0 1 -
// 0x63 BIT 4,E 2 8 0 Z 0 1 -
// 0x64 BIT 4,H 2 8 0 Z 0 1 -
// 0x65 BIT 4,L 2 8 0 Z 0 1 -
// 0x67 BIT 4,A 2 8 0 Z 0 1 -
// 0x68 BIT 5,B 2 8 0 Z 0 1 -
// 0x69 BIT 5,C 2 8 0 Z 0 1 -
// 0x6A BIT 5,D 2 8 0 Z 0 1 -
// 0x6B BIT 5,E 2 8 0 Z 0 1 -
// 0x6C BIT 5,H 2 8 0 Z 0 1 -
// 0x6D BIT 5,L 2 8 0 Z 0 1 -
// 0x6F BIT 5,A 2 8 0 Z 0 1 -
// 0x70 BIT 6,B 2 8 0 Z 0 1 -
// 0x71 BIT 6,C 2 8 0 Z 0 1 -
// 0x72 BIT 6,D 2 8 0 Z 0 1 -
// 0x73 BIT 6,E 2 8 0 Z 0 1 -
// 0x74 BIT 6,H 2 8 0 Z 0 1 -
// 0x75 BIT 6,L 2 8 0 Z 0 1 -
// 0x77 BIT 6,A 2 8 0 Z 0 1 -
// 0x78 BIT 7,B 2 8 0 Z 0 1 -
// 0x79 BIT 7,C 2 8 0 Z 0 1 -
// 0x7A BIT 7,D 2 8 0 Z 0 1 -
// 0x7B BIT 7,E 2 8 0 Z 0 1 -
// 0x7C BIT 7,H 2 8 0 Z 0 1 -
// 0x7D BIT 7,L 2 8 0 Z 0 1 -
// 0x7F BIT 7,A 2 8 0 Z 0 1 -
func (c *cpu) bit_r(opcode uint8, gb *GameBoy) {
	var v uint8
	var mask uint8

	// ew
	switch opcode {
	case 0x40:
		mask = 1 << 0
		v = c.B
	case 0x41:
		mask = 1 << 0
		v = c.C
	case 0x42:
		mask = 1 << 0
		v = c.D
	case 0x43:
		mask = 1 << 0
		v = c.E
	case 0x44:
		mask = 1 << 0
		v = c.H
	case 0x45:
		mask = 1 << 0
		v = c.L
	case 0x47:
		mask = 1 << 0
		v = c.A
	case 0x48:
		mask = 1 << 1
		v = c.B
	case 0x49:
		mask = 1 << 1
		v = c.C
	case 0x4A:
		mask = 1 << 1
		v = c.D
	case 0x4B:
		mask = 1 << 1
		v = c.E
	case 0x4C:
		mask = 1 << 1
		v = c.H
	case 0x4D:
		mask = 1 << 1
		v = c.L
	case 0x4F:
		mask = 1 << 1
		v = c.A
	case 0x50:
		mask = 1 << 2
		v = c.B
	case 0x51:
		mask = 1 << 2
		v = c.C
	case 0x52:
		mask = 1 << 2
		v = c.D
	case 0x53:
		mask = 1 << 2
		v = c.E
	case 0x54:
		mask = 1 << 2
		v = c.H
	case 0x55:
		mask = 1 << 2
		v = c.L
	case 0x57:
		mask = 1 << 2
		v = c.A
	case 0x58:
		mask = 1 << 3
		v = c.B
	case 0x59:
		mask = 1 << 3
		v = c.C
	case 0x5A:
		mask = 1 << 3
		v = c.D
	case 0x5B:
		mask = 1 << 3
		v = c.E
	case 0x5C:
		mask = 1 << 3
		v = c.H
	case 0x5D:
		mask = 1 << 3
		v = c.L
	case 0x5F:
		mask = 1 << 3
		v = c.A
	case 0x60:
		mask = 1 << 4
		v = c.B
	case 0x61:
		mask = 1 << 4
		v = c.C
	case 0x62:
		mask = 1 << 4
		v = c.D
	case 0x63:
		mask = 1 << 4
		v = c.E
	case 0x64:
		mask = 1 << 4
		v = c.H
	case 0x65:
		mask = 1 << 4
		v = c.L
	case 0x67:
		mask = 1 << 4
		v = c.A
	case 0x68:
		mask = 1 << 5
		v = c.B
	case 0x69:
		mask = 1 << 5
		v = c.C
	case 0x6A:
		mask = 1 << 5
		v = c.D
	case 0x6B:
		mask = 1 << 5
		v = c.E
	case 0x6C:
		mask = 1 << 5
		v = c.H
	case 0x6D:
		mask = 1 << 5
		v = c.L
	case 0x6F:
		mask = 1 << 5
		v = c.A
	case 0x70:
		mask = 1 << 6
		v = c.B
	case 0x71:
		mask = 1 << 6
		v = c.C
	case 0x72:
		mask = 1 << 6
		v = c.D
	case 0x73:
		mask = 1 << 6
		v = c.E
	case 0x74:
		mask = 1 << 6
		v = c.H
	case 0x75:
		mask = 1 << 6
		v = c.L
	case 0x77:
		mask = 1 << 6
		v = c.A
	case 0x78:
		mask = 1 << 7
		v = c.B
	case 0x79:
		mask = 1 << 7
		v = c.C
	case 0x7A:
		mask = 1 << 7
		v = c.D
	case 0x7B:
		mask = 1 << 7
		v = c.E
	case 0x7C:
		mask = 1 << 7
		v = c.H
	case 0x7D:
		mask = 1 << 7
		v = c.L
	case 0x7F:
		mask = 1 << 7
		v = c.A
	}

	c.F.set(Z, v&mask == 0)
	c.F.set(N, false)
	c.F.set(H, true)
}

// 0x86 RES 0,(HL)      2 16 0 - - - -
// 0x8E RES 1,(HL)      2 16 0 - - - -
// 0x96 RES 2,(HL)      2 16 0 - - - -
// 0x9E RES 3,(HL)      2 16 0 - - - -
// 0xA6 RES 4,(HL)      2 16 0 - - - -
// 0xAE RES 5,(HL)      2 16 0 - - - -
// 0xB6 RES 6,(HL)      2 16 0 - - - -
// 0xBE RES 7,(HL)      2 16 0 - - - -
func (c *cpu) res_irr(opcode uint8, gb *GameBoy) {
	addr := uint16(c.H)<<8 | uint16(c.L)
	v := c.readFrom(gb, addr)
	var mask uint8

	switch opcode {
	case 0x86:
		mask = 1 << 0
	case 0x8E:
		mask = 1 << 1
	case 0x96:
		mask = 1 << 2
	case 0x9E:
		mask = 1 << 3
	case 0xA6:
		mask = 1 << 4
	case 0xAE:
		mask = 1 << 5
	case 0xB6:
		mask = 1 << 6
	case 0xBE:
		mask = 1 << 7
	}

	v &^= mask
	c.writeTo(gb, addr, v)
}

// 0x80 RES 0,B 2 8 0 - - - -
// 0x81 RES 0,C 2 8 0 - - - -
// 0x82 RES 0,D 2 8 0 - - - -
// 0x83 RES 0,E 2 8 0 - - - -
// 0x84 RES 0,H 2 8 0 - - - -
// 0x85 RES 0,L 2 8 0 - - - -
// 0x87 RES 0,A 2 8 0 - - - -
// 0x88 RES 1,B 2 8 0 - - - -
// 0x89 RES 1,C 2 8 0 - - - -
// 0x8A RES 1,D 2 8 0 - - - -
// 0x8B RES 1,E 2 8 0 - - - -
// 0x8C RES 1,H 2 8 0 - - - -
// 0x8D RES 1,L 2 8 0 - - - -
// 0x8F RES 1,A 2 8 0 - - - -
// 0x90 RES 2,B 2 8 0 - - - -
// 0x91 RES 2,C 2 8 0 - - - -
// 0x92 RES 2,D 2 8 0 - - - -
// 0x93 RES 2,E 2 8 0 - - - -
// 0x94 RES 2,H 2 8 0 - - - -
// 0x95 RES 2,L 2 8 0 - - - -
// 0x97 RES 2,A 2 8 0 - - - -
// 0x98 RES 3,B 2 8 0 - - - -
// 0x99 RES 3,C 2 8 0 - - - -
// 0x9A RES 3,D 2 8 0 - - - -
// 0x9B RES 3,E 2 8 0 - - - -
// 0x9C RES 3,H 2 8 0 - - - -
// 0x9D RES 3,L 2 8 0 - - - -
// 0x9F RES 3,A 2 8 0 - - - -
// 0xA0 RES 4,B 2 8 0 - - - -
// 0xA1 RES 4,C 2 8 0 - - - -
// 0xA2 RES 4,D 2 8 0 - - - -
// 0xA3 RES 4,E 2 8 0 - - - -
// 0xA4 RES 4,H 2 8 0 - - - -
// 0xA5 RES 4,L 2 8 0 - - - -
// 0xA7 RES 4,A 2 8 0 - - - -
// 0xA8 RES 5,B 2 8 0 - - - -
// 0xA9 RES 5,C 2 8 0 - - - -
// 0xAA RES 5,D 2 8 0 - - - -
// 0xAB RES 5,E 2 8 0 - - - -
// 0xAC RES 5,H 2 8 0 - - - -
// 0xAD RES 5,L 2 8 0 - - - -
// 0xAF RES 5,A 2 8 0 - - - -
// 0xB0 RES 6,B 2 8 0 - - - -
// 0xB1 RES 6,C 2 8 0 - - - -
// 0xB2 RES 6,D 2 8 0 - - - -
// 0xB3 RES 6,E 2 8 0 - - - -
// 0xB4 RES 6,H 2 8 0 - - - -
// 0xB5 RES 6,L 2 8 0 - - - -
// 0xB7 RES 6,A 2 8 0 - - - -
// 0xB8 RES 7,B 2 8 0 - - - -
// 0xB9 RES 7,C 2 8 0 - - - -
// 0xBA RES 7,D 2 8 0 - - - -
// 0xBB RES 7,E 2 8 0 - - - -
// 0xBC RES 7,H 2 8 0 - - - -
// 0xBD RES 7,L 2 8 0 - - - -
// 0xBF RES 7,A 2 8 0 - - - -
func (c *cpu) res_r(opcode uint8, gb *GameBoy) {
	var r *uint8
	var mask uint8

	// ew again
	switch opcode {
	case 0x80:
		mask = 1 << 0
		r = &c.B
	case 0x81:
		mask = 1 << 0
		r = &c.C
	case 0x82:
		mask = 1 << 0
		r = &c.D
	case 0x83:
		mask = 1 << 0
		r = &c.E
	case 0x84:
		mask = 1 << 0
		r = &c.H
	case 0x85:
		mask = 1 << 0
		r = &c.L
	case 0x87:
		mask = 1 << 0
		r = &c.A
	case 0x88:
		mask = 1 << 1
		r = &c.B
	case 0x89:
		mask = 1 << 1
		r = &c.C
	case 0x8A:
		mask = 1 << 1
		r = &c.D
	case 0x8B:
		mask = 1 << 1
		r = &c.E
	case 0x8C:
		mask = 1 << 1
		r = &c.H
	case 0x8D:
		mask = 1 << 1
		r = &c.L
	case 0x8F:
		mask = 1 << 1
		r = &c.A
	case 0x90:
		mask = 1 << 2
		r = &c.B
	case 0x91:
		mask = 1 << 2
		r = &c.C
	case 0x92:
		mask = 1 << 2
		r = &c.D
	case 0x93:
		mask = 1 << 2
		r = &c.E
	case 0x94:
		mask = 1 << 2
		r = &c.H
	case 0x95:
		mask = 1 << 2
		r = &c.L
	case 0x97:
		mask = 1 << 2
		r = &c.A
	case 0x98:
		mask = 1 << 3
		r = &c.B
	case 0x99:
		mask = 1 << 3
		r = &c.C
	case 0x9A:
		mask = 1 << 3
		r = &c.D
	case 0x9B:
		mask = 1 << 3
		r = &c.E
	case 0x9C:
		mask = 1 << 3
		r = &c.H
	case 0x9D:
		mask = 1 << 3
		r = &c.L
	case 0x9F:
		mask = 1 << 3
		r = &c.A
	case 0xA0:
		mask = 1 << 4
		r = &c.B
	case 0xA1:
		mask = 1 << 4
		r = &c.C
	case 0xA2:
		mask = 1 << 4
		r = &c.D
	case 0xA3:
		mask = 1 << 4
		r = &c.E
	case 0xA4:
		mask = 1 << 4
		r = &c.H
	case 0xA5:
		mask = 1 << 4
		r = &c.L
	case 0xA7:
		mask = 1 << 4
		r = &c.A
	case 0xA8:
		mask = 1 << 5
		r = &c.B
	case 0xA9:
		mask = 1 << 5
		r = &c.C
	case 0xAA:
		mask = 1 << 5
		r = &c.D
	case 0xAB:
		mask = 1 << 5
		r = &c.E
	case 0xAC:
		mask = 1 << 5
		r = &c.H
	case 0xAD:
		mask = 1 << 5
		r = &c.L
	case 0xAF:
		mask = 1 << 5
		r = &c.A
	case 0xB0:
		mask = 1 << 6
		r = &c.B
	case 0xB1:
		mask = 1 << 6
		r = &c.C
	case 0xB2:
		mask = 1 << 6
		r = &c.D
	case 0xB3:
		mask = 1 << 6
		r = &c.E
	case 0xB4:
		mask = 1 << 6
		r = &c.H
	case 0xB5:
		mask = 1 << 6
		r = &c.L
	case 0xB7:
		mask = 1 << 6
		r = &c.A
	case 0xB8:
		mask = 1 << 7
		r = &c.B
	case 0xB9:
		mask = 1 << 7
		r = &c.C
	case 0xBA:
		mask = 1 << 7
		r = &c.D
	case 0xBB:
		mask = 1 << 7
		r = &c.E
	case 0xBC:
		mask = 1 << 7
		r = &c.H
	case 0xBD:
		mask = 1 << 7
		r = &c.L
	case 0xBF:
		mask = 1 << 7
		r = &c.A
	}

	*r &^= mask
}

// 0x16 RL (HL) 2 16 0 Z 0 0 C
func (c *cpu) rl_irr(opcode uint8, gb *GameBoy) {
	addr := uint16(c.H)<<8 | uint16(c.L)
	v := c.readFrom(gb, addr)

	carryOut := v & 0x80
	var carryIn uint8
	if c.F.has(CY) {
		carryIn = 1
	}

	v = v<<1 | carryIn

	c.F.set(Z, v == 0)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, carryOut > 0)

	c.writeTo(gb, addr, v)
}

// 0x10 RL B    2 8 0 Z 0 0 C
// 0x11 RL C    2 8 0 Z 0 0 C
// 0x12 RL D    2 8 0 Z 0 0 C
// 0x13 RL E    2 8 0 Z 0 0 C
// 0x14 RL H    2 8 0 Z 0 0 C
// 0x15 RL L    2 8 0 Z 0 0 C
// 0x17 RL A    2 8 0 Z 0 0 C
func (c *cpu) rl_r(opcode uint8, gb *GameBoy) {
	var r *uint8

	switch opcode {
	case 0x10:
		r = &c.B
	case 0x11:
		r = &c.C
	case 0x12:
		r = &c.D
	case 0x13:
		r = &c.E
	case 0x14:
		r = &c.H
	case 0x15:
		r = &c.L
	case 0x17:
		r = &c.A
	}

	carryOut := *r & 0x80
	var carryIn uint8
	if c.F.has(CY) {
		carryIn = 1
	}

	*r = *r<<1 | carryIn

	c.F.set(Z, *r == 0)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, carryOut > 0)
}

// 0x06 RLC (HL)        2 16 0 Z 0 0 C
func (c *cpu) rlc_irr(opcode uint8, gb *GameBoy) {
	addr := uint16(c.H)<<8 | uint16(c.L)
	v := c.readFrom(gb, addr)

	carryOut := v & 0x80 >> 7
	v = v<<1 | carryOut

	c.F.set(Z, v == 0)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, carryOut > 0)
	c.writeTo(gb, addr, v)
}

// 0x00 RLC B   2 8 0 Z 0 0 C
// 0x01 RLC C   2 8 0 Z 0 0 C
// 0x02 RLC D   2 8 0 Z 0 0 C
// 0x03 RLC E   2 8 0 Z 0 0 C
// 0x04 RLC H   2 8 0 Z 0 0 C
// 0x05 RLC L   2 8 0 Z 0 0 C
// 0x07 RLC A   2 8 0 Z 0 0 C
func (c *cpu) rlc_r(opcode uint8, gb *GameBoy) {
	var r *uint8

	switch opcode {
	case 0x00:
		r = &c.B
	case 0x01:
		r = &c.C
	case 0x02:
		r = &c.D
	case 0x03:
		r = &c.E
	case 0x04:
		r = &c.H
	case 0x05:
		r = &c.L
	case 0x07:
		r = &c.A
	}

	carryOut := *r & 0x80 >> 7
	*r = *r<<1 | carryOut

	c.F.set(Z, *r == 0)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, carryOut > 0)
}

// 0x1E RR (HL) 2 16 0 Z 0 0 C
func (c *cpu) rr_irr(opcode uint8, gb *GameBoy) {
	addr := uint16(c.H)<<8 | uint16(c.L)
	v := c.readFrom(gb, addr)

	carryOut := v & 0x01
	var carryIn uint8
	if c.F.has(CY) {
		carryIn = 1 << 7
	}

	v = v>>1 | carryIn
	c.F.set(Z, v == 0)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, carryOut > 0)

	c.writeTo(gb, addr, v)
}

// 0x18 RR B    2 8 0 Z 0 0 C
// 0x19 RR C    2 8 0 Z 0 0 C
// 0x1A RR D    2 8 0 Z 0 0 C
// 0x1B RR E    2 8 0 Z 0 0 C
// 0x1C RR H    2 8 0 Z 0 0 C
// 0x1D RR L    2 8 0 Z 0 0 C
// 0x1F RR A    2 8 0 Z 0 0 C
func (c *cpu) rr_r(opcode uint8, gb *GameBoy) {
	var r *uint8

	switch opcode {
	case 0x18:
		r = &c.B
	case 0x19:
		r = &c.C
	case 0x1A:
		r = &c.D
	case 0x1B:
		r = &c.E
	case 0x1C:
		r = &c.H
	case 0x1D:
		r = &c.L
	case 0x1F:
		r = &c.A
	}

	carryOut := *r & 0x01
	var carryIn uint8
	if c.F.has(CY) {
		carryIn = 1 << 7
	}

	*r = *r>>1 | carryIn
	c.F.set(Z, *r == 0)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, carryOut > 0)
}

// 0x0E RRC (HL)        2 16 0 Z 0 0 C
func (c *cpu) rrc_irr(opcode uint8, gb *GameBoy) {
	addr := uint16(c.H)<<8 | uint16(c.L)
	v := c.readFrom(gb, addr)

	carryOut := v & 0x01 << 7
	v = v>>1 | carryOut

	c.F.set(Z, v == 0)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, carryOut > 0)

	c.writeTo(gb, addr, v)
}

// 0x08 RRC B   2 8 0 Z 0 0 C
// 0x09 RRC C   2 8 0 Z 0 0 C
// 0x0A RRC D   2 8 0 Z 0 0 C
// 0x0B RRC E   2 8 0 Z 0 0 C
// 0x0C RRC H   2 8 0 Z 0 0 C
// 0x0D RRC L   2 8 0 Z 0 0 C
// 0x0F RRC A   2 8 0 Z 0 0 C
func (c *cpu) rrc_r(opcode uint8, gb *GameBoy) {
	var r *uint8

	switch opcode {
	case 0x08:
		r = &c.B
	case 0x09:
		r = &c.C
	case 0x0A:
		r = &c.D
	case 0x0B:
		r = &c.E
	case 0x0C:
		r = &c.H
	case 0x0D:
		r = &c.L
	case 0x0F:
		r = &c.A
	}

	carryOut := *r & 0x01 << 7
	*r = *r>>1 | carryOut

	c.F.set(Z, *r == 0)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, carryOut > 0)
}

// 0xC6 SET 0,(HL)      2 16 0 - - - -
// 0xCE SET 1,(HL)      2 16 0 - - - -
// 0xD6 SET 2,(HL)      2 16 0 - - - -
// 0xDE SET 3,(HL)      2 16 0 - - - -
// 0xE6 SET 4,(HL)      2 16 0 - - - -
// 0xEE SET 5,(HL)      2 16 0 - - - -
// 0xF6 SET 6,(HL)      2 16 0 - - - -
// 0xFE SET 7,(HL)      2 16 0 - - - -
func (c *cpu) set_irr(opcode uint8, gb *GameBoy) {
	addr := uint16(c.H)<<8 | uint16(c.L)
	v := c.readFrom(gb, addr)

	var mask uint8
	switch opcode {
	case 0xC6:
		mask = 1 << 0
	case 0xCE:
		mask = 1 << 1
	case 0xD6:
		mask = 1 << 2
	case 0xDE:
		mask = 1 << 3
	case 0xE6:
		mask = 1 << 4
	case 0xEE:
		mask = 1 << 5
	case 0xF6:
		mask = 1 << 6
	case 0xFE:
		mask = 1 << 7
	}

	v |= mask

	c.writeTo(gb, addr, v)
}

// 0xC0 SET 0,B 2 8 0 - - - -
// 0xC1 SET 0,C 2 8 0 - - - -
// 0xC2 SET 0,D 2 8 0 - - - -
// 0xC3 SET 0,E 2 8 0 - - - -
// 0xC4 SET 0,H 2 8 0 - - - -
// 0xC5 SET 0,L 2 8 0 - - - -
// 0xC7 SET 0,A 2 8 0 - - - -
// 0xC8 SET 1,B 2 8 0 - - - -
// 0xC9 SET 1,C 2 8 0 - - - -
// 0xCA SET 1,D 2 8 0 - - - -
// 0xCB SET 1,E 2 8 0 - - - -
// 0xCC SET 1,H 2 8 0 - - - -
// 0xCD SET 1,L 2 8 0 - - - -
// 0xCF SET 1,A 2 8 0 - - - -
// 0xD0 SET 2,B 2 8 0 - - - -
// 0xD1 SET 2,C 2 8 0 - - - -
// 0xD2 SET 2,D 2 8 0 - - - -
// 0xD3 SET 2,E 2 8 0 - - - -
// 0xD4 SET 2,H 2 8 0 - - - -
// 0xD5 SET 2,L 2 8 0 - - - -
// 0xD7 SET 2,A 2 8 0 - - - -
// 0xD8 SET 3,B 2 8 0 - - - -
// 0xD9 SET 3,C 2 8 0 - - - -
// 0xDA SET 3,D 2 8 0 - - - -
// 0xDB SET 3,E 2 8 0 - - - -
// 0xDC SET 3,H 2 8 0 - - - -
// 0xDD SET 3,L 2 8 0 - - - -
// 0xDF SET 3,A 2 8 0 - - - -
// 0xE0 SET 4,B 2 8 0 - - - -
// 0xE1 SET 4,C 2 8 0 - - - -
// 0xE2 SET 4,D 2 8 0 - - - -
// 0xE3 SET 4,E 2 8 0 - - - -
// 0xE4 SET 4,H 2 8 0 - - - -
// 0xE5 SET 4,L 2 8 0 - - - -
// 0xE7 SET 4,A 2 8 0 - - - -
// 0xE8 SET 5,B 2 8 0 - - - -
// 0xE9 SET 5,C 2 8 0 - - - -
// 0xEA SET 5,D 2 8 0 - - - -
// 0xEB SET 5,E 2 8 0 - - - -
// 0xEC SET 5,H 2 8 0 - - - -
// 0xED SET 5,L 2 8 0 - - - -
// 0xEF SET 5,A 2 8 0 - - - -
// 0xF0 SET 6,B 2 8 0 - - - -
// 0xF1 SET 6,C 2 8 0 - - - -
// 0xF2 SET 6,D 2 8 0 - - - -
// 0xF3 SET 6,E 2 8 0 - - - -
// 0xF4 SET 6,H 2 8 0 - - - -
// 0xF5 SET 6,L 2 8 0 - - - -
// 0xF7 SET 6,A 2 8 0 - - - -
// 0xF8 SET 7,B 2 8 0 - - - -
// 0xF9 SET 7,C 2 8 0 - - - -
// 0xFA SET 7,D 2 8 0 - - - -
// 0xFB SET 7,E 2 8 0 - - - -
// 0xFC SET 7,H 2 8 0 - - - -
// 0xFD SET 7,L 2 8 0 - - - -
// 0xFF SET 7,A 2 8 0 - - - -
func (c *cpu) set_r(opcode uint8, gb *GameBoy) {
	var r *uint8
	var mask uint8

	// ew again
	switch opcode {
	case 0xC0:
		mask = 1 << 0
		r = &c.B
	case 0xC1:
		mask = 1 << 0
		r = &c.C
	case 0xC2:
		mask = 1 << 0
		r = &c.D
	case 0xC3:
		mask = 1 << 0
		r = &c.E
	case 0xC4:
		mask = 1 << 0
		r = &c.H
	case 0xC5:
		mask = 1 << 0
		r = &c.L
	case 0xC7:
		mask = 1 << 0
		r = &c.A
	case 0xC8:
		mask = 1 << 1
		r = &c.B
	case 0xC9:
		mask = 1 << 1
		r = &c.C
	case 0xCA:
		mask = 1 << 1
		r = &c.D
	case 0xCB:
		mask = 1 << 1
		r = &c.E
	case 0xCC:
		mask = 1 << 1
		r = &c.H
	case 0xCD:
		mask = 1 << 1
		r = &c.L
	case 0xCF:
		mask = 1 << 1
		r = &c.A
	case 0xD0:
		mask = 1 << 2
		r = &c.B
	case 0xD1:
		mask = 1 << 2
		r = &c.C
	case 0xD2:
		mask = 1 << 2
		r = &c.D
	case 0xD3:
		mask = 1 << 2
		r = &c.E
	case 0xD4:
		mask = 1 << 2
		r = &c.H
	case 0xD5:
		mask = 1 << 2
		r = &c.L
	case 0xD7:
		mask = 1 << 2
		r = &c.A
	case 0xD8:
		mask = 1 << 3
		r = &c.B
	case 0xD9:
		mask = 1 << 3
		r = &c.C
	case 0xDA:
		mask = 1 << 3
		r = &c.D
	case 0xDB:
		mask = 1 << 3
		r = &c.E
	case 0xDC:
		mask = 1 << 3
		r = &c.H
	case 0xDD:
		mask = 1 << 3
		r = &c.L
	case 0xDF:
		mask = 1 << 3
		r = &c.A
	case 0xE0:
		mask = 1 << 4
		r = &c.B
	case 0xE1:
		mask = 1 << 4
		r = &c.C
	case 0xE2:
		mask = 1 << 4
		r = &c.D
	case 0xE3:
		mask = 1 << 4
		r = &c.E
	case 0xE4:
		mask = 1 << 4
		r = &c.H
	case 0xE5:
		mask = 1 << 4
		r = &c.L
	case 0xE7:
		mask = 1 << 4
		r = &c.A
	case 0xE8:
		mask = 1 << 5
		r = &c.B
	case 0xE9:
		mask = 1 << 5
		r = &c.C
	case 0xEA:
		mask = 1 << 5
		r = &c.D
	case 0xEB:
		mask = 1 << 5
		r = &c.E
	case 0xEC:
		mask = 1 << 5
		r = &c.H
	case 0xED:
		mask = 1 << 5
		r = &c.L
	case 0xEF:
		mask = 1 << 5
		r = &c.A
	case 0xF0:
		mask = 1 << 6
		r = &c.B
	case 0xF1:
		mask = 1 << 6
		r = &c.C
	case 0xF2:
		mask = 1 << 6
		r = &c.D
	case 0xF3:
		mask = 1 << 6
		r = &c.E
	case 0xF4:
		mask = 1 << 6
		r = &c.H
	case 0xF5:
		mask = 1 << 6
		r = &c.L
	case 0xF7:
		mask = 1 << 6
		r = &c.A
	case 0xF8:
		mask = 1 << 7
		r = &c.B
	case 0xF9:
		mask = 1 << 7
		r = &c.C
	case 0xFA:
		mask = 1 << 7
		r = &c.D
	case 0xFB:
		mask = 1 << 7
		r = &c.E
	case 0xFC:
		mask = 1 << 7
		r = &c.H
	case 0xFD:
		mask = 1 << 7
		r = &c.L
	case 0xFF:
		mask = 1 << 7
		r = &c.A
	}

	*r |= mask
}

// 0x26 SLA (HL)        2 16 0 Z 0 0 C
func (c *cpu) sla_irr(opcode uint8, gb *GameBoy) {
	addr := uint16(c.H)<<8 | uint16(c.L)
	v := c.readFrom(gb, addr)

	carryOut := v & 0x80
	v <<= 1

	c.F.set(Z, v == 0)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, carryOut > 0)

	c.writeTo(gb, addr, v)
}

// 0x20 SLA B   2 8 0 Z 0 0 C
// 0x21 SLA C   2 8 0 Z 0 0 C
// 0x22 SLA D   2 8 0 Z 0 0 C
// 0x23 SLA E   2 8 0 Z 0 0 C
// 0x24 SLA H   2 8 0 Z 0 0 C
// 0x25 SLA L   2 8 0 Z 0 0 C
// 0x27 SLA A   2 8 0 Z 0 0 C
func (c *cpu) sla_r(opcode uint8, gb *GameBoy) {
	var r *uint8

	switch opcode {
	case 0x20:
		r = &c.B
	case 0x21:
		r = &c.C
	case 0x22:
		r = &c.D
	case 0x23:
		r = &c.E
	case 0x24:
		r = &c.H
	case 0x25:
		r = &c.L
	case 0x27:
		r = &c.A
	}

	carryOut := *r & 0x80
	*r <<= 1

	c.F.set(Z, *r == 0)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, carryOut > 0)
}

// 0x2E SRA (HL)        2 16 0 Z 0 0 0
func (c *cpu) sra_irr(opcode uint8, gb *GameBoy) {
	addr := uint16(c.H)<<8 | uint16(c.L)
	v := c.readFrom(gb, addr)

	carryIn := v & 0x80
	carryOut := v & 0x01
	v = v>>1 | carryIn

	c.F.set(Z, v == 0)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, carryOut > 0) // rednex says this is set, I guess we'll see

	c.writeTo(gb, addr, v)
}

// 0x28 SRA B   2 8 0 Z 0 0 0
// 0x29 SRA C   2 8 0 Z 0 0 0
// 0x2A SRA D   2 8 0 Z 0 0 0
// 0x2B SRA E   2 8 0 Z 0 0 0
// 0x2C SRA H   2 8 0 Z 0 0 0
// 0x2D SRA L   2 8 0 Z 0 0 0
// 0x2F SRA A   2 8 0 Z 0 0 0
func (c *cpu) sra_r(opcode uint8, gb *GameBoy) {
	var r *uint8

	switch opcode {
	case 0x28:
		r = &c.B
	case 0x29:
		r = &c.C
	case 0x2A:
		r = &c.D
	case 0x2B:
		r = &c.E
	case 0x2C:
		r = &c.H
	case 0x2D:
		r = &c.L
	case 0x2F:
		r = &c.A
	}

	carryIn := *r & 0x80
	carryOut := *r & 0x01
	*r = *r>>1 | carryIn

	c.F.set(Z, *r == 0)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, carryOut > 0) // rednex says this is set, I guess we'll see
}

// 0x3E SRL (HL)        2 16 0 Z 0 0 C
func (c *cpu) srl_irr(opcode uint8, gb *GameBoy) {
	addr := uint16(c.H)<<8 | uint16(c.L)
	v := c.readFrom(gb, addr)

	carryOut := v & 0x01
	v >>= 1

	c.F.set(Z, v == 0)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, carryOut > 0)
	c.writeTo(gb, addr, v)
}

// 0x38 SRL B   2 8 0 Z 0 0 C
// 0x39 SRL C   2 8 0 Z 0 0 C
// 0x3A SRL D   2 8 0 Z 0 0 C
// 0x3B SRL E   2 8 0 Z 0 0 C
// 0x3C SRL H   2 8 0 Z 0 0 C
// 0x3D SRL L   2 8 0 Z 0 0 C
// 0x3F SRL A   2 8 0 Z 0 0 C
func (c *cpu) srl_r(opcode uint8, gb *GameBoy) {
	var r *uint8

	switch opcode {
	case 0x38:
		r = &c.B
	case 0x39:
		r = &c.C
	case 0x3A:
		r = &c.D
	case 0x3B:
		r = &c.E
	case 0x3C:
		r = &c.H
	case 0x3D:
		r = &c.L
	case 0x3F:
		r = &c.A
	}

	carryOut := *r & 0x01
	*r >>= 1

	c.F.set(Z, *r == 0)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, carryOut > 0)
}

// 0x36 SWAP (HL)       2 16 0 Z 0 0 0
func (c *cpu) swap_irr(opcode uint8, gb *GameBoy) {
	addr := uint16(c.H)<<8 | uint16(c.L)
	v := c.readFrom(gb, addr)

	lo := v & 0xF0 >> 4
	hi := v & 0x0F << 4
	v = hi | lo

	c.F.set(Z, v == 0)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, false)

	c.writeTo(gb, addr, v)
}

// 0x30 SWAP B  2 8 0 Z 0 0 0
// 0x31 SWAP C  2 8 0 Z 0 0 0
// 0x32 SWAP D  2 8 0 Z 0 0 0
// 0x33 SWAP E  2 8 0 Z 0 0 0
// 0x34 SWAP H  2 8 0 Z 0 0 0
// 0x35 SWAP L  2 8 0 Z 0 0 0
// 0x37 SWAP A  2 8 0 Z 0 0 0
func (c *cpu) swap_r(opcode uint8, gb *GameBoy) {
	var r *uint8

	switch opcode {
	case 0x30:
		r = &c.B
	case 0x31:
		r = &c.C
	case 0x32:
		r = &c.D
	case 0x33:
		r = &c.E
	case 0x34:
		r = &c.H
	case 0x35:
		r = &c.L
	case 0x37:
		r = &c.A
	}

	lo := *r & 0xF0 >> 4
	hi := *r & 0x0F << 4
	*r = hi | lo

	c.F.set(Z, *r == 0)
	c.F.set(N, false)
	c.F.set(H, false)
	c.F.set(CY, false)
}
