package gb

import "fmt"

type sram []uint8

func (r sram) read(addr uint32) uint8 {
	if len(r) == 0 {
		return 0xFF
	}
	return r[addr%uint32(len(r))]
}

func (r sram) write(addr uint32, v uint8) {
	if len(r) == 0 {
		return
	}
	r[addr%uint32(len(r))] = v
}

type mbc interface {
	clock(gb *GameBoy)
	read(addr uint16) uint8
	write(addr uint16, v uint8)
	saveable() bool
	save() []byte
	loadSave(d []byte)
}

type mbcFlags uint8

const (
	none   mbcFlags = 0
	mbcRam mbcFlags = 1 << iota
	mbcBattery
	mbcTimer
)

var mbcs = map[uint8]func(rom, CartridgeInfo) mbc{
	0x00: func(rom rom, _ CartridgeInfo) mbc { return mbc0{rom: rom} },
	0x01: newMbc1(none),
	0x02: newMbc1(mbcRam),
	0x03: newMbc1(mbcRam | mbcBattery),
	0x05: newMbc2(none),
	0x06: newMbc2(mbcBattery),
	// 0x08: mapper0{ram: true},
	// 0x09: mapper0{ram: true, battery: true},
	// 0x0B: mmm01{},
	// 0x0C: mmm01{ram: true},
	// 0x0D: mmm01{ram: true, battery: true},
	// 0x0F: mbc3{timer: true, battery: true},
	// 0x10: mbc3{ram: true, timer: true, battery: true},
	0x11: newMbc3(none),
	0x12: newMbc3(mbcRam),
	0x13: newMbc3(mbcRam | mbcBattery),
	// 0x19: mbc5{},
	// 0x1A: mbc5{ram: true},
	// 0x1B: mbc5{ram: true, battery: true},
	// 0x1C: mbc5{rumble: true},
	// 0x1D: mbc5{ram: true, rumble: true},
	// 0x1E: mbc5{ram: true, battery: true, rumble: true},
	// 0x20: mbc6{ram: true, battery: true},
	// 0x22: mbc7{ram: true, battery: true, accelerometer: true},
	// 0xFC: pocket{camera: true},
	// 0xFD: bandai{tama5: true},
	// 0xFE: huc3{},
	// 0xFF: huc1{ram: true, battery: true},
}

type mbc0 struct {
	rom rom
}

func (mbc0) clock(gb *GameBoy) {}

func (m mbc0) read(addr uint16) uint8 {
	if addr >= 0x0000 && addr <= 0x7FFF {
		return m.rom.read(uint64(addr))
	}

	return 0xff
}

func (m mbc0) write(addr uint16, v uint8) {}

func (mbc0) saveable() bool  { return false }
func (mbc0) save() []byte    { return nil }
func (mbc0) loadSave([]byte) {}

type mbc1 struct {
	rom        rom
	ram        sram
	battery    bool
	ramEnabled bool

	bankMode uint8
	bankLo   uint8
	bankHi   uint8
}

func newMbc1(f mbcFlags) func(rom, CartridgeInfo) mbc {
	return func(rom rom, c CartridgeInfo) mbc {
		v := &mbc1{
			rom:     rom,
			bankLo:  1,
			battery: f&mbcBattery > 0,
		}

		if f&mbcRam > 0 {
			v.ram = make(sram, c.RAMSize)
		}

		return v
	}
}

func (*mbc1) clock(gb *GameBoy) {}

func (m *mbc1) read(addr uint16) uint8 {
	if addr >= 0x0000 && addr <= 0x3FFF {
		var bank uint64
		if m.bankMode == 1 {
			bank = uint64(m.bankHi << 5)
		}
		return m.rom.read(bank*0x4000 + uint64(addr))
	}

	if addr >= 0x4000 && addr <= 0x7FFF {
		bank := uint64(m.bankHi<<5 | m.bankLo)
		return m.rom.read(bank*0x4000 + uint64(addr-0x4000))
	}

	if addr >= 0xA000 && addr <= 0xBFFF {
		if !m.ramEnabled {
			return 0xFF
		}

		var bank uint32
		if m.bankMode == 1 {
			bank = uint32(m.bankHi)
		}

		return m.ram.read(bank*0x2000 + uint32(addr-0xA000))
	}

	return 0xFF
}

func (m *mbc1) write(addr uint16, v uint8) {
	if addr >= 0x0000 && addr <= 0x1FFF {
		m.ramEnabled = v&0x0F == 0x0A
		return
	}

	if addr >= 0x2000 && addr <= 0x3FFF {
		v &= 0x1F
		if v == 0 {
			v++
		}

		m.bankLo = v
		return
	}

	if addr >= 0x4000 && addr <= 0x5FFF {
		v &= 0x03

		m.bankHi = v
		return
	}

	if addr >= 0x6000 && addr <= 0x7FFF {
		v &= 0x01

		m.bankMode = v
		return
	}

	if addr >= 0xA000 && addr <= 0xBFFF {
		if !m.ramEnabled {
			return
		}

		var bank uint32
		if m.bankMode == 1 {
			bank = uint32(m.bankHi)
		}

		m.ram.write(bank*0x2000+uint32(addr-0xA000), v)
		return
	}
}

func (m *mbc1) saveable() bool { return m.battery }
func (m *mbc1) save() []byte   { return m.ram[:] }
func (m *mbc1) loadSave(d []byte) {
	copy(m.ram[:], d)
}

type mbc2 struct {
	rom        rom
	ram        sram
	romBank    uint64
	ramEnabled bool
	battery    bool
}

