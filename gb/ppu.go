package gb

import (
	"image/color"
)

type sprite struct {
	y, x, tile, flags uint8
}

var basePalette = [4]color.RGBA{
	{0xFF, 0xFF, 0xFF, 0xFF},
	{0xCC, 0xCC, 0xCC, 0xFF},
	{0x33, 0x33, 0x33, 0xFF},
	{0x00, 0x00, 0x00, 0xFF},
}

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
	lcdStatMode lcdStat = 3 // Mode Flag       (Mode 0-3, see below) (Read Only)
	//						0: During H-Blank
	//						1: During V-Blank
	//						2: During Searching OAM
	//						3: During Transferring Data to LCD Driver

	lcdStatCoincidenceFlag lcdStat = 1 << (iota + 1) // Coincidence Flag  (0:LYC<>LY, 1:LYC=LY) (Read Only)
	lcdStatHBlank                                    // Mode 0 H-Blank Interrupt     (1=Enable) (Read/Write)
	lcdStatVBlank                                    // Mode 1 V-Blank Interrupt     (1=Enable) (Read/Write)
	lcdStatOAM                                       // Mode 2 OAM Interrupt         (1=Enable) (Read/Write)
	lcdStatCoincidenceInt                            // LYC=LY Coincidence Interrupt (1=Enable) (Read/Write)
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

	VRAM       [8 * KiB]byte
	OAM        [160]byte
	Nametable1 [1 * KiB]byte
	Nametable2 [1 * KiB]byte

	sprites [10]sprite
	frame   [160 * 144 * 4]uint8
	clocks  uint64
}

func (p *ppu) clock(gb *GameBoy) {
	switch {
	case p.LY >= 0 && p.LY <= 143:
		// mode 2 (oam search)
		if p.clocks >= 0 && p.clocks <= 79 {
			p.setMode(2)
			p.oamSearch()
		}

		// mode 3 (draw)
		if p.clocks >= 80 && p.clocks <= 251 {
			p.setMode(3)

			if p.clocks == 251 {
				p.drawLine()
				p.drawSprites()
			}
		}

		// mode 0 (hblank)
		if p.clocks >= 252 && p.clocks <= 455 {
			p.setMode(0)
			if p.clocks == 252 && p.STAT&lcdStatHBlank > 0 {
				gb.interruptCtrl.raise(lcdStatInterrupt)
			}
		}

	// mode 1 (vblank)
	case p.LY >= 144 && p.LY <= 153:
		p.setMode(1)
		if p.clocks == 0 && p.LY == 144 {
			gb.interruptCtrl.raise(vblankInterrupt)
		}
	}

	p.clocks++
	p.clocks %= 456
	if p.clocks == 0 {
		p.LY++
		p.LY %= 154

		if p.STAT&lcdStatCoincidenceInt > 0 {
			if p.STAT&lcdStatCoincidenceFlag == 0 && p.LY != p.LYC {
				gb.interruptCtrl.raise(lcdStatInterrupt)
			}
			if p.STAT&lcdStatCoincidenceFlag == 1 && p.LY == p.LYC {
				gb.interruptCtrl.raise(lcdStatInterrupt)
			}
		}
	}
}

func (p *ppu) drawLine() {
	if p.LCDC&lcdcDisplayEnable == 0 {
		return
	}

	fineY := uint16(p.LY)
	for fineX := uint16(0); fineX < 160; fineX++ {
		row := fineY % 8
		tileIndex := p.tileIndex(fineX, fineY)

		addr := p.tileBaseAddr(tileIndex) + row*2
		tileLo := p.read(addr)
		tileHi := p.read(addr + 1)

		tileHi <<= fineX % 8
		tileLo <<= fineX % 8

		pixelLo := tileLo & 0x80 >> 7
		pixelHi := tileHi & 0x80 >> 7
		paletteIdx := pixelHi<<1 | pixelLo
		colour := p.paletteLookup(paletteIdx, p.BGP)

		offset := int(fineY)*160*4 + int(fineX)*4
		p.frame[offset+0] = colour.R
		p.frame[offset+1] = colour.G
		p.frame[offset+2] = colour.B
		p.frame[offset+3] = colour.A
	}

	if p.LCDC&lcdcObjEnable == 0 {
		return
	}
}

func (p *ppu) tileIndex(x, y uint16) uint16 {
	offset := y / 8 * 32
	x /= 8
	if p.LCDC&lcdcBgSelect == 0 {
		return uint16(p.Nametable1[offset+x]) * 16
	}
	if p.LCDC&lcdcBgSelect > 0 {
		return uint16(p.Nametable2[offset+x]) * 16
	}

	panic("?")
}

