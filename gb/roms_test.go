package gb

import (
	"bytes"
	"os"
	"testing"
)

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

func TestCpuInstr(t *testing.T) {
	tests := []string{
		"01-special.gb",
		// "02-interrupts.gb",
		"03-op sp,hl.gb",
		"04-op r,imm.gb",
		"05-op rp.gb",
		"06-ld r,r.gb",
		"07-jr,jp,call,ret,rst.gb",
		"08-misc instrs.gb",
		"09-op r,r.gb",
		"10-bit ops.gb",
		"11-op a,(hl).gb",
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			f, err := os.Open("../testdata/gb-test-roms/cpu_instrs/individual/" + tt)
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

			for gb.machineCycles < 0x1FFFFFF {
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
		})
	}
}
