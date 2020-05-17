package gb

import (
	"image"
	"image/color"
	"math/bits"
)

type sprite struct {
	y, x, tile uint8
	flags      spriteFlags
}

type spriteFlags uint8

const (
	spriteCGBPalette  = 0x07            // Bit2-0 Palette number  **CGB Mode Only**     (OBP0-7) (Used for both BG and Window. BG color 0 is always behind OBJ)
	spriteCGBVramBank = 1 << (iota + 2) // Bit3   Tile VRAM-Bank  **CGB Mode Only**     (0=Bank 0, 1=Bank 1)
	spriptePalette                      // Bit4   Palette number  **Non CGB Mode Only** (0=OBP0, 1=OBP1)
	spriteFlipX                         // Bit5   X flip          (0=Normal, 1=Horizontally mirrored)
	spriteFlipY                         // Bit6   Y flip          (0=Normal, 1=Vertically mirrored)
	spritePriority                      // Bit7   OBJ-to-BG Priority (0=OBJ Above BG, 1=OBJ Behind BG color 1-3)
)

func (f spriteFlags) String() string {
	buf := make([]rune, 0, 30)

	buf = append(buf, 'C', 'G', '_', 'O', 'B', 'P', rune(48+(f&spriteCGBPalette)))
	buf = append(buf, '|', 'B', 'A', 'N', 'K', rune(48+(f&spriteCGBVramBank)))
	buf = append(buf, '|', 'O', 'B', 'P', rune(48+(f&spriteCGBVramBank)))
	if f&spriteFlipX > 0 {
		buf = append(buf, '|', 'X')
	}
	if f&spriteFlipY > 0 {
		buf = append(buf, '|', 'Y')
	}
	if f&spritePriority > 0 {
		buf = append(buf, '|', 'O', 'V', 'E', 'R')
	}

	return string(buf)
}

// var basePalette = [4]color.RGBA{
// 	{0xFF, 0xFF, 0xFF, 0xFF},
// 	{0xCC, 0xCC, 0xCC, 0xFF},
// 	{0x33, 0x33, 0x33, 0xFF},
// 	{0x00, 0x00, 0x00, 0xFF},
// }

