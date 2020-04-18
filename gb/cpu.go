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
	c // carry
	h // halfCarry
	n // negative
	z // zero
)

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
	F    uint8
	B, C uint8
	D, E uint8
	H, L uint8
	SP   uint16
	PC   uint16

	address uint16
	data    uint8

	status  cpuStatus
	opstack []func()
}

func (c *cpu) clock() {
	if len(c.opstack) == 0 {
		return
	}

	head := len(c.opstack) - 1
	c.opstack[head]()
	c.opstack = c.opstack[:head]
}

func (c *cpu) adc()
func (c *cpu) add()
func (c *cpu) and()
func (c *cpu) bit()
func (c *cpu) call()
func (c *cpu) ccf()
func (c *cpu) cp()
func (c *cpu) cpl()
func (c *cpu) daa()
func (c *cpu) dec()
func (c *cpu) di()
func (c *cpu) ei()
func (c *cpu) halt()
func (c *cpu) inc()
func (c *cpu) jp()
func (c *cpu) jr()
func (c *cpu) ld()
func (c *cpu) ldh()
func (c *cpu) nop()
func (c *cpu) or()
func (c *cpu) pop()
func (c *cpu) prefix()
func (c *cpu) push()
func (c *cpu) res()
func (c *cpu) ret()
func (c *cpu) reti()
func (c *cpu) rl()
func (c *cpu) rla()
func (c *cpu) rlc()
func (c *cpu) rlca()
func (c *cpu) rr()
func (c *cpu) rra()
func (c *cpu) rrc()
func (c *cpu) rrca()
func (c *cpu) rst()
func (c *cpu) sbc()
func (c *cpu) scf()
func (c *cpu) set()
func (c *cpu) sla()
func (c *cpu) sra()
func (c *cpu) srl()
func (c *cpu) stop()
func (c *cpu) sub()
func (c *cpu) swap()
func (c *cpu) xor()