func (p *ppu) tileBaseAddr(tileIdx uint16) uint16 {
	// (0=8800-97FF, 1=8000-8FFF)
	mode := p.LCDC & lcdcBgWindowSelect
	if mode == 0 {
		return uint16(0x9000 + int(int8(tileIdx)))
	}
	if mode > 0 {
		return uint16(0x8000 + int(tileIdx))
	}
	panic("?")
}

func (p *ppu) spriteTileBaseAddr(tileIdx uint16) uint16 {
	return uint16(0x8000) + tileIdx
}

func (p *ppu) setMode(mode uint8) {
	p.STAT &^= 3
	p.STAT |= lcdStat(mode & 3)
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
	}

	if addr >= 0x9800 && addr <= 0x9BFF {
		return p.Nametable1[addr-0x9800]
	}
	if addr >= 0x9C00 && addr <= 0x9FFF {
		return p.Nametable2[addr-0x9C00]
	}

	if addr >= 0x8000 && addr <= 0x8FFF {
		return p.VRAM[addr-0x8000]
	}

	// fmt.Fprintf(os.Stderr, "unhandled ppu read 0x%04X\n", addr)
	// panic(fmt.Sprintf("unhandled ppu read 0x%04X", addr))
	return 0
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
	}

	if addr >= 0x9800 && addr <= 0x9BFF {
		p.Nametable1[addr-0x9800] = v
		return
	}
	if addr >= 0x9C00 && addr <= 0x9FFF {
		p.Nametable2[addr-0x9C00] = v
		return
	}

	if addr >= 0x8000 && addr <= 0x8FFF {
		p.VRAM[addr-0x8000] = v
		return
	}

	if addr >= 0xFE00 && addr <= 0xFE9F {
		p.OAM[addr-0xFE00] = v
		return
	}

	// fmt.Fprintf(os.Stderr, "unhandled ppu write 0x%04X: 0x%02X\n", addr, v)
	// panic(fmt.Sprintf("unhandled ppu write 0x%04X: 0x%02X", addr, v))
	return
}

func (p *ppu) drawSprites() {
	if p.LCDC&lcdcDisplayEnable == 0 {
		return
	}
	if p.LCDC&lcdcObjEnable == 0 {
		return
	}

	for i := 0; i < 40; i++ {
		var (
			spriteY     = p.OAM[i*4+0] - 16
			spriteX     = p.OAM[i*4+1] - 8
			spriteTile  = p.OAM[i*4+2]
			spriteFlags = p.OAM[i*4+3]
		)

		row := int(p.LY) - int(spriteY)
		if row < 0 || row > 7 {
			continue
		}
		addr := 0x8000 + uint16(spriteTile)*16 + uint16(row)*2
		tileLo := p.read(addr)
		tileHi := p.read(addr + 1)

		for fineX := byte(0); fineX < 8; fineX++ {
			x := spriteX + fineX

			pixelLo := tileLo & 0x80 >> 7
			pixelHi := tileHi & 0x80 >> 7
			paletteIdx := pixelHi<<1 | pixelLo

			tileHi <<= 1
			tileLo <<= 1

			if x < 0 || x > 159 {
				continue
			}

			if paletteIdx == 00 {
				continue
			}

			var colour color.RGBA
			switch {
			case spriteFlags&(1<<4) == 0:
				colour = p.paletteLookup(paletteIdx, p.OBP0)
			case spriteFlags&(1<<4) > 0:
				colour = p.paletteLookup(paletteIdx, p.OBP1)
			}

			offset := int(p.LY)*160*4 + int(x)*4
			p.frame[offset+0] = colour.R
			p.frame[offset+1] = colour.G
			p.frame[offset+2] = colour.B
			p.frame[offset+3] = colour.A
		}
	}
}

func (p *ppu) oamSearch() { // TODO
	var idx int
	y := int(p.LY)
	for i := 0; i < 40; i++ {
		screenY := int(p.OAM[i*4+0]) - 16
		if y-screenY < 0 || y-screenY > 7 {
			continue
		}

		p.sprites[idx] = sprite{
			y:     p.OAM[i*4+0],
			x:     p.OAM[i*4+1],
			tile:  p.OAM[i*4+2],
			flags: p.OAM[i*4+3],
		}
		idx++

		if idx == 10 {
			break
		}
	}
}

func (p *ppu) paletteLookup(id, palette uint8) color.RGBA {
	shift := id * 2
	mask := uint8(0x03) << shift
	return basePalette[palette&mask>>shift]
}
