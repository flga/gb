package gb

type cartridge struct {
	mapper *mapper
}

func (c *cartridge) translate(addr uint16) uint16 {
	return c.mapper.translate(addr)
}

// gb probably has a mapper0 too, so this 0 value should be useful
type mapper struct{}

func (m *mapper) translate(addr uint16) uint16 {
	//todo: memory map
	return addr
}
