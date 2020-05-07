package gb

const (
	cpuFreq     = 4194304
	machineFreq = 1048576
	divFreq     = 16384
)

var ioRegs = struct {
	// joypad
	P1 uint16

	// serial
	SB uint16
	SC uint16

	// timer
	DIV  uint16
	TIMA uint16
	TMA  uint16
	TAC  uint16

	// interrupts
	IF uint16
	IE uint16

	// ppu
	LCDC uint16
	STAT uint16
	SCY  uint16
	SCX  uint16
	LY   uint16
	LYC  uint16
	DMA  uint16
	BGP  uint16
	OBP0 uint16
	OBP1 uint16
	WY   uint16
	WX   uint16

	// apu
	NR10             uint16
	NR11             uint16
	NR12             uint16
	NR13             uint16
	NR14             uint16
	NR21             uint16
	NR22             uint16
	NR23             uint16
	NR24             uint16
	NR30             uint16
	NR31             uint16
	NR32             uint16
	NR33             uint16
	NR34             uint16
	WavePatternStart uint16
	WavePatternEnd   uint16
	NR41             uint16
	NR42             uint16
	NR43             uint16
	NR44             uint16
	NR50             uint16
	NR51             uint16
	NR52             uint16
}{
	P1: 0xff00,

	SB: 0xff01,
	SC: 0xff02,

	DIV:  0xff04,
	TIMA: 0xff05,
	TMA:  0xff06,
	TAC:  0xff07,

	IF: 0xff0f,
	IE: 0xffff,

	LCDC: 0xff40,
	STAT: 0xff41,
	SCY:  0xff42,
	SCX:  0xff43,
	LY:   0xff44,
	LYC:  0xff45,
	DMA:  0xff46,
	BGP:  0xff47,
	OBP0: 0xff48,
	OBP1: 0xff49,
	WY:   0xff4a,
	WX:   0xff4b,

	NR10:             0xff10,
	NR11:             0xff11,
	NR12:             0xff12,
	NR13:             0xff13,
	NR14:             0xff14,
	NR21:             0xff16,
	NR22:             0xff17,
	NR23:             0xff18,
	NR24:             0xff19,
	NR30:             0xff1a,
	NR31:             0xff1b,
	NR32:             0xff1c,
	NR33:             0xff1d,
	NR34:             0xff1e,
	WavePatternStart: 0xff30,
	WavePatternEnd:   0xff3f,
	NR41:             0xff20,
	NR42:             0xff21,
	NR43:             0xff22,
	NR44:             0xff23,
	NR50:             0xff24,
	NR51:             0xff25,
	NR52:             0xff26,
}
