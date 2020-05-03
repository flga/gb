package gb

import "fmt"

type lcdc uint8

const (
	lcdcPriority       lcdc = 1 << iota // BG/Window Display/Priority     (0=Off, 1=On)
	lcdcObjEnable                       // OBJ (Sprite) Display Enable    (0=Off, 1=On)
	lcdcObjSize                         // OBJ (Sprite) Size              (0=8x8, 1=8x16)
	lcdcBgSelect                        // BG Tile Map Display Select     (0=9800-9BFF, 1=9C00-9FFF)
	lcdcBgWindowSelect                  // BG & Window Tile Data Select   (0=8800-97FF, 1=8000-8FFF)
	lcdcWindowEnable                    // Window Display Enable          (0=Off, 1=On)
	lcdcWindowSelect                    // Window Tile Map Display Select (0=9800-9BFF, 1=9C00-9FFF)
	lcdcDisplayEnable                   // LCD Display Enable             (0=Off, 1=On)
)

type lcdStat uint8

const (
	lcdStatMode lcdStat = 2 // Mode Flag       (Mode 0-3, see below) (Read Only)
	//						0: During H-Blank
	//						1: During V-Blank
	//						2: During Searching OAM
	//						3: During Transferring Data to LCD Driver

	lcdStatCoincidenceFlag lcdStat = 1<<iota + 2 // Coincidence Flag  (0:LYC<>LY, 1:LYC=LY) (Read Only)
	lcdStatHBlank                                // Mode 0 H-Blank Interrupt     (1=Enable) (Read/Write)
	lcdStatVBlank                                // Mode 1 V-Blank Interrupt     (1=Enable) (Read/Write)
	lcdStatOAM                                   // Mode 2 OAM Interrupt         (1=Enable) (Read/Write)
	lcdStatCoincidenceInt                        // LYC=LY Coincidence Interrupt (1=Enable) (Read/Write)
)

type ppu struct {
	LCDC lcdc    // LCD Control (R/W)
	STAT lcdStat // LCDC Status (R/W)
	SCY  uint8   // Scroll Y (R/W)
	SCX  uint8   // Scroll X (R/W)
	LY   uint8   // LCDC Y-Coordinate (R)
	LYC  uint8   // LY Compare (R/W)
	WY   uint8   // Window Y Position (R/W)
	WX   uint8   // Window X Position minus 7 (R/W)
	BGP  uint8   // BG Palette Data (R/W)
	OBP0 uint8   // Object Palette 0 Data (R/W)
	OBP1 uint8   // Object Palette 1 Data (R/W)
	DMA  uint8   // DMA Transfer and Start Address (R/W)

	cycle uint64
}

func (p *ppu) clock(b bus) {
	//todo: frames
}

func (p *ppu) read(addr uint16) uint8 {
	switch addr {
	case 0xFF40:
		return uint8(p.LCDC)
	case 0xFF41:
		return uint8(p.STAT)
	case 0xFF42:
		return p.SCY
	case 0xFF43:
		return p.SCX
	case 0xFF44:
		return p.LY
	case 0xFF45:
		return p.LYC
	case 0xFF4A:
		return p.WY
	case 0xFF4B:
		return p.WX
	case 0xFF47:
		return p.BGP
	case 0xFF48:
		return p.OBP0
	case 0xFF49:
		return p.OBP1
	case 0xFF46:
		return p.DMA
	}

	panic(fmt.Sprintf("unhandled ppu read 0x%04X", addr))
}

func (p *ppu) write(addr uint16, v uint8) {
	switch addr {
	case 0xFF40:
		p.LCDC = lcdc(v)
		return
	case 0xFF41:
		p.STAT = lcdStat(v) &^ (lcdStatMode | lcdStatCoincidenceFlag)
		return
	case 0xFF42:
		p.SCY = v
		return
	case 0xFF43:
		p.SCX = v
		return
	case 0xFF45:
		p.LYC = v
		return
	case 0xFF4A:
		p.WY = v
		return
	case 0xFF4B:
		p.WX = v
		return
	case 0xFF47:
		p.BGP = v
		return
	case 0xFF48:
		p.OBP0 = v
		return
	case 0xFF49:
		p.OBP1 = v
		return
	case 0xFF46:
		p.DMA = v
		return
	}

	panic(fmt.Sprintf("unhandled ppu write 0x%04X: 0x%02X", addr, v))
}
