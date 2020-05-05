package gb

import (
	"fmt"
	"io"
	"strings"
)

var wroteHeader bool

func disassemble(pc uint16, bus bus, A uint8, F cpuFlags, B, C, D, E, H, L uint8, SP uint16, w io.Writer) {
	if !wroteHeader {
		w.Write([]byte("[PC  ] op                [mem curval +-2] F    A  B  C  D  E  H  L  SP\n"))
		wroteHeader = true
	}
	var (
		firstColLen  = 24
		secondColLen = 41
		wrote        = 0
	)

	read16 := func() uint16 {
		lo := bus.peek(pc)
		pc++
		hi := bus.peek(pc)
		pc++

		return uint16(hi)<<8 | uint16(lo)
	}
	read8 := func() uint8 {
		v := bus.peek(pc)
		pc++

		return v
	}
	readr8 := func() int8 {
		return int8(read8())
	}
	write := func(s string) {
		n, err := w.Write([]byte(s))
		if err != nil {
			panic(err)
		}
		wrote += n
	}
	writePeek := func(addr uint16) {
		n, err := fmt.Fprintf(w, "%s[%02x %02x %02x %02x %02x]", strings.Repeat(" ", firstColLen-wrote), bus.peek(addr-2), bus.peek(addr-1), bus.peek(addr), bus.peek(addr+1), bus.peek(addr+2))
		if err != nil {
			panic(err)
		}
		wrote += n
	}
	writef8 := func(format string, v uint8) {
		n, err := fmt.Fprintf(w, format, v)
		if err != nil {
			panic(err)
		}
		wrote += n
	}
	writefr8 := func(format string, v int8) {
		n, err := fmt.Fprintf(w, format, v)
		if err != nil {
			panic(err)
		}
		wrote += n
	}
	writef16 := func(format string, v uint16) {
		n, err := fmt.Fprintf(w, format, v)
		if err != nil {
			panic(err)
		}
		wrote += n
	}

	writef16("[%04X] ", pc)

printInstr:
	op := bus.peek(pc)
	pc++

	switch op {
	// d16
	case 0x01:
		writef16("LD BC,%04Xh", read16())
	case 0x11:
		writef16("LD DE,%04Xh", read16())
	case 0x21:
		writef16("LD HL,%04Xh", read16())
	case 0x31:
		writef16("LD SP,%04Xh", read16())

	// d8
	case 0x06:
		writef8("LD B,%02Xh", read8())
	case 0x0E:
		writef8("LD C,%02Xh", read8())
	case 0x16:
		writef8("LD D,%02Xh", read8())
	case 0x1E:
		writef8("LD E,%02Xh", read8())
	case 0x26:
		writef8("LD H,%02Xh", read8())
	case 0x2E:
		writef8("LD L,%02Xh", read8())
	case 0x36:
		writef8("LD (HL),%02Xh", read8())
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x3E:
		writef8("LD A,%02Xh", read8())
	case 0xC6:
		writef8("ADD A,%02Xh", read8())
	case 0xCE:
		writef8("ADC A,%02Xh", read8())
	case 0xD6:
		writef8("SUB %02Xh", read8())
	case 0xDE:
		writef8("SBC A,%02Xh", read8())
	case 0xE6:
		writef8("AND %02Xh", read8())
	case 0xEE:
		writef8("XOR %02Xh", read8())
	case 0xF6:
		writef8("OR %02Xh", read8())
	case 0xFE:
		writef8("CP %02Xh", read8())

	// (a8)
	case 0xF0:
		lo := read8()
		writef8("LDH A,(%02Xh)", lo)
		writePeek(0xFF00 | uint16(lo))

	case 0xE0:
		lo := read8()
		writef8("LDH (%02Xh),A", lo)
		writePeek(0xFF | uint16(lo))

	// (a16)
	case 0x08:
		v := read16()
		writef16("LD (%04Xh),SP", v)
		writePeek(v)
	case 0xEA:
		v := read16()
		writef16("LD (%04Xh),A", v)
		writePeek(v)
	case 0xFA:
		v := read16()
		writef16("LD A,(%04Xh)", v)
		writePeek(v)

	// a16
	case 0xC2:
		writef16("JP NZ,%04Xh", read16())
	case 0xC3:
		writef16("JP %04Xh", read16())
	case 0xC4:
		writef16("CALL NZ,%04Xh", read16())
	case 0xCA:
		writef16("JP Z,%04Xh", read16())
	case 0xCC:
		writef16("CALL Z,%04Xh", read16())
	case 0xCD:
		writef16("CALL %04Xh", read16())
	case 0xD2:
		writef16("JP NC,%04Xh", read16())
	case 0xD4:
		writef16("CALL NC,%04Xh", read16())
	case 0xDA:
		writef16("JP C,%04Xh", read16())
	case 0xDC:
		writef16("CALL C,%04Xh", read16())

	// r8
	case 0x18:
		writefr8("JR %d", readr8())
	case 0x20:
		writefr8("JR NZ,%d", readr8())
	case 0x28:
		writefr8("JR Z,%d", readr8())
	case 0x30:
		writefr8("JR NC,%d", readr8())
	case 0x38:
		writefr8("JR C,%d", readr8())
	case 0xE8:
		writefr8("ADD SP,%d", readr8())
	case 0xF8:
		writefr8("LD HL,SP+%d", readr8())

	case 0x00:
		write("NOP")
	case 0x02:
		write("LD (BC),A")
		writePeek(uint16(B)<<8 | uint16(C))
	case 0x03:
		write("INC BC")
	case 0x04:
		write("INC B")
	case 0x05:
		write("DEC B")
	case 0x07:
		write("RLCA")
	case 0x09:
		write("ADD HL,BC")
	case 0x0A:
		write("LD A,(BC)")
		writePeek(uint16(B)<<8 | uint16(C))
	case 0x0B:
		write("DEC BC")
	case 0x0C:
		write("INC C")
	case 0x0D:
		write("DEC C")
	case 0x0F:
		write("RRCA")
	case 0x10:
		write("STOP 0")
	case 0x12:
		write("LD (DE),A")
		writePeek(uint16(D)<<8 | uint16(E))
	case 0x13:
		write("INC DE")
	case 0x14:
		write("INC D")
	case 0x15:
		write("DEC D")
	case 0x17:
		write("RLA")
	case 0x19:
		write("ADD HL,DE")
	case 0x1A:
		write("LD A,(DE)")
		writePeek(uint16(D)<<8 | uint16(E))
	case 0x1B:
		write("DEC DE")
	case 0x1C:
		write("INC E")
	case 0x1D:
		write("DEC E")
	case 0x1F:
		write("RRA")
	case 0x22:
		write("LD (HL+),A")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x23:
		write("INC HL")
	case 0x24:
		write("INC H")
	case 0x25:
		write("DEC H")
	case 0x27:
		write("DAA")
	case 0x29:
		write("ADD HL,HL")
	case 0x2A:
		write("LD A,(HL+)")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x2B:
		write("DEC HL")
	case 0x2C:
		write("INC L")
	case 0x2D:
		write("DEC L")
	case 0x2F:
		write("CPL")
	case 0x32:
		write("LD (HL-),A")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x33:
		write("INC SP")
	case 0x34:
		write("INC (HL)")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x35:
		write("DEC (HL)")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x37:
		write("SCF")
	case 0x39:
		write("ADD HL,SP")
	case 0x3A:
		write("LD A,(HL-)")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x3B:
		write("DEC SP")
	case 0x3C:
		write("INC A")
	case 0x3D:
		write("DEC A")
	case 0x3F:
		write("CCF")
	case 0x40:
		write("LD B,B")
	case 0x41:
		write("LD B,C")
	case 0x42:
		write("LD B,D")
	case 0x43:
		write("LD B,E")
	case 0x44:
		write("LD B,H")
	case 0x45:
		write("LD B,L")
	case 0x46:
		write("LD B,(HL)")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x47:
		write("LD B,A")
	case 0x48:
		write("LD C,B")
	case 0x49:
		write("LD C,C")
	case 0x4A:
		write("LD C,D")
	case 0x4B:
		write("LD C,E")
	case 0x4C:
		write("LD C,H")
	case 0x4D:
		write("LD C,L")
	case 0x4E:
		write("LD C,(HL)")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x4F:
		write("LD C,A")
	case 0x50:
		write("LD D,B")
	case 0x51:
		write("LD D,C")
	case 0x52:
		write("LD D,D")
	case 0x53:
		write("LD D,E")
	case 0x54:
		write("LD D,H")
	case 0x55:
		write("LD D,L")
	case 0x56:
		write("LD D,(HL)")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x57:
		write("LD D,A")
	case 0x58:
		write("LD E,B")
	case 0x59:
		write("LD E,C")
	case 0x5A:
		write("LD E,D")
	case 0x5B:
		write("LD E,E")
	case 0x5C:
		write("LD E,H")
	case 0x5D:
		write("LD E,L")
	case 0x5E:
		write("LD E,(HL)")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x5F:
		write("LD E,A")
	case 0x60:
		write("LD H,B")
	case 0x61:
		write("LD H,C")
	case 0x62:
		write("LD H,D")
	case 0x63:
		write("LD H,E")
	case 0x64:
		write("LD H,H")
	case 0x65:
		write("LD H,L")
	case 0x66:
		write("LD H,(HL)")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x67:
		write("LD H,A")
	case 0x68:
		write("LD L,B")
	case 0x69:
		write("LD L,C")
	case 0x6A:
		write("LD L,D")
	case 0x6B:
		write("LD L,E")
	case 0x6C:
		write("LD L,H")
	case 0x6D:
		write("LD L,L")
	case 0x6E:
		write("LD L,(HL)")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x6F:
		write("LD L,A")
	case 0x70:
		write("LD (HL),B")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x71:
		write("LD (HL),C")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x72:
		write("LD (HL),D")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x73:
		write("LD (HL),E")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x74:
		write("LD (HL),H")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x75:
		write("LD (HL),L")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x76:
		write("HALT")
	case 0x77:
		write("LD (HL),A")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x78:
		write("LD A,B")
	case 0x79:
		write("LD A,C")
	case 0x7A:
		write("LD A,D")
	case 0x7B:
		write("LD A,E")
	case 0x7C:
		write("LD A,H")
	case 0x7D:
		write("LD A,L")
	case 0x7E:
		write("LD A,(HL)")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x7F:
		write("LD A,A")
	case 0x80:
		write("ADD A,B")
	case 0x81:
		write("ADD A,C")
	case 0x82:
		write("ADD A,D")
	case 0x83:
		write("ADD A,E")
	case 0x84:
		write("ADD A,H")
	case 0x85:
		write("ADD A,L")
	case 0x86:
		write("ADD A,(HL)")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x87:
		write("ADD A,A")
	case 0x88:
		write("ADC A,B")
	case 0x89:
		write("ADC A,C")
	case 0x8A:
		write("ADC A,D")
	case 0x8B:
		write("ADC A,E")
	case 0x8C:
		write("ADC A,H")
	case 0x8D:
		write("ADC A,L")
	case 0x8E:
		write("ADC A,(HL)")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x8F:
		write("ADC A,A")
	case 0x90:
		write("SUB B")
	case 0x91:
		write("SUB C")
	case 0x92:
		write("SUB D")
	case 0x93:
		write("SUB E")
	case 0x94:
		write("SUB H")
	case 0x95:
		write("SUB L")
	case 0x96:
		write("SUB (HL)")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x97:
		write("SUB A")
	case 0x98:
		write("SBC A,B")
	case 0x99:
		write("SBC A,C")
	case 0x9A:
		write("SBC A,D")
	case 0x9B:
		write("SBC A,E")
	case 0x9C:
		write("SBC A,H")
	case 0x9D:
		write("SBC A,L")
	case 0x9E:
		write("SBC A,(HL)")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0x9F:
		write("SBC A,A")
	case 0xA0:
		write("AND B")
	case 0xA1:
		write("AND C")
	case 0xA2:
		write("AND D")
	case 0xA3:
		write("AND E")
	case 0xA4:
		write("AND H")
	case 0xA5:
		write("AND L")
	case 0xA6:
		write("AND (HL)")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0xA7:
		write("AND A")
	case 0xA8:
		write("XOR B")
	case 0xA9:
		write("XOR C")
	case 0xAA:
		write("XOR D")
	case 0xAB:
		write("XOR E")
	case 0xAC:
		write("XOR H")
	case 0xAD:
		write("XOR L")
	case 0xAE:
		write("XOR (HL)")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0xAF:
		write("XOR A")
	case 0xB0:
		write("OR B")
	case 0xB1:
		write("OR C")
	case 0xB2:
		write("OR D")
	case 0xB3:
		write("OR E")
	case 0xB4:
		write("OR H")
	case 0xB5:
		write("OR L")
	case 0xB6:
		write("OR (HL)")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0xB7:
		write("OR A")
	case 0xB8:
		write("CP B")
	case 0xB9:
		write("CP C")
	case 0xBA:
		write("CP D")
	case 0xBB:
		write("CP E")
	case 0xBC:
		write("CP H")
	case 0xBD:
		write("CP L")
	case 0xBE:
		write("CP (HL)")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0xBF:
		write("CP A")
	case 0xC0:
		write("RET NZ")
	case 0xC1:
		write("POP BC")
	case 0xC5:
		write("PUSH BC")
	case 0xC7:
		write("RST 00H")
	case 0xC8:
		write("RET Z")
	case 0xC9:
		write("RET")
	case 0xCB:
		write("CB ")
		goto printInstr // :)
	case 0xCF:
		write("RST 08H")
	case 0xD0:
		write("RET NC")
	case 0xD1:
		write("POP DE")
	case 0xD3:
		write("ILLEGAL")
	case 0xD5:
		write("PUSH DE")
	case 0xD7:
		write("RST 10H")
	case 0xD8:
		write("RET C")
	case 0xD9:
		write("RETI")
	case 0xDB:
		write("ILLEGAL")
	case 0xDD:
		write("ILLEGAL")
	case 0xDF:
		write("RST 18H")
	case 0xE1:
		write("POP HL")
	case 0xE2:
		write("LD (C),A")
		writePeek(0xFF00 | uint16(C))
	case 0xE3:
		write("ILLEGAL")
	case 0xE4:
		write("ILLEGAL")
	case 0xE5:
		write("PUSH HL")
	case 0xE7:
		write("RST 20H")
	case 0xE9:
		write("JP (HL)")
		writePeek(uint16(H)<<8 | uint16(L))
	case 0xEB:
		write("ILLEGAL")
	case 0xEC:
		write("ILLEGAL")
	case 0xED:
		write("ILLEGAL")
	case 0xEF:
		write("RST 28H")
	case 0xF1:
		write("POP AF")
	case 0xF2:
		write("LD A,(C)")
		writePeek(0xFF00 | uint16(C))
	case 0xF3:
		write("DI")
	case 0xF4:
		write("ILLEGAL")
	case 0xF5:
		write("PUSH AF")
	case 0xF7:
		write("RST 30H")
	case 0xF9:
		write("LD SP,HL")
	case 0xFB:
		write("EI")
	case 0xFC:
		write("ILLEGAL")
	case 0xFD:
		write("ILLEGAL")
	case 0xFF:
		write("RST 38H")
	}
	fmt.Fprintf(w, "%s %s %02x %02x %02x %02x %02x %02x %02x %04x\n", strings.Repeat(" ", secondColLen-wrote), F, A, B, C, D, E, H, L, SP)
}
