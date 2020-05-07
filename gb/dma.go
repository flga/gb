package gb

type dmaCtrl struct{}

func (d *dmaCtrl) clock(gb *GameBoy) {}

func (d *dmaCtrl) write(addr uint16, v uint8) {
	unmappedWrite("dma controller", addr, v)
}

func (d *dmaCtrl) read(addr uint16) uint8 {
	unmappedRead("dma controller", addr)
	return 0
}
