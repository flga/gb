package gb

import "testing"

func TestRamMap(t *testing.T) {
	var c cpu
	c.init(0x0000)
	c.A = 0xFF
	c.B = 0xFF
	c.C = 0xFF
	c.D = 0xFF
	c.E = 0xFF
	c.H = 0xFF
	c.L = 0xFF

	cart := cartridge{
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

	bus := mmu{
		cpu:       &c,
		wram:      make(memory, 8*KiB),
		hram:      make(memory, 127),
		cartridge: &cart,
	}

	c.clock(&bus) // LD A,0x42
	c.clock(&bus) // LD (0xC000),A
	c.clock(&bus) // LD HL,0xC000
	c.clock(&bus) // LD B,(HL)
	c.clock(&bus) // INC HL
	c.clock(&bus) // LD C,(HL)
	c.clock(&bus) // LD HL,0xE000
	c.clock(&bus) // LD D,(HL)
	c.clock(&bus) // INC HL
	c.clock(&bus) // LD E,(HL)

	// target addr and mirror should have 0x42
	if got, want := c.B, uint8(0x42); got != want {
		t.Errorf("wram: got 0x%02X, want 0x%02X", got, want)
	}
	if got, want := c.D, uint8(0x42); got != want {
		t.Errorf("wram mirror: got 0x%02X, want 0x%02X", got, want)
	}

	// addresses not touched should remain 0x00
	if got, want := c.C, uint8(0x00); got != want {
		t.Errorf("wram: got 0x%02X, want 0x%02X", got, want)
	}
	if got, want := c.E, uint8(0x00); got != want {
		t.Errorf("wram mirror: got 0x%02X, want 0x%02X", got, want)
	}

	c.clock(&bus) // LD (0xFF80),A
	c.clock(&bus) // LD HL,0xFF80
	c.clock(&bus) // LD B,(HL)
	c.clock(&bus) // INC HL
	c.clock(&bus) // LD C,(HL)

	// target addr should have 0x42
	if got, want := c.B, uint8(0x42); got != want {
		t.Errorf("hram: got 0x%02X, want 0x%02X", got, want)
	}

	// addresses not touched should remain 0x00
	if got, want := c.C, uint8(0x00); got != want {
		t.Errorf("hram: got 0x%02X, want 0x%02X", got, want)
	}
}
