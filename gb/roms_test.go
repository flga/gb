package gb

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCpuInstr(t *testing.T) {
	tests := []string{
		testRom("cpu_instrs/individual/01-special.gb"),
		testRom("cpu_instrs/individual/02-interrupts.gb"),
		testRom("cpu_instrs/individual/03-op sp,hl.gb"),
		testRom("cpu_instrs/individual/04-op r,imm.gb"),
		testRom("cpu_instrs/individual/05-op rp.gb"),
		testRom("cpu_instrs/individual/06-ld r,r.gb"),
		testRom("cpu_instrs/individual/07-jr,jp,call,ret,rst.gb"),
		testRom("cpu_instrs/individual/08-misc instrs.gb"),
		testRom("cpu_instrs/individual/09-op r,r.gb"),
		testRom("cpu_instrs/individual/10-bit ops.gb"),
		testRom("cpu_instrs/individual/11-op a,(hl).gb"),
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			blarggTest(tt, t)
		})
	}
}

func TestInstrTiming(t *testing.T) {
	tests := []string{
		testRom("instr_timing/instr_timing.gb"),
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			blarggTest(tt, t)
		})
	}
}

func TestMemTiming(t *testing.T) {
	tests := []string{
		testRom("mem_timing/individual/01-read_timing.gb"),
		testRom("mem_timing/individual/02-write_timing.gb"),
		testRom("mem_timing/individual/03-modify_timing.gb"),
		testRom("mem_timing-2/rom_singles/01-read_timing.gb"),
		testRom("mem_timing-2/rom_singles/02-write_timing.gb"),
		testRom("mem_timing-2/rom_singles/03-modify_timing.gb"),
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			blarggTest(tt, t)
		})
	}
}

func TestInterruptTiming(t *testing.T) {
	t.Skip("cgb")
	tests := []string{
		testRom("interrupt_time/interrupt_time.gb"),
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			blarggTest(tt, t)
		})
	}
}

func TestHaltBug(t *testing.T) {
	t.Skip()
	tests := []string{
		testRom("halt_bug.gb"),
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			blarggTest(tt, t)
		})
	}
}

func testRom(path string) string { return filepath.Join("../testdata/gb-test-roms", path) }

type testSerialCtrl struct {
	buf []byte
}

func (s *testSerialCtrl) clock(gb *GameBoy)      {}
func (s *testSerialCtrl) read(addr uint16) uint8 { return 0 }
func (s *testSerialCtrl) write(addr uint16, v uint8) {
	if addr != 0xFF01 {
		return
	}
	s.buf = append(s.buf, v)
}

func (s *testSerialCtrl) String() string {
	return string(s.buf)
}

func (s *testSerialCtrl) passed() int {
	if bytes.Contains(s.buf, []byte("Passed")) {
		return 1
	} else if bytes.Contains(s.buf, []byte("Failed")) {
		return 0
	}

	return -1
}

func blarggTest(path string, t *testing.T) {
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	cart, err := NewCartridge(f)
	if err != nil {
		t.Fatal(err)
	}

	gb := New(cart, false)

	var ctrl testSerialCtrl
	gb.serial = &ctrl

	gb.PowerOn()

	for gb.machineCycles < 0x8FFFFFF {
		gb.ExecuteInst()

		status := ctrl.passed()
		switch status {
		case -1:
			continue
		case 0:
			t.Log(ctrl.String())
			t.Error("Failed")
			return
		case 1:
			return
		}
	}

	t.Fatal("timeout")
}