var basePalette = [4]color.RGBA{
	// {0xd2, 0xe6, 0xa6,0xff}, // LCD OFF
	{0xc6, 0xde, 0x8c, 0xff},
	{0x84, 0xa5, 0x63, 0xff},
	{0x39, 0x61, 0x39, 0xff},
	{0x08, 0x18, 0x10, 0xff},
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

func (l lcdc) displayEnabled() bool { return l&lcdcDisplayEnable > 0 }
func (l lcdc) tileIDUnsigned() bool { return l&lcdcBgWindowSelect > 0 }
func (l lcdc) spriteEnabled() bool  { return l&lcdcObjEnable > 0 }

func (l lcdc) windowEnabled() bool { // TODO: double check bit 0 behaviour
	if l&lcdcPriority == 0 {
		return false
	}
	return l&lcdcWindowEnable > 0
}

func (l lcdc) spriteHeight() uint8 {
	if l&lcdcObjSize > 0 {
		return 16
	}
	return 8
}

type ppuMode uint8

const (
	modeHblank ppuMode = iota
	modeVblank
	modeOam
	modeTransfer
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

func (l lcdStat) lycIntEnabled() bool { return l&lcdStatCoincidenceInt > 0 }
func (l lcdStat) oamIntEnabled() bool { return l&lcdStatOAM > 0 }
func (l lcdStat) vblIntEnabled() bool { return l&lcdStatVBlank > 0 }
func (l lcdStat) hblIntEnabled() bool { return l&lcdStatHBlank > 0 }

func (l *lcdStat) updateLy(ly, lyc uint8) {
	if ly == lyc {
		*l |= lcdStatCoincidenceFlag
	} else {
		*l &^= lcdStatCoincidenceFlag
	}
}

func (l *lcdStat) setMode(m ppuMode) {
	*l &^= 0x03
	*l |= lcdStat(m) & 0x03
}

func (l *lcdStat) write(l2 uint8) {
	*l = lcdStat(l2) & 0x78
}

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

	nametables     *image.RGBA
	vram           *image.RGBA
	frames         uint64
	hideSprites    bool
	hideBackground bool
	hideWindow     bool
}

func (p *ppu) clock(gb *GameBoy) {
	if !p.LCDC.displayEnabled() {
		// return
	}

	switch {
	case p.LY >= 0 && p.LY <= 143:
		// mode 2 (oam search)
		if p.clocks >= 0 && p.clocks <= 79 {
			p.STAT.setMode(modeOam)
			if p.clocks == 0 {
				if p.STAT.oamIntEnabled() {
					gb.interruptCtrl.raise(lcdStatInterrupt)
				}
				p.oamSearch()
			}
		}

		// mode 3 (draw)
		if p.clocks >= 80 && p.clocks <= 251 {
			if p.clocks == 80 {
				p.STAT.setMode(modeTransfer)
			}

			if p.clocks == 251 {
				p.drawLine()
				p.drawSprites()
			}
		}

		// mode 0 (hblank)
		if p.clocks >= 252 && p.clocks <= 455 {
			p.STAT.setMode(modeHblank)
			if p.clocks == 252 && p.STAT.hblIntEnabled() {
				gb.interruptCtrl.raise(lcdStatInterrupt)
			}
		}

	// mode 1 (vblank)
	case p.LY >= 144 && p.LY <= 153:
		p.STAT.setMode(modeVblank)
		if p.clocks == 0 && p.LY == 144 {
			p.frames++
			p.drawNametables()
			p.drawVram()
			gb.interruptCtrl.raise(vblankInterrupt)
			if p.STAT.vblIntEnabled() {
				gb.interruptCtrl.raise(lcdStatInterrupt)
			}
		}
	}

	p.clocks++
	p.clocks %= 456
	if p.clocks == 0 {
		p.LY++
		p.LY %= 154
		p.STAT.updateLy(p.LY, p.LYC)
		if p.LY == p.LYC && p.STAT.lycIntEnabled() {
			gb.interruptCtrl.raise(lcdStatInterrupt)
		}
	}
}

func (p *ppu) drawLine() {
	wy := p.WY
	wx := p.WX - 7
	fineY := p.LY
	for fineX := uint8(0); fineX < 160; fineX++ {
		var window bool
		var tileIndex uint8
		var row uint8
		var tileLo, tileHi uint8
		if p.LCDC.windowEnabled() && p.LY >= wy && fineX >= wx {
			window = true
			row = (p.LY - wy) % 8
			tileIndex = p.tileIndex(fineX-wx, p.LY-wy, window)
			addr := p.tileBaseAddr(tileIndex) + uint16(row)*2
			tileLo = p.read(addr)
			tileHi = p.read(addr + 1)

			tileHi <<= (fineX - wx) % 8
			tileLo <<= (fineX - wx) % 8
		} else {
			row = (fineY + p.SCY) % 8
			tileIndex = p.tileIndex(fineX+p.SCX, fineY+p.SCY, window)
			addr := p.tileBaseAddr(tileIndex) + uint16(row)*2
			tileLo = p.read(addr)
			tileHi = p.read(addr + 1)

			tileHi <<= (fineX + p.SCX) % 8
			tileLo <<= (fineX + p.SCX) % 8
		}

		pixelLo := tileLo & 0x80 >> 7
		pixelHi := tileHi & 0x80 >> 7
		paletteIdx := pixelHi<<1 | pixelLo
		colour := p.paletteLookup(paletteIdx, p.BGP)

		if (window && p.hideWindow) || (!window && p.hideBackground) {
			colour = color.RGBA{}
		}
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

func (p *ppu) tileIndex(x, y uint8, window bool) uint8 {
	offset := uint16(y/8)*32 + uint16(x/8)

	mask := lcdcBgSelect
	if window {
		mask = lcdcWindowSelect
	}

	if p.LCDC&mask == 0 {
		return p.Nametable1[offset]
	}

	return p.Nametable2[offset]
}

func (p *ppu) tileBaseAddr(tileIdx uint8) uint16 {
	if p.LCDC.tileIDUnsigned() {
		return 0x8000 + uint16(tileIdx)*16
	}

	return 0x8800 + (uint16(int8(tileIdx))+128)*16
}

func (p *ppu) spriteTileBaseAddr(tileIdx uint16) uint16 {
	return uint16(0x8000) + tileIdx
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

	if addr >= 0x8000 && addr <= 0x9FFF {
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
		p.STAT.write(v)
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

	if addr >= 0x8000 && addr <= 0x9FFF {
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
	if !p.LCDC.spriteEnabled() {
		return
	}

	for i := 0; i < 40; i++ {
		var (
			spriteY     = int(p.OAM[i*4+0]) - 16
			spriteX     = p.OAM[i*4+1] - 8
			spriteTile  = p.OAM[i*4+2]
			spriteFlags = spriteFlags(p.OAM[i*4+3])
		)

		row := int(p.LY) - int(spriteY)
		if row < 0 || row > 7 {
			continue
		}
		if spriteFlags&spriteFlipY > 0 {
			row ^= 0x07
		}
		addr := 0x8000 + uint16(spriteTile)*16 + uint16(row)*2
		tileLo := p.read(addr)
		tileHi := p.read(addr + 1)

		if spriteFlags&spriteFlipX > 0 {
			tileLo = bits.Reverse8(tileLo)
			tileHi = bits.Reverse8(tileHi)
		}

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
			case spriteFlags&spriptePalette == 0:
				colour = p.paletteLookup(paletteIdx, p.OBP0)
			case spriteFlags&spriptePalette > 0:
				colour = p.paletteLookup(paletteIdx, p.OBP1)
			}

			if p.hideSprites {
				continue
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
			flags: spriteFlags(p.OAM[i*4+3]),
		}
		idx++

		if idx == 10 {
			break
		}
	}
}

func (p *ppu) paletteLookup(id, palette uint8) color.RGBA {
	shift := id * 2
	return basePalette[palette>>shift&0x03]
}

func (p *ppu) drawNametables() *image.RGBA {
	if p.nametables == nil {
		p.nametables = image.NewRGBA(image.Rect(0, 0, 512, 256))
	}
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			tileId := p.Nametable1[y*32+x]
			addr := p.tileBaseAddr(tileId)
			for fineY := 0; fineY < 8; fineY++ {
				tileLo := p.read(addr)
				addr++
				tileHi := p.read(addr)
				addr++
				for fineX := 0; fineX < 8; fineX++ {
					pixelLo := tileLo & 0x80 >> 7
					pixelHi := tileHi & 0x80 >> 7
					paletteIdx := pixelHi<<1 | pixelLo
					colour := p.paletteLookup(paletteIdx, p.BGP)

					tileLo <<= 1
					tileHi <<= 1

					p.nametables.Set(x*8+fineX, y*8+fineY, colour)
				}
			}
		}
	}

	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			tileId := p.Nametable2[y*32+x]
			addr := p.tileBaseAddr(tileId)
			for fineY := 0; fineY < 8; fineY++ {
				tileLo := p.read(addr)
				addr++
				tileHi := p.read(addr)
				addr++
				for fineX := 0; fineX < 8; fineX++ {
					pixelLo := tileLo & 0x80 >> 7
					pixelHi := tileHi & 0x80 >> 7
					paletteIdx := pixelHi<<1 | pixelLo
					colour := p.paletteLookup(paletteIdx, p.BGP)

					tileLo <<= 1
					tileHi <<= 1

					p.nametables.Set(256+x*8+fineX, y*8+fineY, colour)
				}
			}
		}
	}
	return p.nametables
}

func (p *ppu) drawVram() *image.RGBA {
	if p.vram == nil {
		p.vram = image.NewRGBA(image.Rect(0, 0, 128, 192))
	}

	addr := uint16(0x8000)
	for y := 0; y < 24; y++ {
		for x := 0; x < 16; x++ {
			for fineY := 0; fineY < 8; fineY++ {
				tileLo := p.read(addr)
				addr++
				tileHi := p.read(addr)
				addr++

				for fineX := 0; fineX < 8; fineX++ {
					pixelLo := tileLo & 0x80 >> 7
					pixelHi := tileHi & 0x80 >> 7
					paletteIdx := pixelHi<<1 | pixelLo
					colour := p.paletteLookup(paletteIdx, p.BGP)

					tileLo <<= 1
					tileHi <<= 1

					p.vram.Set(x*8+fineX, y*8+fineY, colour)
					// fmt.Println(y*8+fineY, x*8+fineX)
					// _ = colour
					// p.vram.Set(x*8+fineX, y*8+fineY, color.RGBA{0xff, 0x00, 0x00, 0xff})
				}
			}
		}
	}

	return p.vram
}
