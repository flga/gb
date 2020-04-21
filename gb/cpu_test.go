package gb

import (
	"fmt"
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

func (cd cpuData) String() string {
	return fmt.Sprintf("A[0x%02X] F[%s B0x%02X] C[0x%02X] D[0x%02X] E[0x%02X] H[0x%02X] L[0x%02X] SP[0x%04X] PC[0x%04X]",
		cd.A,
		cd.F,
		cd.B,
		cd.C,
		cd.D,
		cd.E,
		cd.H,
		cd.L,
		cd.SP,
		cd.PC,
	)
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
			pre: cpuData{
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		// 0x01: TODO
		"LD (BC), A": {
			code: []byte{0x02},
			pre: cpuData{
				A:  0x42,
				B:  0x00,
				C:  0x02,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x42,
				B:  0x00,
				C:  0x02,
				PC: 0x8001,
			},
			wantbus: testBus{
				0x0002: 0x42,
			},
		},
		// 0x03: TODO
		"INC B": {
			code: []byte{0x04},
			pre: cpuData{
				B:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x43,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"INC B zero": {
			code: []byte{0x04},
			pre: cpuData{
				B:  0xff,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x00,
				F:  fz | fh,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"INC B half carry": {
			code: []byte{0x04},
			pre: cpuData{
				B:  0x0f,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x10,
				F:  fh,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC B": {
			code: []byte{0x05},
			pre: cpuData{
				B:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x41,
				F:  fn,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC B zero": {
			code: []byte{0x05},
			pre: cpuData{
				B:  0x01,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x00,
				F:  fn | fz,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC B half carry": {
			code: []byte{0x05},
			pre: cpuData{
				B:  0x00,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0xFF,
				F:  fn | fh,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD B, n": {
			code: []byte{0x06, 0x42},
			pre: cpuData{
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x42,
				PC: 0x8002,
			},
			wantbus: testBus{},
		},
		"RLCA": {
			code: []byte{0x07},
			pre: cpuData{
				A:  0x77,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0xEE,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"RLCA carry": {
			code: []byte{0x07},
			pre: cpuData{
				A:  0xF7,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0xEF,
				F:  fc,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD A, (BC)": {
			code: []byte{0x0A},
			pre: cpuData{
				B:  0x00,
				C:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0x42},
			want: cpuData{
				A:  0x42,
				B:  0x00,
				C:  0x02,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x42},
		},
		"INC C": {
			code: []byte{0x0C},
			pre: cpuData{
				C:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				C:  0x43,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"INC C zero": {
			code: []byte{0x0C},
			pre: cpuData{
				C:  0xFF,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				C:  0x00,
				F:  fz | fh,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"INC C half carry": {
			code: []byte{0x0C},
			pre: cpuData{
				C:  0x0F,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				C:  0x10,
				F:  fh,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC C": {
			code: []byte{0x0D},
			pre: cpuData{
				C:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				C:  0x41,
				F:  fn,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC C zero": {
			code: []byte{0x0D},
			pre: cpuData{
				C:  0x01,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				C:  0x00,
				F:  fn | fz,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC C half carry": {
			code: []byte{0x0D},
			pre: cpuData{
				C:  0x00,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				C:  0xFF,
				F:  fn | fh,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD C, n": {
			code: []byte{0x0E, 0x42},
			pre: cpuData{
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				C:  0x42,
				PC: 0x8002,
			},
			wantbus: testBus{},
		},
		"RRCA": {
			code: []byte{0x0F},
			pre: cpuData{
				A:  0xEE,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x77,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"RRCA carry": {
			code: []byte{0x0F},
			pre: cpuData{
				A:  0xEF,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0xF7,
				F:  fc,
				PC: 0x8001,
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
					tt.bus.write(tt.pre.PC+uint16(i), op)
				}

				c.executeInst(tt.bus)

				// "unload rom"
				for i := range tt.code {
					delete(tt.bus, tt.pre.PC+uint16(i))
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
					t.Errorf("cpu.executeInst()")
					t.Logf("_pre %s", tt.pre)
					t.Logf("_got %s", got)
					t.Logf("want %s", tt.want)
				}

				if !reflect.DeepEqual(tt.bus, tt.wantbus) {
					t.Errorf("cpu.executeInst() bus = %v, want %v", tt.bus, tt.wantbus)
				}
			})

		}(mnemonic, tt)
	}
}
