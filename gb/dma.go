package gb

type dmaCtrl struct {
	src    uint16
	target uint16
}

func (d *dmaCtrl) clock(gb *GameBoy) {
	if gb.state&dma > 0 {
		v := gb.read(d.src)
		// fmt.Printf("dma transfer 0x%04X -> 0x%04X (0x%02X)\n", d.src, d.target, v)
		gb.ppu.write(d.target, v)
		d.src++
		d.target++
		if d.target > 0xFE9F {
			gb.state &^= dma
		}
	}
}

func (d *dmaCtrl) write(addr uint16, v uint8) {
	switch addr {
	case ioRegs.DMA:
		d.src = uint16(v) << 8
		d.target = 0xFE00
	}
	unmappedWrite("dma controller", addr, v)
}

func (d *dmaCtrl) read(addr uint16) uint8 {
	unmappedRead("dma controller", addr)
	return 0
}
