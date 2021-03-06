package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"unicode"
)

const raw = `["NOP\n1  4\n- - - -","LD BC,d16\n3  12\n- - - -","LD (BC),A\n1  8\n- - - -","INC BC\n1  8\n- - - -","INC B\n1  4\nZ 0 H -","DEC B\n1  4\nZ 1 H -","LD B,d8\n2  8\n- - - -","RLCA\n1  4\n0 0 0 C","LD (a16),SP\n3  20\n- - - -","ADD HL,BC\n1  8\n- 0 H C","LD A,(BC)\n1  8\n- - - -","DEC BC\n1  8\n- - - -","INC C\n1  4\nZ 0 H -","DEC C\n1  4\nZ 1 H -","LD C,d8\n2  8\n- - - -","RRCA\n1  4\n0 0 0 C","STOP 0\n2  4\n- - - -","LD DE,d16\n3  12\n- - - -","LD (DE),A\n1  8\n- - - -","INC DE\n1  8\n- - - -","INC D\n1  4\nZ 0 H -","DEC D\n1  4\nZ 1 H -","LD D,d8\n2  8\n- - - -","RLA\n1  4\n0 0 0 C","JR r8\n2  12\n- - - -","ADD HL,DE\n1  8\n- 0 H C","LD A,(DE)\n1  8\n- - - -","DEC DE\n1  8\n- - - -","INC E\n1  4\nZ 0 H -","DEC E\n1  4\nZ 1 H -","LD E,d8\n2  8\n- - - -","RRA\n1  4\n0 0 0 C","JR NZ,r8\n2  12/8\n- - - -","LD HL,d16\n3  12\n- - - -","LD (HL+),A\n1  8\n- - - -","INC HL\n1  8\n- - - -","INC H\n1  4\nZ 0 H -","DEC H\n1  4\nZ 1 H -","LD H,d8\n2  8\n- - - -","DAA\n1  4\nZ - 0 C","JR Z,r8\n2  12/8\n- - - -","ADD HL,HL\n1  8\n- 0 H C","LD A,(HL+)\n1  8\n- - - -","DEC HL\n1  8\n- - - -","INC L\n1  4\nZ 0 H -","DEC L\n1  4\nZ 1 H -","LD L,d8\n2  8\n- - - -","CPL\n1  4\n- 1 1 -","JR NC,r8\n2  12/8\n- - - -","LD SP,d16\n3  12\n- - - -","LD (HL-),A\n1  8\n- - - -","INC SP\n1  8\n- - - -","INC (HL)\n1  12\nZ 0 H -","DEC (HL)\n1  12\nZ 1 H -","LD (HL),d8\n2  12\n- - - -","SCF\n1  4\n- 0 0 1","JR C,r8\n2  12/8\n- - - -","ADD HL,SP\n1  8\n- 0 H C","LD A,(HL-)\n1  8\n- - - -","DEC SP\n1  8\n- - - -","INC A\n1  4\nZ 0 H -","DEC A\n1  4\nZ 1 H -","LD A,d8\n2  8\n- - - -","CCF\n1  4\n- 0 0 C","LD B,B\n1  4\n- - - -","LD B,C\n1  4\n- - - -","LD B,D\n1  4\n- - - -","LD B,E\n1  4\n- - - -","LD B,H\n1  4\n- - - -","LD B,L\n1  4\n- - - -","LD B,(HL)\n1  8\n- - - -","LD B,A\n1  4\n- - - -","LD C,B\n1  4\n- - - -","LD C,C\n1  4\n- - - -","LD C,D\n1  4\n- - - -","LD C,E\n1  4\n- - - -","LD C,H\n1  4\n- - - -","LD C,L\n1  4\n- - - -","LD C,(HL)\n1  8\n- - - -","LD C,A\n1  4\n- - - -","LD D,B\n1  4\n- - - -","LD D,C\n1  4\n- - - -","LD D,D\n1  4\n- - - -","LD D,E\n1  4\n- - - -","LD D,H\n1  4\n- - - -","LD D,L\n1  4\n- - - -","LD D,(HL)\n1  8\n- - - -","LD D,A\n1  4\n- - - -","LD E,B\n1  4\n- - - -","LD E,C\n1  4\n- - - -","LD E,D\n1  4\n- - - -","LD E,E\n1  4\n- - - -","LD E,H\n1  4\n- - - -","LD E,L\n1  4\n- - - -","LD E,(HL)\n1  8\n- - - -","LD E,A\n1  4\n- - - -","LD H,B\n1  4\n- - - -","LD H,C\n1  4\n- - - -","LD H,D\n1  4\n- - - -","LD H,E\n1  4\n- - - -","LD H,H\n1  4\n- - - -","LD H,L\n1  4\n- - - -","LD H,(HL)\n1  8\n- - - -","LD H,A\n1  4\n- - - -","LD L,B\n1  4\n- - - -","LD L,C\n1  4\n- - - -","LD L,D\n1  4\n- - - -","LD L,E\n1  4\n- - - -","LD L,H\n1  4\n- - - -","LD L,L\n1  4\n- - - -","LD L,(HL)\n1  8\n- - - -","LD L,A\n1  4\n- - - -","LD (HL),B\n1  8\n- - - -","LD (HL),C\n1  8\n- - - -","LD (HL),D\n1  8\n- - - -","LD (HL),E\n1  8\n- - - -","LD (HL),H\n1  8\n- - - -","LD (HL),L\n1  8\n- - - -","HALT\n1  4\n- - - -","LD (HL),A\n1  8\n- - - -","LD A,B\n1  4\n- - - -","LD A,C\n1  4\n- - - -","LD A,D\n1  4\n- - - -","LD A,E\n1  4\n- - - -","LD A,H\n1  4\n- - - -","LD A,L\n1  4\n- - - -","LD A,(HL)\n1  8\n- - - -","LD A,A\n1  4\n- - - -","ADD A,B\n1  4\nZ 0 H C","ADD A,C\n1  4\nZ 0 H C","ADD A,D\n1  4\nZ 0 H C","ADD A,E\n1  4\nZ 0 H C","ADD A,H\n1  4\nZ 0 H C","ADD A,L\n1  4\nZ 0 H C","ADD A,(HL)\n1  8\nZ 0 H C","ADD A,A\n1  4\nZ 0 H C","ADC A,B\n1  4\nZ 0 H C","ADC A,C\n1  4\nZ 0 H C","ADC A,D\n1  4\nZ 0 H C","ADC A,E\n1  4\nZ 0 H C","ADC A,H\n1  4\nZ 0 H C","ADC A,L\n1  4\nZ 0 H C","ADC A,(HL)\n1  8\nZ 0 H C","ADC A,A\n1  4\nZ 0 H C","SUB B\n1  4\nZ 1 H C","SUB C\n1  4\nZ 1 H C","SUB D\n1  4\nZ 1 H C","SUB E\n1  4\nZ 1 H C","SUB H\n1  4\nZ 1 H C","SUB L\n1  4\nZ 1 H C","SUB (HL)\n1  8\nZ 1 H C","SUB A\n1  4\nZ 1 H C","SBC A,B\n1  4\nZ 1 H C","SBC A,C\n1  4\nZ 1 H C","SBC A,D\n1  4\nZ 1 H C","SBC A,E\n1  4\nZ 1 H C","SBC A,H\n1  4\nZ 1 H C","SBC A,L\n1  4\nZ 1 H C","SBC A,(HL)\n1  8\nZ 1 H C","SBC A,A\n1  4\nZ 1 H C","AND B\n1  4\nZ 0 1 0","AND C\n1  4\nZ 0 1 0","AND D\n1  4\nZ 0 1 0","AND E\n1  4\nZ 0 1 0","AND H\n1  4\nZ 0 1 0","AND L\n1  4\nZ 0 1 0","AND (HL)\n1  8\nZ 0 1 0","AND A\n1  4\nZ 0 1 0","XOR B\n1  4\nZ 0 0 0","XOR C\n1  4\nZ 0 0 0","XOR D\n1  4\nZ 0 0 0","XOR E\n1  4\nZ 0 0 0","XOR H\n1  4\nZ 0 0 0","XOR L\n1  4\nZ 0 0 0","XOR (HL)\n1  8\nZ 0 0 0","XOR A\n1  4\nZ 0 0 0","OR B\n1  4\nZ 0 0 0","OR C\n1  4\nZ 0 0 0","OR D\n1  4\nZ 0 0 0","OR E\n1  4\nZ 0 0 0","OR H\n1  4\nZ 0 0 0","OR L\n1  4\nZ 0 0 0","OR (HL)\n1  8\nZ 0 0 0","OR A\n1  4\nZ 0 0 0","CP B\n1  4\nZ 1 H C","CP C\n1  4\nZ 1 H C","CP D\n1  4\nZ 1 H C","CP E\n1  4\nZ 1 H C","CP H\n1  4\nZ 1 H C","CP L\n1  4\nZ 1 H C","CP (HL)\n1  8\nZ 1 H C","CP A\n1  4\nZ 1 H C","RET NZ\n1  20/8\n- - - -","POP BC\n1  12\n- - - -","JP NZ,a16\n3  16/12\n- - - -","JP a16\n3  16\n- - - -","CALL NZ,a16\n3  24/12\n- - - -","PUSH BC\n1  16\n- - - -","ADD A,d8\n2  8\nZ 0 H C","RST 00H\n1  16\n- - - -","RET Z\n1  20/8\n- - - -","RET\n1  16\n- - - -","JP Z,a16\n3  16/12\n- - - -","PREFIX CB\n1  4\n- - - -","CALL Z,a16\n3  24/12\n- - - -","CALL a16\n3  24\n- - - -","ADC A,d8\n2  8\nZ 0 H C","RST 08H\n1  16\n- - - -","RET NC\n1  20/8\n- - - -","POP DE\n1  12\n- - - -","JP NC,a16\n3  16/12\n- - - -"," ","CALL NC,a16\n3  24/12\n- - - -","PUSH DE\n1  16\n- - - -","SUB d8\n2  8\nZ 1 H C","RST 10H\n1  16\n- - - -","RET C\n1  20/8\n- - - -","RETI\n1  16\n- - - -","JP C,a16\n3  16/12\n- - - -"," ","CALL C,a16\n3  24/12\n- - - -"," ","SBC A,d8\n2  8\nZ 1 H C","RST 18H\n1  16\n- - - -","LDH (a8),A\n2  12\n- - - -","POP HL\n1  12\n- - - -","LD (C),A\n2  8\n- - - -"," "," ","PUSH HL\n1  16\n- - - -","AND d8\n2  8\nZ 0 1 0","RST 20H\n1  16\n- - - -","ADD SP,r8\n2  16\n0 0 H C","JP (HL)\n1  4\n- - - -","LD (a16),A\n3  16\n- - - -"," "," "," ","XOR d8\n2  8\nZ 0 0 0","RST 28H\n1  16\n- - - -","LDH A,(a8)\n2  12\n- - - -","POP AF\n1  12\nZ N H C","LD A,(C)\n2  8\n- - - -","DI\n1  4\n- - - -"," ","PUSH AF\n1  16\n- - - -","OR d8\n2  8\nZ 0 0 0","RST 30H\n1  16\n- - - -","LD HL,SP+r8\n2  12\n0 0 H C","LD SP,HL\n1  8\n- - - -","LD A,(a16)\n3  16\n- - - -","EI\n1  4\n- - - -"," "," ","CP d8\n2  8\nZ 1 H C","RST 38H\n1  16\n- - - -"]`
const rawPreffix = `["RLC B\n2  8\nZ 0 0 C","RLC C\n2  8\nZ 0 0 C","RLC D\n2  8\nZ 0 0 C","RLC E\n2  8\nZ 0 0 C","RLC H\n2  8\nZ 0 0 C","RLC L\n2  8\nZ 0 0 C","RLC (HL)\n2  16\nZ 0 0 C","RLC A\n2  8\nZ 0 0 C","RRC B\n2  8\nZ 0 0 C","RRC C\n2  8\nZ 0 0 C","RRC D\n2  8\nZ 0 0 C","RRC E\n2  8\nZ 0 0 C","RRC H\n2  8\nZ 0 0 C","RRC L\n2  8\nZ 0 0 C","RRC (HL)\n2  16\nZ 0 0 C","RRC A\n2  8\nZ 0 0 C","RL B\n2  8\nZ 0 0 C","RL C\n2  8\nZ 0 0 C","RL D\n2  8\nZ 0 0 C","RL E\n2  8\nZ 0 0 C","RL H\n2  8\nZ 0 0 C","RL L\n2  8\nZ 0 0 C","RL (HL)\n2  16\nZ 0 0 C","RL A\n2  8\nZ 0 0 C","RR B\n2  8\nZ 0 0 C","RR C\n2  8\nZ 0 0 C","RR D\n2  8\nZ 0 0 C","RR E\n2  8\nZ 0 0 C","RR H\n2  8\nZ 0 0 C","RR L\n2  8\nZ 0 0 C","RR (HL)\n2  16\nZ 0 0 C","RR A\n2  8\nZ 0 0 C","SLA B\n2  8\nZ 0 0 C","SLA C\n2  8\nZ 0 0 C","SLA D\n2  8\nZ 0 0 C","SLA E\n2  8\nZ 0 0 C","SLA H\n2  8\nZ 0 0 C","SLA L\n2  8\nZ 0 0 C","SLA (HL)\n2  16\nZ 0 0 C","SLA A\n2  8\nZ 0 0 C","SRA B\n2  8\nZ 0 0 0","SRA C\n2  8\nZ 0 0 0","SRA D\n2  8\nZ 0 0 0","SRA E\n2  8\nZ 0 0 0","SRA H\n2  8\nZ 0 0 0","SRA L\n2  8\nZ 0 0 0","SRA (HL)\n2  16\nZ 0 0 0","SRA A\n2  8\nZ 0 0 0","SWAP B\n2  8\nZ 0 0 0","SWAP C\n2  8\nZ 0 0 0","SWAP D\n2  8\nZ 0 0 0","SWAP E\n2  8\nZ 0 0 0","SWAP H\n2  8\nZ 0 0 0","SWAP L\n2  8\nZ 0 0 0","SWAP (HL)\n2  16\nZ 0 0 0","SWAP A\n2  8\nZ 0 0 0","SRL B\n2  8\nZ 0 0 C","SRL C\n2  8\nZ 0 0 C","SRL D\n2  8\nZ 0 0 C","SRL E\n2  8\nZ 0 0 C","SRL H\n2  8\nZ 0 0 C","SRL L\n2  8\nZ 0 0 C","SRL (HL)\n2  16\nZ 0 0 C","SRL A\n2  8\nZ 0 0 C","BIT 0,B\n2  8\nZ 0 1 -","BIT 0,C\n2  8\nZ 0 1 -","BIT 0,D\n2  8\nZ 0 1 -","BIT 0,E\n2  8\nZ 0 1 -","BIT 0,H\n2  8\nZ 0 1 -","BIT 0,L\n2  8\nZ 0 1 -","BIT 0,(HL)\n2  16\nZ 0 1 -","BIT 0,A\n2  8\nZ 0 1 -","BIT 1,B\n2  8\nZ 0 1 -","BIT 1,C\n2  8\nZ 0 1 -","BIT 1,D\n2  8\nZ 0 1 -","BIT 1,E\n2  8\nZ 0 1 -","BIT 1,H\n2  8\nZ 0 1 -","BIT 1,L\n2  8\nZ 0 1 -","BIT 1,(HL)\n2  16\nZ 0 1 -","BIT 1,A\n2  8\nZ 0 1 -","BIT 2,B\n2  8\nZ 0 1 -","BIT 2,C\n2  8\nZ 0 1 -","BIT 2,D\n2  8\nZ 0 1 -","BIT 2,E\n2  8\nZ 0 1 -","BIT 2,H\n2  8\nZ 0 1 -","BIT 2,L\n2  8\nZ 0 1 -","BIT 2,(HL)\n2  16\nZ 0 1 -","BIT 2,A\n2  8\nZ 0 1 -","BIT 3,B\n2  8\nZ 0 1 -","BIT 3,C\n2  8\nZ 0 1 -","BIT 3,D\n2  8\nZ 0 1 -","BIT 3,E\n2  8\nZ 0 1 -","BIT 3,H\n2  8\nZ 0 1 -","BIT 3,L\n2  8\nZ 0 1 -","BIT 3,(HL)\n2  16\nZ 0 1 -","BIT 3,A\n2  8\nZ 0 1 -","BIT 4,B\n2  8\nZ 0 1 -","BIT 4,C\n2  8\nZ 0 1 -","BIT 4,D\n2  8\nZ 0 1 -","BIT 4,E\n2  8\nZ 0 1 -","BIT 4,H\n2  8\nZ 0 1 -","BIT 4,L\n2  8\nZ 0 1 -","BIT 4,(HL)\n2  16\nZ 0 1 -","BIT 4,A\n2  8\nZ 0 1 -","BIT 5,B\n2  8\nZ 0 1 -","BIT 5,C\n2  8\nZ 0 1 -","BIT 5,D\n2  8\nZ 0 1 -","BIT 5,E\n2  8\nZ 0 1 -","BIT 5,H\n2  8\nZ 0 1 -","BIT 5,L\n2  8\nZ 0 1 -","BIT 5,(HL)\n2  16\nZ 0 1 -","BIT 5,A\n2  8\nZ 0 1 -","BIT 6,B\n2  8\nZ 0 1 -","BIT 6,C\n2  8\nZ 0 1 -","BIT 6,D\n2  8\nZ 0 1 -","BIT 6,E\n2  8\nZ 0 1 -","BIT 6,H\n2  8\nZ 0 1 -","BIT 6,L\n2  8\nZ 0 1 -","BIT 6,(HL)\n2  16\nZ 0 1 -","BIT 6,A\n2  8\nZ 0 1 -","BIT 7,B\n2  8\nZ 0 1 -","BIT 7,C\n2  8\nZ 0 1 -","BIT 7,D\n2  8\nZ 0 1 -","BIT 7,E\n2  8\nZ 0 1 -","BIT 7,H\n2  8\nZ 0 1 -","BIT 7,L\n2  8\nZ 0 1 -","BIT 7,(HL)\n2  16\nZ 0 1 -","BIT 7,A\n2  8\nZ 0 1 -","RES 0,B\n2  8\n- - - -","RES 0,C\n2  8\n- - - -","RES 0,D\n2  8\n- - - -","RES 0,E\n2  8\n- - - -","RES 0,H\n2  8\n- - - -","RES 0,L\n2  8\n- - - -","RES 0,(HL)\n2  16\n- - - -","RES 0,A\n2  8\n- - - -","RES 1,B\n2  8\n- - - -","RES 1,C\n2  8\n- - - -","RES 1,D\n2  8\n- - - -","RES 1,E\n2  8\n- - - -","RES 1,H\n2  8\n- - - -","RES 1,L\n2  8\n- - - -","RES 1,(HL)\n2  16\n- - - -","RES 1,A\n2  8\n- - - -","RES 2,B\n2  8\n- - - -","RES 2,C\n2  8\n- - - -","RES 2,D\n2  8\n- - - -","RES 2,E\n2  8\n- - - -","RES 2,H\n2  8\n- - - -","RES 2,L\n2  8\n- - - -","RES 2,(HL)\n2  16\n- - - -","RES 2,A\n2  8\n- - - -","RES 3,B\n2  8\n- - - -","RES 3,C\n2  8\n- - - -","RES 3,D\n2  8\n- - - -","RES 3,E\n2  8\n- - - -","RES 3,H\n2  8\n- - - -","RES 3,L\n2  8\n- - - -","RES 3,(HL)\n2  16\n- - - -","RES 3,A\n2  8\n- - - -","RES 4,B\n2  8\n- - - -","RES 4,C\n2  8\n- - - -","RES 4,D\n2  8\n- - - -","RES 4,E\n2  8\n- - - -","RES 4,H\n2  8\n- - - -","RES 4,L\n2  8\n- - - -","RES 4,(HL)\n2  16\n- - - -","RES 4,A\n2  8\n- - - -","RES 5,B\n2  8\n- - - -","RES 5,C\n2  8\n- - - -","RES 5,D\n2  8\n- - - -","RES 5,E\n2  8\n- - - -","RES 5,H\n2  8\n- - - -","RES 5,L\n2  8\n- - - -","RES 5,(HL)\n2  16\n- - - -","RES 5,A\n2  8\n- - - -","RES 6,B\n2  8\n- - - -","RES 6,C\n2  8\n- - - -","RES 6,D\n2  8\n- - - -","RES 6,E\n2  8\n- - - -","RES 6,H\n2  8\n- - - -","RES 6,L\n2  8\n- - - -","RES 6,(HL)\n2  16\n- - - -","RES 6,A\n2  8\n- - - -","RES 7,B\n2  8\n- - - -","RES 7,C\n2  8\n- - - -","RES 7,D\n2  8\n- - - -","RES 7,E\n2  8\n- - - -","RES 7,H\n2  8\n- - - -","RES 7,L\n2  8\n- - - -","RES 7,(HL)\n2  16\n- - - -","RES 7,A\n2  8\n- - - -","SET 0,B\n2  8\n- - - -","SET 0,C\n2  8\n- - - -","SET 0,D\n2  8\n- - - -","SET 0,E\n2  8\n- - - -","SET 0,H\n2  8\n- - - -","SET 0,L\n2  8\n- - - -","SET 0,(HL)\n2  16\n- - - -","SET 0,A\n2  8\n- - - -","SET 1,B\n2  8\n- - - -","SET 1,C\n2  8\n- - - -","SET 1,D\n2  8\n- - - -","SET 1,E\n2  8\n- - - -","SET 1,H\n2  8\n- - - -","SET 1,L\n2  8\n- - - -","SET 1,(HL)\n2  16\n- - - -","SET 1,A\n2  8\n- - - -","SET 2,B\n2  8\n- - - -","SET 2,C\n2  8\n- - - -","SET 2,D\n2  8\n- - - -","SET 2,E\n2  8\n- - - -","SET 2,H\n2  8\n- - - -","SET 2,L\n2  8\n- - - -","SET 2,(HL)\n2  16\n- - - -","SET 2,A\n2  8\n- - - -","SET 3,B\n2  8\n- - - -","SET 3,C\n2  8\n- - - -","SET 3,D\n2  8\n- - - -","SET 3,E\n2  8\n- - - -","SET 3,H\n2  8\n- - - -","SET 3,L\n2  8\n- - - -","SET 3,(HL)\n2  16\n- - - -","SET 3,A\n2  8\n- - - -","SET 4,B\n2  8\n- - - -","SET 4,C\n2  8\n- - - -","SET 4,D\n2  8\n- - - -","SET 4,E\n2  8\n- - - -","SET 4,H\n2  8\n- - - -","SET 4,L\n2  8\n- - - -","SET 4,(HL)\n2  16\n- - - -","SET 4,A\n2  8\n- - - -","SET 5,B\n2  8\n- - - -","SET 5,C\n2  8\n- - - -","SET 5,D\n2  8\n- - - -","SET 5,E\n2  8\n- - - -","SET 5,H\n2  8\n- - - -","SET 5,L\n2  8\n- - - -","SET 5,(HL)\n2  16\n- - - -","SET 5,A\n2  8\n- - - -","SET 6,B\n2  8\n- - - -","SET 6,C\n2  8\n- - - -","SET 6,D\n2  8\n- - - -","SET 6,E\n2  8\n- - - -","SET 6,H\n2  8\n- - - -","SET 6,L\n2  8\n- - - -","SET 6,(HL)\n2  16\n- - - -","SET 6,A\n2  8\n- - - -","SET 7,B\n2  8\n- - - -","SET 7,C\n2  8\n- - - -","SET 7,D\n2  8\n- - - -","SET 7,E\n2  8\n- - - -","SET 7,H\n2  8\n- - - -","SET 7,L\n2  8\n- - - -","SET 7,(HL)\n2  16\n- - - -","SET 7,A\n2  8\n- - - -"]`

