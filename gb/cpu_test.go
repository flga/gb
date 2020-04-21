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
	return fmt.Sprintf("A[0x%02X] F[%s] B[0x%02X] C[0x%02X] D[0x%02X] E[0x%02X] H[0x%02X] L[0x%02X] SP[0x%04X] PC[0x%04X]",
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
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0x10_0x1F(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"LD (DE), A": {
			code: []byte{0x12},
			pre: cpuData{
				A:  0x42,
				D:  0x00,
				E:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x02: 0x00},
			want: cpuData{
				A:  0x42,
				D:  0x00,
				E:  0x02,
				PC: 0x8001,
			},
			wantbus: testBus{0x02: 0x42},
		},
		"INC D": {
			code: []byte{0x14},
			pre: cpuData{
				D:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				D:  0x43,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"INC D zero": {
			code: []byte{0x14},
			pre: cpuData{
				D:  0xFF,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				D:  0x00,
				F:  fz | fh,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"INC D half carry": {
			code: []byte{0x14},
			pre: cpuData{
				D:  0x0F,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				D:  0x10,
				F:  fh,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC D": {
			code: []byte{0x15},
			pre: cpuData{
				D:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				D:  0x41,
				F:  fn,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC D zero": {
			code: []byte{0x15},
			pre: cpuData{
				D:  0x01,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				D:  0x00,
				F:  fn | fz,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC D half carry": {
			code: []byte{0x15},
			pre: cpuData{
				D:  0x00,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				D:  0xFF,
				F:  fn | fh,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD D, n": {
			code: []byte{0x16, 0x42},
			pre: cpuData{
				D:  0x00,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				D:  0x42,
				PC: 0x8002,
			},
			wantbus: testBus{},
		},
		"RLA": {
			code: []byte{0x17},
			pre: cpuData{
				A:  0x55,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0xAA,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"RLA carry": {
			code: []byte{0x17},
			pre: cpuData{
				A:  0xAA,
				F:  fc,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x55,
				F:  fc,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD A, (DE)": {
			code: []byte{0x1A},
			pre: cpuData{
				A:  0x00,
				D:  0x00,
				E:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0x42},
			want: cpuData{
				A:  0x42,
				D:  0x00,
				E:  0x02,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x42},
		},
		"INC E": {
			code: []byte{0x1C},
			pre: cpuData{
				E:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				E:  0x43,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"INC E zero": {
			code: []byte{0x1C},
			pre: cpuData{
				E:  0xFF,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				E:  0x00,
				F:  fz | fh,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"INC E half carry": {
			code: []byte{0x1C},
			pre: cpuData{
				E:  0x0F,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				E:  0x10,
				F:  fh,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC E": {
			code: []byte{0x1D},
			pre: cpuData{
				E:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				E:  0x41,
				F:  fn,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC E zero": {
			code: []byte{0x1D},
			pre: cpuData{
				E:  0x01,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				E:  0x00,
				F:  fn | fz,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC E half carry": {
			code: []byte{0x1D},
			pre: cpuData{
				E:  0x00,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				E:  0xFF,
				F:  fn | fh,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD E, n": {
			code: []byte{0x1E, 0x42},
			pre: cpuData{
				E:  0x00,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				E:  0x42,
				PC: 0x8002,
			},
			wantbus: testBus{},
		},
		"RRA": {
			code: []byte{0x1F},
			pre: cpuData{
				A:  0xAA,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x55,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"RRA carry": {
			code: []byte{0x1F},
			pre: cpuData{
				A:  0x55,
				F:  fc,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0xAA,
				F:  fc,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0x20_0x2F(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"JR NZ, e": {
			code: []byte{0x20, 0x02},
			pre: cpuData{
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				PC: 0x8004,
			},
			wantbus: testBus{},
		},
		"JR NZ, e negative offset": {
			code: []byte{0x20, 0xFD}, // 0xFD = -3
			pre: cpuData{
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				PC: 0x7FFF,
			},
			wantbus: testBus{},
		},
		"JR NZ, NO JUMP": {
			code: []byte{0x20, 0xFD}, // 0xFD = -3
			pre: cpuData{
				F:  fz,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				F:  fz,
				PC: 0x8002,
			},
			wantbus: testBus{},
		},
		"LDI (HL+), A": {
			code: []byte{0x22},
			pre: cpuData{
				A:  0x42,
				H:  0x00,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0x00},
			want: cpuData{
				A:  0x42,
				H:  0x00,
				L:  0x03,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x42},
		},
		"INC H": {
			code: []byte{0x24},
			pre: cpuData{
				H:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				H:  0x43,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"INC H zero": {
			code: []byte{0x24},
			pre: cpuData{
				H:  0xFF,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				H:  0x00,
				F:  fz | fh,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"INC H half carry": {
			code: []byte{0x24},
			pre: cpuData{
				H:  0x0F,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				H:  0x10,
				F:  fh,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC H": {
			code: []byte{0x25},
			pre: cpuData{
				H:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				H:  0x41,
				F:  fn,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC H zero": {
			code: []byte{0x25},
			pre: cpuData{
				H:  0x01,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				H:  0x00,
				F:  fz | fn,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC H half carry": {
			code: []byte{0x25},
			pre: cpuData{
				H:  0x00,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				H:  0xFF,
				F:  fn | fh,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD H, n": {
			code: []byte{0x26, 0x42},
			pre: cpuData{
				H:  0x00,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				H:  0x42,
				PC: 0x8002,
			},
			wantbus: testBus{},
		},
		"JR Z, e": {
			code: []byte{0x28, 0x02},
			pre: cpuData{
				F:  fz,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				F:  fz,
				PC: 0x8004,
			},
			wantbus: testBus{},
		},
		"JR Z, e negative offset": {
			code: []byte{0x28, 0xFD}, // 0xFD = -3
			pre: cpuData{
				F:  fz,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				F:  fz,
				PC: 0x7FFF,
			},
			wantbus: testBus{},
		},
		"JR Z, NO JUMP": {
			code: []byte{0x28, 0xFD}, // 0xFD = -3
			pre: cpuData{
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				PC: 0x8002,
			},
			wantbus: testBus{},
		},
		"LD A, (HL+)": {
			code: []byte{0x2A},
			pre: cpuData{
				A:  0x00,
				H:  0x00,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0x42},
			want: cpuData{
				A:  0x42,
				H:  0x00,
				L:  0x03,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x42},
		},
		"INC L": {
			code: []byte{0x2C},
			pre: cpuData{
				L:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				L:  0x43,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"INC L zero": {
			code: []byte{0x2C},
			pre: cpuData{
				L:  0xFF,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				L:  0x00,
				F:  fz | fh,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"INC L half carry": {
			code: []byte{0x2C},
			pre: cpuData{
				L:  0x0F,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				L:  0x10,
				F:  fh,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC L": {
			code: []byte{0x2D},
			pre: cpuData{
				L:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				L:  0x41,
				F:  fn,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC L zero": {
			code: []byte{0x2D},
			pre: cpuData{
				L:  0x01,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				L:  0x00,
				F:  fn | fz,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC L half carry": {
			code: []byte{0x2D},
			pre: cpuData{
				L:  0x00,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				L:  0xFF,
				F:  fn | fh,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD L, n": {
			code: []byte{0x2E, 0x42},
			pre: cpuData{
				L:  0x00,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				L:  0x42,
				PC: 0x8002,
			},
			wantbus: testBus{},
		},
		"CPL": {
			code: []byte{0x2F},
			pre: cpuData{
				A:  0xAA,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x55,
				F:  fn | fh,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0x30_0x3F(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"JR NC, e": {
			code: []byte{0x30, 0x02},
			pre: cpuData{
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				PC: 0x8004,
			},
			wantbus: testBus{},
		},
		"JR NC, e negative offset": {
			code: []byte{0x30, 0xFD}, // 0xFD = -3
			pre: cpuData{
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				PC: 0x7FFF,
			},
			wantbus: testBus{},
		},
		"JR NC, NO JUMP": {
			code: []byte{0x30, 0xFD}, // 0xFD = -3
			pre: cpuData{
				F:  fc,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				F:  fc,
				PC: 0x8002,
			},
			wantbus: testBus{},
		},
		"LDD (HL-), A": {
			code: []byte{0x32},
			pre: cpuData{
				A:  0x42,
				H:  0x00,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0x00},
			want: cpuData{
				A:  0x42,
				H:  0x00,
				L:  0x01,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x42},
		},
		"INC (HL)": {
			code: []byte{0x34},
			pre: cpuData{
				H:  0x00,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0x42},
			want: cpuData{
				H:  0x00,
				L:  0x02,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x43},
		},
		"INC (HL) zero": {
			code: []byte{0x34},
			pre: cpuData{
				H:  0x00,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0xFF},
			want: cpuData{
				H:  0x00,
				L:  0x02,
				F:  fz | fh,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x00},
		},
		"INC (HL) half carry": {
			code: []byte{0x34},
			pre: cpuData{
				H:  0x00,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0x0F},
			want: cpuData{
				H:  0x00,
				L:  0x02,
				F:  fh,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x10},
		},
		"DEC (HL)": {
			code: []byte{0x35},
			pre: cpuData{
				H:  0x00,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0x42},
			want: cpuData{
				H:  0x00,
				L:  0x02,
				F:  fn,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x41},
		},
		"DEC (HL) zero": {
			code: []byte{0x35},
			pre: cpuData{
				H:  0x00,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0x01},
			want: cpuData{
				H:  0x00,
				L:  0x02,
				F:  fn | fz,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x00},
		},
		"DEC (HL) half carry": {
			code: []byte{0x35},
			pre: cpuData{
				H:  0x00,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0x00},
			want: cpuData{
				H:  0x00,
				L:  0x02,
				F:  fn | fh,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0xFF},
		},
		"LD (HL), n": {
			code: []byte{0x36, 0x42},
			pre: cpuData{
				H:  0x00,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0x00},
			want: cpuData{
				H:  0x00,
				L:  0x02,
				PC: 0x8002,
			},
			wantbus: testBus{0x0002: 0x42},
		},
		"SCF": {
			code: []byte{0x37},
			pre: cpuData{
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				F:  fc,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"SCF carry": {
			code: []byte{0x37},
			pre: cpuData{
				F:  fc,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				F:  fc,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"JR C, e": {
			code: []byte{0x38, 0x02},
			pre: cpuData{
				F:  fc,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				F:  fc,
				PC: 0x8004,
			},
			wantbus: testBus{},
		},
		"JR C, e negative offset": {
			code: []byte{0x38, 0xFD}, // 0xFD = -3
			pre: cpuData{
				F:  fc,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				F:  fc,
				PC: 0x7FFF,
			},
			wantbus: testBus{},
		},
		"JR C, NO JUMP": {
			code: []byte{0x38, 0xFD}, // 0xFD = -3
			pre: cpuData{
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				PC: 0x8002,
			},
			wantbus: testBus{},
		},
		"LD A, (HL-)": {
			code: []byte{0x3A},
			pre: cpuData{
				A:  0x00,
				H:  0x00,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0x42},
			want: cpuData{
				A:  0x42,
				H:  0x00,
				L:  0x01,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x42},
		},
		"INC A": {
			code: []byte{0x3C},
			pre: cpuData{
				A:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x43,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"INC A zero": {
			code: []byte{0x3C},
			pre: cpuData{
				A:  0xFF,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x00,
				F:  fz | fh,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"INC A half carry": {
			code: []byte{0x3C},
			pre: cpuData{
				A:  0x0F,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x10,
				F:  fh,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC A": {
			code: []byte{0x3D},
			pre: cpuData{
				A:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x41,
				F:  fn,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC A zero": {
			code: []byte{0x3D},
			pre: cpuData{
				A:  0x01,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x00,
				F:  fn | fz,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC A half carry": {
			code: []byte{0x3D},
			pre: cpuData{
				A:  0x00,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0xFF,
				F:  fn | fh,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD A, n": {
			code: []byte{0x3E, 0x42},
			pre: cpuData{
				A:  0x00,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x42,
				PC: 0x8002,
			},
			wantbus: testBus{},
		},
		"CCF": {
			code: []byte{0x3F},
			pre: cpuData{
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				F:  fc,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"CCF carry": {
			code: []byte{0x3F},
			pre: cpuData{
				F:  fc,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func testInst(mnemonic string, tt cpuSingleTest, t *testing.T) {
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
}
