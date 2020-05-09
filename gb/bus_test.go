package gb

import (
	"testing"
)

func TestRamMap(t *testing.T) {
	cart := &Cartridge{
		rom: []byte{
			0x3E, 0x42, //       LD A,0x42
			0xEA, 0x00, 0xC0, // LD (0xC000),A
			0x21, 0x00, 0xC0, // LD HL,0xC000
			0x46,             // LD B,(HL)
			0x23,             // INC HL
			0x4E,             // LD C,(HL)
			0x21, 0x00, 0xE0, // LD HL,0xE000
			0x56, //             LD D,(HL)
			0x23, //             INC HL
			0x5E, //             LD E,(HL)

			0xEA, 0x80, 0xFF, // LD (0xFF80),A
			0x21, 0x80, 0xFF, // LD HL,0xFF80
			0x46, //             LD B,(HL)
			0x23, //             INC HL
			0x4E, //             LD C,(HL)
		},
		mapper: mapper0{},
	}

	gb := New(cart, false)

	gb.PowerOn()
	gb.cpu.PC = 0x0000

	gb.ExecuteInst() // LD A,0x42
	gb.ExecuteInst() // LD (0xC000),A
	gb.ExecuteInst() // LD HL,0xC000
	gb.ExecuteInst() // LD B,(HL)
	gb.ExecuteInst() // INC HL
	gb.ExecuteInst() // LD C,(HL)
	gb.ExecuteInst() // LD HL,0xE000
	gb.ExecuteInst() // LD D,(HL)
	gb.ExecuteInst() // INC HL
	gb.ExecuteInst() // LD E,(HL)

	// target addr and mirror should have 0x42
	if got, want := gb.cpu.B, uint8(0x42); got != want {
		t.Errorf("wram: got 0x%02X, want 0x%02X", got, want)
	}
	if got, want := gb.cpu.D, uint8(0x42); got != want {
		t.Errorf("wram mirror: got 0x%02X, want 0x%02X", got, want)
	}

	// addresses not touched should remain 0x00
	if got, want := gb.cpu.C, uint8(0x00); got != want {
		t.Errorf("wram: got 0x%02X, want 0x%02X", got, want)
	}
	if got, want := gb.cpu.E, uint8(0x00); got != want {
		t.Errorf("wram mirror: got 0x%02X, want 0x%02X", got, want)
	}

	gb.ExecuteInst() // LD (0xFF80),A
	gb.ExecuteInst() // LD HL,0xFF80
	gb.ExecuteInst() // LD B,(HL)
	gb.ExecuteInst() // INC HL
	gb.ExecuteInst() // LD C,(HL)

	// target addr should have 0x42
	if got, want := gb.cpu.B, uint8(0x42); got != want {
		t.Errorf("hram: got 0x%02X, want 0x%02X", got, want)
	}

	// addresses not touched should remain 0x00
	if got, want := gb.cpu.C, uint8(0x00); got != want {
		t.Errorf("hram: got 0x%02X, want 0x%02X", got, want)
	}
}