type instruction struct {
	code        uint8
	op          string
	mnemonic    string
	size        uint8
	cycles      uint8
	cyclesExtra uint8
	flags       string
}

func main() {
	gen := flag.String("gen", "data", "what to generate, data|funcs|table")
	preffix := flag.Bool("p", false, "use preffix table")
	flag.Parse()

	switch *gen {
	case "data":
		data(*preffix)
	case "funcs":
		funcs(*preffix)
	case "table":
		table(*preffix)
	}
}

func data(preffix bool) {
	instructions := parse(preffix)

	tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	defer tw.Flush()

	fmt.Fprintln(tw, "CODE\tMNEMONIC\tOP\tSIZE\tCYCLES\tCYCLESEXTRA\tFLAGS")

	for _, inst := range instructions {
		fmt.Fprintf(tw, "0x%02X\t%s\t%s\t%d\t%d\t%d\t%s\n", inst.code, inst.mnemonic, inst.op, inst.size, inst.cycles, inst.cyclesExtra, inst.flags)
	}
}

func table(preffix bool) {
	instructions := parse(preffix)

	fmt.Println("func (c *cpu) genTable() {")
	fmt.Println("	c.table = [256]op{")
	for i, inst := range instructions {
		fmt.Printf("c.%s,", inst.op)
		if i > 0 && (i+1)%16 == 0 {
			fmt.Println("		")
		}
	}
	fmt.Println()
	fmt.Println("	}")
	fmt.Println("}")
}

