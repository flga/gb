package gb

import (
	"reflect"
	"testing"
)

type testBus map[uint16]byte

func (b testBus) read(addr uint16) uint8 {
	return b[addr]
}

func (b testBus) write(addr uint16, v uint8) {
	b[addr] = v
}

type cpuData struct {
	A    uint8
	F    cpuFlags
	B, C uint8
	D, E uint8
	H, L uint8
	SP   uint16
	PC   uint16
}

type cpuSingleTest struct {
	code    []byte
	pre     cpuData
	bus     testBus
	want    cpuData
	wantbus testBus
}

// tests shamelessly stolen from mooney-gb because I'm lazy

func TestCpuOps0x00_0x0F(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"NOP": {
			code: []byte{0x00},
			pre:  cpuData{},
			bus:  testBus{},
			want: cpuData{
				PC: 0x0001,
			},
			wantbus: testBus{},
		},
		// 0x01: TODO
		"LD (BC), A": {
			code: []byte{0x02, 0xed, 0x00},
			pre: cpuData{
				A: 0x42,
				B: 0x80,
				C: 0x02,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x42,
				B:  0x80,
				C:  0x02,
				PC: 0x0001,
			},
			wantbus: testBus{
				0x8002: 0x42,
			},
		},
		// 0x03: TODO
		"INC B": {
			code: []byte{0x04},
			pre: cpuData{
				B: 0x42,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x43,
				PC: 0x0001,
			},
			wantbus: testBus{},
		},
		"INC B zero": {
			code: []byte{0x04},
			pre: cpuData{
				B: 0xff,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x00,
				F:  fz | fh,
				PC: 0x0001,
			},
			wantbus: testBus{},
		},
		"INC B half carry": {
			code: []byte{0x04},
			pre: cpuData{
				B: 0x0f,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x10,
				F:  fh,
				PC: 0x0001,
			},
			wantbus: testBus{},
		},
	}

	for mnemonic, tt := range tests {
		func(mnemonic string, tt cpuSingleTest) {
			t.Run(mnemonic, func(t *testing.T) {
				c := &cpu{
					A:  tt.pre.A,
					F:  tt.pre.F,
					B:  tt.pre.B,
					C:  tt.pre.C,
					D:  tt.pre.D,
					E:  tt.pre.E,
					H:  tt.pre.H,
					L:  tt.pre.L,
					SP: tt.pre.SP,
					PC: tt.pre.PC,
				}
				c.init()

				// "load rom"
				for i, op := range tt.code {
					tt.bus.write(uint16(i), op)
				}

				c.executeInst(tt.bus)

				// "unload rom"
				for i := range tt.code {
					delete(tt.bus, uint16(i))
				}

				got := cpuData{
					A:  c.A,
					F:  c.F,
					B:  c.B,
					C:  c.C,
					D:  c.D,
					E:  c.E,
					H:  c.H,
					L:  c.L,
					SP: c.SP,
					PC: c.PC,
				}

				if got != tt.want {
					t.Errorf("cpu.executeInst() = %+v, want %+v", got, tt.want)
				}

				if !reflect.DeepEqual(tt.bus, tt.wantbus) {
					t.Errorf("cpu.executeInst() bus = %v, want %v", tt.bus, tt.wantbus)
				}
			})

		}(mnemonic, tt)
	}
}