func newMbc2(f mbcFlags) func(rom, CartridgeInfo) mbc {
	return func(rom rom, c CartridgeInfo) mbc {
		return &mbc2{
			rom:     rom,
			ram:     make(sram, 512),
			romBank: 1,
			battery: f&mbcBattery > 0,
		}
	}
}

func (*mbc2) clock(gb *GameBoy) {}

func (m *mbc2) read(addr uint16) uint8 {
	if addr >= 0x0000 && addr <= 0x3FFF {
		return m.rom.read(uint64(addr))
	}

	if addr >= 0x4000 && addr <= 0x7FFF {
		bank := uint64(m.romBank)
		return m.rom.read(bank*0x4000 + uint64(addr-0x4000))
	}

	if m.ramEnabled && addr >= 0xA000 && addr <= 0xBFFF {
		return m.ram.read(uint32(addr-0xA000)) | 0xF0
	}

	return 0xFF
}

func (m *mbc2) write(addr uint16, v uint8) {
	if addr >= 0x0000 && addr <= 0x3FFF {
		if addr&0x0100 == 0 {
			m.ramEnabled = v&0x0F == 0x0A
			return
		}

		m.romBank = uint64(v) & 0x0F
		if m.romBank == 0 {
			m.romBank++
		}

		return
	}

	if m.ramEnabled && addr >= 0xA000 && addr <= 0xBFFF {
		m.ram.write(uint32(addr-0xA000), v&0x0F)
		return
	}
}

func (m *mbc2) saveable() bool { return m.battery }
func (m *mbc2) save() []byte   { return m.ram[:] }
func (m *mbc2) loadSave(d []byte) {
	copy(m.ram[:], d)
}

type mbc3 struct {
	rom     rom
	ram     sram
	battery bool

	ramRTC bool

	romBank    uint8
	ramRtcBank uint8

	RTCS  uint8 // 08h Seconds   0-59 (0-3Bh)
	RTCM  uint8 // 09h Minutes   0-59 (0-3Bh)
	RTCH  uint8 // 0Ah Hours     0-23 (0-17h)
	RTCDL uint8 // 0Bh Lower 8 bits of Day Counter (0-FFh)
	RTCDH uint8 // 0Ch Upper 1 bit of Day Counter, Carry Bit, Halt Flag
	//                 Bit 0  Most significant bit of Day Counter (Bit 8)
	//                 Bit 6  Halt (0=Active, 1=Stop Timer)
	//                 Bit 7  Day Counter Carry Bit (1=Counter Overflow)

	prevRtcWrite uint8
	clocks       uint64
}

func newMbc3(f mbcFlags) func(rom, CartridgeInfo) mbc {
	return func(rom rom, c CartridgeInfo) mbc {
		v := &mbc3{
			rom:     rom,
			romBank: 1,
			battery: f&mbcBattery > 0,
		}

		if f&mbcRam > 0 {
			v.ram = make(sram, c.RAMSize)
		}

		return v
	}
}

func (m *mbc3) clock(gb *GameBoy) {
	if m.clocks%128 == 0 {
		// rtc
	}
	m.clocks++
}

func (m *mbc3) read(addr uint16) uint8 {
	// ROM Bank 0
	if addr >= 0x0000 && addr <= 0x3FFF {
		return m.rom.read(uint64(addr))
	}

	// Switchable ROM bank
	if addr >= 0x4000 && addr <= 0x7FFF {
		bank := uint64(m.romBank) * 0x4000
		return m.rom.read(bank + uint64(addr-0x4000))
	}

	// RAM Bank/RTC Register
	if addr >= 0xA000 && addr <= 0xBFFF {
		if m.ramRtcBank >= 0x00 && m.ramRtcBank <= 0x07 {
			bank := uint32(m.ramRtcBank)
			// return m.ram[(bank+int(addr-0xA000))%len(m.ram)]
			return m.ram.read(bank*0x2000 + uint32(addr-0xA000))
		} else {
			// rtc
		}

		return 0xff
	}

	panic(fmt.Sprintf("hm? %04x", addr))
	return 0xff
}

func (m *mbc3) write(addr uint16, v uint8) {
	// RAM and RTC Registers Enable
	if addr >= 0x0000 && addr <= 0x1FFF {
		m.ramRTC = v&0x0F == 0x0A
		return
	}

	// ROM Bank
	if addr >= 0x2000 && addr <= 0x3FFF {
		v &= 0x7F
		if v == 0 {
			v++
		}
		m.romBank = v
		return
	}

	// RAM Bank/RTC Register
	if addr >= 0x4000 && addr <= 0x5FFF {
		v &= 0x0F
		m.ramRtcBank = v
		return
	}

	// Latch Clock Data
	if addr >= 0x6000 && addr <= 0x7FFF {
		if v == 1 && m.prevRtcWrite == 0 {
			m.latch()
		}

		m.prevRtcWrite = v
		return
	}

	// RAM Bank/RTC Register
	if addr >= 0xA000 && addr <= 0xBFFF {
		if m.ramRtcBank >= 0x00 && m.ramRtcBank <= 0x07 {
			bank := uint32(m.ramRtcBank)
			m.ram.write(bank*0x2000+uint32(addr-0xA000), v)
			return
		} else {
			// rtc
		}

		return
	}

	panic(fmt.Sprintf("hm? %04x %02x", addr, v))
}

func (m *mbc3) saveable() bool { return m.battery }
func (m *mbc3) save() []byte   { return m.ram[:] }
func (m *mbc3) loadSave(d []byte) {
	copy(m.ram[:], d)
}

func (m *mbc3) latch() {}