func funcs(preffix bool) {
	instructions := parse(preffix)

	funcs := make(map[string][]string)
	for _, inst := range instructions {
		funcs[inst.op] = append(funcs[inst.op], fmt.Sprintf("0x%02X\t%s\t%d %d %d %s", inst.code, inst.mnemonic, inst.size, inst.cycles, inst.cyclesExtra, inst.flags))
	}

	var keys []string
	for k := range funcs {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		v := funcs[k]
		for _, mn := range v {
			fmt.Printf("// %s\n", mn)
		}
		fmt.Printf("func (c *cpu) %s(opcode uint8, b *bus)    {panic(\"not implemented\")}\n", k)
	}
}

func parse(preffix bool) []instruction {
	var data []string
	target := raw
	if preffix {
		target = rawPreffix
	}
	if err := json.Unmarshal([]byte(target), &data); err != nil {
		panic(err)
	}

	var instructions []instruction
	for i, s := range data {
		s = strings.TrimSpace(s)
		if len(s) == 0 {
			instructions = append(instructions, instruction{
				op: "illegal",
			})
			continue
		}

		lines := strings.Split(s, "\n")
		opLine := lines[0]
		sizesLine := lines[1]
		flagsLine := lines[2]

		var inst instruction
		inst.code = uint8(i)
		inst.mnemonic = opLine

		op := strings.Split(inst.mnemonic, " ")
		inst.op = strings.ToLower(op[0])

		operands := strings.FieldsFunc(opLine, unicode.IsSpace)[1:]
		if len(operands) > 0 {
			parts := strings.Split(operands[0], ",")
			var operand string
			switch parts[0] {
			case "A", "B", "C", "D", "E", "F", "H", "L":
				operand = "r"
			case "AF", "BC", "DE", "HL":
				operand = "rr"
			case "SP":
				operand = "sp"
			case "a16":
				operand = "a16"
			case "d8":
				operand = "d8"
			case "r8":
				operand = "r8"
			case "Z":
				operand = "Z"
			case "NZ":
				operand = "NZ"
			case "NC":
				operand = "NC"
			case "(A)", "(B)", "(C)", "(D)", "(E)", "(F)", "(H)", "(L)":
				operand = "ir"
			case "(AF)", "(BC)", "(DE)", "(HL)":
				operand = "irr"
			case "(a16)":
				operand = "ia16"
			case "(a8)":
				operand = "ia8"
			case "(HL+)", "(HL-)":
				operand = "hlid"
			}
			if operand != "" {
				inst.op += "_" + operand
			}

			if len(parts) > 1 {
				var operand string
				switch parts[1] {
				case "A", "B", "C", "D", "E", "F", "H", "L":
					operand = "r"
				case "AF", "BC", "DE", "HL":
					operand = "rr"
				case "SP":
					operand = "sp"
				case "a16":
					operand = "a16"
				case "d16":
					operand = "d16"
				case "d8":
					operand = "d8"
				case "r8":
					operand = "r8"
				case "SP+r8":
					operand = "SP_r8"
				case "(A)", "(B)", "(C)", "(D)", "(E)", "(F)", "(H)", "(L)":
					operand = "ir"
				case "(AF)", "(BC)", "(DE)", "(HL)":
					operand = "irr"
				case "(a16)":
					operand = "ia16"
				case "(a8)":
					operand = "ia8"
				case "(HL+)", "(HL-)":
					operand = "hlid"
				}
				if operand != "" {
					inst.op += "_" + operand
				}
			}
		}

		sizeAndCycles := strings.FieldsFunc(sizesLine, unicode.IsSpace)
		size, err := strconv.Atoi(sizeAndCycles[0])
		if err != nil {
			panic(err)
		}

		cyclesStr := strings.Split(sizeAndCycles[1], "/")
		cycles, err := strconv.Atoi(cyclesStr[0])
		if err != nil {
			panic(err)
		}
		var cyclesExtra int
		if len(cyclesStr) > 1 {
			cyclesExtra, err = strconv.Atoi(cyclesStr[1])
			if err != nil {
				panic(err)
			}
		}

		inst.size = uint8(size)
		inst.cycles = uint8(cycles)
		inst.cyclesExtra = uint8(cyclesExtra)
		inst.flags = flagsLine

		instructions = append(instructions, inst)
	}

	return instructions
}

func digits(s string) string {
	return strings.Map(func(r rune) rune {
		if !unicode.IsDigit(r) {
			return -1
		}
		return r
	}, s)
}
