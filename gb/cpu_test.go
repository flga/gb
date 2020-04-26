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

	IME bool
}

func (cd cpuData) String() string {
	return fmt.Sprintf("F[%s] A[0x%02X] B[0x%02X] C[0x%02X] D[0x%02X] E[0x%02X] H[0x%02X] L[0x%02X] SP[0x%04X] PC[0x%04X]",
		cd.F,
		cd.A,
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
			code:    []byte{0x00},
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{PC: 0x8001},
			wantbus: testBus{},
		},
		"LD BC,d16": {
			code:    []byte{0x01, 0x41, 0x42},
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x42, C: 0x41, PC: 0x8003},
			wantbus: testBus{},
		},
		"LD (BC), A": {
			code:    []byte{0x02},
			pre:     cpuData{A: 0x42, B: 0x00, C: 0x02, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, B: 0x00, C: 0x02, PC: 0x8001},
			wantbus: testBus{0x0002: 0x42},
		},
		"INC BC": {
			code:    []byte{0x03},
			pre:     cpuData{B: 0x41, C: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x41, C: 0x43, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC BC zero": {
			code:    []byte{0x03},
			pre:     cpuData{B: 0xFF, C: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x00, C: 0x00, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC BC half carry": {
			code:    []byte{0x03},
			pre:     cpuData{B: 0x00, C: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x01, C: 0x00, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC B": {
			code:    []byte{0x04},
			pre:     cpuData{B: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x43, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC B zero": {
			code:    []byte{0x04},
			pre:     cpuData{B: 0xff, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x00, F: Z | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC B half carry": {
			code:    []byte{0x04},
			pre:     cpuData{B: 0x0f, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x10, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC B": {
			code:    []byte{0x05},
			pre:     cpuData{B: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x41, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC B zero": {
			code:    []byte{0x05},
			pre:     cpuData{B: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x00, F: N | Z, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC B half carry": {
			code:    []byte{0x05},
			pre:     cpuData{B: 0x00, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0xFF, F: N | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD B, n": {
			code:    []byte{0x06, 0x42},
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x42, PC: 0x8002},
			wantbus: testBus{},
		},
		"RLCA": {
			code:    []byte{0x07},
			pre:     cpuData{A: 0x77, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0xEE, PC: 0x8001},
			wantbus: testBus{},
		},
		"RLCA carry": {
			code:    []byte{0x07},
			pre:     cpuData{A: 0xF7, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0xEF, F: CY, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD (a16),SP": {
			code:    []byte{0x08, 0x41, 0x42},
			pre:     cpuData{SP: 0x1424, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x1424, PC: 0x8003},
			wantbus: testBus{0x4241: 0x24, 0x4242: 0x14},
		},
		"ADD HL,BC": {
			code:    []byte{0x09},
			pre:     cpuData{B: 0x0F, C: 0xFC, H: 0x00, L: 0x03, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x0F, C: 0xFC, H: 0x0F, L: 0xFF, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD HL,BC carry": {
			code:    []byte{0x09},
			pre:     cpuData{B: 0xB7, C: 0xFD, H: 0x50, L: 0x02, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0xB7, C: 0xFD, H: 0x07, L: 0xFF, F: CY, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD HL,BC half carry": {
			code:    []byte{0x09},
			pre:     cpuData{B: 0x06, C: 0x05, H: 0x8A, L: 0x23, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x06, C: 0x05, H: 0x90, L: 0x28, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD A, (BC)": {
			code:    []byte{0x0A},
			pre:     cpuData{B: 0x00, C: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x42},
			want:    cpuData{A: 0x42, B: 0x00, C: 0x02, PC: 0x8001},
			wantbus: testBus{0x0002: 0x42},
		},
		"DEC BC": {
			code:    []byte{0x0B},
			pre:     cpuData{B: 0x00, C: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x00, C: 0x41, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC BC zero": {
			code:    []byte{0x0B},
			pre:     cpuData{B: 0x00, C: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x00, C: 0x00, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC BC half carry": {
			code:    []byte{0x0B},
			pre:     cpuData{B: 0x01, C: 0x00, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x00, C: 0xFF, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC C": {
			code:    []byte{0x0C},
			pre:     cpuData{C: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{C: 0x43, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC C zero": {
			code:    []byte{0x0C},
			pre:     cpuData{C: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{C: 0x00, F: Z | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC C half carry": {
			code:    []byte{0x0C},
			pre:     cpuData{C: 0x0F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{C: 0x10, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC C": {
			code:    []byte{0x0D},
			pre:     cpuData{C: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{C: 0x41, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC C zero": {
			code:    []byte{0x0D},
			pre:     cpuData{C: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{C: 0x00, F: N | Z, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC C half carry": {
			code:    []byte{0x0D},
			pre:     cpuData{C: 0x00, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{C: 0xFF, F: N | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD C, n": {
			code:    []byte{0x0E, 0x42},
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{C: 0x42, PC: 0x8002},
			wantbus: testBus{},
		},
		"RRCA": {
			code:    []byte{0x0F},
			pre:     cpuData{A: 0xEE, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x77, PC: 0x8001},
			wantbus: testBus{},
		},
		"RRCA carry": {
			code:    []byte{0x0F},
			pre:     cpuData{A: 0xEF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0xF7, F: CY, PC: 0x8001},
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
			code:    []byte{0x11, 0x41, 0x42},
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{D: 0x42, E: 0x41, PC: 0x8003},
			wantbus: testBus{},
		},
		"LD (DE), A": {
			code:    []byte{0x12},
			pre:     cpuData{A: 0x42, D: 0x00, E: 0x02, PC: 0x8000},
			bus:     testBus{0x02: 0x00},
			want:    cpuData{A: 0x42, D: 0x00, E: 0x02, PC: 0x8001},
			wantbus: testBus{0x02: 0x42},
		},
		"INC DE": {
			code:    []byte{0x13},
			pre:     cpuData{D: 0x41, E: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{D: 0x41, E: 0x43, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC DE zero": {
			code:    []byte{0x13},
			pre:     cpuData{D: 0xFF, E: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{D: 0x00, E: 0x00, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC DE half carry": {
			code:    []byte{0x13},
			pre:     cpuData{D: 0x00, E: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{D: 0x01, E: 0x00, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC D": {
			code:    []byte{0x14},
			pre:     cpuData{D: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{D: 0x43, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC D zero": {
			code:    []byte{0x14},
			pre:     cpuData{D: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{D: 0x00, F: Z | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC D half carry": {
			code:    []byte{0x14},
			pre:     cpuData{D: 0x0F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{D: 0x10, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC D": {
			code:    []byte{0x15},
			pre:     cpuData{D: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{D: 0x41, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC D zero": {
			code:    []byte{0x15},
			pre:     cpuData{D: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{D: 0x00, F: N | Z, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC D half carry": {
			code:    []byte{0x15},
			pre:     cpuData{D: 0x00, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{D: 0xFF, F: N | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD D, n": {
			code:    []byte{0x16, 0x42},
			pre:     cpuData{D: 0x00, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{D: 0x42, PC: 0x8002},
			wantbus: testBus{},
		},
		"RLA": {
			code:    []byte{0x17},
			pre:     cpuData{A: 0x55, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0xAA, PC: 0x8001},
			wantbus: testBus{},
		},
		"RLA carry": {
			code:    []byte{0x17},
			pre:     cpuData{A: 0xAA, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x55, F: CY, PC: 0x8001},
			wantbus: testBus{},
		},
		"JR e": {
			code:    []byte{0x18, 0x02},
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{PC: 0x8004},
			wantbus: testBus{},
		},
		"JR e negative offset": {
			code:    []byte{0x18, 0xFD}, // 0xFD = -3
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{PC: 0x7FFF},
			wantbus: testBus{},
		},
		"ADD HL,DE": {
			code:    []byte{0x19},
			pre:     cpuData{D: 0x0F, E: 0xFC, H: 0x00, L: 0x03, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{D: 0x0F, E: 0xFC, H: 0x0F, L: 0xFF, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD HL,DE carry": {
			code:    []byte{0x19},
			pre:     cpuData{D: 0xB7, E: 0xFD, H: 0x50, L: 0x02, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{D: 0xB7, E: 0xFD, H: 0x07, L: 0xFF, F: CY, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD HL,DE half carry": {
			code:    []byte{0x19},
			pre:     cpuData{D: 0x06, E: 0x05, H: 0x8A, L: 0x23, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{D: 0x06, E: 0x05, H: 0x90, L: 0x28, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD A, (DE)": {
			code:    []byte{0x1A},
			pre:     cpuData{A: 0x00, D: 0x00, E: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x42},
			want:    cpuData{A: 0x42, D: 0x00, E: 0x02, PC: 0x8001},
			wantbus: testBus{0x0002: 0x42},
		},
		"DEC DE": {
			code:    []byte{0x1B},
			pre:     cpuData{D: 0x00, E: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{D: 0x00, E: 0x41, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC DE zero": {
			code:    []byte{0x1B},
			pre:     cpuData{D: 0x00, E: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{D: 0x00, E: 0x00, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC DE half carry": {
			code:    []byte{0x1B},
			pre:     cpuData{D: 0x01, E: 0x00, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{D: 0x00, E: 0xFF, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC E": {
			code:    []byte{0x1C},
			pre:     cpuData{E: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{E: 0x43, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC E zero": {
			code:    []byte{0x1C},
			pre:     cpuData{E: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{E: 0x00, F: Z | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC E half carry": {
			code:    []byte{0x1C},
			pre:     cpuData{E: 0x0F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{E: 0x10, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC E": {
			code:    []byte{0x1D},
			pre:     cpuData{E: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{E: 0x41, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC E zero": {
			code:    []byte{0x1D},
			pre:     cpuData{E: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{E: 0x00, F: N | Z, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC E half carry": {
			code:    []byte{0x1D},
			pre:     cpuData{E: 0x00, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{E: 0xFF, F: N | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD E, n": {
			code:    []byte{0x1E, 0x42},
			pre:     cpuData{E: 0x00, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{E: 0x42, PC: 0x8002},
			wantbus: testBus{},
		},
		"RRA": {
			code:    []byte{0x1F},
			pre:     cpuData{A: 0xAA, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x55, PC: 0x8001},
			wantbus: testBus{},
		},
		"RRA carry": {
			code:    []byte{0x1F},
			pre:     cpuData{A: 0x55, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0xAA, F: CY, PC: 0x8001},
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
			code:    []byte{0x20, 0x02},
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{PC: 0x8004},
			wantbus: testBus{},
		},
		"JR NZ, e negative offset": {
			code:    []byte{0x20, 0xFD}, // 0xFD = -3
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{PC: 0x7FFF},
			wantbus: testBus{},
		},
		"JR NZ, NO JUMP": {
			code:    []byte{0x20, 0xFD}, // 0xFD = -3
			pre:     cpuData{F: Z, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: Z, PC: 0x8002},
			wantbus: testBus{},
		},
		"LD HL,d16": {
			code:    []byte{0x21, 0x41, 0x42},
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0x42, L: 0x41, PC: 0x8003},
			wantbus: testBus{},
		},
		"LDI (HL+), A": {
			code:    []byte{0x22},
			pre:     cpuData{A: 0x42, H: 0x00, L: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x00},
			want:    cpuData{A: 0x42, H: 0x00, L: 0x03, PC: 0x8001},
			wantbus: testBus{0x0002: 0x42},
		},
		"INC HL": {
			code:    []byte{0x23},
			pre:     cpuData{H: 0x41, L: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0x41, L: 0x43, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC HL zero": {
			code:    []byte{0x23},
			pre:     cpuData{H: 0xFF, L: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0x00, L: 0x00, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC HL half carry": {
			code:    []byte{0x23},
			pre:     cpuData{H: 0x00, L: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0x01, L: 0x00, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC H": {
			code:    []byte{0x24},
			pre:     cpuData{H: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0x43, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC H zero": {
			code:    []byte{0x24},
			pre:     cpuData{H: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0x00, F: Z | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC H half carry": {
			code:    []byte{0x24},
			pre:     cpuData{H: 0x0F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0x10, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC H": {
			code:    []byte{0x25},
			pre:     cpuData{H: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0x41, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC H zero": {
			code:    []byte{0x25},
			pre:     cpuData{H: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0x00, F: Z | N, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC H half carry": {
			code:    []byte{0x25},
			pre:     cpuData{H: 0x00, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0xFF, F: N | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD H, n": {
			code:    []byte{0x26, 0x42},
			pre:     cpuData{H: 0x00, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0x42, PC: 0x8002},
			wantbus: testBus{},
		},
		// TODO: DAA
		"JR Z, e": {
			code:    []byte{0x28, 0x02},
			pre:     cpuData{F: Z, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: Z, PC: 0x8004},
			wantbus: testBus{},
		},
		"JR Z, e negative offset": {
			code:    []byte{0x28, 0xFD}, // 0xFD = -3
			pre:     cpuData{F: Z, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: Z, PC: 0x7FFF},
			wantbus: testBus{},
		},
		"JR Z, NO JUMP": {
			code:    []byte{0x28, 0xFD}, // 0xFD = -3
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{PC: 0x8002},
			wantbus: testBus{},
		},
		"ADD HL,HL": {
			code:    []byte{0x29},
			pre:     cpuData{H: 0x02, L: 0xAA, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0x05, L: 0x54, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD HL,HL carry": {
			code:    []byte{0x29},
			pre:     cpuData{H: 0x80, L: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0x00, L: 0x02, F: CY, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD HL,HL half carry": {
			code:    []byte{0x29},
			pre:     cpuData{H: 0x0F, L: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0x1F, L: 0xFE, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD A, (HL+)": {
			code:    []byte{0x2A},
			pre:     cpuData{A: 0x00, H: 0x00, L: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x42},
			want:    cpuData{A: 0x42, H: 0x00, L: 0x03, PC: 0x8001},
			wantbus: testBus{0x0002: 0x42},
		},
		"DEC HL": {
			code:    []byte{0x2B},
			pre:     cpuData{H: 0x00, L: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0x00, L: 0x41, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC HL zero": {
			code:    []byte{0x2B},
			pre:     cpuData{H: 0x00, L: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0x00, L: 0x00, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC HL half carry": {
			code:    []byte{0x2B},
			pre:     cpuData{H: 0x01, L: 0x00, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0x00, L: 0xFF, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC L": {
			code:    []byte{0x2C},
			pre:     cpuData{L: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{L: 0x43, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC L zero": {
			code:    []byte{0x2C},
			pre:     cpuData{L: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{L: 0x00, F: Z | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC L half carry": {
			code:    []byte{0x2C},
			pre:     cpuData{L: 0x0F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{L: 0x10, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC L": {
			code:    []byte{0x2D},
			pre:     cpuData{L: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{L: 0x41, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC L zero": {
			code:    []byte{0x2D},
			pre:     cpuData{L: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{L: 0x00, F: N | Z, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC L half carry": {
			code:    []byte{0x2D},
			pre:     cpuData{L: 0x00, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{L: 0xFF, F: N | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD L, n": {
			code:    []byte{0x2E, 0x42},
			pre:     cpuData{L: 0x00, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{L: 0x42, PC: 0x8002},
			wantbus: testBus{},
		},
		"CPL": {
			code:    []byte{0x2F},
			pre:     cpuData{A: 0xAA, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x55, F: N | H, PC: 0x8001},
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
			code:    []byte{0x30, 0x02},
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{PC: 0x8004},
			wantbus: testBus{},
		},
		"JR NC, e negative offset": {
			code:    []byte{0x30, 0xFD}, // 0xFD = -3
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{PC: 0x7FFF},
			wantbus: testBus{},
		},
		"JR NC, NO JUMP": {
			code:    []byte{0x30, 0xFD}, // 0xFD = -3
			pre:     cpuData{F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: CY, PC: 0x8002},
			wantbus: testBus{},
		},
		"LD SP,d16": {
			code:    []byte{0x31, 0x41, 0x42},
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x4241, PC: 0x8003},
			wantbus: testBus{},
		},
		"LDD (HL-), A": {
			code:    []byte{0x32},
			pre:     cpuData{A: 0x42, H: 0x00, L: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x00},
			want:    cpuData{A: 0x42, H: 0x00, L: 0x01, PC: 0x8001},
			wantbus: testBus{0x0002: 0x42},
		},
		"INC SP": {
			code:    []byte{0x33},
			pre:     cpuData{SP: 0x4142, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x4143, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC SP zero": {
			code:    []byte{0x33},
			pre:     cpuData{SP: 0xFFFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x0000, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC SP half carry": {
			code:    []byte{0x33},
			pre:     cpuData{SP: 0x00FF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x0100, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC (HL)": {
			code:    []byte{0x34},
			pre:     cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x42},
			want:    cpuData{H: 0x00, L: 0x02, PC: 0x8001},
			wantbus: testBus{0x0002: 0x43},
		},
		"INC (HL) zero": {
			code:    []byte{0x34},
			pre:     cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0xFF},
			want:    cpuData{H: 0x00, L: 0x02, F: Z | H, PC: 0x8001},
			wantbus: testBus{0x0002: 0x00},
		},
		"INC (HL) half carry": {
			code:    []byte{0x34},
			pre:     cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x0F},
			want:    cpuData{H: 0x00, L: 0x02, F: H, PC: 0x8001},
			wantbus: testBus{0x0002: 0x10},
		},
		"DEC (HL)": {
			code:    []byte{0x35},
			pre:     cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x42},
			want:    cpuData{H: 0x00, L: 0x02, F: N, PC: 0x8001},
			wantbus: testBus{0x0002: 0x41},
		},
		"DEC (HL) zero": {
			code:    []byte{0x35},
			pre:     cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x01},
			want:    cpuData{H: 0x00, L: 0x02, F: N | Z, PC: 0x8001},
			wantbus: testBus{0x0002: 0x00},
		},
		"DEC (HL) half carry": {
			code:    []byte{0x35},
			pre:     cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x00},
			want:    cpuData{H: 0x00, L: 0x02, F: N | H, PC: 0x8001},
			wantbus: testBus{0x0002: 0xFF},
		},
		"LD (HL), n": {
			code:    []byte{0x36, 0x42},
			pre:     cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x00},
			want:    cpuData{H: 0x00, L: 0x02, PC: 0x8002},
			wantbus: testBus{0x0002: 0x42},
		},
		"SCF": {
			code:    []byte{0x37},
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: CY, PC: 0x8001},
			wantbus: testBus{},
		},
		"SCF carry": {
			code:    []byte{0x37},
			pre:     cpuData{F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: CY, PC: 0x8001},
			wantbus: testBus{},
		},
		"JR C, e": {
			code:    []byte{0x38, 0x02},
			pre:     cpuData{F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: CY, PC: 0x8004},
			wantbus: testBus{},
		},
		"JR C, e negative offset": {
			code:    []byte{0x38, 0xFD}, // 0xFD = -3
			pre:     cpuData{F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: CY, PC: 0x7FFF},
			wantbus: testBus{},
		},
		"JR C, NO JUMP": {
			code:    []byte{0x38, 0xFD}, // 0xFD = -3
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{PC: 0x8002},
			wantbus: testBus{},
		},
		"ADD HL,SP": {
			code:    []byte{0x39},
			pre:     cpuData{SP: 0x0FFC, H: 0x00, L: 0x03, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x0FFC, H: 0x0F, L: 0xFF, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD HL,SP carry": {
			code:    []byte{0x39},
			pre:     cpuData{SP: 0xB7FD, H: 0x50, L: 0x02, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0xB7FD, H: 0x07, L: 0xFF, F: CY, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD HL,SP half carry": {
			code:    []byte{0x39},
			pre:     cpuData{SP: 0x0605, H: 0x8A, L: 0x23, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x0605, H: 0x90, L: 0x28, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD A, (HL-)": {
			code:    []byte{0x3A},
			pre:     cpuData{A: 0x00, H: 0x00, L: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x42},
			want:    cpuData{A: 0x42, H: 0x00, L: 0x01, PC: 0x8001},
			wantbus: testBus{0x0002: 0x42},
		},
		"DEC SP": {
			code:    []byte{0x3B},
			pre:     cpuData{SP: 0x0042, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x0041, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC SP zero": {
			code:    []byte{0x3B},
			pre:     cpuData{SP: 0x0001, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x0000, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC SP half carry": {
			code:    []byte{0x3B},
			pre:     cpuData{SP: 0x0100, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x00FF, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC A": {
			code:    []byte{0x3C},
			pre:     cpuData{A: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x43, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC A zero": {
			code:    []byte{0x3C},
			pre:     cpuData{A: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, F: Z | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"INC A half carry": {
			code:    []byte{0x3C},
			pre:     cpuData{A: 0x0F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x10, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC A": {
			code:    []byte{0x3D},
			pre:     cpuData{A: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x41, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC A zero": {
			code:    []byte{0x3D},
			pre:     cpuData{A: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, F: N | Z, PC: 0x8001},
			wantbus: testBus{},
		},
		"DEC A half carry": {
			code:    []byte{0x3D},
			pre:     cpuData{A: 0x00, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0xFF, F: N | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD A, n": {
			code:    []byte{0x3E, 0x42},
			pre:     cpuData{A: 0x00, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, PC: 0x8002},
			wantbus: testBus{},
		},
		"CCF": {
			code:    []byte{0x3F},
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: CY, PC: 0x8001},
			wantbus: testBus{},
		},
		"CCF carry": {
			code:    []byte{0x3F},
			pre:     cpuData{F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{PC: 0x8001},
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
			code:    []byte{0x40},
			pre:     cpuData{B: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD B, C": {
			code:    []byte{0x41},
			pre:     cpuData{B: 0x00, C: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x42, C: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD B, D": {
			code:    []byte{0x42},
			pre:     cpuData{B: 0x00, D: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x42, D: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD B, E": {
			code:    []byte{0x43},
			pre:     cpuData{B: 0x00, E: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x42, E: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD B, H": {
			code:    []byte{0x44},
			pre:     cpuData{B: 0x00, H: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x42, H: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD B, L": {
			code:    []byte{0x45},
			pre:     cpuData{B: 0x00, L: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x42, L: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD B, (HL)": {
			code:    []byte{0x46},
			pre:     cpuData{B: 0x00, H: 0x00, L: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x42},
			want:    cpuData{B: 0x42, H: 0x00, L: 0x02, PC: 0x8001},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD B, A": {
			code: []byte{0x47},
			pre:  cpuData{A: 0x42, B: 0x00, PC: 0x8000}, bus: testBus{},
			want:    cpuData{A: 0x42, B: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD C, B": {
			code: []byte{0x48},
			pre:  cpuData{B: 0x42, C: 0x00, PC: 0x8000}, bus: testBus{},
			want:    cpuData{B: 0x42, C: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD C, C": {
			code:    []byte{0x49},
			pre:     cpuData{C: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{C: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD C, D": {
			code: []byte{0x4A},
			pre:  cpuData{C: 0x00, D: 0x42, PC: 0x8000}, bus: testBus{},
			want:    cpuData{C: 0x42, D: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD C, E": {
			code: []byte{0x4B},
			pre:  cpuData{C: 0x00, E: 0x42, PC: 0x8000}, bus: testBus{},
			want:    cpuData{C: 0x42, E: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD C, H": {
			code: []byte{0x4C},
			pre:  cpuData{C: 0x00, H: 0x42, PC: 0x8000}, bus: testBus{},
			want:    cpuData{C: 0x42, H: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD C, L": {
			code: []byte{0x4D},
			pre:  cpuData{C: 0x00, L: 0x42, PC: 0x8000}, bus: testBus{},
			want:    cpuData{C: 0x42, L: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD C, (HL)": {
			code: []byte{0x4E},
			pre:  cpuData{C: 0x00, H: 0x00, L: 0x02, PC: 0x8000}, bus: testBus{0x0002: 0x42},
			want:    cpuData{C: 0x42, H: 0x00, L: 0x02, PC: 0x8001},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD C, A": {
			code: []byte{0x4F},
			pre:  cpuData{A: 0x42, C: 0x00, PC: 0x8000}, bus: testBus{},
			want:    cpuData{A: 0x42, C: 0x42, PC: 0x8001},
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
			code:    []byte{0x50},
			pre:     cpuData{B: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x42, D: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD D, C": {
			code:    []byte{0x51},
			pre:     cpuData{C: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{C: 0x42, D: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD D, D": {
			code:    []byte{0x52},
			pre:     cpuData{D: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{D: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD D, E": {
			code:    []byte{0x53},
			pre:     cpuData{E: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{E: 0x42, D: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD D, H": {
			code:    []byte{0x54},
			pre:     cpuData{H: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0x42, D: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD D, L": {
			code:    []byte{0x55},
			pre:     cpuData{L: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{L: 0x42, D: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD D, (HL)": {
			code:    []byte{0x56},
			pre:     cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x42},
			want:    cpuData{D: 0x42, H: 0x00, L: 0x02, PC: 0x8001},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD D, A": {
			code:    []byte{0x57},
			pre:     cpuData{A: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, D: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD E, B": {
			code:    []byte{0x58},
			pre:     cpuData{B: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x42, E: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD E, C": {
			code:    []byte{0x59},
			pre:     cpuData{C: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{C: 0x42, E: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD E, D": {
			code:    []byte{0x5A},
			pre:     cpuData{D: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{D: 0x42, E: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD E, E": {
			code:    []byte{0x5B},
			pre:     cpuData{E: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{E: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD E, H": {
			code:    []byte{0x5C},
			pre:     cpuData{H: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0x42, E: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD E, L": {
			code:    []byte{0x5D},
			pre:     cpuData{L: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{L: 0x42, E: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD E, (HL)": {
			code:    []byte{0x5E},
			pre:     cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x42},
			want:    cpuData{H: 0x00, L: 0x02, E: 0x42, PC: 0x8001},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD E, A": {
			code:    []byte{0x5F},
			pre:     cpuData{A: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, E: 0x42, PC: 0x8001},
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
			code:    []byte{0x60},
			pre:     cpuData{B: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x42, H: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD H, C": {
			code:    []byte{0x61},
			pre:     cpuData{C: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{C: 0x42, H: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD H, D": {
			code:    []byte{0x62},
			pre:     cpuData{D: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{D: 0x42, H: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD H, E": {
			code:    []byte{0x63},
			pre:     cpuData{E: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{E: 0x42, H: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD H, H": {
			code:    []byte{0x64},
			pre:     cpuData{H: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD H, L": {
			code:    []byte{0x65},
			pre:     cpuData{L: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{L: 0x42, H: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD H, (HL)": {
			code:    []byte{0x66},
			pre:     cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x42},
			want:    cpuData{H: 0x42, L: 0x02, PC: 0x8001},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD H, A": {
			code:    []byte{0x67},
			pre:     cpuData{A: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, H: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD L, B": {
			code:    []byte{0x68},
			pre:     cpuData{B: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x42, L: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD L, C": {
			code:    []byte{0x69},
			pre:     cpuData{C: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{C: 0x42, L: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD L, D": {
			code:    []byte{0x6A},
			pre:     cpuData{D: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{D: 0x42, L: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD L, E": {
			code:    []byte{0x6B},
			pre:     cpuData{E: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{E: 0x42, L: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD L, H": {
			code:    []byte{0x6C},
			pre:     cpuData{H: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0x42, L: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD L, L": {
			code:    []byte{0x6D},
			pre:     cpuData{L: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{L: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD L, (HL)": {
			code:    []byte{0x6E},
			pre:     cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x42},
			want:    cpuData{H: 0x00, L: 0x42, PC: 0x8001},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD L, A": {
			code:    []byte{0x6F},
			pre:     cpuData{A: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, L: 0x42, PC: 0x8001},
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
			code:    []byte{0x70},
			pre:     cpuData{B: 0x42, H: 0x00, L: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x00},
			want:    cpuData{B: 0x42, H: 0x00, L: 0x02, PC: 0x8001},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD (HL), C": {
			code:    []byte{0x71},
			pre:     cpuData{C: 0x42, H: 0x00, L: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x00},
			want:    cpuData{C: 0x42, H: 0x00, L: 0x02, PC: 0x8001},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD (HL), D": {
			code:    []byte{0x72},
			pre:     cpuData{D: 0x42, H: 0x00, L: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x00},
			want:    cpuData{D: 0x42, H: 0x00, L: 0x02, PC: 0x8001},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD (HL), E": {
			code:    []byte{0x73},
			pre:     cpuData{E: 0x42, H: 0x00, L: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x00},
			want:    cpuData{E: 0x42, H: 0x00, L: 0x02, PC: 0x8001},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD (HL), H": {
			code:    []byte{0x74},
			pre:     cpuData{H: 0x42, L: 0x02, PC: 0x8000},
			bus:     testBus{0x4202: 0x42},
			want:    cpuData{H: 0x42, L: 0x02, PC: 0x8001},
			wantbus: testBus{0x4202: 0x42},
		},
		"LD (HL), L": {
			code:    []byte{0x75},
			pre:     cpuData{H: 0x00, L: 0x42, PC: 0x8000},
			bus:     testBus{0x0042: 0x00},
			want:    cpuData{H: 0x00, L: 0x42, PC: 0x8001},
			wantbus: testBus{0x0042: 0x42},
		},
		// TODO: halt
		"LD (HL), A": {
			code:    []byte{0x77},
			pre:     cpuData{A: 0x42, H: 0x00, L: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x00},
			want:    cpuData{A: 0x42, H: 0x00, L: 0x02, PC: 0x8001},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD A, B": {
			code:    []byte{0x78},
			pre:     cpuData{B: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, B: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD A, C": {
			code:    []byte{0x79},
			pre:     cpuData{C: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, C: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD A, D": {
			code:    []byte{0x7A},
			pre:     cpuData{D: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, D: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD A, E": {
			code:    []byte{0x7B},
			pre:     cpuData{E: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, E: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD A, H": {
			code:    []byte{0x7C},
			pre:     cpuData{H: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, H: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD A, L": {
			code:    []byte{0x7D},
			pre:     cpuData{L: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, L: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
		"LD A, (HL)": {
			code:    []byte{0x7E},
			pre:     cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:     testBus{0x0002: 0x42},
			want:    cpuData{A: 0x42, H: 0x00, L: 0x02, PC: 0x8001},
			wantbus: testBus{0x0002: 0x42},
		},
		"LD A, A": {
			code:    []byte{0x7F},
			pre:     cpuData{A: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, PC: 0x8001},
			wantbus: testBus{},
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0x80_0x8F(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"ADD A,B": {
			code:    []byte{0x80},
			pre:     cpuData{A: 0x01, B: 0x41, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, B: 0x41, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,B zero": {
			code:    []byte{0x80},
			pre:     cpuData{A: 0xFF, B: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, B: 0x01, F: Z | H | CY, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,B half carry": {
			code:    []byte{0x80},
			pre:     cpuData{A: 0x0F, B: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x10, B: 0x01, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,B carry in": {
			code:    []byte{0x80},
			pre:     cpuData{A: 0x01, B: 0x41, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, B: 0x41, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,B carry out": {
			code:    []byte{0x80},
			pre:     cpuData{A: 0x3C, B: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x3B, B: 0xFF, F: CY | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,C": {
			code:    []byte{0x81},
			pre:     cpuData{A: 0x01, C: 0x41, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, C: 0x41, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,C zero": {
			code:    []byte{0x81},
			pre:     cpuData{A: 0xFF, C: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, C: 0x01, F: Z | H | CY, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,C half carry": {
			code:    []byte{0x81},
			pre:     cpuData{A: 0x0F, C: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x10, C: 0x01, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,C carry in": {
			code:    []byte{0x81},
			pre:     cpuData{A: 0x01, C: 0x41, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, C: 0x41, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,C carry out": {
			code:    []byte{0x81},
			pre:     cpuData{A: 0x3C, C: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x3B, C: 0xFF, F: CY | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,D": {
			code:    []byte{0x82},
			pre:     cpuData{A: 0x01, D: 0x41, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, D: 0x41, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,D zero": {
			code:    []byte{0x82},
			pre:     cpuData{A: 0xFF, D: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, D: 0x01, F: Z | H | CY, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,D half carry": {
			code:    []byte{0x82},
			pre:     cpuData{A: 0x0F, D: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x10, D: 0x01, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,D carry in": {
			code:    []byte{0x82},
			pre:     cpuData{A: 0x01, D: 0x41, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, D: 0x41, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,D carry out": {
			code:    []byte{0x82},
			pre:     cpuData{A: 0x3C, D: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x3B, D: 0xFF, F: CY | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,E": {
			code:    []byte{0x83},
			pre:     cpuData{A: 0x01, E: 0x41, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, E: 0x41, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,E zero": {
			code:    []byte{0x83},
			pre:     cpuData{A: 0xFF, E: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, E: 0x01, F: Z | H | CY, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,E half carry": {
			code:    []byte{0x83},
			pre:     cpuData{A: 0x0F, E: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x10, E: 0x01, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,E carry in": {
			code:    []byte{0x83},
			pre:     cpuData{A: 0x01, E: 0x41, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, E: 0x41, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,E carry out": {
			code:    []byte{0x83},
			pre:     cpuData{A: 0x3C, E: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x3B, E: 0xFF, F: CY | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,H": {
			code:    []byte{0x84},
			pre:     cpuData{A: 0x01, H: 0x41, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, H: 0x41, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,H zero": {
			code:    []byte{0x84},
			pre:     cpuData{A: 0xFF, H: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, H: 0x01, F: Z | H | CY, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,H half carry": {
			code:    []byte{0x84},
			pre:     cpuData{A: 0x0F, H: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x10, H: 0x01, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,H carry in": {
			code:    []byte{0x84},
			pre:     cpuData{A: 0x01, H: 0x41, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, H: 0x41, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,H carry out": {
			code:    []byte{0x84},
			pre:     cpuData{A: 0x3C, H: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x3B, H: 0xFF, F: CY | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,L": {
			code:    []byte{0x85},
			pre:     cpuData{A: 0x01, L: 0x41, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, L: 0x41, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,L zero": {
			code:    []byte{0x85},
			pre:     cpuData{A: 0xFF, L: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, L: 0x01, F: Z | H | CY, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,L half carry": {
			code:    []byte{0x85},
			pre:     cpuData{A: 0x0F, L: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x10, L: 0x01, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,L carry in": {
			code:    []byte{0x85},
			pre:     cpuData{A: 0x01, L: 0x41, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, L: 0x41, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,L carry out": {
			code:    []byte{0x85},
			pre:     cpuData{A: 0x3C, L: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x3B, L: 0xFF, F: CY | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,(HL)": {
			code:    []byte{0x86},
			pre:     cpuData{A: 0x01, H: 0x10, L: 0x42, PC: 0x8000},
			bus:     testBus{0x1042: 0x41},
			want:    cpuData{A: 0x42, H: 0x10, L: 0x42, PC: 0x8001},
			wantbus: testBus{0x1042: 0x41},
		},
		"ADD A,(HL) zero": {
			code:    []byte{0x86},
			pre:     cpuData{A: 0xFF, H: 0x10, L: 0x42, PC: 0x8000},
			bus:     testBus{0x1042: 0x01},
			want:    cpuData{A: 0x00, H: 0x10, L: 0x42, F: Z | H | CY, PC: 0x8001},
			wantbus: testBus{0x1042: 0x01},
		},
		"ADD A,(HL) half carry": {
			code:    []byte{0x86},
			pre:     cpuData{A: 0x0F, H: 0x10, L: 0x42, PC: 0x8000},
			bus:     testBus{0x1042: 0x01},
			want:    cpuData{A: 0x10, H: 0x10, L: 0x42, F: H, PC: 0x8001},
			wantbus: testBus{0x1042: 0x01},
		},
		"ADD A,(HL) carry in": {
			code:    []byte{0x86},
			pre:     cpuData{A: 0x01, H: 0x10, L: 0x42, F: CY, PC: 0x8000},
			bus:     testBus{0x1042: 0x41},
			want:    cpuData{A: 0x42, H: 0x10, L: 0x42, PC: 0x8001},
			wantbus: testBus{0x1042: 0x41},
		},
		"ADD A,(HL) carry out": {
			code:    []byte{0x86},
			pre:     cpuData{A: 0x3C, H: 0x10, L: 0x42, PC: 0x8000},
			bus:     testBus{0x1042: 0xFF},
			want:    cpuData{A: 0x3B, H: 0x10, L: 0x42, F: CY | H, PC: 0x8001},
			wantbus: testBus{0x1042: 0xFF},
		},
		"ADD A,A": {
			code:    []byte{0x87},
			pre:     cpuData{A: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x02, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,A zero": {
			code:    []byte{0x87},
			pre:     cpuData{A: 0x80, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, F: Z | CY, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,A half carry": {
			code:    []byte{0x87},
			pre:     cpuData{A: 0x0F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x1E, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,A carry in": {
			code:    []byte{0x87},
			pre:     cpuData{A: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x02, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADD A,A carry out": {
			code:    []byte{0x87},
			pre:     cpuData{A: 0xFE, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0xFC, F: CY | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,B": {
			code:    []byte{0x88},
			pre:     cpuData{A: 0x41, B: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, B: 0x01, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,B zero": {
			code:    []byte{0x88},
			pre:     cpuData{A: 0xFF, B: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, B: 0x01, F: Z | CY | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,B carry in": {
			code:    []byte{0x88},
			pre:     cpuData{A: 0x41, B: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x43, B: 0x01, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,B half carry": {
			code:    []byte{0x88},
			pre:     cpuData{A: 0x0F, B: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x10, B: 0x01, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,B carry out": {
			code:    []byte{0x88},
			pre:     cpuData{A: 0x3C, B: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x3B, B: 0xFF, F: CY | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,C": {
			code:    []byte{0x89},
			pre:     cpuData{A: 0x41, C: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, C: 0x01, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,C zero carry": {
			code:    []byte{0x89},
			pre:     cpuData{A: 0xFF, C: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, C: 0x01, F: Z | CY | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,C carry in": {
			code:    []byte{0x89},
			pre:     cpuData{A: 0x41, C: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x43, C: 0x01, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,C half carry": {
			code:    []byte{0x89},
			pre:     cpuData{A: 0x0F, C: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x10, C: 0x01, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,C carry out": {
			code:    []byte{0x89},
			pre:     cpuData{A: 0x3C, C: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x3B, C: 0xFF, F: CY | H, PC: 0x8001},
			wantbus: testBus{},
		},

		"ADC A,D": {
			code:    []byte{0x8A},
			pre:     cpuData{A: 0x41, D: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, D: 0x01, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,D zero carry": {
			code:    []byte{0x8A},
			pre:     cpuData{A: 0xFF, D: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, D: 0x01, F: Z | CY | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,D carry in": {
			code:    []byte{0x8A},
			pre:     cpuData{A: 0x41, D: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x43, D: 0x01, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,D half carry": {
			code:    []byte{0x8A},
			pre:     cpuData{A: 0x0F, D: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x10, D: 0x01, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,D carry out": {
			code:    []byte{0x8A},
			pre:     cpuData{A: 0x3C, D: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x3B, D: 0xFF, F: CY | H, PC: 0x8001},
			wantbus: testBus{},
		},

		"ADC A,E": {
			code:    []byte{0x8B},
			pre:     cpuData{A: 0x41, E: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, E: 0x01, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,E zero carry": {
			code:    []byte{0x8B},
			pre:     cpuData{A: 0xFF, E: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, E: 0x01, F: Z | CY | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,E carry in": {
			code:    []byte{0x8B},
			pre:     cpuData{A: 0x41, E: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x43, E: 0x01, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,E half carry": {
			code:    []byte{0x8B},
			pre:     cpuData{A: 0x0F, E: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x10, E: 0x01, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,E carry out": {
			code:    []byte{0x8B},
			pre:     cpuData{A: 0x3C, E: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x3B, E: 0xFF, F: CY | H, PC: 0x8001},
			wantbus: testBus{},
		},

		"ADC A,H": {
			code:    []byte{0x8C},
			pre:     cpuData{A: 0x41, H: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, H: 0x01, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,H zero carry": {
			code:    []byte{0x8C},
			pre:     cpuData{A: 0xFF, H: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, H: 0x01, F: Z | CY | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,H carry in": {
			code:    []byte{0x8C},
			pre:     cpuData{A: 0x41, H: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x43, H: 0x01, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,H half carry": {
			code:    []byte{0x8C},
			pre:     cpuData{A: 0x0F, H: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x10, H: 0x01, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,H carry out": {
			code:    []byte{0x8C},
			pre:     cpuData{A: 0x3C, H: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x3B, H: 0xFF, F: CY | H, PC: 0x8001},
			wantbus: testBus{},
		},

		"ADC A,L": {
			code:    []byte{0x8D},
			pre:     cpuData{A: 0x41, L: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, L: 0x01, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,L zero carry": {
			code:    []byte{0x8D},
			pre:     cpuData{A: 0xFF, L: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, L: 0x01, F: Z | CY | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,L carry in": {
			code:    []byte{0x8D},
			pre:     cpuData{A: 0x41, L: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x43, L: 0x01, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,L half carry": {
			code:    []byte{0x8D},
			pre:     cpuData{A: 0x0F, L: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x10, L: 0x01, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,L carry out": {
			code:    []byte{0x8D},
			pre:     cpuData{A: 0x3C, L: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x3B, L: 0xFF, F: CY | H, PC: 0x8001},
			wantbus: testBus{},
		},

		"ADC A,(HL)": {
			code:    []byte{0x8E},
			pre:     cpuData{A: 0x41, H: 0x41, L: 0x01, PC: 0x8000},
			bus:     testBus{0x4101: 0x01},
			want:    cpuData{A: 0x42, H: 0x41, L: 0x01, PC: 0x8001},
			wantbus: testBus{0x4101: 0x01},
		},
		"ADC A,(HL) zero carry": {
			code:    []byte{0x8E},
			pre:     cpuData{A: 0xFF, H: 0x41, L: 0x01, PC: 0x8000},
			bus:     testBus{0x4101: 0x01},
			want:    cpuData{A: 0x00, H: 0x41, L: 0x01, F: Z | CY | H, PC: 0x8001},
			wantbus: testBus{0x4101: 0x01},
		},
		"ADC A,(HL) carry in": {
			code:    []byte{0x8E},
			pre:     cpuData{A: 0x41, H: 0x41, L: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{0x4101: 0x01},
			want:    cpuData{A: 0x43, H: 0x41, L: 0x01, PC: 0x8001},
			wantbus: testBus{0x4101: 0x01},
		},
		"ADC A,(HL) half carry": {
			code:    []byte{0x8E},
			pre:     cpuData{A: 0x0F, H: 0x41, L: 0x01, PC: 0x8000},
			bus:     testBus{0x4101: 0x01},
			want:    cpuData{A: 0x10, H: 0x41, L: 0x01, F: H, PC: 0x8001},
			wantbus: testBus{0x4101: 0x01},
		},
		"ADC A,(HL) carry out": {
			code:    []byte{0x8E},
			pre:     cpuData{A: 0x3C, H: 0x41, L: 0x01, PC: 0x8000},
			bus:     testBus{0x4101: 0xFF},
			want:    cpuData{A: 0x3B, H: 0x41, L: 0x01, F: CY | H, PC: 0x8001},
			wantbus: testBus{0x4101: 0xFF},
		},

		"ADC A,A": {
			code:    []byte{0x8F},
			pre:     cpuData{A: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x02, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,A zero carry": {
			code:    []byte{0x8F},
			pre:     cpuData{A: 0x80, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, F: Z | CY, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,A carry in": {
			code:    []byte{0x8F},
			pre:     cpuData{A: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x03, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,A half carry": {
			code:    []byte{0x8F},
			pre:     cpuData{A: 0x0F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x1E, F: H, PC: 0x8001},
			wantbus: testBus{},
		},
		"ADC A,A carry out": {
			code:    []byte{0x8F},
			pre:     cpuData{A: 0xFE, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0xFC, F: CY | H, PC: 0x8001},
			wantbus: testBus{},
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0x90_0x9F(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"SUB B": {
			code:    []byte{0x90},
			pre:     cpuData{A: 0x02, B: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, B: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB B zero": {
			code:    []byte{0x90},
			pre:     cpuData{A: 0x3E, B: 0x3E, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, B: 0x3E, F: N | Z, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB B half carry": {
			code:    []byte{0x90},
			pre:     cpuData{A: 0x3E, B: 0x0F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x2F, B: 0x0F, F: N | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB B carry in": {
			code:    []byte{0x90},
			pre:     cpuData{A: 0x02, B: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, B: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB B carry out": {
			code:    []byte{0x90},
			pre:     cpuData{A: 0x3E, B: 0x40, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0xFE, B: 0x40, F: N | CY, PC: 0x8001},
			wantbus: testBus{},
		},

		"SUB C": {
			code:    []byte{0x91},
			pre:     cpuData{A: 0x02, C: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, C: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB C zero": {
			code:    []byte{0x91},
			pre:     cpuData{A: 0x3E, C: 0x3E, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, C: 0x3E, F: N | Z, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB C half carry": {
			code:    []byte{0x91},
			pre:     cpuData{A: 0x3E, C: 0x0F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x2F, C: 0x0F, F: N | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB C carry in": {
			code:    []byte{0x91},
			pre:     cpuData{A: 0x02, C: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, C: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB C carry out": {
			code:    []byte{0x91},
			pre:     cpuData{A: 0x3E, C: 0x40, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0xFE, C: 0x40, F: N | CY, PC: 0x8001},
			wantbus: testBus{},
		},

		"SUB D": {
			code:    []byte{0x92},
			pre:     cpuData{A: 0x02, D: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, D: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB D zero": {
			code:    []byte{0x92},
			pre:     cpuData{A: 0x3E, D: 0x3E, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, D: 0x3E, F: N | Z, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB D half carry": {
			code:    []byte{0x92},
			pre:     cpuData{A: 0x3E, D: 0x0F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x2F, D: 0x0F, F: N | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB D carry in": {
			code:    []byte{0x92},
			pre:     cpuData{A: 0x02, D: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, D: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB D carry out": {
			code:    []byte{0x92},
			pre:     cpuData{A: 0x3E, D: 0x40, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0xFE, D: 0x40, F: N | CY, PC: 0x8001},
			wantbus: testBus{},
		},

		"SUB E": {
			code:    []byte{0x93},
			pre:     cpuData{A: 0x02, E: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, E: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB E zero": {
			code:    []byte{0x93},
			pre:     cpuData{A: 0x3E, E: 0x3E, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, E: 0x3E, F: N | Z, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB E half carry": {
			code:    []byte{0x93},
			pre:     cpuData{A: 0x3E, E: 0x0F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x2F, E: 0x0F, F: N | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB E carry in": {
			code:    []byte{0x93},
			pre:     cpuData{A: 0x02, E: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, E: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB E carry out": {
			code:    []byte{0x93},
			pre:     cpuData{A: 0x3E, E: 0x40, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0xFE, E: 0x40, F: N | CY, PC: 0x8001},
			wantbus: testBus{},
		},

		"SUB H": {
			code:    []byte{0x94},
			pre:     cpuData{A: 0x02, H: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, H: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB H zero": {
			code:    []byte{0x94},
			pre:     cpuData{A: 0x3E, H: 0x3E, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, H: 0x3E, F: N | Z, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB H half carry": {
			code:    []byte{0x94},
			pre:     cpuData{A: 0x3E, H: 0x0F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x2F, H: 0x0F, F: N | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB H carry in": {
			code:    []byte{0x94},
			pre:     cpuData{A: 0x02, H: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, H: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB H carry out": {
			code:    []byte{0x94},
			pre:     cpuData{A: 0x3E, H: 0x40, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0xFE, H: 0x40, F: N | CY, PC: 0x8001},
			wantbus: testBus{},
		},

		"SUB L": {
			code:    []byte{0x95},
			pre:     cpuData{A: 0x02, L: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, L: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB L zero": {
			code:    []byte{0x95},
			pre:     cpuData{A: 0x3E, L: 0x3E, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, L: 0x3E, F: N | Z, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB L half carry": {
			code:    []byte{0x95},
			pre:     cpuData{A: 0x3E, L: 0x0F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x2F, L: 0x0F, F: N | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB L carry in": {
			code:    []byte{0x95},
			pre:     cpuData{A: 0x02, L: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, L: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB L carry out": {
			code:    []byte{0x95},
			pre:     cpuData{A: 0x3E, L: 0x40, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0xFE, L: 0x40, F: N | CY, PC: 0x8001},
			wantbus: testBus{},
		},

		"SUB (HL)": {
			code:    []byte{0x96},
			pre:     cpuData{A: 0x02, H: 0x40, L: 0x01, PC: 0x8000},
			bus:     testBus{0x4001: 0x01},
			want:    cpuData{A: 0x01, H: 0x40, L: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{0x4001: 0x01},
		},
		"SUB (HL) zero": {
			code:    []byte{0x96},
			pre:     cpuData{A: 0x3E, H: 0x40, L: 0x01, PC: 0x8000},
			bus:     testBus{0x4001: 0x3E},
			want:    cpuData{A: 0x00, H: 0x40, L: 0x01, F: N | Z, PC: 0x8001},
			wantbus: testBus{0x4001: 0x3E},
		},
		"SUB (HL) half carry": {
			code:    []byte{0x96},
			pre:     cpuData{A: 0x3E, H: 0x40, L: 0x01, PC: 0x8000},
			bus:     testBus{0x4001: 0x0F},
			want:    cpuData{A: 0x2F, H: 0x40, L: 0x01, F: N | H, PC: 0x8001},
			wantbus: testBus{0x4001: 0x0F},
		},
		"SUB (HL) carry in": {
			code:    []byte{0x96},
			pre:     cpuData{A: 0x02, H: 0x40, L: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{0x4001: 0x01},
			want:    cpuData{A: 0x01, H: 0x40, L: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{0x4001: 0x01},
		},
		"SUB (HL) carry out": {
			code:    []byte{0x96},
			pre:     cpuData{A: 0x3E, H: 0x40, L: 0x01, PC: 0x8000},
			bus:     testBus{0x4001: 0x40},
			want:    cpuData{A: 0xFE, H: 0x40, L: 0x01, F: N | CY, PC: 0x8001},
			wantbus: testBus{0x4001: 0x40},
		},

		"SUB A": {
			code:    []byte{0x97},
			pre:     cpuData{A: 0x02, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, F: N | Z, PC: 0x8001},
			wantbus: testBus{},
		},
		"SUB A carry in": {
			code:    []byte{0x97},
			pre:     cpuData{A: 0x02, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, F: N | Z, PC: 0x8001},
			wantbus: testBus{},
		},

		"SBC B": {
			code:    []byte{0x98},
			pre:     cpuData{A: 0x02, B: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, B: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"SBC B zero": {
			code:    []byte{0x98},
			pre:     cpuData{A: 0x3E, B: 0x3E, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, B: 0x3E, F: N | Z, PC: 0x8001},
			wantbus: testBus{},
		},
		"SBC B half carry": {
			code:    []byte{0x98},
			pre:     cpuData{A: 0x3E, B: 0x0F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x2F, B: 0x0F, F: N | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"SBC B carry in": {
			code:    []byte{0x98},
			pre:     cpuData{A: 0x03, B: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, B: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},

		"SBC C": {
			code:    []byte{0x99},
			pre:     cpuData{A: 0x02, C: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, C: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"SBC C zero": {
			code:    []byte{0x99},
			pre:     cpuData{A: 0x3E, C: 0x3E, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, C: 0x3E, F: N | Z, PC: 0x8001},
			wantbus: testBus{},
		},
		"SBC C half carry": {
			code:    []byte{0x99},
			pre:     cpuData{A: 0x3E, C: 0x0F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x2F, C: 0x0F, F: N | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"SBC C carry in": {
			code:    []byte{0x99},
			pre:     cpuData{A: 0x03, C: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, C: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},

		"SBC D": {
			code:    []byte{0x9A},
			pre:     cpuData{A: 0x02, D: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, D: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"SBC D zero": {
			code:    []byte{0x9A},
			pre:     cpuData{A: 0x3E, D: 0x3E, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, D: 0x3E, F: N | Z, PC: 0x8001},
			wantbus: testBus{},
		},
		"SBC D half carry": {
			code:    []byte{0x9A},
			pre:     cpuData{A: 0x3E, D: 0x0F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x2F, D: 0x0F, F: N | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"SBC D carry in": {
			code:    []byte{0x9A},
			pre:     cpuData{A: 0x03, D: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, D: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},

		"SBC E": {
			code:    []byte{0x9B},
			pre:     cpuData{A: 0x02, E: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, E: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"SBC E zero": {
			code:    []byte{0x9B},
			pre:     cpuData{A: 0x3E, E: 0x3E, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, E: 0x3E, F: N | Z, PC: 0x8001},
			wantbus: testBus{},
		},
		"SBC E half carry": {
			code:    []byte{0x9B},
			pre:     cpuData{A: 0x3E, E: 0x0F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x2F, E: 0x0F, F: N | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"SBC E carry in": {
			code:    []byte{0x9B},
			pre:     cpuData{A: 0x03, E: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, E: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},

		"SBC H": {
			code:    []byte{0x9C},
			pre:     cpuData{A: 0x02, H: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, H: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"SBC H zero": {
			code:    []byte{0x9C},
			pre:     cpuData{A: 0x3E, H: 0x3E, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, H: 0x3E, F: N | Z, PC: 0x8001},
			wantbus: testBus{},
		},
		"SBC H half carry": {
			code:    []byte{0x9C},
			pre:     cpuData{A: 0x3E, H: 0x0F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x2F, H: 0x0F, F: N | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"SBC H carry in": {
			code:    []byte{0x9C},
			pre:     cpuData{A: 0x03, H: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, H: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},

		"SBC L": {
			code:    []byte{0x9D},
			pre:     cpuData{A: 0x02, L: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, L: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},
		"SBC L zero": {
			code:    []byte{0x9D},
			pre:     cpuData{A: 0x3E, L: 0x3E, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, L: 0x3E, F: N | Z, PC: 0x8001},
			wantbus: testBus{},
		},
		"SBC L half carry": {
			code:    []byte{0x9D},
			pre:     cpuData{A: 0x3E, L: 0x0F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x2F, L: 0x0F, F: N | H, PC: 0x8001},
			wantbus: testBus{},
		},
		"SBC L carry in": {
			code:    []byte{0x9D},
			pre:     cpuData{A: 0x03, L: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, L: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{},
		},

		"SBC (HL)": {
			code:    []byte{0x9E},
			pre:     cpuData{A: 0x02, H: 0x40, L: 0x01, PC: 0x8000},
			bus:     testBus{0x4001: 0x01},
			want:    cpuData{A: 0x01, H: 0x40, L: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{0x4001: 0x01},
		},
		"SBC (HL) zero": {
			code:    []byte{0x9E},
			pre:     cpuData{A: 0x3E, H: 0x40, L: 0x01, PC: 0x8000},
			bus:     testBus{0x4001: 0x3E},
			want:    cpuData{A: 0x00, H: 0x40, L: 0x01, F: N | Z, PC: 0x8001},
			wantbus: testBus{0x4001: 0x3E},
		},
		"SBC (HL) half carry": {
			code:    []byte{0x9E},
			pre:     cpuData{A: 0x3E, H: 0x40, L: 0x01, PC: 0x8000},
			bus:     testBus{0x4001: 0x0F},
			want:    cpuData{A: 0x2F, H: 0x40, L: 0x01, F: N | H, PC: 0x8001},
			wantbus: testBus{0x4001: 0x0F},
		},
		"SBC (HL) carry in": {
			code:    []byte{0x9E},
			pre:     cpuData{A: 0x03, H: 0x40, L: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{0x4001: 0x01},
			want:    cpuData{A: 0x01, H: 0x40, L: 0x01, F: N, PC: 0x8001},
			wantbus: testBus{0x4001: 0x01},
		},

		"SBC A": {
			code:    []byte{0x9F},
			pre:     cpuData{A: 0x02, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x0, F: N | Z, PC: 0x8001},
			wantbus: testBus{},
		},
		"SBC A carry in": {
			code:    []byte{0x9F},
			pre:     cpuData{A: 0x03, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0xFF, F: N | H | CY, PC: 0x8001},
			wantbus: testBus{},
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0xA0_0xAF(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"AND B 00": {code: []byte{0xA0}, pre: cpuData{A: 0x00, B: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, B: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}},
		"AND B 10": {code: []byte{0xA0}, pre: cpuData{A: 0x01, B: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, B: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}},
		"AND B 01": {code: []byte{0xA0}, pre: cpuData{A: 0x00, B: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, B: 0x01, F: H | Z, PC: 0x8001}, wantbus: testBus{}},
		"AND B 11": {code: []byte{0xA0}, pre: cpuData{A: 0x01, B: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, B: 0x01, F: H, PC: 0x8001}, wantbus: testBus{}},

		"AND C 00": {code: []byte{0xA1}, pre: cpuData{A: 0x00, C: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, C: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}},
		"AND C 10": {code: []byte{0xA1}, pre: cpuData{A: 0x01, C: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, C: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}},
		"AND C 01": {code: []byte{0xA1}, pre: cpuData{A: 0x00, C: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, C: 0x01, F: H | Z, PC: 0x8001}, wantbus: testBus{}},
		"AND C 11": {code: []byte{0xA1}, pre: cpuData{A: 0x01, C: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, C: 0x01, F: H, PC: 0x8001}, wantbus: testBus{}},

		"AND D 00": {code: []byte{0xA2}, pre: cpuData{A: 0x00, D: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, D: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}},
		"AND D 10": {code: []byte{0xA2}, pre: cpuData{A: 0x01, D: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, D: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}},
		"AND D 01": {code: []byte{0xA2}, pre: cpuData{A: 0x00, D: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, D: 0x01, F: H | Z, PC: 0x8001}, wantbus: testBus{}},
		"AND D 11": {code: []byte{0xA2}, pre: cpuData{A: 0x01, D: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, D: 0x01, F: H, PC: 0x8001}, wantbus: testBus{}},

		"AND E 00": {code: []byte{0xA3}, pre: cpuData{A: 0x00, E: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, E: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}},
		"AND E 10": {code: []byte{0xA3}, pre: cpuData{A: 0x01, E: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, E: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}},
		"AND E 01": {code: []byte{0xA3}, pre: cpuData{A: 0x00, E: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, E: 0x01, F: H | Z, PC: 0x8001}, wantbus: testBus{}},
		"AND E 11": {code: []byte{0xA3}, pre: cpuData{A: 0x01, E: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, E: 0x01, F: H, PC: 0x8001}, wantbus: testBus{}},

		"AND H 00": {code: []byte{0xA4}, pre: cpuData{A: 0x00, H: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, H: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}},
		"AND H 10": {code: []byte{0xA4}, pre: cpuData{A: 0x01, H: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, H: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}},
		"AND H 01": {code: []byte{0xA4}, pre: cpuData{A: 0x00, H: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, H: 0x01, F: H | Z, PC: 0x8001}, wantbus: testBus{}},
		"AND H 11": {code: []byte{0xA4}, pre: cpuData{A: 0x01, H: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, H: 0x01, F: H, PC: 0x8001}, wantbus: testBus{}},

		"AND L 00": {code: []byte{0xA5}, pre: cpuData{A: 0x00, L: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, L: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}},
		"AND L 10": {code: []byte{0xA5}, pre: cpuData{A: 0x01, L: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, L: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}},
		"AND L 01": {code: []byte{0xA5}, pre: cpuData{A: 0x00, L: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, L: 0x01, F: H | Z, PC: 0x8001}, wantbus: testBus{}},
		"AND L 11": {code: []byte{0xA5}, pre: cpuData{A: 0x01, L: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, L: 0x01, F: H, PC: 0x8001}, wantbus: testBus{}},

		"AND (HL) 00": {code: []byte{0xA6}, pre: cpuData{A: 0x00, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x00}, want: cpuData{A: 0x00, H: 0x40, L: 0x01, F: H | Z, PC: 0x8001}, wantbus: testBus{0x4001: 0x00}},
		"AND (HL) 10": {code: []byte{0xA6}, pre: cpuData{A: 0x01, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x00}, want: cpuData{A: 0x00, H: 0x40, L: 0x01, F: H | Z, PC: 0x8001}, wantbus: testBus{0x4001: 0x00}},
		"AND (HL) 01": {code: []byte{0xA6}, pre: cpuData{A: 0x00, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x01}, want: cpuData{A: 0x00, H: 0x40, L: 0x01, F: H | Z, PC: 0x8001}, wantbus: testBus{0x4001: 0x01}},
		"AND (HL) 11": {code: []byte{0xA6}, pre: cpuData{A: 0x01, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x01}, want: cpuData{A: 0x01, H: 0x40, L: 0x01, F: H, PC: 0x8001}, wantbus: testBus{0x4001: 0x01}},

		"AND A":      {code: []byte{0xA7}, pre: cpuData{A: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, F: H, PC: 0x8001}, wantbus: testBus{}},
		"AND A zero": {code: []byte{0xA7}, pre: cpuData{A: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}},

		"XOR B 00": {code: []byte{0xA8}, pre: cpuData{A: 0x00, B: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, B: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}},
		"XOR B 10": {code: []byte{0xA8}, pre: cpuData{A: 0x01, B: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, B: 0x00, PC: 0x8001}, wantbus: testBus{}},
		"XOR B 01": {code: []byte{0xA8}, pre: cpuData{A: 0x00, B: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, B: 0x01, PC: 0x8001}, wantbus: testBus{}},
		"XOR B 11": {code: []byte{0xA8}, pre: cpuData{A: 0x01, B: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, B: 0x01, F: Z, PC: 0x8001}, wantbus: testBus{}},

		"XOR C 00": {code: []byte{0xA9}, pre: cpuData{A: 0x00, C: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, C: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}},
		"XOR C 10": {code: []byte{0xA9}, pre: cpuData{A: 0x01, C: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, C: 0x00, PC: 0x8001}, wantbus: testBus{}},
		"XOR C 01": {code: []byte{0xA9}, pre: cpuData{A: 0x00, C: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, C: 0x01, PC: 0x8001}, wantbus: testBus{}},
		"XOR C 11": {code: []byte{0xA9}, pre: cpuData{A: 0x01, C: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, C: 0x01, F: Z, PC: 0x8001}, wantbus: testBus{}},

		"XOR D 00": {code: []byte{0xAA}, pre: cpuData{A: 0x00, D: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, D: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}},
		"XOR D 10": {code: []byte{0xAA}, pre: cpuData{A: 0x01, D: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, D: 0x00, PC: 0x8001}, wantbus: testBus{}},
		"XOR D 01": {code: []byte{0xAA}, pre: cpuData{A: 0x00, D: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, D: 0x01, PC: 0x8001}, wantbus: testBus{}},
		"XOR D 11": {code: []byte{0xAA}, pre: cpuData{A: 0x01, D: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, D: 0x01, F: Z, PC: 0x8001}, wantbus: testBus{}},

		"XOR E 00": {code: []byte{0xAB}, pre: cpuData{A: 0x00, E: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, E: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}},
		"XOR E 10": {code: []byte{0xAB}, pre: cpuData{A: 0x01, E: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, E: 0x00, PC: 0x8001}, wantbus: testBus{}},
		"XOR E 01": {code: []byte{0xAB}, pre: cpuData{A: 0x00, E: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, E: 0x01, PC: 0x8001}, wantbus: testBus{}},
		"XOR E 11": {code: []byte{0xAB}, pre: cpuData{A: 0x01, E: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, E: 0x01, F: Z, PC: 0x8001}, wantbus: testBus{}},

		"XOR H 00": {code: []byte{0xAC}, pre: cpuData{A: 0x00, H: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, H: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}},
		"XOR H 10": {code: []byte{0xAC}, pre: cpuData{A: 0x01, H: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, H: 0x00, PC: 0x8001}, wantbus: testBus{}},
		"XOR H 01": {code: []byte{0xAC}, pre: cpuData{A: 0x00, H: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, H: 0x01, PC: 0x8001}, wantbus: testBus{}},
		"XOR H 11": {code: []byte{0xAC}, pre: cpuData{A: 0x01, H: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, H: 0x01, F: Z, PC: 0x8001}, wantbus: testBus{}},

		"XOR L 00": {code: []byte{0xAD}, pre: cpuData{A: 0x00, L: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, L: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}},
		"XOR L 10": {code: []byte{0xAD}, pre: cpuData{A: 0x01, L: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, L: 0x00, PC: 0x8001}, wantbus: testBus{}},
		"XOR L 01": {code: []byte{0xAD}, pre: cpuData{A: 0x00, L: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, L: 0x01, PC: 0x8001}, wantbus: testBus{}},
		"XOR L 11": {code: []byte{0xAD}, pre: cpuData{A: 0x01, L: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, L: 0x01, F: Z, PC: 0x8001}, wantbus: testBus{}},

		"XOR (HL) 00": {code: []byte{0xAE}, pre: cpuData{A: 0x00, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x00}, want: cpuData{A: 0x00, H: 0x40, L: 0x01, F: Z, PC: 0x8001}, wantbus: testBus{0x4001: 0x00}},
		"XOR (HL) 10": {code: []byte{0xAE}, pre: cpuData{A: 0x01, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x00}, want: cpuData{A: 0x01, H: 0x40, L: 0x01, PC: 0x8001}, wantbus: testBus{0x4001: 0x00}},
		"XOR (HL) 01": {code: []byte{0xAE}, pre: cpuData{A: 0x00, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x01}, want: cpuData{A: 0x01, H: 0x40, L: 0x01, PC: 0x8001}, wantbus: testBus{0x4001: 0x01}},
		"XOR (HL) 11": {code: []byte{0xAE}, pre: cpuData{A: 0x01, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x01}, want: cpuData{A: 0x00, H: 0x40, L: 0x01, F: Z, PC: 0x8001}, wantbus: testBus{0x4001: 0x01}},

		"XOR A":      {code: []byte{0xAF}, pre: cpuData{A: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}},
		"XOR A zero": {code: []byte{0xAF}, pre: cpuData{A: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0xB0_0xBF(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"OR B 00": {code: []byte{0xB0}, pre: cpuData{A: 0x00, B: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, B: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}},
		"OR B 10": {code: []byte{0xB0}, pre: cpuData{A: 0x01, B: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, B: 0x00, PC: 0x8001}, wantbus: testBus{}},
		"OR B 01": {code: []byte{0xB0}, pre: cpuData{A: 0x00, B: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, B: 0x01, PC: 0x8001}, wantbus: testBus{}},
		"OR B 11": {code: []byte{0xB0}, pre: cpuData{A: 0x01, B: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, B: 0x01, PC: 0x8001}, wantbus: testBus{}},

		"OR C 00": {code: []byte{0xB1}, pre: cpuData{A: 0x00, C: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, C: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}},
		"OR C 10": {code: []byte{0xB1}, pre: cpuData{A: 0x01, C: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, C: 0x00, PC: 0x8001}, wantbus: testBus{}},
		"OR C 01": {code: []byte{0xB1}, pre: cpuData{A: 0x00, C: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, C: 0x01, PC: 0x8001}, wantbus: testBus{}},
		"OR C 11": {code: []byte{0xB1}, pre: cpuData{A: 0x01, C: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, C: 0x01, PC: 0x8001}, wantbus: testBus{}},

		"OR D 00": {code: []byte{0xB2}, pre: cpuData{A: 0x00, D: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, D: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}},
		"OR D 10": {code: []byte{0xB2}, pre: cpuData{A: 0x01, D: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, D: 0x00, PC: 0x8001}, wantbus: testBus{}},
		"OR D 01": {code: []byte{0xB2}, pre: cpuData{A: 0x00, D: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, D: 0x01, PC: 0x8001}, wantbus: testBus{}},
		"OR D 11": {code: []byte{0xB2}, pre: cpuData{A: 0x01, D: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, D: 0x01, PC: 0x8001}, wantbus: testBus{}},

		"OR E 00": {code: []byte{0xB3}, pre: cpuData{A: 0x00, E: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, E: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}},
		"OR E 10": {code: []byte{0xB3}, pre: cpuData{A: 0x01, E: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, E: 0x00, PC: 0x8001}, wantbus: testBus{}},
		"OR E 01": {code: []byte{0xB3}, pre: cpuData{A: 0x00, E: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, E: 0x01, PC: 0x8001}, wantbus: testBus{}},
		"OR E 11": {code: []byte{0xB3}, pre: cpuData{A: 0x01, E: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, E: 0x01, PC: 0x8001}, wantbus: testBus{}},

		"OR H 00": {code: []byte{0xB4}, pre: cpuData{A: 0x00, H: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, H: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}},
		"OR H 10": {code: []byte{0xB4}, pre: cpuData{A: 0x01, H: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, H: 0x00, PC: 0x8001}, wantbus: testBus{}},
		"OR H 01": {code: []byte{0xB4}, pre: cpuData{A: 0x00, H: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, H: 0x01, PC: 0x8001}, wantbus: testBus{}},
		"OR H 11": {code: []byte{0xB4}, pre: cpuData{A: 0x01, H: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, H: 0x01, PC: 0x8001}, wantbus: testBus{}},

		"OR L 00": {code: []byte{0xB5}, pre: cpuData{A: 0x00, L: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, L: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}},
		"OR L 10": {code: []byte{0xB5}, pre: cpuData{A: 0x01, L: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, L: 0x00, PC: 0x8001}, wantbus: testBus{}},
		"OR L 01": {code: []byte{0xB5}, pre: cpuData{A: 0x00, L: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, L: 0x01, PC: 0x8001}, wantbus: testBus{}},
		"OR L 11": {code: []byte{0xB5}, pre: cpuData{A: 0x01, L: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, L: 0x01, PC: 0x8001}, wantbus: testBus{}},

		"OR (HL) 00": {code: []byte{0xB6}, pre: cpuData{A: 0x00, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x00}, want: cpuData{A: 0x00, H: 0x40, L: 0x01, F: Z, PC: 0x8001}, wantbus: testBus{0x4001: 0x00}},
		"OR (HL) 10": {code: []byte{0xB6}, pre: cpuData{A: 0x01, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x00}, want: cpuData{A: 0x01, H: 0x40, L: 0x01, PC: 0x8001}, wantbus: testBus{0x4001: 0x00}},
		"OR (HL) 01": {code: []byte{0xB6}, pre: cpuData{A: 0x00, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x01}, want: cpuData{A: 0x01, H: 0x40, L: 0x01, PC: 0x8001}, wantbus: testBus{0x4001: 0x01}},
		"OR (HL) 11": {code: []byte{0xB6}, pre: cpuData{A: 0x01, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x01}, want: cpuData{A: 0x01, H: 0x40, L: 0x01, PC: 0x8001}, wantbus: testBus{0x4001: 0x01}},

		"OR A":      {code: []byte{0xB7}, pre: cpuData{A: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, PC: 0x8001}, wantbus: testBus{}},
		"OR A zero": {code: []byte{0xB7}, pre: cpuData{A: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}},

		"CP B lt": {code: []byte{0xB8}, pre: cpuData{A: 0x42, B: 0x41, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, B: 0x41, F: N | H, PC: 0x8001}, wantbus: testBus{}},
		"CP B eq": {code: []byte{0xB8}, pre: cpuData{A: 0x42, B: 0x42, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, B: 0x42, F: N | Z, PC: 0x8001}, wantbus: testBus{}},
		"CP B gt": {code: []byte{0xB8}, pre: cpuData{A: 0x42, B: 0x43, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, B: 0x43, F: N | CY, PC: 0x8001}, wantbus: testBus{}},

		"CP C lt": {code: []byte{0xB9}, pre: cpuData{A: 0x42, C: 0x41, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, C: 0x41, F: N | H, PC: 0x8001}, wantbus: testBus{}},
		"CP C eq": {code: []byte{0xB9}, pre: cpuData{A: 0x42, C: 0x42, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, C: 0x42, F: N | Z, PC: 0x8001}, wantbus: testBus{}},
		"CP C gt": {code: []byte{0xB9}, pre: cpuData{A: 0x42, C: 0x43, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, C: 0x43, F: N | CY, PC: 0x8001}, wantbus: testBus{}},

		"CP D lt": {code: []byte{0xBA}, pre: cpuData{A: 0x42, D: 0x41, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, D: 0x41, F: N | H, PC: 0x8001}, wantbus: testBus{}},
		"CP D eq": {code: []byte{0xBA}, pre: cpuData{A: 0x42, D: 0x42, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, D: 0x42, F: N | Z, PC: 0x8001}, wantbus: testBus{}},
		"CP D gt": {code: []byte{0xBA}, pre: cpuData{A: 0x42, D: 0x43, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, D: 0x43, F: N | CY, PC: 0x8001}, wantbus: testBus{}},

		"CP E lt": {code: []byte{0xBB}, pre: cpuData{A: 0x42, E: 0x41, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, E: 0x41, F: N | H, PC: 0x8001}, wantbus: testBus{}},
		"CP E eq": {code: []byte{0xBB}, pre: cpuData{A: 0x42, E: 0x42, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, E: 0x42, F: N | Z, PC: 0x8001}, wantbus: testBus{}},
		"CP E gt": {code: []byte{0xBB}, pre: cpuData{A: 0x42, E: 0x43, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, E: 0x43, F: N | CY, PC: 0x8001}, wantbus: testBus{}},

		"CP H lt": {code: []byte{0xBC}, pre: cpuData{A: 0x42, H: 0x41, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, H: 0x41, F: N | H, PC: 0x8001}, wantbus: testBus{}},
		"CP H eq": {code: []byte{0xBC}, pre: cpuData{A: 0x42, H: 0x42, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, H: 0x42, F: N | Z, PC: 0x8001}, wantbus: testBus{}},
		"CP H gt": {code: []byte{0xBC}, pre: cpuData{A: 0x42, H: 0x43, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, H: 0x43, F: N | CY, PC: 0x8001}, wantbus: testBus{}},

		"CP L lt": {code: []byte{0xBD}, pre: cpuData{A: 0x42, L: 0x41, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, L: 0x41, F: N | H, PC: 0x8001}, wantbus: testBus{}},
		"CP L eq": {code: []byte{0xBD}, pre: cpuData{A: 0x42, L: 0x42, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, L: 0x42, F: N | Z, PC: 0x8001}, wantbus: testBus{}},
		"CP L gt": {code: []byte{0xBD}, pre: cpuData{A: 0x42, L: 0x43, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, L: 0x43, F: N | CY, PC: 0x8001}, wantbus: testBus{}},

		"CP (HL) lt": {code: []byte{0xBE}, pre: cpuData{A: 0x42, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x41}, want: cpuData{A: 0x42, H: 0x40, L: 0x01, F: N | H, PC: 0x8001}, wantbus: testBus{0x4001: 0x41}},
		"CP (HL) eq": {code: []byte{0xBE}, pre: cpuData{A: 0x42, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x42}, want: cpuData{A: 0x42, H: 0x40, L: 0x01, F: N | Z, PC: 0x8001}, wantbus: testBus{0x4001: 0x42}},
		"CP (HL) gt": {code: []byte{0xBE}, pre: cpuData{A: 0x42, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x43}, want: cpuData{A: 0x42, H: 0x40, L: 0x01, F: N | CY, PC: 0x8001}, wantbus: testBus{0x4001: 0x43}},

		"CP A eq": {code: []byte{0xBF}, pre: cpuData{A: 0x42, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, F: N | Z, PC: 0x8001}, wantbus: testBus{}},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0xC0_0xCF(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"RET NZ": {
			code:    []byte{0xC0},
			pre:     cpuData{SP: 0x0000, PC: 0x8000},
			bus:     testBus{0x0000: 0x01, 0x0001: 0x40},
			want:    cpuData{SP: 0x0002, PC: 0x4001},
			wantbus: testBus{0x0000: 0x01, 0x0001: 0x40},
		},
		"RET NZ zero": {
			code:    []byte{0xC0},
			pre:     cpuData{F: Z, SP: 0x0000, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: Z, SP: 0x0000, PC: 0x8001},
			wantbus: testBus{},
		},

		"POP BC": {
			code:    []byte{0xC1},
			pre:     cpuData{SP: 0x0000, PC: 0x8000},
			bus:     testBus{0x0000: 0x01, 0x0001: 0x40},
			want:    cpuData{B: 0x40, C: 0x01, SP: 0x0002, PC: 0x8001},
			wantbus: testBus{0x0000: 0x01, 0x0001: 0x40},
		},

		"JP NZ,a16": {
			code:    []byte{0xC2, 0x01, 0x40},
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{PC: 0x4001},
			wantbus: testBus{},
		},
		"JP NZ,a16 zero": {
			code:    []byte{0xC2, 0x01, 0x40},
			pre:     cpuData{F: Z, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: Z, PC: 0x8003},
			wantbus: testBus{},
		},

		"JP a16": {
			code:    []byte{0xC3, 0x01, 0x40},
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{PC: 0x4001},
			wantbus: testBus{},
		},

		"CALL NZ,a16": {
			code:    []byte{0xC4, 0x01, 0x40},
			pre:     cpuData{SP: 0x0008, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x0006, PC: 0x4001},
			wantbus: testBus{0x0006: 0x03, 0x0007: 0x80},
		},
		"CALL NZ,a16 zero": {
			code:    []byte{0xC4, 0x01, 0x40},
			pre:     cpuData{F: Z, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: Z, PC: 0x8003},
			wantbus: testBus{},
		},

		"PUSH BC": {
			code:    []byte{0xC5},
			pre:     cpuData{B: 0x40, C: 0x1, SP: 0x0009, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{B: 0x40, C: 0x1, SP: 0x0007, PC: 0x8001},
			wantbus: testBus{0x0008: 0x40, 0x0007: 0x1},
		},

		"ADD A,d8": {
			code:    []byte{0xC6, 0x41},
			pre:     cpuData{A: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, PC: 0x8002},
			wantbus: testBus{},
		},
		"ADD A,d8 zero": {
			code:    []byte{0xC6, 0x01},
			pre:     cpuData{A: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, F: Z | H | CY, PC: 0x8002},
			wantbus: testBus{},
		},
		"ADD A,d8 half carry": {
			code:    []byte{0xC6, 0x01},
			pre:     cpuData{A: 0x0F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x10, F: H, PC: 0x8002},
			wantbus: testBus{},
		},
		"ADD A,d8 carry in": {
			code:    []byte{0xC6, 0x41},
			pre:     cpuData{A: 0x01, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, PC: 0x8002},
			wantbus: testBus{},
		},
		"ADD A,d8 carry out": {
			code:    []byte{0xC6, 0xFF},
			pre:     cpuData{A: 0x3C, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x3B, F: CY | H, PC: 0x8002},
			wantbus: testBus{},
		},

		"RST 00H": {
			code:    []byte{0xC7},
			pre:     cpuData{SP: 0x0009, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x0007, PC: 0x0000},
			wantbus: testBus{0x0008: 0x80, 0x0007: 0x01},
		},

		"RET Z": {
			code:    []byte{0xC8},
			pre:     cpuData{SP: 0x0000, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x0000, PC: 0x8001},
			wantbus: testBus{},
		},
		"RET Z zero": {
			code:    []byte{0xC8},
			pre:     cpuData{F: Z, SP: 0x0000, PC: 0x8000},
			bus:     testBus{0x0000: 0x01, 0x0001: 0x40},
			want:    cpuData{F: Z, SP: 0x0002, PC: 0x4001},
			wantbus: testBus{0x0000: 0x01, 0x0001: 0x40},
		},

		"RET": {
			code:    []byte{0xC9},
			pre:     cpuData{SP: 0x0000, PC: 0x8000},
			bus:     testBus{0x0000: 0x01, 0x0001: 0x40},
			want:    cpuData{SP: 0x0002, PC: 0x4001},
			wantbus: testBus{0x0000: 0x01, 0x0001: 0x40},
		},

		"JP Z,a16": {
			code:    []byte{0xCA, 0x01, 0x40},
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{PC: 0x8003},
			wantbus: testBus{},
		},
		"JP Z,a16 zero": {
			code:    []byte{0xCA, 0x01, 0x40},
			pre:     cpuData{F: Z, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: Z, PC: 0x4001},
			wantbus: testBus{},
		},
		// TODO: prefix
		"CALL Z,a16": {
			code:    []byte{0xCC, 0x01, 0x40},
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{PC: 0x8003},
			wantbus: testBus{},
		},
		"CALL Z,a16 zero": {
			code:    []byte{0xCC, 0x01, 0x40},
			pre:     cpuData{F: Z, SP: 0x0008, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: Z, SP: 0x0006, PC: 0x4001},
			wantbus: testBus{0x0006: 0x03, 0x0007: 0x80},
		},

		"CALL a16": {
			code:    []byte{0xCD, 0x01, 0x40},
			pre:     cpuData{SP: 0x0008, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x0006, PC: 0x4001},
			wantbus: testBus{0x0006: 0x03, 0x0007: 0x80},
		},

		"ADC A,d8": {
			code:    []byte{0xCE, 0x01},
			pre:     cpuData{A: 0x41, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x42, PC: 0x8002},
			wantbus: testBus{},
		},
		"ADC A,d8 zero": {
			code:    []byte{0xCE, 0x01},
			pre:     cpuData{A: 0xFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, F: Z | CY | H, PC: 0x8002},
			wantbus: testBus{},
		},
		"ADC A,d8 carry in": {
			code:    []byte{0xCE, 0x01},
			pre:     cpuData{A: 0x41, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x43, PC: 0x8002},
			wantbus: testBus{},
		},
		"ADC A,d8 half carry": {
			code:    []byte{0xCE, 0x01},
			pre:     cpuData{A: 0x0F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x10, F: H, PC: 0x8002},
			wantbus: testBus{},
		},
		"ADC A,d8 carry out": {
			code:    []byte{0xCE, 0xFF},
			pre:     cpuData{A: 0x3C, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x3B, F: CY | H, PC: 0x8002},
			wantbus: testBus{},
		},

		"RST 08H": {
			code:    []byte{0xCF},
			pre:     cpuData{SP: 0x0009, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x0007, PC: 0x0008},
			wantbus: testBus{0x0008: 0x80, 0x0007: 0x01},
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0xD0_0xDF(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"RET NC": {
			code:    []byte{0xD0},
			pre:     cpuData{SP: 0x0000, PC: 0x8000},
			bus:     testBus{0x0000: 0x01, 0x0001: 0x40},
			want:    cpuData{SP: 0x0002, PC: 0x4001},
			wantbus: testBus{0x0000: 0x01, 0x0001: 0x40},
		},
		"RET NC carry": {
			code:    []byte{0xD0},
			pre:     cpuData{F: CY, SP: 0x0000, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: CY, SP: 0x0000, PC: 0x8001},
			wantbus: testBus{},
		},

		"POP DE": {
			code:    []byte{0xD1},
			pre:     cpuData{SP: 0x0000, PC: 0x8000},
			bus:     testBus{0x0000: 0x01, 0x0001: 0x40},
			want:    cpuData{D: 0x40, E: 0x01, SP: 0x0002, PC: 0x8001},
			wantbus: testBus{0x0000: 0x01, 0x0001: 0x40},
		},

		"JP NC,a16": {
			code:    []byte{0xD2, 0x01, 0x40},
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{PC: 0x4001},
			wantbus: testBus{},
		},
		"JP NC,a16 carry": {
			code:    []byte{0xD2, 0x01, 0x40},
			pre:     cpuData{F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: CY, PC: 0x8003},
			wantbus: testBus{},
		},
		// TODO: illegal
		"CALL NC,a16": {
			code:    []byte{0xD4, 0x01, 0x40},
			pre:     cpuData{SP: 0x0008, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x0006, PC: 0x4001},
			wantbus: testBus{0x0006: 0x03, 0x0007: 0x80},
		},
		"CALL NC,a16 carry": {
			code:    []byte{0xD4, 0x01, 0x40},
			pre:     cpuData{F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: CY, PC: 0x8003},
			wantbus: testBus{},
		},

		"PUSH DE": {
			code:    []byte{0xD5},
			pre:     cpuData{D: 0x40, E: 0x1, SP: 0x0009, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{D: 0x40, E: 0x1, SP: 0x0007, PC: 0x8001},
			wantbus: testBus{0x0008: 0x40, 0x0007: 0x1},
		},

		"SUB d8": {
			code:    []byte{0xD6, 0x01},
			pre:     cpuData{A: 0x02, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, F: N, PC: 0x8002},
			wantbus: testBus{},
		},
		"SUB d8 zero": {
			code:    []byte{0xD6, 0x3E},
			pre:     cpuData{A: 0x3E, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, F: N | Z, PC: 0x8002},
			wantbus: testBus{},
		},
		"SUB d8 half carry": {
			code:    []byte{0xD6, 0x0F},
			pre:     cpuData{A: 0x3E, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x2F, F: N | H, PC: 0x8002},
			wantbus: testBus{},
		},
		"SUB d8 carry in": {
			code:    []byte{0xD6, 0x01},
			pre:     cpuData{A: 0x02, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, F: N, PC: 0x8002},
			wantbus: testBus{},
		},
		"SUB d8 carry out": {
			code:    []byte{0xD6, 0x40},
			pre:     cpuData{A: 0x3E, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0xFE, F: N | CY, PC: 0x8002},
			wantbus: testBus{},
		},

		"RST 10H": {
			code:    []byte{0xD7},
			pre:     cpuData{SP: 0x0009, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x0007, PC: 0x0010},
			wantbus: testBus{0x0008: 0x80, 0x0007: 0x01},
		},

		"RET C": {
			code:    []byte{0xD8},
			pre:     cpuData{SP: 0x0000, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x0000, PC: 0x8001},
			wantbus: testBus{},
		},
		"RET C carry": {
			code:    []byte{0xD8},
			pre:     cpuData{F: CY, SP: 0x0000, PC: 0x8000},
			bus:     testBus{0x0000: 0x01, 0x0001: 0x40},
			want:    cpuData{F: CY, SP: 0x0002, PC: 0x4001},
			wantbus: testBus{0x0000: 0x01, 0x0001: 0x40},
		},

		"RETI": {
			code:    []byte{0xD9},
			pre:     cpuData{SP: 0x0000, PC: 0x8000},
			bus:     testBus{0x0000: 0x01, 0x0001: 0x40},
			want:    cpuData{SP: 0x0002, PC: 0x4001, IME: true},
			wantbus: testBus{0x0000: 0x01, 0x0001: 0x40},
		},

		"JP C,a16": {
			code:    []byte{0xDA, 0x01, 0x40},
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{PC: 0x8003},
			wantbus: testBus{},
		},
		"JP C,a16 carry": {
			code:    []byte{0xDA, 0x01, 0x40},
			pre:     cpuData{F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: CY, PC: 0x4001},
			wantbus: testBus{},
		},
		// TODO: illegal
		"CALL C,a16": {
			code:    []byte{0xDC, 0x01, 0x40},
			pre:     cpuData{PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{PC: 0x8003},
			wantbus: testBus{},
		},
		"CALL C,a16 carry": {
			code:    []byte{0xDC, 0x01, 0x40},
			pre:     cpuData{F: CY, SP: 0x0008, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: CY, SP: 0x0006, PC: 0x4001},
			wantbus: testBus{0x0006: 0x03, 0x0007: 0x80},
		},

		// TODO: illegal

		"SBC d8": {
			code:    []byte{0xDE, 0x01},
			pre:     cpuData{A: 0x02, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, F: N, PC: 0x8002},
			wantbus: testBus{},
		},
		"SBC d8 zero": {
			code:    []byte{0xDE, 0x3E},
			pre:     cpuData{A: 0x3E, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, F: N | Z, PC: 0x8002},
			wantbus: testBus{},
		},
		"SBC d8 half carry": {
			code:    []byte{0xDE, 0x0F},
			pre:     cpuData{A: 0x3E, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x2F, F: N | H, PC: 0x8002},
			wantbus: testBus{},
		},
		"SBC d8 carry in": {
			code:    []byte{0xDE, 0x01},
			pre:     cpuData{A: 0x03, F: CY, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, F: N, PC: 0x8002},
			wantbus: testBus{},
		},

		"RST 18H": {
			code:    []byte{0xDF},
			pre:     cpuData{SP: 0x0009, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x0007, PC: 0x0018},
			wantbus: testBus{0x0008: 0x80, 0x0007: 0x01},
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0xE0_0xEF(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"LDH (a8),A": {
			code:    []byte{0xE0, 0x42},
			pre:     cpuData{A: 0x41, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x41, PC: 0x8002},
			wantbus: testBus{0xFF42: 0x41},
		},

		"POP HL": {
			code:    []byte{0xE1},
			pre:     cpuData{SP: 0x0000, PC: 0x8000},
			bus:     testBus{0x0000: 0x01, 0x0001: 0x40},
			want:    cpuData{H: 0x40, L: 0x01, SP: 0x0002, PC: 0x8001},
			wantbus: testBus{0x0000: 0x01, 0x0001: 0x40},
		},

		"LD (C),A": {
			code:    []byte{0xE2},
			pre:     cpuData{A: 0x41, C: 0x42, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x41, C: 0x42, PC: 0x8001},
			wantbus: testBus{0xFF42: 0x41},
		},

		// TODO: illegal
		// TODO: illegal

		"PUSH HL": {
			code:    []byte{0xE5},
			pre:     cpuData{H: 0x40, L: 0x1, SP: 0x0009, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0x40, L: 0x1, SP: 0x0007, PC: 0x8001},
			wantbus: testBus{0x0008: 0x40, 0x0007: 0x1},
		},

		"AND d8 00": {
			code:    []byte{0xE6, 0x00},
			pre:     cpuData{A: 0x00, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, F: H | Z, PC: 0x8002},
			wantbus: testBus{},
		},
		"AND d8 10": {
			code:    []byte{0xE6, 0x00},
			pre:     cpuData{A: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, F: H | Z, PC: 0x8002},
			wantbus: testBus{},
		},
		"AND d8 01": {
			code:    []byte{0xE6, 0x01},
			pre:     cpuData{A: 0x00, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x00, F: H | Z, PC: 0x8002},
			wantbus: testBus{},
		},
		"AND d8 11": {
			code:    []byte{0xE6, 0x01},
			pre:     cpuData{A: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x01, F: H, PC: 0x8002},
			wantbus: testBus{},
		},

		"RST 20H": {
			code:    []byte{0xE7},
			pre:     cpuData{SP: 0x0009, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x0007, PC: 0x0020},
			wantbus: testBus{0x0008: 0x80, 0x0007: 0x01},
		},

		"ADD SP,r8 +42": {
			code:    []byte{0xE8, 0x42},
			pre:     cpuData{SP: 0x2000, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x2042, PC: 0x8002},
			wantbus: testBus{},
		},
		"ADD SP,r8 -42": {
			code:    []byte{0xE8, 0x42 ^ 0xFF + 1},
			pre:     cpuData{SP: 0x2000, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x1FBE, PC: 0x8002},
			wantbus: testBus{},
		},
		"ADD SP,r8 zero": {
			code:    []byte{0xE8, 0x00},
			pre:     cpuData{SP: 0x0000, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x0000, PC: 0x8002},
			wantbus: testBus{},
		},
		"ADD SP,r8 + half carry": {
			code:    []byte{0xE8, 0x01},
			pre:     cpuData{SP: 0x200F, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: H, SP: 0x2010, PC: 0x8002},
			wantbus: testBus{},
		},
		"ADD SP,r8 - half carry": {
			code:    []byte{0xE8, 0xEF},
			pre:     cpuData{SP: 0x0001, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: H, SP: 0xFFF0, PC: 0x8002},
			wantbus: testBus{},
		},
		"ADD SP,r8 127": {
			code:    []byte{0xE8, 0x7F},
			pre:     cpuData{SP: 0x2000, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x207F, PC: 0x8002},
			wantbus: testBus{},
		},
		"ADD SP,r8 -1": {
			code:    []byte{0xE8, 0xFF},
			pre:     cpuData{SP: 0x0001, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: H | CY, SP: 0x0000, PC: 0x8002},
			wantbus: testBus{},
		},
		"ADD SP,r8 -1 carry": {
			code:    []byte{0xE8, 0xFF},
			pre:     cpuData{SP: 0x0000, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0xFFFF, PC: 0x8002},
			wantbus: testBus{},
		},
		"ADD SP,r8 +1 carry": {
			code:    []byte{0xE8, 0x01},
			pre:     cpuData{SP: 0xFFFF, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{F: H | CY, SP: 0x0000, PC: 0x8002},
			wantbus: testBus{},
		},

		"JP (HL)": {
			code:    []byte{0xE9},
			pre:     cpuData{H: 0x40, L: 0x01, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{H: 0x40, L: 0x01, PC: 0x4001},
			wantbus: testBus{},
		},

		"LD (a16),A": {
			code:    []byte{0xEA, 0x01, 0x42},
			pre:     cpuData{A: 0x41, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{A: 0x41, PC: 0x8003},
			wantbus: testBus{0x4201: 0x41},
		},

		// TODO: illegal
		// TODO: illegal
		// TODO: illegal

		"XOR d8 00": {code: []byte{0xEE, 0x00}, pre: cpuData{A: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, F: Z, PC: 0x8002}, wantbus: testBus{}},
		"XOR d8 10": {code: []byte{0xEE, 0x00}, pre: cpuData{A: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, PC: 0x8002}, wantbus: testBus{}},
		"XOR d8 01": {code: []byte{0xEE, 0x01}, pre: cpuData{A: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, PC: 0x8002}, wantbus: testBus{}},
		"XOR d8 11": {code: []byte{0xEE, 0x01}, pre: cpuData{A: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, F: Z, PC: 0x8002}, wantbus: testBus{}},

		"RST 28H": {
			code:    []byte{0xEF},
			pre:     cpuData{SP: 0x0009, PC: 0x8000},
			bus:     testBus{},
			want:    cpuData{SP: 0x0007, PC: 0x0028},
			wantbus: testBus{0x0008: 0x80, 0x0007: 0x01},
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
			A:   c.A,
			F:   c.F,
			B:   c.B,
			C:   c.C,
			D:   c.D,
			E:   c.E,
			H:   c.H,
			L:   c.L,
			SP:  c.SP,
			PC:  c.PC,
			IME: c.IME,
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
