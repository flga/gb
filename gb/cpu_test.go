package gb

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

// I know this is is ugly but map literals are so nice for testing
const testBusCycleAddr uint16 = 0xFFFF

type testBus map[uint16]byte

func (b testBus) read(addr uint16) uint8 {
	if addr == testBusCycleAddr {
		panic("testBusCycleAddr is reserved")
	}
	b[testBusCycleAddr]++
	return b[addr]
}

func (b testBus) write(addr uint16, v uint8) {
	if addr == testBusCycleAddr {
		panic("testBusCycleAddr is reserved")
	}
	b[testBusCycleAddr]++
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
	debug      bool
	code       []byte
	pre        cpuData
	bus        testBus
	want       cpuData
	wantbus    testBus
	wantCycles uint8
}

// tests shamelessly stolen from mooney-gb because I'm lazy

func TestCpuOps0x00_0x0F(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"NOP": {
			code:       []byte{0x00},
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD BC,d16": {
			code:       []byte{0x01, 0x41, 0x42},
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x42, C: 0x41, PC: 0x8003},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"LD (BC), A": {
			code:       []byte{0x02},
			pre:        cpuData{A: 0x42, B: 0x00, C: 0x02, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, B: 0x00, C: 0x02, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x42},
			wantCycles: 2,
		},
		"INC BC": {
			code:       []byte{0x03},
			pre:        cpuData{B: 0x41, C: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x41, C: 0x43, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"INC BC zero": {
			code:       []byte{0x03},
			pre:        cpuData{B: 0xFF, C: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x00, C: 0x00, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"INC BC half carry": {
			code:       []byte{0x03},
			pre:        cpuData{B: 0x00, C: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x01, C: 0x00, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"INC B": {
			code:       []byte{0x04},
			pre:        cpuData{B: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x43, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"INC B zero": {
			code:       []byte{0x04},
			pre:        cpuData{B: 0xff, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x00, F: Z | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"INC B half carry": {
			code:       []byte{0x04},
			pre:        cpuData{B: 0x0f, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x10, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"DEC B": {
			code:       []byte{0x05},
			pre:        cpuData{B: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x41, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"DEC B zero": {
			code:       []byte{0x05},
			pre:        cpuData{B: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x00, F: N | Z, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"DEC B half carry": {
			code:       []byte{0x05},
			pre:        cpuData{B: 0x00, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0xFF, F: N | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD B, n": {
			code:       []byte{0x06, 0x42},
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x42, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"RLCA": {
			code:       []byte{0x07},
			pre:        cpuData{A: 0x77, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0xEE, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"RLCA carry": {
			code:       []byte{0x07},
			pre:        cpuData{A: 0xF7, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0xEF, F: CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD (a16),SP": {
			code:       []byte{0x08, 0x41, 0x42},
			pre:        cpuData{SP: 0x1424, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x1424, PC: 0x8003},
			wantbus:    testBus{0x4241: 0x24, 0x4242: 0x14},
			wantCycles: 5,
		},
		"ADD HL,BC": {
			code:       []byte{0x09},
			pre:        cpuData{B: 0x0F, C: 0xFC, H: 0x00, L: 0x03, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x0F, C: 0xFC, H: 0x0F, L: 0xFF, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"ADD HL,BC carry": {
			code:       []byte{0x09},
			pre:        cpuData{B: 0xB7, C: 0xFD, H: 0x50, L: 0x02, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0xB7, C: 0xFD, H: 0x07, L: 0xFF, F: CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"ADD HL,BC half carry": {
			code:       []byte{0x09},
			pre:        cpuData{B: 0x06, C: 0x05, H: 0x8A, L: 0x23, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x06, C: 0x05, H: 0x90, L: 0x28, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"LD A, (BC)": {
			code:       []byte{0x0A},
			pre:        cpuData{B: 0x00, C: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x42},
			want:       cpuData{A: 0x42, B: 0x00, C: 0x02, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x42},
			wantCycles: 2,
		},
		"DEC BC": {
			code:       []byte{0x0B},
			pre:        cpuData{B: 0x00, C: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x00, C: 0x41, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"DEC BC zero": {
			code:       []byte{0x0B},
			pre:        cpuData{B: 0x00, C: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x00, C: 0x00, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"DEC BC half carry": {
			code:       []byte{0x0B},
			pre:        cpuData{B: 0x01, C: 0x00, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x00, C: 0xFF, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"INC C": {
			code:       []byte{0x0C},
			pre:        cpuData{C: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{C: 0x43, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"INC C zero": {
			code:       []byte{0x0C},
			pre:        cpuData{C: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{C: 0x00, F: Z | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"INC C half carry": {
			code:       []byte{0x0C},
			pre:        cpuData{C: 0x0F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{C: 0x10, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"DEC C": {
			code:       []byte{0x0D},
			pre:        cpuData{C: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{C: 0x41, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"DEC C zero": {
			code:       []byte{0x0D},
			pre:        cpuData{C: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{C: 0x00, F: N | Z, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"DEC C half carry": {
			code:       []byte{0x0D},
			pre:        cpuData{C: 0x00, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{C: 0xFF, F: N | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD C, n": {
			code:       []byte{0x0E, 0x42},
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{C: 0x42, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"RRCA": {
			code:       []byte{0x0F},
			pre:        cpuData{A: 0xEE, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x77, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"RRCA carry": {
			code:       []byte{0x0F},
			pre:        cpuData{A: 0xEF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0xF7, F: CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
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
			code:       []byte{0x11, 0x41, 0x42},
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{D: 0x42, E: 0x41, PC: 0x8003},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"LD (DE), A": {
			code:       []byte{0x12},
			pre:        cpuData{A: 0x42, D: 0x00, E: 0x02, PC: 0x8000},
			bus:        testBus{0x02: 0x00},
			want:       cpuData{A: 0x42, D: 0x00, E: 0x02, PC: 0x8001},
			wantbus:    testBus{0x02: 0x42},
			wantCycles: 2,
		},
		"INC DE": {
			code:       []byte{0x13},
			pre:        cpuData{D: 0x41, E: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{D: 0x41, E: 0x43, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"INC DE zero": {
			code:       []byte{0x13},
			pre:        cpuData{D: 0xFF, E: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{D: 0x00, E: 0x00, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"INC DE half carry": {
			code:       []byte{0x13},
			pre:        cpuData{D: 0x00, E: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{D: 0x01, E: 0x00, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"INC D": {
			code:       []byte{0x14},
			pre:        cpuData{D: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{D: 0x43, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"INC D zero": {
			code:       []byte{0x14},
			pre:        cpuData{D: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{D: 0x00, F: Z | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"INC D half carry": {
			code:       []byte{0x14},
			pre:        cpuData{D: 0x0F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{D: 0x10, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"DEC D": {
			code:       []byte{0x15},
			pre:        cpuData{D: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{D: 0x41, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"DEC D zero": {
			code:       []byte{0x15},
			pre:        cpuData{D: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{D: 0x00, F: N | Z, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"DEC D half carry": {
			code:       []byte{0x15},
			pre:        cpuData{D: 0x00, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{D: 0xFF, F: N | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD D, n": {
			code:       []byte{0x16, 0x42},
			pre:        cpuData{D: 0x00, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{D: 0x42, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"RLA": {
			code:       []byte{0x17},
			pre:        cpuData{A: 0x55, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0xAA, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"RLA carry": {
			code:       []byte{0x17},
			pre:        cpuData{A: 0xAA, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x55, F: CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"JR e": {
			code:       []byte{0x18, 0x02},
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{PC: 0x8004},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"JR e negative offset": {
			code:       []byte{0x18, 0xFD}, // 0xFD = -3
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{PC: 0x7FFF},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"ADD HL,DE": {
			code:       []byte{0x19},
			pre:        cpuData{D: 0x0F, E: 0xFC, H: 0x00, L: 0x03, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{D: 0x0F, E: 0xFC, H: 0x0F, L: 0xFF, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"ADD HL,DE carry": {
			code:       []byte{0x19},
			pre:        cpuData{D: 0xB7, E: 0xFD, H: 0x50, L: 0x02, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{D: 0xB7, E: 0xFD, H: 0x07, L: 0xFF, F: CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"ADD HL,DE half carry": {
			code:       []byte{0x19},
			pre:        cpuData{D: 0x06, E: 0x05, H: 0x8A, L: 0x23, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{D: 0x06, E: 0x05, H: 0x90, L: 0x28, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"LD A, (DE)": {
			code:       []byte{0x1A},
			pre:        cpuData{A: 0x00, D: 0x00, E: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x42},
			want:       cpuData{A: 0x42, D: 0x00, E: 0x02, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x42},
			wantCycles: 2,
		},
		"DEC DE": {
			code:       []byte{0x1B},
			pre:        cpuData{D: 0x00, E: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{D: 0x00, E: 0x41, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"DEC DE zero": {
			code:       []byte{0x1B},
			pre:        cpuData{D: 0x00, E: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{D: 0x00, E: 0x00, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"DEC DE half carry": {
			code:       []byte{0x1B},
			pre:        cpuData{D: 0x01, E: 0x00, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{D: 0x00, E: 0xFF, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"INC E": {
			code:       []byte{0x1C},
			pre:        cpuData{E: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{E: 0x43, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"INC E zero": {
			code:       []byte{0x1C},
			pre:        cpuData{E: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{E: 0x00, F: Z | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"INC E half carry": {
			code:       []byte{0x1C},
			pre:        cpuData{E: 0x0F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{E: 0x10, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"DEC E": {
			code:       []byte{0x1D},
			pre:        cpuData{E: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{E: 0x41, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"DEC E zero": {
			code:       []byte{0x1D},
			pre:        cpuData{E: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{E: 0x00, F: N | Z, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"DEC E half carry": {
			code:       []byte{0x1D},
			pre:        cpuData{E: 0x00, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{E: 0xFF, F: N | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD E, n": {
			code:       []byte{0x1E, 0x42},
			pre:        cpuData{E: 0x00, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{E: 0x42, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"RRA": {
			code:       []byte{0x1F},
			pre:        cpuData{A: 0xAA, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x55, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"RRA carry": {
			code:       []byte{0x1F},
			pre:        cpuData{A: 0x55, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0xAA, F: CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0x20_0x2F(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"JR NZ, e": {
			code:       []byte{0x20, 0x02},
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{PC: 0x8004},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"JR NZ, e negative offset": {
			code:       []byte{0x20, 0xFD}, // 0xFD = -3
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{PC: 0x7FFF},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"JR NZ, NO JUMP": {
			code:       []byte{0x20, 0xFD}, // 0xFD = -3
			pre:        cpuData{F: Z, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: Z, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"LD HL,d16": {
			code:       []byte{0x21, 0x41, 0x42},
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x42, L: 0x41, PC: 0x8003},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"LDI (HL+), A": {
			code:       []byte{0x22},
			pre:        cpuData{A: 0x42, H: 0x00, L: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x00},
			want:       cpuData{A: 0x42, H: 0x00, L: 0x03, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x42},
			wantCycles: 2,
		},
		"INC HL": {
			code:       []byte{0x23},
			pre:        cpuData{H: 0x41, L: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x41, L: 0x43, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"INC HL zero": {
			code:       []byte{0x23},
			pre:        cpuData{H: 0xFF, L: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x00, L: 0x00, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"INC HL half carry": {
			code:       []byte{0x23},
			pre:        cpuData{H: 0x00, L: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x01, L: 0x00, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"INC H": {
			code:       []byte{0x24},
			pre:        cpuData{H: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x43, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"INC H zero": {
			code:       []byte{0x24},
			pre:        cpuData{H: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x00, F: Z | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"INC H half carry": {
			code:       []byte{0x24},
			pre:        cpuData{H: 0x0F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x10, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"DEC H": {
			code:       []byte{0x25},
			pre:        cpuData{H: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x41, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"DEC H zero": {
			code:       []byte{0x25},
			pre:        cpuData{H: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x00, F: Z | N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"DEC H half carry": {
			code:       []byte{0x25},
			pre:        cpuData{H: 0x00, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0xFF, F: N | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD H, n": {
			code:       []byte{0x26, 0x42},
			pre:        cpuData{H: 0x00, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x42, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		// TODO: DAA
		"JR Z, e": {
			code:       []byte{0x28, 0x02},
			pre:        cpuData{F: Z, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: Z, PC: 0x8004},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"JR Z, e negative offset": {
			code:       []byte{0x28, 0xFD}, // 0xFD = -3
			pre:        cpuData{F: Z, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: Z, PC: 0x7FFF},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"JR Z, NO JUMP": {
			code:       []byte{0x28, 0xFD}, // 0xFD = -3
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"ADD HL,HL": {
			code:       []byte{0x29},
			pre:        cpuData{H: 0x02, L: 0xAA, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x05, L: 0x54, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"ADD HL,HL carry": {
			code:       []byte{0x29},
			pre:        cpuData{H: 0x80, L: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x00, L: 0x02, F: CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"ADD HL,HL half carry": {
			code:       []byte{0x29},
			pre:        cpuData{H: 0x0F, L: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x1F, L: 0xFE, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"LD A, (HL+)": {
			code:       []byte{0x2A},
			pre:        cpuData{A: 0x00, H: 0x00, L: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x42},
			want:       cpuData{A: 0x42, H: 0x00, L: 0x03, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x42},
			wantCycles: 2,
		},
		"DEC HL": {
			code:       []byte{0x2B},
			pre:        cpuData{H: 0x00, L: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x00, L: 0x41, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"DEC HL zero": {
			code:       []byte{0x2B},
			pre:        cpuData{H: 0x00, L: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x00, L: 0x00, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"DEC HL half carry": {
			code:       []byte{0x2B},
			pre:        cpuData{H: 0x01, L: 0x00, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x00, L: 0xFF, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"INC L": {
			code:       []byte{0x2C},
			pre:        cpuData{L: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{L: 0x43, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"INC L zero": {
			code:       []byte{0x2C},
			pre:        cpuData{L: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{L: 0x00, F: Z | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"INC L half carry": {
			code:       []byte{0x2C},
			pre:        cpuData{L: 0x0F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{L: 0x10, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"DEC L": {
			code:       []byte{0x2D},
			pre:        cpuData{L: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{L: 0x41, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"DEC L zero": {
			code:       []byte{0x2D},
			pre:        cpuData{L: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{L: 0x00, F: N | Z, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"DEC L half carry": {
			code:       []byte{0x2D},
			pre:        cpuData{L: 0x00, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{L: 0xFF, F: N | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD L, n": {
			code:       []byte{0x2E, 0x42},
			pre:        cpuData{L: 0x00, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{L: 0x42, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"CPL": {
			code:       []byte{0x2F},
			pre:        cpuData{A: 0xAA, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x55, F: N | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0x30_0x3F(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"JR NC, e": {
			code:       []byte{0x30, 0x02},
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{PC: 0x8004},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"JR NC, e negative offset": {
			code:       []byte{0x30, 0xFD}, // 0xFD = -3
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{PC: 0x7FFF},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"JR NC, NO JUMP": {
			code:       []byte{0x30, 0xFD}, // 0xFD = -3
			pre:        cpuData{F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: CY, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"LD SP,d16": {
			code:       []byte{0x31, 0x41, 0x42},
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x4241, PC: 0x8003},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"LDD (HL-), A": {
			code:       []byte{0x32},
			pre:        cpuData{A: 0x42, H: 0x00, L: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x00},
			want:       cpuData{A: 0x42, H: 0x00, L: 0x01, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x42},
			wantCycles: 2,
		},
		"INC SP": {
			code:       []byte{0x33},
			pre:        cpuData{SP: 0x4142, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x4143, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"INC SP zero": {
			code:       []byte{0x33},
			pre:        cpuData{SP: 0xFFFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x0000, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"INC SP half carry": {
			code:       []byte{0x33},
			pre:        cpuData{SP: 0x00FF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x0100, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"INC (HL)": {
			code:       []byte{0x34},
			pre:        cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x42},
			want:       cpuData{H: 0x00, L: 0x02, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x43},
			wantCycles: 3,
		},
		"INC (HL) zero": {
			code:       []byte{0x34},
			pre:        cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0xFF},
			want:       cpuData{H: 0x00, L: 0x02, F: Z | H, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x00},
			wantCycles: 3,
		},
		"INC (HL) half carry": {
			code:       []byte{0x34},
			pre:        cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x0F},
			want:       cpuData{H: 0x00, L: 0x02, F: H, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x10},
			wantCycles: 3,
		},
		"DEC (HL)": {
			code:       []byte{0x35},
			pre:        cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x42},
			want:       cpuData{H: 0x00, L: 0x02, F: N, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x41},
			wantCycles: 3,
		},
		"DEC (HL) zero": {
			code:       []byte{0x35},
			pre:        cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x01},
			want:       cpuData{H: 0x00, L: 0x02, F: N | Z, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x00},
			wantCycles: 3,
		},
		"DEC (HL) half carry": {
			code:       []byte{0x35},
			pre:        cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x00},
			want:       cpuData{H: 0x00, L: 0x02, F: N | H, PC: 0x8001},
			wantbus:    testBus{0x0002: 0xFF},
			wantCycles: 3,
		},
		"LD (HL), n": {
			code:       []byte{0x36, 0x42},
			pre:        cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x00},
			want:       cpuData{H: 0x00, L: 0x02, PC: 0x8002},
			wantbus:    testBus{0x0002: 0x42},
			wantCycles: 3,
		},
		"SCF": {
			code:       []byte{0x37},
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SCF carry": {
			code:       []byte{0x37},
			pre:        cpuData{F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"JR C, e": {
			code:       []byte{0x38, 0x02},
			pre:        cpuData{F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: CY, PC: 0x8004},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"JR C, e negative offset": {
			code:       []byte{0x38, 0xFD}, // 0xFD = -3
			pre:        cpuData{F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: CY, PC: 0x7FFF},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"JR C, NO JUMP": {
			code:       []byte{0x38, 0xFD}, // 0xFD = -3
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"ADD HL,SP": {
			code:       []byte{0x39},
			pre:        cpuData{SP: 0x0FFC, H: 0x00, L: 0x03, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x0FFC, H: 0x0F, L: 0xFF, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"ADD HL,SP carry": {
			code:       []byte{0x39},
			pre:        cpuData{SP: 0xB7FD, H: 0x50, L: 0x02, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0xB7FD, H: 0x07, L: 0xFF, F: CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"ADD HL,SP half carry": {
			code:       []byte{0x39},
			pre:        cpuData{SP: 0x0605, H: 0x8A, L: 0x23, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x0605, H: 0x90, L: 0x28, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"LD A, (HL-)": {
			code:       []byte{0x3A},
			pre:        cpuData{A: 0x00, H: 0x00, L: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x42},
			want:       cpuData{A: 0x42, H: 0x00, L: 0x01, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x42},
			wantCycles: 2,
		},
		"DEC SP": {
			code:       []byte{0x3B},
			pre:        cpuData{SP: 0x0042, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x0041, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"DEC SP zero": {
			code:       []byte{0x3B},
			pre:        cpuData{SP: 0x0001, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x0000, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"DEC SP half carry": {
			code:       []byte{0x3B},
			pre:        cpuData{SP: 0x0100, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x00FF, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"INC A": {
			code:       []byte{0x3C},
			pre:        cpuData{A: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x43, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"INC A zero": {
			code:       []byte{0x3C},
			pre:        cpuData{A: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, F: Z | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"INC A half carry": {
			code:       []byte{0x3C},
			pre:        cpuData{A: 0x0F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x10, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"DEC A": {
			code:       []byte{0x3D},
			pre:        cpuData{A: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x41, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"DEC A zero": {
			code:       []byte{0x3D},
			pre:        cpuData{A: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, F: N | Z, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"DEC A half carry": {
			code:       []byte{0x3D},
			pre:        cpuData{A: 0x00, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0xFF, F: N | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD A, n": {
			code:       []byte{0x3E, 0x42},
			pre:        cpuData{A: 0x00, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"CCF": {
			code:       []byte{0x3F},
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"CCF carry": {
			code:       []byte{0x3F},
			pre:        cpuData{F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0x40_0x4F(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"LD B, B": {
			code:       []byte{0x40},
			pre:        cpuData{B: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD B, C": {
			code:       []byte{0x41},
			pre:        cpuData{B: 0x00, C: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x42, C: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD B, D": {
			code:       []byte{0x42},
			pre:        cpuData{B: 0x00, D: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x42, D: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD B, E": {
			code:       []byte{0x43},
			pre:        cpuData{B: 0x00, E: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x42, E: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD B, H": {
			code:       []byte{0x44},
			pre:        cpuData{B: 0x00, H: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x42, H: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD B, L": {
			code:       []byte{0x45},
			pre:        cpuData{B: 0x00, L: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x42, L: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD B, (HL)": {
			code:       []byte{0x46},
			pre:        cpuData{B: 0x00, H: 0x00, L: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x42},
			want:       cpuData{B: 0x42, H: 0x00, L: 0x02, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x42},
			wantCycles: 2,
		},
		"LD B, A": {
			code: []byte{0x47},
			pre:  cpuData{A: 0x42, B: 0x00, PC: 0x8000}, bus: testBus{},
			want:       cpuData{A: 0x42, B: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD C, B": {
			code: []byte{0x48},
			pre:  cpuData{B: 0x42, C: 0x00, PC: 0x8000}, bus: testBus{},
			want:       cpuData{B: 0x42, C: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD C, C": {
			code:       []byte{0x49},
			pre:        cpuData{C: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{C: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD C, D": {
			code: []byte{0x4A},
			pre:  cpuData{C: 0x00, D: 0x42, PC: 0x8000}, bus: testBus{},
			want:       cpuData{C: 0x42, D: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD C, E": {
			code: []byte{0x4B},
			pre:  cpuData{C: 0x00, E: 0x42, PC: 0x8000}, bus: testBus{},
			want:       cpuData{C: 0x42, E: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD C, H": {
			code: []byte{0x4C},
			pre:  cpuData{C: 0x00, H: 0x42, PC: 0x8000}, bus: testBus{},
			want:       cpuData{C: 0x42, H: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD C, L": {
			code: []byte{0x4D},
			pre:  cpuData{C: 0x00, L: 0x42, PC: 0x8000}, bus: testBus{},
			want:       cpuData{C: 0x42, L: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD C, (HL)": {
			code: []byte{0x4E},
			pre:  cpuData{C: 0x00, H: 0x00, L: 0x02, PC: 0x8000}, bus: testBus{0x0002: 0x42},
			want:       cpuData{C: 0x42, H: 0x00, L: 0x02, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x42},
			wantCycles: 2,
		},
		"LD C, A": {
			code: []byte{0x4F},
			pre:  cpuData{A: 0x42, C: 0x00, PC: 0x8000}, bus: testBus{},
			want:       cpuData{A: 0x42, C: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0x50_0x5F(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"LD D, B": {
			code:       []byte{0x50},
			pre:        cpuData{B: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x42, D: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD D, C": {
			code:       []byte{0x51},
			pre:        cpuData{C: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{C: 0x42, D: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD D, D": {
			code:       []byte{0x52},
			pre:        cpuData{D: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{D: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD D, E": {
			code:       []byte{0x53},
			pre:        cpuData{E: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{E: 0x42, D: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD D, H": {
			code:       []byte{0x54},
			pre:        cpuData{H: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x42, D: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD D, L": {
			code:       []byte{0x55},
			pre:        cpuData{L: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{L: 0x42, D: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD D, (HL)": {
			code:       []byte{0x56},
			pre:        cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x42},
			want:       cpuData{D: 0x42, H: 0x00, L: 0x02, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x42},
			wantCycles: 2,
		},
		"LD D, A": {
			code:       []byte{0x57},
			pre:        cpuData{A: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, D: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD E, B": {
			code:       []byte{0x58},
			pre:        cpuData{B: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x42, E: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD E, C": {
			code:       []byte{0x59},
			pre:        cpuData{C: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{C: 0x42, E: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD E, D": {
			code:       []byte{0x5A},
			pre:        cpuData{D: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{D: 0x42, E: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD E, E": {
			code:       []byte{0x5B},
			pre:        cpuData{E: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{E: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD E, H": {
			code:       []byte{0x5C},
			pre:        cpuData{H: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x42, E: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD E, L": {
			code:       []byte{0x5D},
			pre:        cpuData{L: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{L: 0x42, E: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD E, (HL)": {
			code:       []byte{0x5E},
			pre:        cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x42},
			want:       cpuData{H: 0x00, L: 0x02, E: 0x42, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x42},
			wantCycles: 2,
		},
		"LD E, A": {
			code:       []byte{0x5F},
			pre:        cpuData{A: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, E: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0x60_0x6F(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"LD H, B": {
			code:       []byte{0x60},
			pre:        cpuData{B: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x42, H: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD H, C": {
			code:       []byte{0x61},
			pre:        cpuData{C: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{C: 0x42, H: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD H, D": {
			code:       []byte{0x62},
			pre:        cpuData{D: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{D: 0x42, H: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD H, E": {
			code:       []byte{0x63},
			pre:        cpuData{E: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{E: 0x42, H: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD H, H": {
			code:       []byte{0x64},
			pre:        cpuData{H: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD H, L": {
			code:       []byte{0x65},
			pre:        cpuData{L: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{L: 0x42, H: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD H, (HL)": {
			code:       []byte{0x66},
			pre:        cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x42},
			want:       cpuData{H: 0x42, L: 0x02, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x42},
			wantCycles: 2,
		},
		"LD H, A": {
			code:       []byte{0x67},
			pre:        cpuData{A: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, H: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD L, B": {
			code:       []byte{0x68},
			pre:        cpuData{B: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x42, L: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD L, C": {
			code:       []byte{0x69},
			pre:        cpuData{C: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{C: 0x42, L: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD L, D": {
			code:       []byte{0x6A},
			pre:        cpuData{D: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{D: 0x42, L: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD L, E": {
			code:       []byte{0x6B},
			pre:        cpuData{E: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{E: 0x42, L: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD L, H": {
			code:       []byte{0x6C},
			pre:        cpuData{H: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x42, L: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD L, L": {
			code:       []byte{0x6D},
			pre:        cpuData{L: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{L: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD L, (HL)": {
			code:       []byte{0x6E},
			pre:        cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x42},
			want:       cpuData{H: 0x00, L: 0x42, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x42},
			wantCycles: 2,
		},
		"LD L, A": {
			code:       []byte{0x6F},
			pre:        cpuData{A: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, L: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0x70_0x7F(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"LD (HL), B": {
			code:       []byte{0x70},
			pre:        cpuData{B: 0x42, H: 0x00, L: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x00},
			want:       cpuData{B: 0x42, H: 0x00, L: 0x02, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x42},
			wantCycles: 2,
		},
		"LD (HL), C": {
			code:       []byte{0x71},
			pre:        cpuData{C: 0x42, H: 0x00, L: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x00},
			want:       cpuData{C: 0x42, H: 0x00, L: 0x02, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x42},
			wantCycles: 2,
		},
		"LD (HL), D": {
			code:       []byte{0x72},
			pre:        cpuData{D: 0x42, H: 0x00, L: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x00},
			want:       cpuData{D: 0x42, H: 0x00, L: 0x02, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x42},
			wantCycles: 2,
		},
		"LD (HL), E": {
			code:       []byte{0x73},
			pre:        cpuData{E: 0x42, H: 0x00, L: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x00},
			want:       cpuData{E: 0x42, H: 0x00, L: 0x02, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x42},
			wantCycles: 2,
		},
		"LD (HL), H": {
			code:       []byte{0x74},
			pre:        cpuData{H: 0x42, L: 0x02, PC: 0x8000},
			bus:        testBus{0x4202: 0x42},
			want:       cpuData{H: 0x42, L: 0x02, PC: 0x8001},
			wantbus:    testBus{0x4202: 0x42},
			wantCycles: 2,
		},
		"LD (HL), L": {
			code:       []byte{0x75},
			pre:        cpuData{H: 0x00, L: 0x42, PC: 0x8000},
			bus:        testBus{0x0042: 0x00},
			want:       cpuData{H: 0x00, L: 0x42, PC: 0x8001},
			wantbus:    testBus{0x0042: 0x42},
			wantCycles: 2,
		},
		// TODO: halt
		"LD (HL), A": {
			code:       []byte{0x77},
			pre:        cpuData{A: 0x42, H: 0x00, L: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x00},
			want:       cpuData{A: 0x42, H: 0x00, L: 0x02, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x42},
			wantCycles: 2,
		},
		"LD A, B": {
			code:       []byte{0x78},
			pre:        cpuData{B: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, B: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD A, C": {
			code:       []byte{0x79},
			pre:        cpuData{C: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, C: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD A, D": {
			code:       []byte{0x7A},
			pre:        cpuData{D: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, D: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD A, E": {
			code:       []byte{0x7B},
			pre:        cpuData{E: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, E: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD A, H": {
			code:       []byte{0x7C},
			pre:        cpuData{H: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, H: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD A, L": {
			code:       []byte{0x7D},
			pre:        cpuData{L: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, L: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"LD A, (HL)": {
			code:       []byte{0x7E},
			pre:        cpuData{H: 0x00, L: 0x02, PC: 0x8000},
			bus:        testBus{0x0002: 0x42},
			want:       cpuData{A: 0x42, H: 0x00, L: 0x02, PC: 0x8001},
			wantbus:    testBus{0x0002: 0x42},
			wantCycles: 2,
		},
		"LD A, A": {
			code:       []byte{0x7F},
			pre:        cpuData{A: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0x80_0x8F(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"ADD A,B": {
			code:       []byte{0x80},
			pre:        cpuData{A: 0x01, B: 0x41, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, B: 0x41, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,B zero": {
			code:       []byte{0x80},
			pre:        cpuData{A: 0xFF, B: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, B: 0x01, F: Z | H | CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,B half carry": {
			code:       []byte{0x80},
			pre:        cpuData{A: 0x0F, B: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x10, B: 0x01, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,B carry in": {
			code:       []byte{0x80},
			pre:        cpuData{A: 0x01, B: 0x41, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, B: 0x41, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,B carry out": {
			code:       []byte{0x80},
			pre:        cpuData{A: 0x3C, B: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x3B, B: 0xFF, F: CY | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,C": {
			code:       []byte{0x81},
			pre:        cpuData{A: 0x01, C: 0x41, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, C: 0x41, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,C zero": {
			code:       []byte{0x81},
			pre:        cpuData{A: 0xFF, C: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, C: 0x01, F: Z | H | CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,C half carry": {
			code:       []byte{0x81},
			pre:        cpuData{A: 0x0F, C: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x10, C: 0x01, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,C carry in": {
			code:       []byte{0x81},
			pre:        cpuData{A: 0x01, C: 0x41, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, C: 0x41, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,C carry out": {
			code:       []byte{0x81},
			pre:        cpuData{A: 0x3C, C: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x3B, C: 0xFF, F: CY | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,D": {
			code:       []byte{0x82},
			pre:        cpuData{A: 0x01, D: 0x41, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, D: 0x41, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,D zero": {
			code:       []byte{0x82},
			pre:        cpuData{A: 0xFF, D: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, D: 0x01, F: Z | H | CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,D half carry": {
			code:       []byte{0x82},
			pre:        cpuData{A: 0x0F, D: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x10, D: 0x01, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,D carry in": {
			code:       []byte{0x82},
			pre:        cpuData{A: 0x01, D: 0x41, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, D: 0x41, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,D carry out": {
			code:       []byte{0x82},
			pre:        cpuData{A: 0x3C, D: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x3B, D: 0xFF, F: CY | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,E": {
			code:       []byte{0x83},
			pre:        cpuData{A: 0x01, E: 0x41, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, E: 0x41, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,E zero": {
			code:       []byte{0x83},
			pre:        cpuData{A: 0xFF, E: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, E: 0x01, F: Z | H | CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,E half carry": {
			code:       []byte{0x83},
			pre:        cpuData{A: 0x0F, E: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x10, E: 0x01, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,E carry in": {
			code:       []byte{0x83},
			pre:        cpuData{A: 0x01, E: 0x41, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, E: 0x41, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,E carry out": {
			code:       []byte{0x83},
			pre:        cpuData{A: 0x3C, E: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x3B, E: 0xFF, F: CY | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,H": {
			code:       []byte{0x84},
			pre:        cpuData{A: 0x01, H: 0x41, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, H: 0x41, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,H zero": {
			code:       []byte{0x84},
			pre:        cpuData{A: 0xFF, H: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, H: 0x01, F: Z | H | CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,H half carry": {
			code:       []byte{0x84},
			pre:        cpuData{A: 0x0F, H: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x10, H: 0x01, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,H carry in": {
			code:       []byte{0x84},
			pre:        cpuData{A: 0x01, H: 0x41, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, H: 0x41, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,H carry out": {
			code:       []byte{0x84},
			pre:        cpuData{A: 0x3C, H: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x3B, H: 0xFF, F: CY | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,L": {
			code:       []byte{0x85},
			pre:        cpuData{A: 0x01, L: 0x41, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, L: 0x41, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,L zero": {
			code:       []byte{0x85},
			pre:        cpuData{A: 0xFF, L: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, L: 0x01, F: Z | H | CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,L half carry": {
			code:       []byte{0x85},
			pre:        cpuData{A: 0x0F, L: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x10, L: 0x01, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,L carry in": {
			code:       []byte{0x85},
			pre:        cpuData{A: 0x01, L: 0x41, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, L: 0x41, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,L carry out": {
			code:       []byte{0x85},
			pre:        cpuData{A: 0x3C, L: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x3B, L: 0xFF, F: CY | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,(HL)": {
			code:       []byte{0x86},
			pre:        cpuData{A: 0x01, H: 0x10, L: 0x42, PC: 0x8000},
			bus:        testBus{0x1042: 0x41},
			want:       cpuData{A: 0x42, H: 0x10, L: 0x42, PC: 0x8001},
			wantbus:    testBus{0x1042: 0x41},
			wantCycles: 2,
		},
		"ADD A,(HL) zero": {
			code:       []byte{0x86},
			pre:        cpuData{A: 0xFF, H: 0x10, L: 0x42, PC: 0x8000},
			bus:        testBus{0x1042: 0x01},
			want:       cpuData{A: 0x00, H: 0x10, L: 0x42, F: Z | H | CY, PC: 0x8001},
			wantbus:    testBus{0x1042: 0x01},
			wantCycles: 2,
		},
		"ADD A,(HL) half carry": {
			code:       []byte{0x86},
			pre:        cpuData{A: 0x0F, H: 0x10, L: 0x42, PC: 0x8000},
			bus:        testBus{0x1042: 0x01},
			want:       cpuData{A: 0x10, H: 0x10, L: 0x42, F: H, PC: 0x8001},
			wantbus:    testBus{0x1042: 0x01},
			wantCycles: 2,
		},
		"ADD A,(HL) carry in": {
			code:       []byte{0x86},
			pre:        cpuData{A: 0x01, H: 0x10, L: 0x42, F: CY, PC: 0x8000},
			bus:        testBus{0x1042: 0x41},
			want:       cpuData{A: 0x42, H: 0x10, L: 0x42, PC: 0x8001},
			wantbus:    testBus{0x1042: 0x41},
			wantCycles: 2,
		},
		"ADD A,(HL) carry out": {
			code:       []byte{0x86},
			pre:        cpuData{A: 0x3C, H: 0x10, L: 0x42, PC: 0x8000},
			bus:        testBus{0x1042: 0xFF},
			want:       cpuData{A: 0x3B, H: 0x10, L: 0x42, F: CY | H, PC: 0x8001},
			wantbus:    testBus{0x1042: 0xFF},
			wantCycles: 2,
		},
		"ADD A,A": {
			code:       []byte{0x87},
			pre:        cpuData{A: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x02, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,A zero": {
			code:       []byte{0x87},
			pre:        cpuData{A: 0x80, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, F: Z | CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,A half carry": {
			code:       []byte{0x87},
			pre:        cpuData{A: 0x0F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x1E, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,A carry in": {
			code:       []byte{0x87},
			pre:        cpuData{A: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x02, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADD A,A carry out": {
			code:       []byte{0x87},
			pre:        cpuData{A: 0xFE, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0xFC, F: CY | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,B": {
			code:       []byte{0x88},
			pre:        cpuData{A: 0x41, B: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, B: 0x01, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,B zero": {
			code:       []byte{0x88},
			pre:        cpuData{A: 0xFF, B: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, B: 0x01, F: Z | CY | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,B carry in": {
			code:       []byte{0x88},
			pre:        cpuData{A: 0x41, B: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x43, B: 0x01, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,B half carry": {
			code:       []byte{0x88},
			pre:        cpuData{A: 0x0F, B: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x10, B: 0x01, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,B carry out": {
			code:       []byte{0x88},
			pre:        cpuData{A: 0x3C, B: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x3B, B: 0xFF, F: CY | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,C": {
			code:       []byte{0x89},
			pre:        cpuData{A: 0x41, C: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, C: 0x01, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,C zero carry": {
			code:       []byte{0x89},
			pre:        cpuData{A: 0xFF, C: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, C: 0x01, F: Z | CY | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,C carry in": {
			code:       []byte{0x89},
			pre:        cpuData{A: 0x41, C: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x43, C: 0x01, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,C half carry": {
			code:       []byte{0x89},
			pre:        cpuData{A: 0x0F, C: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x10, C: 0x01, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,C carry out": {
			code:       []byte{0x89},
			pre:        cpuData{A: 0x3C, C: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x3B, C: 0xFF, F: CY | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},

		"ADC A,D": {
			code:       []byte{0x8A},
			pre:        cpuData{A: 0x41, D: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, D: 0x01, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,D zero carry": {
			code:       []byte{0x8A},
			pre:        cpuData{A: 0xFF, D: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, D: 0x01, F: Z | CY | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,D carry in": {
			code:       []byte{0x8A},
			pre:        cpuData{A: 0x41, D: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x43, D: 0x01, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,D half carry": {
			code:       []byte{0x8A},
			pre:        cpuData{A: 0x0F, D: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x10, D: 0x01, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,D carry out": {
			code:       []byte{0x8A},
			pre:        cpuData{A: 0x3C, D: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x3B, D: 0xFF, F: CY | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},

		"ADC A,E": {
			code:       []byte{0x8B},
			pre:        cpuData{A: 0x41, E: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, E: 0x01, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,E zero carry": {
			code:       []byte{0x8B},
			pre:        cpuData{A: 0xFF, E: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, E: 0x01, F: Z | CY | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,E carry in": {
			code:       []byte{0x8B},
			pre:        cpuData{A: 0x41, E: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x43, E: 0x01, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,E half carry": {
			code:       []byte{0x8B},
			pre:        cpuData{A: 0x0F, E: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x10, E: 0x01, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,E carry out": {
			code:       []byte{0x8B},
			pre:        cpuData{A: 0x3C, E: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x3B, E: 0xFF, F: CY | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},

		"ADC A,H": {
			code:       []byte{0x8C},
			pre:        cpuData{A: 0x41, H: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, H: 0x01, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,H zero carry": {
			code:       []byte{0x8C},
			pre:        cpuData{A: 0xFF, H: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, H: 0x01, F: Z | CY | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,H carry in": {
			code:       []byte{0x8C},
			pre:        cpuData{A: 0x41, H: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x43, H: 0x01, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,H half carry": {
			code:       []byte{0x8C},
			pre:        cpuData{A: 0x0F, H: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x10, H: 0x01, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,H carry out": {
			code:       []byte{0x8C},
			pre:        cpuData{A: 0x3C, H: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x3B, H: 0xFF, F: CY | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},

		"ADC A,L": {
			code:       []byte{0x8D},
			pre:        cpuData{A: 0x41, L: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, L: 0x01, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,L zero carry": {
			code:       []byte{0x8D},
			pre:        cpuData{A: 0xFF, L: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, L: 0x01, F: Z | CY | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,L carry in": {
			code:       []byte{0x8D},
			pre:        cpuData{A: 0x41, L: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x43, L: 0x01, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,L half carry": {
			code:       []byte{0x8D},
			pre:        cpuData{A: 0x0F, L: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x10, L: 0x01, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,L carry out": {
			code:       []byte{0x8D},
			pre:        cpuData{A: 0x3C, L: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x3B, L: 0xFF, F: CY | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},

		"ADC A,(HL)": {
			code:       []byte{0x8E},
			pre:        cpuData{A: 0x41, H: 0x41, L: 0x01, PC: 0x8000},
			bus:        testBus{0x4101: 0x01},
			want:       cpuData{A: 0x42, H: 0x41, L: 0x01, PC: 0x8001},
			wantbus:    testBus{0x4101: 0x01},
			wantCycles: 2,
		},
		"ADC A,(HL) zero carry": {
			code:       []byte{0x8E},
			pre:        cpuData{A: 0xFF, H: 0x41, L: 0x01, PC: 0x8000},
			bus:        testBus{0x4101: 0x01},
			want:       cpuData{A: 0x00, H: 0x41, L: 0x01, F: Z | CY | H, PC: 0x8001},
			wantbus:    testBus{0x4101: 0x01},
			wantCycles: 2,
		},
		"ADC A,(HL) carry in": {
			code:       []byte{0x8E},
			pre:        cpuData{A: 0x41, H: 0x41, L: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{0x4101: 0x01},
			want:       cpuData{A: 0x43, H: 0x41, L: 0x01, PC: 0x8001},
			wantbus:    testBus{0x4101: 0x01},
			wantCycles: 2,
		},
		"ADC A,(HL) half carry": {
			code:       []byte{0x8E},
			pre:        cpuData{A: 0x0F, H: 0x41, L: 0x01, PC: 0x8000},
			bus:        testBus{0x4101: 0x01},
			want:       cpuData{A: 0x10, H: 0x41, L: 0x01, F: H, PC: 0x8001},
			wantbus:    testBus{0x4101: 0x01},
			wantCycles: 2,
		},
		"ADC A,(HL) carry out": {
			code:       []byte{0x8E},
			pre:        cpuData{A: 0x3C, H: 0x41, L: 0x01, PC: 0x8000},
			bus:        testBus{0x4101: 0xFF},
			want:       cpuData{A: 0x3B, H: 0x41, L: 0x01, F: CY | H, PC: 0x8001},
			wantbus:    testBus{0x4101: 0xFF},
			wantCycles: 2,
		},

		"ADC A,A": {
			code:       []byte{0x8F},
			pre:        cpuData{A: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x02, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,A zero carry": {
			code:       []byte{0x8F},
			pre:        cpuData{A: 0x80, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, F: Z | CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,A carry in": {
			code:       []byte{0x8F},
			pre:        cpuData{A: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x03, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,A half carry": {
			code:       []byte{0x8F},
			pre:        cpuData{A: 0x0F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x1E, F: H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"ADC A,A carry out": {
			code:       []byte{0x8F},
			pre:        cpuData{A: 0xFE, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0xFC, F: CY | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0x90_0x9F(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"SUB B": {
			code:       []byte{0x90},
			pre:        cpuData{A: 0x02, B: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, B: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB B zero": {
			code:       []byte{0x90},
			pre:        cpuData{A: 0x3E, B: 0x3E, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, B: 0x3E, F: N | Z, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB B half carry": {
			code:       []byte{0x90},
			pre:        cpuData{A: 0x3E, B: 0x0F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x2F, B: 0x0F, F: N | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB B carry in": {
			code:       []byte{0x90},
			pre:        cpuData{A: 0x02, B: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, B: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB B carry out": {
			code:       []byte{0x90},
			pre:        cpuData{A: 0x3E, B: 0x40, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0xFE, B: 0x40, F: N | CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},

		"SUB C": {
			code:       []byte{0x91},
			pre:        cpuData{A: 0x02, C: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, C: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB C zero": {
			code:       []byte{0x91},
			pre:        cpuData{A: 0x3E, C: 0x3E, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, C: 0x3E, F: N | Z, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB C half carry": {
			code:       []byte{0x91},
			pre:        cpuData{A: 0x3E, C: 0x0F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x2F, C: 0x0F, F: N | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB C carry in": {
			code:       []byte{0x91},
			pre:        cpuData{A: 0x02, C: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, C: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB C carry out": {
			code:       []byte{0x91},
			pre:        cpuData{A: 0x3E, C: 0x40, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0xFE, C: 0x40, F: N | CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},

		"SUB D": {
			code:       []byte{0x92},
			pre:        cpuData{A: 0x02, D: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, D: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB D zero": {
			code:       []byte{0x92},
			pre:        cpuData{A: 0x3E, D: 0x3E, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, D: 0x3E, F: N | Z, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB D half carry": {
			code:       []byte{0x92},
			pre:        cpuData{A: 0x3E, D: 0x0F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x2F, D: 0x0F, F: N | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB D carry in": {
			code:       []byte{0x92},
			pre:        cpuData{A: 0x02, D: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, D: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB D carry out": {
			code:       []byte{0x92},
			pre:        cpuData{A: 0x3E, D: 0x40, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0xFE, D: 0x40, F: N | CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},

		"SUB E": {
			code:       []byte{0x93},
			pre:        cpuData{A: 0x02, E: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, E: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB E zero": {
			code:       []byte{0x93},
			pre:        cpuData{A: 0x3E, E: 0x3E, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, E: 0x3E, F: N | Z, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB E half carry": {
			code:       []byte{0x93},
			pre:        cpuData{A: 0x3E, E: 0x0F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x2F, E: 0x0F, F: N | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB E carry in": {
			code:       []byte{0x93},
			pre:        cpuData{A: 0x02, E: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, E: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB E carry out": {
			code:       []byte{0x93},
			pre:        cpuData{A: 0x3E, E: 0x40, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0xFE, E: 0x40, F: N | CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},

		"SUB H": {
			code:       []byte{0x94},
			pre:        cpuData{A: 0x02, H: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, H: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB H zero": {
			code:       []byte{0x94},
			pre:        cpuData{A: 0x3E, H: 0x3E, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, H: 0x3E, F: N | Z, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB H half carry": {
			code:       []byte{0x94},
			pre:        cpuData{A: 0x3E, H: 0x0F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x2F, H: 0x0F, F: N | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB H carry in": {
			code:       []byte{0x94},
			pre:        cpuData{A: 0x02, H: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, H: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB H carry out": {
			code:       []byte{0x94},
			pre:        cpuData{A: 0x3E, H: 0x40, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0xFE, H: 0x40, F: N | CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},

		"SUB L": {
			code:       []byte{0x95},
			pre:        cpuData{A: 0x02, L: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, L: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB L zero": {
			code:       []byte{0x95},
			pre:        cpuData{A: 0x3E, L: 0x3E, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, L: 0x3E, F: N | Z, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB L half carry": {
			code:       []byte{0x95},
			pre:        cpuData{A: 0x3E, L: 0x0F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x2F, L: 0x0F, F: N | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB L carry in": {
			code:       []byte{0x95},
			pre:        cpuData{A: 0x02, L: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, L: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB L carry out": {
			code:       []byte{0x95},
			pre:        cpuData{A: 0x3E, L: 0x40, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0xFE, L: 0x40, F: N | CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},

		"SUB (HL)": {
			code:       []byte{0x96},
			pre:        cpuData{A: 0x02, H: 0x40, L: 0x01, PC: 0x8000},
			bus:        testBus{0x4001: 0x01},
			want:       cpuData{A: 0x01, H: 0x40, L: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{0x4001: 0x01},
			wantCycles: 2,
		},
		"SUB (HL) zero": {
			code:       []byte{0x96},
			pre:        cpuData{A: 0x3E, H: 0x40, L: 0x01, PC: 0x8000},
			bus:        testBus{0x4001: 0x3E},
			want:       cpuData{A: 0x00, H: 0x40, L: 0x01, F: N | Z, PC: 0x8001},
			wantbus:    testBus{0x4001: 0x3E},
			wantCycles: 2,
		},
		"SUB (HL) half carry": {
			code:       []byte{0x96},
			pre:        cpuData{A: 0x3E, H: 0x40, L: 0x01, PC: 0x8000},
			bus:        testBus{0x4001: 0x0F},
			want:       cpuData{A: 0x2F, H: 0x40, L: 0x01, F: N | H, PC: 0x8001},
			wantbus:    testBus{0x4001: 0x0F},
			wantCycles: 2,
		},
		"SUB (HL) carry in": {
			code:       []byte{0x96},
			pre:        cpuData{A: 0x02, H: 0x40, L: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{0x4001: 0x01},
			want:       cpuData{A: 0x01, H: 0x40, L: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{0x4001: 0x01},
			wantCycles: 2,
		},
		"SUB (HL) carry out": {
			code:       []byte{0x96},
			pre:        cpuData{A: 0x3E, H: 0x40, L: 0x01, PC: 0x8000},
			bus:        testBus{0x4001: 0x40},
			want:       cpuData{A: 0xFE, H: 0x40, L: 0x01, F: N | CY, PC: 0x8001},
			wantbus:    testBus{0x4001: 0x40},
			wantCycles: 2,
		},

		"SUB A": {
			code:       []byte{0x97},
			pre:        cpuData{A: 0x02, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, F: N | Z, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SUB A carry in": {
			code:       []byte{0x97},
			pre:        cpuData{A: 0x02, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, F: N | Z, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},

		"SBC B": {
			code:       []byte{0x98},
			pre:        cpuData{A: 0x02, B: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, B: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SBC B zero": {
			code:       []byte{0x98},
			pre:        cpuData{A: 0x3E, B: 0x3E, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, B: 0x3E, F: N | Z, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SBC B half carry": {
			code:       []byte{0x98},
			pre:        cpuData{A: 0x3E, B: 0x0F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x2F, B: 0x0F, F: N | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SBC B carry in": {
			code:       []byte{0x98},
			pre:        cpuData{A: 0x03, B: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, B: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},

		"SBC C": {
			code:       []byte{0x99},
			pre:        cpuData{A: 0x02, C: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, C: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SBC C zero": {
			code:       []byte{0x99},
			pre:        cpuData{A: 0x3E, C: 0x3E, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, C: 0x3E, F: N | Z, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SBC C half carry": {
			code:       []byte{0x99},
			pre:        cpuData{A: 0x3E, C: 0x0F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x2F, C: 0x0F, F: N | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SBC C carry in": {
			code:       []byte{0x99},
			pre:        cpuData{A: 0x03, C: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, C: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},

		"SBC D": {
			code:       []byte{0x9A},
			pre:        cpuData{A: 0x02, D: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, D: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SBC D zero": {
			code:       []byte{0x9A},
			pre:        cpuData{A: 0x3E, D: 0x3E, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, D: 0x3E, F: N | Z, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SBC D half carry": {
			code:       []byte{0x9A},
			pre:        cpuData{A: 0x3E, D: 0x0F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x2F, D: 0x0F, F: N | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SBC D carry in": {
			code:       []byte{0x9A},
			pre:        cpuData{A: 0x03, D: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, D: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},

		"SBC E": {
			code:       []byte{0x9B},
			pre:        cpuData{A: 0x02, E: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, E: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SBC E zero": {
			code:       []byte{0x9B},
			pre:        cpuData{A: 0x3E, E: 0x3E, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, E: 0x3E, F: N | Z, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SBC E half carry": {
			code:       []byte{0x9B},
			pre:        cpuData{A: 0x3E, E: 0x0F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x2F, E: 0x0F, F: N | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SBC E carry in": {
			code:       []byte{0x9B},
			pre:        cpuData{A: 0x03, E: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, E: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},

		"SBC H": {
			code:       []byte{0x9C},
			pre:        cpuData{A: 0x02, H: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, H: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SBC H zero": {
			code:       []byte{0x9C},
			pre:        cpuData{A: 0x3E, H: 0x3E, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, H: 0x3E, F: N | Z, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SBC H half carry": {
			code:       []byte{0x9C},
			pre:        cpuData{A: 0x3E, H: 0x0F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x2F, H: 0x0F, F: N | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SBC H carry in": {
			code:       []byte{0x9C},
			pre:        cpuData{A: 0x03, H: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, H: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},

		"SBC L": {
			code:       []byte{0x9D},
			pre:        cpuData{A: 0x02, L: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, L: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SBC L zero": {
			code:       []byte{0x9D},
			pre:        cpuData{A: 0x3E, L: 0x3E, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, L: 0x3E, F: N | Z, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SBC L half carry": {
			code:       []byte{0x9D},
			pre:        cpuData{A: 0x3E, L: 0x0F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x2F, L: 0x0F, F: N | H, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SBC L carry in": {
			code:       []byte{0x9D},
			pre:        cpuData{A: 0x03, L: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, L: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},

		"SBC (HL)": {
			code:       []byte{0x9E},
			pre:        cpuData{A: 0x02, H: 0x40, L: 0x01, PC: 0x8000},
			bus:        testBus{0x4001: 0x01},
			want:       cpuData{A: 0x01, H: 0x40, L: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{0x4001: 0x01},
			wantCycles: 2,
		},
		"SBC (HL) zero": {
			code:       []byte{0x9E},
			pre:        cpuData{A: 0x3E, H: 0x40, L: 0x01, PC: 0x8000},
			bus:        testBus{0x4001: 0x3E},
			want:       cpuData{A: 0x00, H: 0x40, L: 0x01, F: N | Z, PC: 0x8001},
			wantbus:    testBus{0x4001: 0x3E},
			wantCycles: 2,
		},
		"SBC (HL) half carry": {
			code:       []byte{0x9E},
			pre:        cpuData{A: 0x3E, H: 0x40, L: 0x01, PC: 0x8000},
			bus:        testBus{0x4001: 0x0F},
			want:       cpuData{A: 0x2F, H: 0x40, L: 0x01, F: N | H, PC: 0x8001},
			wantbus:    testBus{0x4001: 0x0F},
			wantCycles: 2,
		},
		"SBC (HL) carry in": {
			code:       []byte{0x9E},
			pre:        cpuData{A: 0x03, H: 0x40, L: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{0x4001: 0x01},
			want:       cpuData{A: 0x01, H: 0x40, L: 0x01, F: N, PC: 0x8001},
			wantbus:    testBus{0x4001: 0x01},
			wantCycles: 2,
		},

		"SBC A": {
			code:       []byte{0x9F},
			pre:        cpuData{A: 0x02, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x0, F: N | Z, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		"SBC A carry in": {
			code:       []byte{0x9F},
			pre:        cpuData{A: 0x03, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0xFF, F: N | H | CY, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 1,
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0xA0_0xAF(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"AND B 00": {code: []byte{0xA0}, pre: cpuData{A: 0x00, B: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, B: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"AND B 10": {code: []byte{0xA0}, pre: cpuData{A: 0x01, B: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, B: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"AND B 01": {code: []byte{0xA0}, pre: cpuData{A: 0x00, B: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, B: 0x01, F: H | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"AND B 11": {code: []byte{0xA0}, pre: cpuData{A: 0x01, B: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, B: 0x01, F: H, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"AND C 00": {code: []byte{0xA1}, pre: cpuData{A: 0x00, C: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, C: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"AND C 10": {code: []byte{0xA1}, pre: cpuData{A: 0x01, C: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, C: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"AND C 01": {code: []byte{0xA1}, pre: cpuData{A: 0x00, C: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, C: 0x01, F: H | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"AND C 11": {code: []byte{0xA1}, pre: cpuData{A: 0x01, C: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, C: 0x01, F: H, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"AND D 00": {code: []byte{0xA2}, pre: cpuData{A: 0x00, D: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, D: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"AND D 10": {code: []byte{0xA2}, pre: cpuData{A: 0x01, D: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, D: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"AND D 01": {code: []byte{0xA2}, pre: cpuData{A: 0x00, D: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, D: 0x01, F: H | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"AND D 11": {code: []byte{0xA2}, pre: cpuData{A: 0x01, D: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, D: 0x01, F: H, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"AND E 00": {code: []byte{0xA3}, pre: cpuData{A: 0x00, E: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, E: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"AND E 10": {code: []byte{0xA3}, pre: cpuData{A: 0x01, E: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, E: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"AND E 01": {code: []byte{0xA3}, pre: cpuData{A: 0x00, E: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, E: 0x01, F: H | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"AND E 11": {code: []byte{0xA3}, pre: cpuData{A: 0x01, E: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, E: 0x01, F: H, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"AND H 00": {code: []byte{0xA4}, pre: cpuData{A: 0x00, H: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, H: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"AND H 10": {code: []byte{0xA4}, pre: cpuData{A: 0x01, H: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, H: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"AND H 01": {code: []byte{0xA4}, pre: cpuData{A: 0x00, H: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, H: 0x01, F: H | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"AND H 11": {code: []byte{0xA4}, pre: cpuData{A: 0x01, H: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, H: 0x01, F: H, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"AND L 00": {code: []byte{0xA5}, pre: cpuData{A: 0x00, L: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, L: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"AND L 10": {code: []byte{0xA5}, pre: cpuData{A: 0x01, L: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, L: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"AND L 01": {code: []byte{0xA5}, pre: cpuData{A: 0x00, L: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, L: 0x01, F: H | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"AND L 11": {code: []byte{0xA5}, pre: cpuData{A: 0x01, L: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, L: 0x01, F: H, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"AND (HL) 00": {code: []byte{0xA6}, pre: cpuData{A: 0x00, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x00}, want: cpuData{A: 0x00, H: 0x40, L: 0x01, F: H | Z, PC: 0x8001}, wantbus: testBus{0x4001: 0x00}, wantCycles: 2},
		"AND (HL) 10": {code: []byte{0xA6}, pre: cpuData{A: 0x01, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x00}, want: cpuData{A: 0x00, H: 0x40, L: 0x01, F: H | Z, PC: 0x8001}, wantbus: testBus{0x4001: 0x00}, wantCycles: 2},
		"AND (HL) 01": {code: []byte{0xA6}, pre: cpuData{A: 0x00, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x01}, want: cpuData{A: 0x00, H: 0x40, L: 0x01, F: H | Z, PC: 0x8001}, wantbus: testBus{0x4001: 0x01}, wantCycles: 2},
		"AND (HL) 11": {code: []byte{0xA6}, pre: cpuData{A: 0x01, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x01}, want: cpuData{A: 0x01, H: 0x40, L: 0x01, F: H, PC: 0x8001}, wantbus: testBus{0x4001: 0x01}, wantCycles: 2},

		"AND A":      {code: []byte{0xA7}, pre: cpuData{A: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, F: H, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"AND A zero": {code: []byte{0xA7}, pre: cpuData{A: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, F: H | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"XOR B 00": {code: []byte{0xA8}, pre: cpuData{A: 0x00, B: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, B: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"XOR B 10": {code: []byte{0xA8}, pre: cpuData{A: 0x01, B: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, B: 0x00, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"XOR B 01": {code: []byte{0xA8}, pre: cpuData{A: 0x00, B: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, B: 0x01, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"XOR B 11": {code: []byte{0xA8}, pre: cpuData{A: 0x01, B: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, B: 0x01, F: Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"XOR C 00": {code: []byte{0xA9}, pre: cpuData{A: 0x00, C: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, C: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"XOR C 10": {code: []byte{0xA9}, pre: cpuData{A: 0x01, C: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, C: 0x00, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"XOR C 01": {code: []byte{0xA9}, pre: cpuData{A: 0x00, C: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, C: 0x01, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"XOR C 11": {code: []byte{0xA9}, pre: cpuData{A: 0x01, C: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, C: 0x01, F: Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"XOR D 00": {code: []byte{0xAA}, pre: cpuData{A: 0x00, D: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, D: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"XOR D 10": {code: []byte{0xAA}, pre: cpuData{A: 0x01, D: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, D: 0x00, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"XOR D 01": {code: []byte{0xAA}, pre: cpuData{A: 0x00, D: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, D: 0x01, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"XOR D 11": {code: []byte{0xAA}, pre: cpuData{A: 0x01, D: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, D: 0x01, F: Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"XOR E 00": {code: []byte{0xAB}, pre: cpuData{A: 0x00, E: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, E: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"XOR E 10": {code: []byte{0xAB}, pre: cpuData{A: 0x01, E: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, E: 0x00, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"XOR E 01": {code: []byte{0xAB}, pre: cpuData{A: 0x00, E: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, E: 0x01, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"XOR E 11": {code: []byte{0xAB}, pre: cpuData{A: 0x01, E: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, E: 0x01, F: Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"XOR H 00": {code: []byte{0xAC}, pre: cpuData{A: 0x00, H: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, H: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"XOR H 10": {code: []byte{0xAC}, pre: cpuData{A: 0x01, H: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, H: 0x00, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"XOR H 01": {code: []byte{0xAC}, pre: cpuData{A: 0x00, H: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, H: 0x01, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"XOR H 11": {code: []byte{0xAC}, pre: cpuData{A: 0x01, H: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, H: 0x01, F: Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"XOR L 00": {code: []byte{0xAD}, pre: cpuData{A: 0x00, L: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, L: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"XOR L 10": {code: []byte{0xAD}, pre: cpuData{A: 0x01, L: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, L: 0x00, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"XOR L 01": {code: []byte{0xAD}, pre: cpuData{A: 0x00, L: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, L: 0x01, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"XOR L 11": {code: []byte{0xAD}, pre: cpuData{A: 0x01, L: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, L: 0x01, F: Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"XOR (HL) 00": {code: []byte{0xAE}, pre: cpuData{A: 0x00, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x00}, want: cpuData{A: 0x00, H: 0x40, L: 0x01, F: Z, PC: 0x8001}, wantbus: testBus{0x4001: 0x00}, wantCycles: 2},
		"XOR (HL) 10": {code: []byte{0xAE}, pre: cpuData{A: 0x01, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x00}, want: cpuData{A: 0x01, H: 0x40, L: 0x01, PC: 0x8001}, wantbus: testBus{0x4001: 0x00}, wantCycles: 2},
		"XOR (HL) 01": {code: []byte{0xAE}, pre: cpuData{A: 0x00, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x01}, want: cpuData{A: 0x01, H: 0x40, L: 0x01, PC: 0x8001}, wantbus: testBus{0x4001: 0x01}, wantCycles: 2},
		"XOR (HL) 11": {code: []byte{0xAE}, pre: cpuData{A: 0x01, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x01}, want: cpuData{A: 0x00, H: 0x40, L: 0x01, F: Z, PC: 0x8001}, wantbus: testBus{0x4001: 0x01}, wantCycles: 2},

		"XOR A":      {code: []byte{0xAF}, pre: cpuData{A: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"XOR A zero": {code: []byte{0xAF}, pre: cpuData{A: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0xB0_0xBF(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"OR B 00": {code: []byte{0xB0}, pre: cpuData{A: 0x00, B: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, B: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"OR B 10": {code: []byte{0xB0}, pre: cpuData{A: 0x01, B: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, B: 0x00, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"OR B 01": {code: []byte{0xB0}, pre: cpuData{A: 0x00, B: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, B: 0x01, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"OR B 11": {code: []byte{0xB0}, pre: cpuData{A: 0x01, B: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, B: 0x01, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"OR C 00": {code: []byte{0xB1}, pre: cpuData{A: 0x00, C: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, C: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"OR C 10": {code: []byte{0xB1}, pre: cpuData{A: 0x01, C: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, C: 0x00, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"OR C 01": {code: []byte{0xB1}, pre: cpuData{A: 0x00, C: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, C: 0x01, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"OR C 11": {code: []byte{0xB1}, pre: cpuData{A: 0x01, C: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, C: 0x01, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"OR D 00": {code: []byte{0xB2}, pre: cpuData{A: 0x00, D: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, D: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"OR D 10": {code: []byte{0xB2}, pre: cpuData{A: 0x01, D: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, D: 0x00, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"OR D 01": {code: []byte{0xB2}, pre: cpuData{A: 0x00, D: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, D: 0x01, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"OR D 11": {code: []byte{0xB2}, pre: cpuData{A: 0x01, D: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, D: 0x01, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"OR E 00": {code: []byte{0xB3}, pre: cpuData{A: 0x00, E: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, E: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"OR E 10": {code: []byte{0xB3}, pre: cpuData{A: 0x01, E: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, E: 0x00, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"OR E 01": {code: []byte{0xB3}, pre: cpuData{A: 0x00, E: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, E: 0x01, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"OR E 11": {code: []byte{0xB3}, pre: cpuData{A: 0x01, E: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, E: 0x01, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"OR H 00": {code: []byte{0xB4}, pre: cpuData{A: 0x00, H: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, H: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"OR H 10": {code: []byte{0xB4}, pre: cpuData{A: 0x01, H: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, H: 0x00, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"OR H 01": {code: []byte{0xB4}, pre: cpuData{A: 0x00, H: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, H: 0x01, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"OR H 11": {code: []byte{0xB4}, pre: cpuData{A: 0x01, H: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, H: 0x01, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"OR L 00": {code: []byte{0xB5}, pre: cpuData{A: 0x00, L: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, L: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"OR L 10": {code: []byte{0xB5}, pre: cpuData{A: 0x01, L: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, L: 0x00, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"OR L 01": {code: []byte{0xB5}, pre: cpuData{A: 0x00, L: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, L: 0x01, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"OR L 11": {code: []byte{0xB5}, pre: cpuData{A: 0x01, L: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, L: 0x01, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"OR (HL) 00": {code: []byte{0xB6}, pre: cpuData{A: 0x00, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x00}, want: cpuData{A: 0x00, H: 0x40, L: 0x01, F: Z, PC: 0x8001}, wantbus: testBus{0x4001: 0x00}, wantCycles: 2},
		"OR (HL) 10": {code: []byte{0xB6}, pre: cpuData{A: 0x01, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x00}, want: cpuData{A: 0x01, H: 0x40, L: 0x01, PC: 0x8001}, wantbus: testBus{0x4001: 0x00}, wantCycles: 2},
		"OR (HL) 01": {code: []byte{0xB6}, pre: cpuData{A: 0x00, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x01}, want: cpuData{A: 0x01, H: 0x40, L: 0x01, PC: 0x8001}, wantbus: testBus{0x4001: 0x01}, wantCycles: 2},
		"OR (HL) 11": {code: []byte{0xB6}, pre: cpuData{A: 0x01, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x01}, want: cpuData{A: 0x01, H: 0x40, L: 0x01, PC: 0x8001}, wantbus: testBus{0x4001: 0x01}, wantCycles: 2},

		"OR A":      {code: []byte{0xB7}, pre: cpuData{A: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"OR A zero": {code: []byte{0xB7}, pre: cpuData{A: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, F: Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"CP B lt": {code: []byte{0xB8}, pre: cpuData{A: 0x42, B: 0x41, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, B: 0x41, F: N | H, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"CP B eq": {code: []byte{0xB8}, pre: cpuData{A: 0x42, B: 0x42, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, B: 0x42, F: N | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"CP B gt": {code: []byte{0xB8}, pre: cpuData{A: 0x42, B: 0x43, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, B: 0x43, F: N | CY, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"CP C lt": {code: []byte{0xB9}, pre: cpuData{A: 0x42, C: 0x41, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, C: 0x41, F: N | H, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"CP C eq": {code: []byte{0xB9}, pre: cpuData{A: 0x42, C: 0x42, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, C: 0x42, F: N | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"CP C gt": {code: []byte{0xB9}, pre: cpuData{A: 0x42, C: 0x43, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, C: 0x43, F: N | CY, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"CP D lt": {code: []byte{0xBA}, pre: cpuData{A: 0x42, D: 0x41, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, D: 0x41, F: N | H, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"CP D eq": {code: []byte{0xBA}, pre: cpuData{A: 0x42, D: 0x42, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, D: 0x42, F: N | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"CP D gt": {code: []byte{0xBA}, pre: cpuData{A: 0x42, D: 0x43, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, D: 0x43, F: N | CY, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"CP E lt": {code: []byte{0xBB}, pre: cpuData{A: 0x42, E: 0x41, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, E: 0x41, F: N | H, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"CP E eq": {code: []byte{0xBB}, pre: cpuData{A: 0x42, E: 0x42, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, E: 0x42, F: N | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"CP E gt": {code: []byte{0xBB}, pre: cpuData{A: 0x42, E: 0x43, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, E: 0x43, F: N | CY, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"CP H lt": {code: []byte{0xBC}, pre: cpuData{A: 0x42, H: 0x41, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, H: 0x41, F: N | H, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"CP H eq": {code: []byte{0xBC}, pre: cpuData{A: 0x42, H: 0x42, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, H: 0x42, F: N | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"CP H gt": {code: []byte{0xBC}, pre: cpuData{A: 0x42, H: 0x43, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, H: 0x43, F: N | CY, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"CP L lt": {code: []byte{0xBD}, pre: cpuData{A: 0x42, L: 0x41, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, L: 0x41, F: N | H, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"CP L eq": {code: []byte{0xBD}, pre: cpuData{A: 0x42, L: 0x42, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, L: 0x42, F: N | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
		"CP L gt": {code: []byte{0xBD}, pre: cpuData{A: 0x42, L: 0x43, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, L: 0x43, F: N | CY, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},

		"CP (HL) lt": {code: []byte{0xBE}, pre: cpuData{A: 0x42, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x41}, want: cpuData{A: 0x42, H: 0x40, L: 0x01, F: N | H, PC: 0x8001}, wantbus: testBus{0x4001: 0x41}, wantCycles: 2},
		"CP (HL) eq": {code: []byte{0xBE}, pre: cpuData{A: 0x42, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x42}, want: cpuData{A: 0x42, H: 0x40, L: 0x01, F: N | Z, PC: 0x8001}, wantbus: testBus{0x4001: 0x42}, wantCycles: 2},
		"CP (HL) gt": {code: []byte{0xBE}, pre: cpuData{A: 0x42, H: 0x40, L: 0x01, PC: 0x8000}, bus: testBus{0x4001: 0x43}, want: cpuData{A: 0x42, H: 0x40, L: 0x01, F: N | CY, PC: 0x8001}, wantbus: testBus{0x4001: 0x43}, wantCycles: 2},

		"CP A eq": {code: []byte{0xBF}, pre: cpuData{A: 0x42, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, F: N | Z, PC: 0x8001}, wantbus: testBus{}, wantCycles: 1},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0xC0_0xCF(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"RET NZ": {
			code:       []byte{0xC0},
			pre:        cpuData{SP: 0x0000, PC: 0x8000},
			bus:        testBus{0x0000: 0x01, 0x0001: 0x40},
			want:       cpuData{SP: 0x0002, PC: 0x4001},
			wantbus:    testBus{0x0000: 0x01, 0x0001: 0x40},
			wantCycles: 5,
		},
		"RET NZ zero": {
			code:       []byte{0xC0},
			pre:        cpuData{F: Z, SP: 0x0000, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: Z, SP: 0x0000, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},

		"POP BC": {
			code:       []byte{0xC1},
			pre:        cpuData{SP: 0x0000, PC: 0x8000},
			bus:        testBus{0x0000: 0x01, 0x0001: 0x40},
			want:       cpuData{B: 0x40, C: 0x01, SP: 0x0002, PC: 0x8001},
			wantbus:    testBus{0x0000: 0x01, 0x0001: 0x40},
			wantCycles: 3,
		},

		"JP NZ,a16": {
			code:       []byte{0xC2, 0x01, 0x40},
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{PC: 0x4001},
			wantbus:    testBus{},
			wantCycles: 4,
		},
		"JP NZ,a16 zero": {
			code:       []byte{0xC2, 0x01, 0x40},
			pre:        cpuData{F: Z, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: Z, PC: 0x8003},
			wantbus:    testBus{},
			wantCycles: 3,
		},

		"JP a16": {
			code:       []byte{0xC3, 0x01, 0x40},
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{PC: 0x4001},
			wantbus:    testBus{},
			wantCycles: 4,
		},

		"CALL NZ,a16": {
			code:       []byte{0xC4, 0x01, 0x40},
			pre:        cpuData{SP: 0x0008, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x0006, PC: 0x4001},
			wantbus:    testBus{0x0006: 0x03, 0x0007: 0x80},
			wantCycles: 6,
		},
		"CALL NZ,a16 zero": {
			code:       []byte{0xC4, 0x01, 0x40},
			pre:        cpuData{F: Z, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: Z, PC: 0x8003},
			wantbus:    testBus{},
			wantCycles: 3,
		},

		"PUSH BC": {
			code:       []byte{0xC5},
			pre:        cpuData{B: 0x40, C: 0x1, SP: 0x0009, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{B: 0x40, C: 0x1, SP: 0x0007, PC: 0x8001},
			wantbus:    testBus{0x0008: 0x40, 0x0007: 0x1},
			wantCycles: 4,
		},

		"ADD A,d8": {
			code:       []byte{0xC6, 0x41},
			pre:        cpuData{A: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"ADD A,d8 zero": {
			code:       []byte{0xC6, 0x01},
			pre:        cpuData{A: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, F: Z | H | CY, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"ADD A,d8 half carry": {
			code:       []byte{0xC6, 0x01},
			pre:        cpuData{A: 0x0F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x10, F: H, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"ADD A,d8 carry in": {
			code:       []byte{0xC6, 0x41},
			pre:        cpuData{A: 0x01, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"ADD A,d8 carry out": {
			code:       []byte{0xC6, 0xFF},
			pre:        cpuData{A: 0x3C, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x3B, F: CY | H, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},

		"RST 00H": {
			code:       []byte{0xC7},
			pre:        cpuData{SP: 0x00F9, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x00F7, PC: 0x0000},
			wantbus:    testBus{0x00F8: 0x80, 0x00F7: 0x01},
			wantCycles: 4,
		},

		"RET Z": {
			code:       []byte{0xC8},
			pre:        cpuData{SP: 0x0000, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x0000, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"RET Z zero": {
			code:       []byte{0xC8},
			pre:        cpuData{F: Z, SP: 0x0000, PC: 0x8000},
			bus:        testBus{0x0000: 0x01, 0x0001: 0x40},
			want:       cpuData{F: Z, SP: 0x0002, PC: 0x4001},
			wantbus:    testBus{0x0000: 0x01, 0x0001: 0x40},
			wantCycles: 5,
		},

		"RET": {
			code:       []byte{0xC9},
			pre:        cpuData{SP: 0x0000, PC: 0x8000},
			bus:        testBus{0x0000: 0x01, 0x0001: 0x40},
			want:       cpuData{SP: 0x0002, PC: 0x4001},
			wantbus:    testBus{0x0000: 0x01, 0x0001: 0x40},
			wantCycles: 4,
		},

		"JP Z,a16": {
			code:       []byte{0xCA, 0x01, 0x40},
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{PC: 0x8003},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"JP Z,a16 zero": {
			code:       []byte{0xCA, 0x01, 0x40},
			pre:        cpuData{F: Z, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: Z, PC: 0x4001},
			wantbus:    testBus{},
			wantCycles: 4,
		},
		// TODO: prefix
		"CALL Z,a16": {
			code:       []byte{0xCC, 0x01, 0x40},
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{PC: 0x8003},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"CALL Z,a16 zero": {
			code:       []byte{0xCC, 0x01, 0x40},
			pre:        cpuData{F: Z, SP: 0x0008, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: Z, SP: 0x0006, PC: 0x4001},
			wantbus:    testBus{0x0006: 0x03, 0x0007: 0x80},
			wantCycles: 6,
		},

		"CALL a16": {
			code:       []byte{0xCD, 0x01, 0x40},
			pre:        cpuData{SP: 0x0008, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x0006, PC: 0x4001},
			wantbus:    testBus{0x0006: 0x03, 0x0007: 0x80},
			wantCycles: 6,
		},

		"ADC A,d8": {
			code:       []byte{0xCE, 0x01},
			pre:        cpuData{A: 0x41, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x42, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"ADC A,d8 zero": {
			code:       []byte{0xCE, 0x01},
			pre:        cpuData{A: 0xFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, F: Z | CY | H, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"ADC A,d8 carry in": {
			code:       []byte{0xCE, 0x01},
			pre:        cpuData{A: 0x41, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x43, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"ADC A,d8 half carry": {
			code:       []byte{0xCE, 0x01},
			pre:        cpuData{A: 0x0F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x10, F: H, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"ADC A,d8 carry out": {
			code:       []byte{0xCE, 0xFF},
			pre:        cpuData{A: 0x3C, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x3B, F: CY | H, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},

		"RST 08H": {
			code:       []byte{0xCF},
			pre:        cpuData{SP: 0x00F9, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x00F7, PC: 0x0008},
			wantbus:    testBus{0x00F8: 0x80, 0x00F7: 0x01},
			wantCycles: 4,
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0xD0_0xDF(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"RET NC": {
			code:       []byte{0xD0},
			pre:        cpuData{SP: 0x0000, PC: 0x8000},
			bus:        testBus{0x0000: 0x01, 0x0001: 0x40},
			want:       cpuData{SP: 0x0002, PC: 0x4001},
			wantbus:    testBus{0x0000: 0x01, 0x0001: 0x40},
			wantCycles: 5,
		},
		"RET NC carry": {
			code:       []byte{0xD0},
			pre:        cpuData{F: CY, SP: 0x0000, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: CY, SP: 0x0000, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},

		"POP DE": {
			code:       []byte{0xD1},
			pre:        cpuData{SP: 0x0000, PC: 0x8000},
			bus:        testBus{0x0000: 0x01, 0x0001: 0x40},
			want:       cpuData{D: 0x40, E: 0x01, SP: 0x0002, PC: 0x8001},
			wantbus:    testBus{0x0000: 0x01, 0x0001: 0x40},
			wantCycles: 3,
		},

		"JP NC,a16": {
			code:       []byte{0xD2, 0x01, 0x40},
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{PC: 0x4001},
			wantbus:    testBus{},
			wantCycles: 4,
		},
		"JP NC,a16 carry": {
			code:       []byte{0xD2, 0x01, 0x40},
			pre:        cpuData{F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: CY, PC: 0x8003},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		// TODO: illegal
		"CALL NC,a16": {
			code:       []byte{0xD4, 0x01, 0x40},
			pre:        cpuData{SP: 0x0008, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x0006, PC: 0x4001},
			wantbus:    testBus{0x0006: 0x03, 0x0007: 0x80},
			wantCycles: 6,
		},
		"CALL NC,a16 carry": {
			code:       []byte{0xD4, 0x01, 0x40},
			pre:        cpuData{F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: CY, PC: 0x8003},
			wantbus:    testBus{},
			wantCycles: 3,
		},

		"PUSH DE": {
			code:       []byte{0xD5},
			pre:        cpuData{D: 0x40, E: 0x1, SP: 0x0009, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{D: 0x40, E: 0x1, SP: 0x0007, PC: 0x8001},
			wantbus:    testBus{0x0008: 0x40, 0x0007: 0x1},
			wantCycles: 4,
		},

		"SUB d8": {
			code:       []byte{0xD6, 0x01},
			pre:        cpuData{A: 0x02, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, F: N, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"SUB d8 zero": {
			code:       []byte{0xD6, 0x3E},
			pre:        cpuData{A: 0x3E, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, F: N | Z, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"SUB d8 half carry": {
			code:       []byte{0xD6, 0x0F},
			pre:        cpuData{A: 0x3E, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x2F, F: N | H, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"SUB d8 carry in": {
			code:       []byte{0xD6, 0x01},
			pre:        cpuData{A: 0x02, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, F: N, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"SUB d8 carry out": {
			code:       []byte{0xD6, 0x40},
			pre:        cpuData{A: 0x3E, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0xFE, F: N | CY, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},

		"RST 10H": {
			code:       []byte{0xD7},
			pre:        cpuData{SP: 0x00F9, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x00F7, PC: 0x0010},
			wantbus:    testBus{0x00F8: 0x80, 0x00F7: 0x01},
			wantCycles: 4,
		},

		"RET C": {
			code:       []byte{0xD8},
			pre:        cpuData{SP: 0x0000, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x0000, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"RET C carry": {
			code:       []byte{0xD8},
			pre:        cpuData{F: CY, SP: 0x0000, PC: 0x8000},
			bus:        testBus{0x0000: 0x01, 0x0001: 0x40},
			want:       cpuData{F: CY, SP: 0x0002, PC: 0x4001},
			wantbus:    testBus{0x0000: 0x01, 0x0001: 0x40},
			wantCycles: 5,
		},

		"RETI": {
			code:       []byte{0xD9},
			pre:        cpuData{SP: 0x0000, PC: 0x8000},
			bus:        testBus{0x0000: 0x01, 0x0001: 0x40},
			want:       cpuData{SP: 0x0002, PC: 0x4001, IME: true},
			wantbus:    testBus{0x0000: 0x01, 0x0001: 0x40},
			wantCycles: 4,
		},

		"JP C,a16": {
			code:       []byte{0xDA, 0x01, 0x40},
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{PC: 0x8003},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"JP C,a16 carry": {
			code:       []byte{0xDA, 0x01, 0x40},
			pre:        cpuData{F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: CY, PC: 0x4001},
			wantbus:    testBus{},
			wantCycles: 4,
		},
		// TODO: illegal
		"CALL C,a16": {
			code:       []byte{0xDC, 0x01, 0x40},
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{PC: 0x8003},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"CALL C,a16 carry": {
			code:       []byte{0xDC, 0x01, 0x40},
			pre:        cpuData{F: CY, SP: 0x0008, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: CY, SP: 0x0006, PC: 0x4001},
			wantbus:    testBus{0x0006: 0x03, 0x0007: 0x80},
			wantCycles: 6,
		},

		// TODO: illegal

		"SBC d8": {
			code:       []byte{0xDE, 0x01},
			pre:        cpuData{A: 0x02, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, F: N, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"SBC d8 zero": {
			code:       []byte{0xDE, 0x3E},
			pre:        cpuData{A: 0x3E, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x00, F: N | Z, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"SBC d8 half carry": {
			code:       []byte{0xDE, 0x0F},
			pre:        cpuData{A: 0x3E, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x2F, F: N | H, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},
		"SBC d8 carry in": {
			code:       []byte{0xDE, 0x01},
			pre:        cpuData{A: 0x03, F: CY, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x01, F: N, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 2,
		},

		"RST 18H": {
			code:       []byte{0xDF},
			pre:        cpuData{SP: 0x00F9, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x00F7, PC: 0x0018},
			wantbus:    testBus{0x00F8: 0x80, 0x00F7: 0x01},
			wantCycles: 4,
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0xE0_0xEF(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"LDH (a8),A": {
			code:       []byte{0xE0, 0x42},
			pre:        cpuData{A: 0x41, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x41, PC: 0x8002},
			wantbus:    testBus{0xFF42: 0x41},
			wantCycles: 3,
		},

		"POP HL": {
			code:       []byte{0xE1},
			pre:        cpuData{SP: 0x0000, PC: 0x8000},
			bus:        testBus{0x0000: 0x01, 0x0001: 0x40},
			want:       cpuData{H: 0x40, L: 0x01, SP: 0x0002, PC: 0x8001},
			wantbus:    testBus{0x0000: 0x01, 0x0001: 0x40},
			wantCycles: 3,
		},

		"LD (C),A": {
			code:       []byte{0xE2},
			pre:        cpuData{A: 0x41, C: 0x42, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x41, C: 0x42, PC: 0x8001},
			wantbus:    testBus{0xFF42: 0x41},
			wantCycles: 2,
		},

		// TODO: illegal
		// TODO: illegal

		"PUSH HL": {
			code:       []byte{0xE5},
			pre:        cpuData{H: 0x40, L: 0x1, SP: 0x0009, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x40, L: 0x1, SP: 0x0007, PC: 0x8001},
			wantbus:    testBus{0x0008: 0x40, 0x0007: 0x1},
			wantCycles: 4,
		},

		"AND d8 00": {code: []byte{0xE6, 0x00}, pre: cpuData{A: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, F: H | Z, PC: 0x8002}, wantbus: testBus{}, wantCycles: 2},
		"AND d8 10": {code: []byte{0xE6, 0x00}, pre: cpuData{A: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, F: H | Z, PC: 0x8002}, wantbus: testBus{}, wantCycles: 2},
		"AND d8 01": {code: []byte{0xE6, 0x01}, pre: cpuData{A: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, F: H | Z, PC: 0x8002}, wantbus: testBus{}, wantCycles: 2},
		"AND d8 11": {code: []byte{0xE6, 0x01}, pre: cpuData{A: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, F: H, PC: 0x8002}, wantbus: testBus{}, wantCycles: 2},

		"RST 20H": {
			code:       []byte{0xE7},
			pre:        cpuData{SP: 0x00F9, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x00F7, PC: 0x0020},
			wantbus:    testBus{0x00F8: 0x80, 0x00F7: 0x01},
			wantCycles: 4,
		},

		"ADD SP,r8 +42": {
			code:       []byte{0xE8, 0x42},
			pre:        cpuData{SP: 0x2000, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x2042, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 4,
		},
		"ADD SP,r8 -42": {
			code:       []byte{0xE8, 0x42 ^ 0xFF + 1},
			pre:        cpuData{SP: 0x2000, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x1FBE, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 4,
		},
		"ADD SP,r8 zero": {
			code:       []byte{0xE8, 0x00},
			pre:        cpuData{SP: 0x0000, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x0000, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 4,
		},
		"ADD SP,r8 + half carry": {
			code:       []byte{0xE8, 0x01},
			pre:        cpuData{SP: 0x200F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: H, SP: 0x2010, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 4,
		},
		"ADD SP,r8 - half carry": {
			code:       []byte{0xE8, 0xEF},
			pre:        cpuData{SP: 0x0001, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: H, SP: 0xFFF0, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 4,
		},
		"ADD SP,r8 127": {
			code:       []byte{0xE8, 0x7F},
			pre:        cpuData{SP: 0x2000, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x207F, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 4,
		},
		"ADD SP,r8 -1": {
			code:       []byte{0xE8, 0xFF},
			pre:        cpuData{SP: 0x0001, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: H | CY, SP: 0x0000, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 4,
		},
		"ADD SP,r8 -1 carry": {
			code:       []byte{0xE8, 0xFF},
			pre:        cpuData{SP: 0x0000, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0xFFFF, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 4,
		},
		"ADD SP,r8 +1 carry": {
			code:       []byte{0xE8, 0x01},
			pre:        cpuData{SP: 0xFFFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{F: H | CY, SP: 0x0000, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 4,
		},

		"JP (HL)": {
			code:       []byte{0xE9},
			pre:        cpuData{H: 0x40, L: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x40, L: 0x01, PC: 0x4001},
			wantbus:    testBus{},
			wantCycles: 1,
		},

		"LD (a16),A": {
			code:       []byte{0xEA, 0x01, 0x42},
			pre:        cpuData{A: 0x41, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x41, PC: 0x8003},
			wantbus:    testBus{0x4201: 0x41},
			wantCycles: 4,
		},

		// TODO: illegal
		// TODO: illegal
		// TODO: illegal

		"XOR d8 00": {code: []byte{0xEE, 0x00}, pre: cpuData{A: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, F: Z, PC: 0x8002}, wantbus: testBus{}, wantCycles: 2},
		"XOR d8 10": {code: []byte{0xEE, 0x00}, pre: cpuData{A: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, PC: 0x8002}, wantbus: testBus{}, wantCycles: 2},
		"XOR d8 01": {code: []byte{0xEE, 0x01}, pre: cpuData{A: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, PC: 0x8002}, wantbus: testBus{}, wantCycles: 2},
		"XOR d8 11": {code: []byte{0xEE, 0x01}, pre: cpuData{A: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, F: Z, PC: 0x8002}, wantbus: testBus{}, wantCycles: 2},

		"RST 28H": {
			code:       []byte{0xEF},
			pre:        cpuData{SP: 0x00F9, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x00F7, PC: 0x0028},
			wantbus:    testBus{0x00F8: 0x80, 0x00F7: 0x01},
			wantCycles: 4,
		},
	}

	for mnemonic, tt := range tests {
		testInst(mnemonic, tt, t)
	}
}

func TestCpuOps0xF0_0xFF(t *testing.T) {
	tests := map[string]cpuSingleTest{
		"LDH A,(a8)": {
			code:       []byte{0xF0, 0x42},
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{0xFF42: 0x41},
			want:       cpuData{A: 0x41, PC: 0x8002},
			wantbus:    testBus{0xFF42: 0x41},
			wantCycles: 3,
		},

		"POP AF": {
			code:       []byte{0xF1},
			pre:        cpuData{SP: 0x0000, PC: 0x8000},
			bus:        testBus{0x0000: 0x01, 0x0001: 0x40},
			want:       cpuData{A: 0x40, F: 0x01, SP: 0x0002, PC: 0x8001},
			wantbus:    testBus{0x0000: 0x01, 0x0001: 0x40},
			wantCycles: 3,
		},

		"LD A,(C)": {
			code:       []byte{0xF2},
			pre:        cpuData{C: 0x42, PC: 0x8000},
			bus:        testBus{0xFF42: 0x41},
			want:       cpuData{A: 0x41, C: 0x42, PC: 0x8001},
			wantbus:    testBus{0xFF42: 0x41},
			wantCycles: 2,
		},

		"DI": {
			code:       []byte{0xF3},
			pre:        cpuData{PC: 0x8000, IME: true},
			bus:        testBus{},
			want:       cpuData{PC: 0x8001, IME: false},
			wantbus:    testBus{},
			wantCycles: 1,
		},
		// TODO: illegal

		"PUSH AF": {
			code:       []byte{0xF5},
			pre:        cpuData{A: 0x40, F: 0x1, SP: 0x0009, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{A: 0x40, F: 0x1, SP: 0x0007, PC: 0x8001},
			wantbus:    testBus{0x0008: 0x40, 0x0007: 0x1},
			wantCycles: 4,
		},

		"OR d8 00": {code: []byte{0xF6, 0x00}, pre: cpuData{A: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x00, F: Z, PC: 0x8002}, wantbus: testBus{}, wantCycles: 2},
		"OR d8 10": {code: []byte{0xF6, 0x00}, pre: cpuData{A: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, PC: 0x8002}, wantbus: testBus{}, wantCycles: 2},
		"OR d8 01": {code: []byte{0xF6, 0x01}, pre: cpuData{A: 0x00, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, PC: 0x8002}, wantbus: testBus{}, wantCycles: 2},
		"OR d8 11": {code: []byte{0xF6, 0x01}, pre: cpuData{A: 0x01, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x01, PC: 0x8002}, wantbus: testBus{}, wantCycles: 2},

		"RST 30H": {
			code:       []byte{0xF7},
			pre:        cpuData{SP: 0x00F9, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x00F7, PC: 0x0030},
			wantbus:    testBus{0x00F8: 0x80, 0x00F7: 0x01},
			wantCycles: 4,
		},

		"LD HL,SP+r8 +42": {
			code:       []byte{0xF8, 0x42},
			pre:        cpuData{SP: 0x2000, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x20, L: 0x42, SP: 0x2000, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"LD HL,SP+r8 -42": {
			code:       []byte{0xF8, 0x42 ^ 0xFF + 1},
			pre:        cpuData{SP: 0x2000, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x1F, L: 0xBE, SP: 0x2000, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"LD HL,SP+r8 zero": {
			code:       []byte{0xF8, 0x00},
			pre:        cpuData{SP: 0x0000, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x00, L: 0x00, SP: 0x0000, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"LD HL,SP+r8 + half carry": {
			code:       []byte{0xF8, 0x01},
			pre:        cpuData{SP: 0x200F, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x20, L: 0x10, F: H, SP: 0x200F, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"LD HL,SP+r8 - half carry": {
			code:       []byte{0xF8, 0xEF},
			pre:        cpuData{SP: 0x0001, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0xFF, L: 0xF0, F: H, SP: 0x0001, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"LD HL,SP+r8 127": {
			code:       []byte{0xF8, 0x7F},
			pre:        cpuData{SP: 0x2000, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x20, L: 0x7F, SP: 0x2000, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"LD HL,SP+r8 -1": {
			code:       []byte{0xF8, 0xFF},
			pre:        cpuData{SP: 0x0001, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x00, L: 0x00, F: H | CY, SP: 0x0001, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"LD HL,SP+r8 -1 carry": {
			code:       []byte{0xF8, 0xFF},
			pre:        cpuData{SP: 0x0000, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0xFF, L: 0xFF, SP: 0x0000, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 3,
		},
		"LD HL,SP+r8 +1 carry": {
			code:       []byte{0xF8, 0x01},
			pre:        cpuData{SP: 0xFFFF, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x00, L: 0x00, F: H | CY, SP: 0xFFFF, PC: 0x8002},
			wantbus:    testBus{},
			wantCycles: 3,
		},

		"LD SP,HL": {
			code:       []byte{0xF9},
			pre:        cpuData{H: 0x40, L: 0x01, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{H: 0x40, L: 0x01, SP: 0x4001, PC: 0x8001},
			wantbus:    testBus{},
			wantCycles: 2,
		},

		"LD A,(a16)": {
			code:       []byte{0xFA, 0x01, 0x42},
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{0x4201: 0x41},
			want:       cpuData{A: 0x41, PC: 0x8003},
			wantbus:    testBus{0x4201: 0x41},
			wantCycles: 4,
		},

		"EI": {
			code:       []byte{0xFB},
			pre:        cpuData{PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{PC: 0x8001, IME: true}, // TODO: ime should be enabled on the *next* cycle
			wantbus:    testBus{},
			wantCycles: 1,
		},

		// TODO: illegal
		// TODO: illegal

		"CP d8 lt": {code: []byte{0xFE, 0x41}, pre: cpuData{A: 0x42, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, F: N | H, PC: 0x8002}, wantbus: testBus{}, wantCycles: 2},
		"CP d8 eq": {code: []byte{0xFE, 0x42}, pre: cpuData{A: 0x42, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, F: N | Z, PC: 0x8002}, wantbus: testBus{}, wantCycles: 2},
		"CP d8 gt": {code: []byte{0xFE, 0x43}, pre: cpuData{A: 0x42, PC: 0x8000}, bus: testBus{}, want: cpuData{A: 0x42, F: N | CY, PC: 0x8002}, wantbus: testBus{}, wantCycles: 2},

		"RST 38H": {
			code:       []byte{0xFF},
			pre:        cpuData{SP: 0x00F9, PC: 0x8000},
			bus:        testBus{},
			want:       cpuData{SP: 0x00F7, PC: 0x0038},
			wantbus:    testBus{0x00F8: 0x80, 0x00F7: 0x01},
			wantCycles: 4,
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

		if tt.debug {
			runtime.Breakpoint()
		}

		// "load rom"
		for i, op := range tt.code {
			tt.bus[tt.pre.PC+uint16(i)] = op
		}

		// runs the cpu until PC points to void
		for {
			c.clock(tt.bus)
			if _, ok := tt.bus[c.PC]; !ok {
				break
			}
		}

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
			t.Errorf("cpu.executeInst(0x%02X)", tt.code[0])
			t.Logf("_pre %s", tt.pre)
			t.Logf("_got %s", got)
			t.Logf("want %s", tt.want)
		}

		cycles := tt.bus[testBusCycleAddr]
		if cycles != tt.wantCycles {
			t.Errorf("cpu.executeInst(0x%02X) cycles = %v, want %v", tt.code[0], cycles, tt.wantCycles)
		}
		delete(tt.bus, testBusCycleAddr)

		if !reflect.DeepEqual(tt.bus, tt.wantbus) {
			t.Errorf("cpu.executeInst(0x%02X) bus = %v, want %v", tt.code[0], tt.bus, tt.wantbus)
		}
	})
}
