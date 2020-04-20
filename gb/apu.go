package gb

type apu struct{}

func (a *apu) clock(b bus) {
	//todo: audio samples
}

func (a *apu) read(addr uint16) uint8 {
	return 0 // todo
}

func (a *apu) write(addr uint16, v uint8) {
	// todo
}

type pulse struct{}
type wave struct{}
type noise struct{}
