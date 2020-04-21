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
		"LD BC,d16": {
			code: []byte{0x01, 0x41, 0x42},
			pre: cpuData{
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x42,
				C:  0x41,
				PC: 0x8003,
			},
			wantbus: testBus{},
		},
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
			wantbus: testBus{0x0002: 0x42},
		},
		"INC BC": {
			code: []byte{0x03},
			pre: cpuData{
				B:  0x41,
				C:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x41,
				C:  0x43,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"INC BC zero": {
			code: []byte{0x03},
			pre: cpuData{
				B:  0xFF,
				C:  0xFF,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x00,
				C:  0x00,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"INC BC half carry": {
			code: []byte{0x03},
			pre: cpuData{
				B:  0x00,
				C:  0xFF,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x01,
				C:  0x00,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
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
		"LD (a16),SP": {
			code: []byte{0x08, 0x41, 0x42},
			pre: cpuData{
				SP: 0x1424,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				SP: 0x1424,
				PC: 0x8003,
			},
			wantbus: testBus{0x4241: 0x24, 0x4242: 0x14},
		},
		"ADD HL,BC": {
			code: []byte{0x09},
			pre: cpuData{
				B:  0x0F,
				C:  0xFC,
				H:  0x00,
				L:  0x03,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x0F,
				C:  0xFC,
				H:  0x0F,
				L:  0xFF,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"ADD HL,BC carry": {
			code: []byte{0x09},
			pre: cpuData{
				B:  0xB7,
				C:  0xFD,
				H:  0x50,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0xB7,
				C:  0xFD,
				H:  0x07,
				L:  0xFF,
				F:  fc,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"ADD HL,BC half carry": {
			code: []byte{0x09},
			pre: cpuData{
				B:  0x06,
				C:  0x05,
				H:  0x8A,
				L:  0x23,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x06,
				C:  0x05,
				H:  0x90,
				L:  0x28,
				F:  fh,
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
		"DEC BC": {
			code: []byte{0x0B},
			pre: cpuData{
				B:  0x00,
				C:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x00,
				C:  0x41,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC BC zero": {
			code: []byte{0x0B},
			pre: cpuData{
				B:  0x00,
				C:  0x01,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x00,
				C:  0x00,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC BC half carry": {
			code: []byte{0x0B},
			pre: cpuData{
				B:  0x01,
				C:  0x00,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x00,
				C:  0xFF,
				PC: 0x8001,
			},
			wantbus: testBus{},
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
		// TODO: STOP
		"LD DE,d16": {
			code: []byte{0x11, 0x41, 0x42},
			pre: cpuData{
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				D:  0x42,
				E:  0x41,
				PC: 0x8003,
			},
			wantbus: testBus{},
		},
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
		"INC DE": {
			code: []byte{0x13},
			pre: cpuData{
				D:  0x41,
				E:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				D:  0x41,
				E:  0x43,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"INC DE zero": {
			code: []byte{0x13},
			pre: cpuData{
				D:  0xFF,
				E:  0xFF,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				D:  0x00,
				E:  0x00,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"INC DE half carry": {
			code: []byte{0x13},
			pre: cpuData{
				D:  0x00,
				E:  0xFF,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				D:  0x01,
				E:  0x00,
				PC: 0x8001,
			},
			wantbus: testBus{},
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
		"JR e": {
			code: []byte{0x18, 0x02},
			pre: cpuData{
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				PC: 0x8004,
			},
			wantbus: testBus{},
		},
		"JR e negative offset": {
			code: []byte{0x18, 0xFD}, // 0xFD = -3
			pre: cpuData{
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				PC: 0x7FFF,
			},
			wantbus: testBus{},
		},
		"ADD HL,DE": {
			code: []byte{0x19},
			pre: cpuData{
				D:  0x0F,
				E:  0xFC,
				H:  0x00,
				L:  0x03,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				D:  0x0F,
				E:  0xFC,
				H:  0x0F,
				L:  0xFF,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"ADD HL,DE carry": {
			code: []byte{0x19},
			pre: cpuData{
				D:  0xB7,
				E:  0xFD,
				H:  0x50,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				D:  0xB7,
				E:  0xFD,
				H:  0x07,
				L:  0xFF,
				F:  fc,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"ADD HL,DE half carry": {
			code: []byte{0x19},
			pre: cpuData{
				D:  0x06,
				E:  0x05,
				H:  0x8A,
				L:  0x23,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				D:  0x06,
				E:  0x05,
				H:  0x90,
				L:  0x28,
				F:  fh,
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
		"DEC DE": {
			code: []byte{0x1B},
			pre: cpuData{
				D:  0x00,
				E:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				D:  0x00,
				E:  0x41,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC DE zero": {
			code: []byte{0x1B},
			pre: cpuData{
				D:  0x00,
				E:  0x01,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				D:  0x00,
				E:  0x00,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC DE half carry": {
			code: []byte{0x1B},
			pre: cpuData{
				D:  0x01,
				E:  0x00,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				D:  0x00,
				E:  0xFF,
				PC: 0x8001,
			},
			wantbus: testBus{},
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
		"LD HL,d16": {
			code: []byte{0x21, 0x41, 0x42},
			pre: cpuData{
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				H:  0x42,
				L:  0x41,
				PC: 0x8003,
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
		"INC HL": {
			code: []byte{0x23},
			pre: cpuData{
				H:  0x41,
				L:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				H:  0x41,
				L:  0x43,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"INC HL zero": {
			code: []byte{0x23},
			pre: cpuData{
				H:  0xFF,
				L:  0xFF,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				H:  0x00,
				L:  0x00,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"INC HL half carry": {
			code: []byte{0x23},
			pre: cpuData{
				H:  0x00,
				L:  0xFF,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				H:  0x01,
				L:  0x00,
				PC: 0x8001,
			},
			wantbus: testBus{},
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
		// TODO: DAA
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
		"ADD HL,HL": {
			code: []byte{0x29},
			pre: cpuData{
				H:  0x02,
				L:  0xAA,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				H:  0x05,
				L:  0x54,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"ADD HL,HL carry": {
			code: []byte{0x29},
			pre: cpuData{
				H:  0x80,
				L:  0x01,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				H:  0x00,
				L:  0x02,
				F:  fc,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"ADD HL,HL half carry": {
			code: []byte{0x29},
			pre: cpuData{
				H:  0x0F,
				L:  0xFF,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				H:  0x1F,
				L:  0xFE,
				F:  fh,
				PC: 0x8001,
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
		"DEC HL": {
			code: []byte{0x2B},
			pre: cpuData{
				H:  0x00,
				L:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				H:  0x00,
				L:  0x41,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC HL zero": {
			code: []byte{0x2B},
			pre: cpuData{
				H:  0x00,
				L:  0x01,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				H:  0x00,
				L:  0x00,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC HL half carry": {
			code: []byte{0x2B},
			pre: cpuData{
				H:  0x01,
				L:  0x00,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				H:  0x00,
				L:  0xFF,
				PC: 0x8001,
			},
			wantbus: testBus{},
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
		"LD SP,d16": {
			code: []byte{0x31, 0x41, 0x42},
			pre: cpuData{
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				SP: 0x4241,
				PC: 0x8003,
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
		"INC SP": {
			code: []byte{0x33},
			pre: cpuData{
				SP: 0x4142,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				SP: 0x4143,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"INC SP zero": {
			code: []byte{0x33},
			pre: cpuData{
				SP: 0xFFFF,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				SP: 0x0000,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"INC SP half carry": {
			code: []byte{0x33},
			pre: cpuData{
				SP: 0x00FF,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				SP: 0x0100,
				PC: 0x8001,
			},
			wantbus: testBus{},
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
		"ADD HL,SP": {
			code: []byte{0x39},
			pre: cpuData{
				SP: 0x0FFC,
				H:  0x00,
				L:  0x03,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				SP: 0x0FFC,
				H:  0x0F,
				L:  0xFF,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"ADD HL,SP carry": {
			code: []byte{0x39},
			pre: cpuData{
				SP: 0xB7FD,
				H:  0x50,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				SP: 0xB7FD,
				H:  0x07,
				L:  0xFF,
				F:  fc,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"ADD HL,SP half carry": {
			code: []byte{0x39},
			pre: cpuData{
				SP: 0x0605,
				H:  0x8A,
				L:  0x23,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				SP: 0x0605,
				H:  0x90,
				L:  0x28,
				F:  fh,
				PC: 0x8001,
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
		"DEC SP": {
			code: []byte{0x3B},
			pre: cpuData{
				SP: 0x0042,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				SP: 0x0041,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC SP zero": {
			code: []byte{0x3B},
			pre: cpuData{
				SP: 0x0001,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				SP: 0x0000,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"DEC SP half carry": {
			code: []byte{0x3B},
			pre: cpuData{
				SP: 0x0100,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				SP: 0x00FF,
				PC: 0x8001,
			},
			wantbus: testBus{},
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

func TestCpuOps0x40_0x4F(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"LD B, B": {
			code: []byte{0x40},
			pre: cpuData{
				B:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD B, C": {
			code: []byte{0x41},
			pre: cpuData{
				B:  0x00,
				C:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x42,
				C:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD B, D": {
			code: []byte{0x42},
			pre: cpuData{
				B:  0x00,
				D:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x42,
				D:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD B, E": {
			code: []byte{0x43},
			pre: cpuData{
				B:  0x00,
				E:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x42,
				E:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD B, H": {
			code: []byte{0x44},
			pre: cpuData{
				B:  0x00,
				H:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x42,
				H:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD B, L": {
			code: []byte{0x45},
			pre: cpuData{
				B:  0x00,
				L:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x42,
				L:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD B, (HL)": {
			code: []byte{0x46},
			pre: cpuData{
				B:  0x00,
				H:  0x00,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0x42},
			want: cpuData{
				B:  0x42,
				H:  0x00,
				L:  0x02,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD B, A": {
			code: []byte{0x47},
			pre: cpuData{
				A:  0x42,
				B:  0x00,
				PC: 0x8000,
			}, bus: testBus{},
			want: cpuData{
				A:  0x42,
				B:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD C, B": {
			code: []byte{0x48},
			pre: cpuData{
				B:  0x42,
				C:  0x00,
				PC: 0x8000,
			}, bus: testBus{},
			want: cpuData{
				B:  0x42,
				C:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD C, C": {
			code: []byte{0x49},
			pre: cpuData{
				C:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				C:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD C, D": {
			code: []byte{0x4A},
			pre: cpuData{
				C:  0x00,
				D:  0x42,
				PC: 0x8000,
			}, bus: testBus{},
			want: cpuData{
				C:  0x42,
				D:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD C, E": {
			code: []byte{0x4B},
			pre: cpuData{
				C:  0x00,
				E:  0x42,
				PC: 0x8000,
			}, bus: testBus{},
			want: cpuData{
				C:  0x42,
				E:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD C, H": {
			code: []byte{0x4C},
			pre: cpuData{
				C:  0x00,
				H:  0x42,
				PC: 0x8000,
			}, bus: testBus{},
			want: cpuData{
				C:  0x42,
				H:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD C, L": {
			code: []byte{0x4D},
			pre: cpuData{
				C:  0x00,
				L:  0x42,
				PC: 0x8000,
			}, bus: testBus{},
			want: cpuData{
				C:  0x42,
				L:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD C, (HL)": {
			code: []byte{0x4E},
			pre: cpuData{
				C:  0x00,
				H:  0x00,
				L:  0x02,
				PC: 0x8000,
			}, bus: testBus{0x0002: 0x42},
			want: cpuData{
				C:  0x42,
				H:  0x00,
				L:  0x02,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD C, A": {
			code: []byte{0x4F},
			pre: cpuData{
				A:  0x42,
				C:  0x00,
				PC: 0x8000,
			}, bus: testBus{},
			want: cpuData{
				A:  0x42,
				C:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0x50_0x5F(t *testing.T) {
	tests := map[string]cpuSingleTest{

		"LD D, B": {
			code: []byte{0x50},
			pre: cpuData{
				B:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x42,
				D:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD D, C": {
			code: []byte{0x51},
			pre: cpuData{
				C:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				C:  0x42,
				D:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD D, D": {
			code: []byte{0x52},
			pre: cpuData{
				D:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				D:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD D, E": {
			code: []byte{0x53},
			pre: cpuData{
				E:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				E:  0x42,
				D:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD D, H": {
			code: []byte{0x54},
			pre: cpuData{
				H:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				H:  0x42,
				D:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD D, L": {
			code: []byte{0x55},
			pre: cpuData{
				L:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				L:  0x42,
				D:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD D, (HL)": {
			code: []byte{0x56},
			pre: cpuData{
				H:  0x00,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0x42},
			want: cpuData{
				D:  0x42,
				H:  0x00,
				L:  0x02,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD D, A": {
			code: []byte{0x57},
			pre: cpuData{
				A:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x42,
				D:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD E, B": {
			code: []byte{0x58},
			pre: cpuData{
				B:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x42,
				E:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD E, C": {
			code: []byte{0x59},
			pre: cpuData{
				C:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				C:  0x42,
				E:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD E, D": {
			code: []byte{0x5A},
			pre: cpuData{
				D:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				D:  0x42,
				E:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD E, E": {
			code: []byte{0x5B},
			pre: cpuData{
				E:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				E:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD E, H": {
			code: []byte{0x5C},
			pre: cpuData{
				H:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				H:  0x42,
				E:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD E, L": {
			code: []byte{0x5D},
			pre: cpuData{
				L:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				L:  0x42,
				E:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD E, (HL)": {
			code: []byte{0x5E},
			pre: cpuData{
				H:  0x00,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0x42},
			want: cpuData{
				H:  0x00,
				L:  0x02,
				E:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD E, A": {
			code: []byte{0x5F},
			pre: cpuData{
				A:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x42,
				E:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0x60_0x6F(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"LD H, B": {
			code: []byte{0x60},
			pre: cpuData{
				B:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x42,
				H:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD H, C": {
			code: []byte{0x61},
			pre: cpuData{
				C:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				C:  0x42,
				H:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD H, D": {
			code: []byte{0x62},
			pre: cpuData{
				D:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				D:  0x42,
				H:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD H, E": {
			code: []byte{0x63},
			pre: cpuData{
				E:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				E:  0x42,
				H:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD H, H": {
			code: []byte{0x64},
			pre: cpuData{
				H:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				H:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD H, L": {
			code: []byte{0x65},
			pre: cpuData{
				L:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				L:  0x42,
				H:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD H, (HL)": {
			code: []byte{0x66},
			pre: cpuData{
				H:  0x00,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0x42},
			want: cpuData{
				H:  0x42,
				L:  0x02,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD H, A": {
			code: []byte{0x67},
			pre: cpuData{
				A:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x42,
				H:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD L, B": {
			code: []byte{0x68},
			pre: cpuData{
				B:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				B:  0x42,
				L:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD L, C": {
			code: []byte{0x69},
			pre: cpuData{
				C:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				C:  0x42,
				L:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD L, D": {
			code: []byte{0x6A},
			pre: cpuData{
				D:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				D:  0x42,
				L:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD L, E": {
			code: []byte{0x6B},
			pre: cpuData{
				E:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				E:  0x42,
				L:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD L, H": {
			code: []byte{0x6C},
			pre: cpuData{
				H:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				H:  0x42,
				L:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD L, L": {
			code: []byte{0x6D},
			pre: cpuData{
				L:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				L:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD L, (HL)": {
			code: []byte{0x6E},
			pre: cpuData{
				H:  0x00,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0x42},
			want: cpuData{
				H:  0x00,
				L:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD L, A": {
			code: []byte{0x6F},
			pre: cpuData{
				A:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x42,
				L:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0x70_0x7F(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"LD (HL), B": {
			code: []byte{0x70},
			pre: cpuData{
				B:  0x42,
				H:  0x00,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0x00},
			want: cpuData{
				B:  0x42,
				H:  0x00,
				L:  0x02,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD (HL), C": {
			code: []byte{0x71},
			pre: cpuData{
				C:  0x42,
				H:  0x00,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0x00},
			want: cpuData{
				C:  0x42,
				H:  0x00,
				L:  0x02,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD (HL), D": {
			code: []byte{0x72},
			pre: cpuData{
				D:  0x42,
				H:  0x00,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0x00},
			want: cpuData{
				D:  0x42,
				H:  0x00,
				L:  0x02,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD (HL), E": {
			code: []byte{0x73},
			pre: cpuData{
				E:  0x42,
				H:  0x00,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0x00},
			want: cpuData{
				E:  0x42,
				H:  0x00,
				L:  0x02,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD (HL), H": {
			code: []byte{0x74},
			pre: cpuData{
				H:  0x42,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x4202: 0x42},
			want: cpuData{
				H:  0x42,
				L:  0x02,
				PC: 0x8001,
			},
			wantbus: testBus{0x4202: 0x42},
		},
		"LD (HL), L": {
			code: []byte{0x75},
			pre: cpuData{
				H:  0x00,
				L:  0x42,
				PC: 0x8000,
			},
			bus: testBus{0x0042: 0x00},
			want: cpuData{
				H:  0x00,
				L:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{0x0042: 0x42},
		},
		// TODO: halt
		"LD (HL), A": {
			code: []byte{0x77},
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
				L:  0x02,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD A, B": {
			code: []byte{0x78},
			pre: cpuData{
				B:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x42,
				B:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD A, C": {
			code: []byte{0x79},
			pre: cpuData{
				C:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x42,
				C:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD A, D": {
			code: []byte{0x7A},
			pre: cpuData{
				D:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x42,
				D:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD A, E": {
			code: []byte{0x7B},
			pre: cpuData{
				E:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x42,
				E:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD A, H": {
			code: []byte{0x7C},
			pre: cpuData{
				H:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x42,
				H:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD A, L": {
			code: []byte{0x7D},
			pre: cpuData{
				L:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x42,
				L:  0x42,
				PC: 0x8001,
			},
			wantbus: testBus{},
		},
		"LD A, (HL)": {
			code: []byte{0x7E},
			pre: cpuData{
				H:  0x00,
				L:  0x02,
				PC: 0x8000,
			},
			bus: testBus{0x0002: 0x42},
			want: cpuData{
				A:  0x42,
				H:  0x00,
				L:  0x02,
				PC: 0x8001,
			},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD A, A": {
			code: []byte{0x7F},
			pre: cpuData{
				A:  0x42,
				PC: 0x8000,
			},
			bus: testBus{},
			want: cpuData{
				A:  0x42,
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
