package gb

import (
	"fmt"
	"io"
)

func disassemble(pc uint16, bus bus, w io.Writer) {
	fmt.Fprintf(w, "0x%04X: ", pc)
	op := bus.peek(pc)
	pc++
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

	switch op {
	// d16
	case 0x01:
		fmt.Fprintf(w, "LD BC,%04Xh\n", read16())
	case 0x11:
		fmt.Fprintf(w, "LD DE,%04Xh\n", read16())
	case 0x21:
		fmt.Fprintf(w, "LD HL,%04Xh\n", read16())
	case 0x31:
		fmt.Fprintf(w, "LD SP,%04Xh\n", read16())

	// d8
	case 0x06:
		fmt.Fprintf(w, "LD B,%02Xh\n", read8())
	case 0x0E:
		fmt.Fprintf(w, "LD C,%02Xh\n", read8())
	case 0x16:
		fmt.Fprintf(w, "LD D,%02Xh\n", read8())
	case 0x1E:
		fmt.Fprintf(w, "LD E,%02Xh\n", read8())
	case 0x26:
		fmt.Fprintf(w, "LD H,%02Xh\n", read8())
	case 0x2E:
		fmt.Fprintf(w, "LD L,%02Xh\n", read8())
	case 0x36:
		fmt.Fprintf(w, "LD (HL),%02Xh\n", read8())
	case 0x3E:
		fmt.Fprintf(w, "LD A,%02Xh\n", read8())
	case 0xC6:
		fmt.Fprintf(w, "ADD A,%02Xh\n", read8())
	case 0xCE:
		fmt.Fprintf(w, "ADC A,%02Xh\n", read8())
	case 0xD6:
		fmt.Fprintf(w, "SUB %02Xh\n", read8())
	case 0xDE:
		fmt.Fprintf(w, "SBC A,%02Xh\n", read8())
	case 0xE6:
		fmt.Fprintf(w, "AND %02Xh\n", read8())
	case 0xEE:
		fmt.Fprintf(w, "XOR %02Xh\n", read8())
	case 0xF6:
		fmt.Fprintf(w, "OR %02Xh\n", read8())
	case 0xFE:
		fmt.Fprintf(w, "CP %02Xh\n", read8())

	// (a8)
	case 0xF0:
		fmt.Fprintf(w, "LDH A,(%02Xh)\n", read8())
	case 0xE0:
		fmt.Fprintf(w, "LDH (%02Xh),A\n", read8())

	// (a16)
	case 0x08:
		fmt.Fprintf(w, "LD (%04Xh),SP\n", read16())
	case 0xEA:
		fmt.Fprintf(w, "LD (%04Xh),A\n", read16())
	case 0xFA:
		fmt.Fprintf(w, "LD A,(%04Xh)\n", read16())

	// a16
	case 0xC2:
		fmt.Fprintf(w, "JP NZ,%04Xh\n", read16())
	case 0xC3:
		fmt.Fprintf(w, "JP %04Xh\n", read16())
	case 0xC4:
		fmt.Fprintf(w, "CALL NZ,%04Xh\n", read16())
	case 0xCA:
		fmt.Fprintf(w, "JP Z,%04Xh\n", read16())
	case 0xCC:
		fmt.Fprintf(w, "CALL Z,%04Xh\n", read16())
	case 0xCD:
		fmt.Fprintf(w, "CALL %04Xh\n", read16())
	case 0xD2:
		fmt.Fprintf(w, "JP NC,%04Xh\n", read16())
	case 0xD4:
		fmt.Fprintf(w, "CALL NC,%04Xh\n", read16())
	case 0xDA:
		fmt.Fprintf(w, "JP C,%04Xh\n", read16())
	case 0xDC:
		fmt.Fprintf(w, "CALL C,%04Xh\n", read16())

	// r8
	case 0x18:
		fmt.Fprintf(w, "JR %d\n", readr8())
	case 0x20:
		fmt.Fprintf(w, "JR NZ,%d\n", readr8())
	case 0x28:
		fmt.Fprintf(w, "JR Z,%d\n", readr8())
	case 0x30:
		fmt.Fprintf(w, "JR NC,%d\n", readr8())
	case 0x38:
		fmt.Fprintf(w, "JR C,%d\n", readr8())
	case 0xE8:
		fmt.Fprintf(w, "ADD SP,%d\n", readr8())
	case 0xF8:
		fmt.Fprintf(w, "LD HL,SP+%d\n", readr8())

	case 0x00:
		w.Write([]byte("NOP\n"))
	case 0x02:
		w.Write([]byte("LD (BC),A\n"))
	case 0x03:
		w.Write([]byte("INC BC\n"))
	case 0x04:
		w.Write([]byte("INC B\n"))
	case 0x05:
		w.Write([]byte("DEC B\n"))
	case 0x07:
		w.Write([]byte("RLCA\n"))
	case 0x09:
		w.Write([]byte("ADD HL,BC\n"))
	case 0x0A:
		w.Write([]byte("LD A,(BC)\n"))
	case 0x0B:
		w.Write([]byte("DEC BC\n"))
	case 0x0C:
		w.Write([]byte("INC C\n"))
	case 0x0D:
		w.Write([]byte("DEC C\n"))
	case 0x0F:
		w.Write([]byte("RRCA\n"))
	case 0x10:
		w.Write([]byte("STOP 0\n"))
	case 0x12:
		w.Write([]byte("LD (DE),A\n"))
	case 0x13:
		w.Write([]byte("INC DE\n"))
	case 0x14:
		w.Write([]byte("INC D\n"))
	case 0x15:
		w.Write([]byte("DEC D\n"))
	case 0x17:
		w.Write([]byte("RLA\n"))
	case 0x19:
		w.Write([]byte("ADD HL,DE\n"))
	case 0x1A:
		w.Write([]byte("LD A,(DE)\n"))
	case 0x1B:
		w.Write([]byte("DEC DE\n"))
	case 0x1C:
		w.Write([]byte("INC E\n"))
	case 0x1D:
		w.Write([]byte("DEC E\n"))
	case 0x1F:
		w.Write([]byte("RRA\n"))
	case 0x22:
		w.Write([]byte("LD (HL+),A\n"))
	case 0x23:
		w.Write([]byte("INC HL\n"))
	case 0x24:
		w.Write([]byte("INC H\n"))
	case 0x25:
		w.Write([]byte("DEC H\n"))
	case 0x27:
		w.Write([]byte("DAA\n"))
	case 0x29:
		w.Write([]byte("ADD HL,HL\n"))
	case 0x2A:
		w.Write([]byte("LD A,(HL+)\n"))
	case 0x2B:
		w.Write([]byte("DEC HL\n"))
	case 0x2C:
		w.Write([]byte("INC L\n"))
	case 0x2D:
		w.Write([]byte("DEC L\n"))
	case 0x2F:
		w.Write([]byte("CPL\n"))
	case 0x32:
		w.Write([]byte("LD (HL-),A\n"))
	case 0x33:
		w.Write([]byte("INC SP\n"))
	case 0x34:
		w.Write([]byte("INC (HL)\n"))
	case 0x35:
		w.Write([]byte("DEC (HL)\n"))
	case 0x37:
		w.Write([]byte("SCF\n"))
	case 0x39:
		w.Write([]byte("ADD HL,SP\n"))
	case 0x3A:
		w.Write([]byte("LD A,(HL-)\n"))
	case 0x3B:
		w.Write([]byte("DEC SP\n"))
	case 0x3C:
		w.Write([]byte("INC A\n"))
	case 0x3D:
		w.Write([]byte("DEC A\n"))
	case 0x3F:
		w.Write([]byte("CCF\n"))
	case 0x40:
		w.Write([]byte("LD B,B\n"))
	case 0x41:
		w.Write([]byte("LD B,C\n"))
	case 0x42:
		w.Write([]byte("LD B,D\n"))
	case 0x43:
		w.Write([]byte("LD B,E\n"))
	case 0x44:
		w.Write([]byte("LD B,H\n"))
	case 0x45:
		w.Write([]byte("LD B,L\n"))
	case 0x46:
		w.Write([]byte("LD B,(HL)\n"))
	case 0x47:
		w.Write([]byte("LD B,A\n"))
	case 0x48:
		w.Write([]byte("LD C,B\n"))
	case 0x49:
		w.Write([]byte("LD C,C\n"))
	case 0x4A:
		w.Write([]byte("LD C,D\n"))
	case 0x4B:
		w.Write([]byte("LD C,E\n"))
	case 0x4C:
		w.Write([]byte("LD C,H\n"))
	case 0x4D:
		w.Write([]byte("LD C,L\n"))
	case 0x4E:
		w.Write([]byte("LD C,(HL)\n"))
	case 0x4F:
		w.Write([]byte("LD C,A\n"))
	case 0x50:
		w.Write([]byte("LD D,B\n"))
	case 0x51:
		w.Write([]byte("LD D,C\n"))
	case 0x52:
		w.Write([]byte("LD D,D\n"))
	case 0x53:
		w.Write([]byte("LD D,E\n"))
	case 0x54:
		w.Write([]byte("LD D,H\n"))
	case 0x55:
		w.Write([]byte("LD D,L\n"))
	case 0x56:
		w.Write([]byte("LD D,(HL)\n"))
	case 0x57:
		w.Write([]byte("LD D,A\n"))
	case 0x58:
		w.Write([]byte("LD E,B\n"))
	case 0x59:
		w.Write([]byte("LD E,C\n"))
	case 0x5A:
		w.Write([]byte("LD E,D\n"))
	case 0x5B:
		w.Write([]byte("LD E,E\n"))
	case 0x5C:
		w.Write([]byte("LD E,H\n"))
	case 0x5D:
		w.Write([]byte("LD E,L\n"))
	case 0x5E:
		w.Write([]byte("LD E,(HL)\n"))
	case 0x5F:
		w.Write([]byte("LD E,A\n"))
	case 0x60:
		w.Write([]byte("LD H,B\n"))
	case 0x61:
		w.Write([]byte("LD H,C\n"))
	case 0x62:
		w.Write([]byte("LD H,D\n"))
	case 0x63:
		w.Write([]byte("LD H,E\n"))
	case 0x64:
		w.Write([]byte("LD H,H\n"))
	case 0x65:
		w.Write([]byte("LD H,L\n"))
	case 0x66:
		w.Write([]byte("LD H,(HL)\n"))
	case 0x67:
		w.Write([]byte("LD H,A\n"))
	case 0x68:
		w.Write([]byte("LD L,B\n"))
	case 0x69:
		w.Write([]byte("LD L,C\n"))
	case 0x6A:
		w.Write([]byte("LD L,D\n"))
	case 0x6B:
		w.Write([]byte("LD L,E\n"))
	case 0x6C:
		w.Write([]byte("LD L,H\n"))
	case 0x6D:
		w.Write([]byte("LD L,L\n"))
	case 0x6E:
		w.Write([]byte("LD L,(HL)\n"))
	case 0x6F:
		w.Write([]byte("LD L,A\n"))
	case 0x70:
		w.Write([]byte("LD (HL),B\n"))
	case 0x71:
		w.Write([]byte("LD (HL),C\n"))
	case 0x72:
		w.Write([]byte("LD (HL),D\n"))
	case 0x73:
		w.Write([]byte("LD (HL),E\n"))
	case 0x74:
		w.Write([]byte("LD (HL),H\n"))
	case 0x75:
		w.Write([]byte("LD (HL),L\n"))
	case 0x76:
		w.Write([]byte("HALT\n"))
	case 0x77:
		w.Write([]byte("LD (HL),A\n"))
	case 0x78:
		w.Write([]byte("LD A,B\n"))
	case 0x79:
		w.Write([]byte("LD A,C\n"))
	case 0x7A:
		w.Write([]byte("LD A,D\n"))
	case 0x7B:
		w.Write([]byte("LD A,E\n"))
	case 0x7C:
		w.Write([]byte("LD A,H\n"))
	case 0x7D:
		w.Write([]byte("LD A,L\n"))
	case 0x7E:
		w.Write([]byte("LD A,(HL)\n"))
	case 0x7F:
		w.Write([]byte("LD A,A\n"))
	case 0x80:
		w.Write([]byte("ADD A,B\n"))
	case 0x81:
		w.Write([]byte("ADD A,C\n"))
	case 0x82:
		w.Write([]byte("ADD A,D\n"))
	case 0x83:
		w.Write([]byte("ADD A,E\n"))
	case 0x84:
		w.Write([]byte("ADD A,H\n"))
	case 0x85:
		w.Write([]byte("ADD A,L\n"))
	case 0x86:
		w.Write([]byte("ADD A,(HL)\n"))
	case 0x87:
		w.Write([]byte("ADD A,A\n"))
	case 0x88:
		w.Write([]byte("ADC A,B\n"))
	case 0x89:
		w.Write([]byte("ADC A,C\n"))
	case 0x8A:
		w.Write([]byte("ADC A,D\n"))
	case 0x8B:
		w.Write([]byte("ADC A,E\n"))
	case 0x8C:
		w.Write([]byte("ADC A,H\n"))
	case 0x8D:
		w.Write([]byte("ADC A,L\n"))
	case 0x8E:
		w.Write([]byte("ADC A,(HL)\n"))
	case 0x8F:
		w.Write([]byte("ADC A,A\n"))
	case 0x90:
		w.Write([]byte("SUB B\n"))
	case 0x91:
		w.Write([]byte("SUB C\n"))
	case 0x92:
		w.Write([]byte("SUB D\n"))
	case 0x93:
		w.Write([]byte("SUB E\n"))
	case 0x94:
		w.Write([]byte("SUB H\n"))
	case 0x95:
		w.Write([]byte("SUB L\n"))
	case 0x96:
		w.Write([]byte("SUB (HL)\n"))
	case 0x97:
		w.Write([]byte("SUB A\n"))
	case 0x98:
		w.Write([]byte("SBC A,B\n"))
	case 0x99:
		w.Write([]byte("SBC A,C\n"))
	case 0x9A:
		w.Write([]byte("SBC A,D\n"))
	case 0x9B:
		w.Write([]byte("SBC A,E\n"))
	case 0x9C:
		w.Write([]byte("SBC A,H\n"))
	case 0x9D:
		w.Write([]byte("SBC A,L\n"))
	case 0x9E:
		w.Write([]byte("SBC A,(HL)\n"))
	case 0x9F:
		w.Write([]byte("SBC A,A\n"))
	case 0xA0:
		w.Write([]byte("AND B\n"))
	case 0xA1:
		w.Write([]byte("AND C\n"))
	case 0xA2:
		w.Write([]byte("AND D\n"))
	case 0xA3:
		w.Write([]byte("AND E\n"))
	case 0xA4:
		w.Write([]byte("AND H\n"))
	case 0xA5:
		w.Write([]byte("AND L\n"))
	case 0xA6:
		w.Write([]byte("AND (HL)\n"))
	case 0xA7:
		w.Write([]byte("AND A\n"))
	case 0xA8:
		w.Write([]byte("XOR B\n"))
	case 0xA9:
		w.Write([]byte("XOR C\n"))
	case 0xAA:
		w.Write([]byte("XOR D\n"))
	case 0xAB:
		w.Write([]byte("XOR E\n"))
	case 0xAC:
		w.Write([]byte("XOR H\n"))
	case 0xAD:
		w.Write([]byte("XOR L\n"))
	case 0xAE:
		w.Write([]byte("XOR (HL)\n"))
	case 0xAF:
		w.Write([]byte("XOR A\n"))
	case 0xB0:
		w.Write([]byte("OR B\n"))
	case 0xB1:
		w.Write([]byte("OR C\n"))
	case 0xB2:
		w.Write([]byte("OR D\n"))
	case 0xB3:
		w.Write([]byte("OR E\n"))
	case 0xB4:
		w.Write([]byte("OR H\n"))
	case 0xB5:
		w.Write([]byte("OR L\n"))
	case 0xB6:
		w.Write([]byte("OR (HL)\n"))
	case 0xB7:
		w.Write([]byte("OR A\n"))
	case 0xB8:
		w.Write([]byte("CP B\n"))
	case 0xB9:
		w.Write([]byte("CP C\n"))
	case 0xBA:
		w.Write([]byte("CP D\n"))
	case 0xBB:
		w.Write([]byte("CP E\n"))
	case 0xBC:
		w.Write([]byte("CP H\n"))
	case 0xBD:
		w.Write([]byte("CP L\n"))
	case 0xBE:
		w.Write([]byte("CP (HL)\n"))
	case 0xBF:
		w.Write([]byte("CP A\n"))
	case 0xC0:
		w.Write([]byte("RET NZ\n"))
	case 0xC1:
		w.Write([]byte("POP BC\n"))
	case 0xC5:
		w.Write([]byte("PUSH BC\n"))
	case 0xC7:
		w.Write([]byte("RST 00H\n"))
	case 0xC8:
		w.Write([]byte("RET Z\n"))
	case 0xC9:
		w.Write([]byte("RET\n"))
	case 0xCB:
		w.Write([]byte("PREFIX CB\n"))
	case 0xCF:
		w.Write([]byte("RST 08H\n"))
	case 0xD0:
		w.Write([]byte("RET NC\n"))
	case 0xD1:
		w.Write([]byte("POP DE\n"))
	case 0xD3:
		w.Write([]byte("ILLEGAL\n"))
	case 0xD5:
		w.Write([]byte("PUSH DE\n"))
	case 0xD7:
		w.Write([]byte("RST 10H\n"))
	case 0xD8:
		w.Write([]byte("RET C\n"))
	case 0xD9:
		w.Write([]byte("RETI\n"))
	case 0xDB:
		w.Write([]byte("ILLEGAL\n"))
	case 0xDD:
		w.Write([]byte("ILLEGAL\n"))
	case 0xDF:
		w.Write([]byte("RST 18H\n"))
	case 0xE1:
		w.Write([]byte("POP HL\n"))
	case 0xE2:
		w.Write([]byte("LD (C),A\n"))
	case 0xE3:
		w.Write([]byte("ILLEGAL\n"))
	case 0xE4:
		w.Write([]byte("ILLEGAL\n"))
	case 0xE5:
		w.Write([]byte("PUSH HL\n"))
	case 0xE7:
		w.Write([]byte("RST 20H\n"))
	case 0xE9:
		w.Write([]byte("JP (HL)\n"))
	case 0xEB:
		w.Write([]byte("ILLEGAL\n"))
	case 0xEC:
		w.Write([]byte("ILLEGAL\n"))
	case 0xED:
		w.Write([]byte("ILLEGAL\n"))
	case 0xEF:
		w.Write([]byte("RST 28H\n"))
	case 0xF1:
		w.Write([]byte("POP AF\n"))
	case 0xF2:
		w.Write([]byte("LD A,(C)\n"))
	case 0xF3:
		w.Write([]byte("DI\n"))
	case 0xF4:
		w.Write([]byte("ILLEGAL\n"))
	case 0xF5:
		w.Write([]byte("PUSH AF\n"))
	case 0xF7:
		w.Write([]byte("RST 30H\n"))
	case 0xF9:
		w.Write([]byte("LD SP,HL\n"))
	case 0xFB:
		w.Write([]byte("EI\n"))
	case 0xFC:
		w.Write([]byte("ILLEGAL\n"))
	case 0xFD:
		w.Write([]byte("ILLEGAL\n"))
	case 0xFF:
		w.Write([]byte("RST 38H\n"))
	}
}
