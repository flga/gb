package gb

type apu struct {
	ChannelControl uint8 // 0xFF24 - NR50 - Channel control / ON-OFF / Volume (R/W)
	OutputTerminal uint8 // 0xFF25 - NR51 - Selection of Sound output terminal (R/W)
	OnOff          uint8 // 0xFF26 - NR52 - Sound on/off

	p1    *pulse
	p2    *pulse
	wave  *wave
	noise *noise
}

func (a *apu) clock(b bus) {
	a.p1.clock(b)
	a.p2.clock(b)
	a.wave.clock(b)
	a.noise.clock(b)
}

func (a *apu) sample() float64 {
	p1 := a.p1.sample()
	p2 := a.p2.sample()
	wave := a.wave.sample()
	noise := a.noise.sample()

	_ = p1
	_ = p2
	_ = wave
	_ = noise
	// TODO: combine samples
	return 0
}

func (a *apu) read(addr uint16) uint8 {
	// apu ctrl
	switch addr {
	case 0xFF24:
		return a.ChannelControl
	case 0xFF25:
		return a.OutputTerminal
	case 0xFF26:
		return a.OnOff
	}

	// pulse1
	if addr >= 0xFF10 && addr <= 0xFF14 {
		return a.p1.read(addr)
	}

	// pulse2
	if addr >= 0xFF16 && addr <= 0xFF19 {
		return a.p2.read(addr)
	}

	// wave
	if addr >= 0xFF1A && addr <= 0xFF1E {
		return a.wave.read(addr)
	}

	// wave pattern
	if addr >= 0xFF30 && addr <= 0xFF3F {
		return a.wave.read(addr)
	}

	// noise
	if addr >= 0xFF20 && addr <= 0xFF23 {
		return a.noise.read(addr)
	}

	return 0 // todo
}

func (a *apu) write(addr uint16, v uint8) {
	// todo
}

type pulse struct {
	Sweep          uint8 // 0xFF10 - NR10 - Channel 1 Sweep register (R/W)
	Length         uint8 // 0xFF11/0xFF16 - NR11/NR21 - Channel 1/2 Sound length/Wave pattern duty (R/W)
	VolumeEnvelope uint8 // 0xFF12/0xFF17 - NR12/NR22 - Channel 1/2 Volume Envelope (R/W)
	FreqLo         uint8 // 0xFF13/0xFF18 - NR13/NR23 - Channel 1/2 Frequency lo (Write Only)
	FreqHi         uint8 // 0xFF14/0xFF19 - NR14/NR24 - Channel 1/2 Frequency hi (R/W)

	isPulse1 bool
}

func (p *pulse) clock(b bus)     {}
func (p *pulse) sample() float64 { return 0 }

func (p *pulse) read(addr uint16) uint8 {
	if p.isPulse1 {
		switch addr {
		case 0xFF10:
			return p.Sweep
		case 0xFF11:
			return p.Length
		case 0xFF12:
			return p.VolumeEnvelope
		case 0xFF14:
			return p.FreqHi
		}
		// panic(fmt.Sprintf("unhandled pulse1 read 0x%04X", addr))
	}

	switch addr {
	case 0xFF16:
		return p.Length
	case 0xFF17:
		return p.VolumeEnvelope
	case 0xFF19:
		return p.FreqHi
	}
	// panic(fmt.Sprintf("unhandled pulse2 read 0x%04X", addr))
	return 0
}

func (p *pulse) write(addr uint16, v uint8) {
	if p.isPulse1 {
		switch addr {
		case 0xFF10:
			p.Sweep = v
			return
		case 0xFF11:
			p.Length = v
			return
		case 0xFF12:
			p.VolumeEnvelope = v
			return
		case 0xFF13:
			p.FreqLo = v
			return
		case 0xFF14:
			p.FreqHi = v
			return
		}
		// panic(fmt.Sprintf("unhandled pulse1 write 0x%04X: 0x%02X", addr, v))
	}

	switch addr {
	case 0xFF16:
		p.Length = v
		return
	case 0xFF17:
		p.VolumeEnvelope = v
		return
	case 0xFF18:
		p.FreqLo = v
		return
	case 0xFF19:
		p.FreqHi = v
		return
	}
	// panic(fmt.Sprintf("unhandled pulse2 write 0x%04X: 0x%02X", addr, v))
}

type wave struct {
	OnOff       uint8     // 0xFF1A - NR30 - Channel 3 Sound on/off (R/W)
	Length      uint8     // 0xFF1B - NR31 - Channel 3 Sound Length
	OutputLevel uint8     // 0xFF1C - NR32 - Channel 3 Select output level (R/W)
	FreqLo      uint8     // 0xFF1D - NR33 - Channel 3 Frequency's lower data (W)
	FreqHi      uint8     // 0xFF1E - NR34 - Channel 3 Frequency's higher data (R/W)
	Pattern     [16]uint8 // 0xFF30-0xFF3F - Wave Pattern RAM
}

func (w *wave) clock(b bus)     {}
func (w *wave) sample() float64 { return 0 }

func (w *wave) read(addr uint16) uint8 {
	if addr >= 0xFF30 && addr <= 0xFF3F {
		return w.Pattern[addr-0xFF30]
	}

	switch addr {
	case 0xFF1A:
		return w.OnOff
	case 0xFF1B:
		return w.Length
	case 0xFF1C:
		return w.OutputLevel
	case 0xFF1E:
		return w.FreqHi
	}

	// panic(fmt.Sprintf("unhandled wave read 0x%04X", addr))
	return 0
}

func (w *wave) write(addr uint16, v uint8) {
	if addr >= 0xFF30 && addr <= 0xFF3F {
		w.Pattern[0xFF30-addr] = v
		return
	}

	switch addr {
	case 0xFF1A:
		w.OnOff = v
		return
	case 0xFF1B:
		w.Length = v
		return
	case 0xFF1C:
		w.OutputLevel = v
		return
	case 0xFF1D:
		w.FreqLo = v
		return
	case 0xFF1E:
		w.FreqHi = v
		return
	}

	// panic(fmt.Sprintf("unhandled wave write 0x%04X: 0x%02X", addr, v))
}

type noise struct {
	Length         uint8 // 0xFF20 - NR41 - Channel 4 Sound Length (R/W)
	VolumeEnvelope uint8 // 0xFF21 - NR42 - Channel 4 Volume Envelope (R/W)
	Counter        uint8 // 0xFF22 - NR43 - Channel 4 Polynomial Counter (R/W)
	CounterLoad    uint8 // 0xFF23 - NR44 - Channel 4 Counter/consecutive; Inital (R/W)
}

func (n *noise) clock(b bus)     {}
func (n *noise) sample() float64 { return 0 }

func (n *noise) read(addr uint16) uint8 {
	switch addr {
	case 0xFF20:
		return n.Length
	case 0xFF21:
		return n.VolumeEnvelope
	case 0xFF22:
		return n.Counter
	case 0xFF23:
		return n.CounterLoad
	}

	// panic(fmt.Sprintf("unhandled noise read 0x%04X", addr))
	return 0

}

func (n *noise) write(addr uint16, v uint8) {
	switch addr {
	case 0xFF20:
		n.Length = v
		return
	case 0xFF21:
		n.VolumeEnvelope = v
		return
	case 0xFF22:
		n.Counter = v
		return
	case 0xFF23:
		n.CounterLoad = v
		return
	}

	// panic(fmt.Sprintf("unhandled noise write 0x%04X: 0x%02X", addr, v))
}
