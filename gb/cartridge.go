package gb

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

type rom []byte

func (r rom) read(addr uint16) byte {
	return r[int(addr)%cap(r)]
}

type Cartridge struct {
	rom    rom
	mapper mapper

	info CartridgeInfo
}

func NewCartridge(r io.Reader) (*Cartridge, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("unable to read rom: %w", err)
	}

	ret := &Cartridge{
		rom:    rom(data),
		info:   CartridgeInfo{},
		mapper: mapper0{},
	}

	ret.info.parseFrom(data)

	return ret, nil
}

func (c *Cartridge) translateRead(addr uint16) uint16 {
	if c == nil {
		return addr // TODO: this is hack to avoid the race condition of reg initialization when cart not present, need to rethink that
	}
	return c.mapper.translateRead(addr)
}

func (c *Cartridge) translateWrite(addr uint16) uint16 {
	if c == nil {
		return addr // TODO: this is hack to avoid the race condition of reg initialization when cart not present, need to rethink that
	}
	return c.mapper.translateWrite(addr)
}

func (c *Cartridge) read(addr uint16) uint8 {
	if addr >= 0xA000 && addr <= 0xBFFF {
		return 0
	}
	return c.rom.read(addr)
}

func (c *Cartridge) write(addr uint16, v uint8) {
	panic("write to rom")
	// todo
}

type mapper interface {
	translateRead(addr uint16) uint16
	translateWrite(addr uint16) uint16
}

type mapper0 struct{}

func (mapper0) translateRead(addr uint16) uint16  { return addr }
func (mapper0) translateWrite(addr uint16) uint16 { return addr }

type CartridgeInfo struct {
	Title                string
	ManufacturerCode     string
	CGBFlag              uint8
	NewLicenseeCode      string
	SGBFlag              uint8
	CartridgeType        uint8
	ROMSize              size
	RAMSize              size
	DestinationCode      uint8
	OldLicenseeCode      uint8
	MaskROMVersionNumber uint8
	HeaderChecksum       uint8
	GlobalChecksum       []uint8
}

func (cm *CartridgeInfo) parseFrom(data []byte) {
	cm.Title = strings.TrimRight(string(data[0x0134:0x0143+1]), "\x00")
	cm.ManufacturerCode = strings.TrimRight(string(data[0x013F:0x0142+1]), "\x00")
	cm.CGBFlag = data[0x0143]
	cm.NewLicenseeCode = strings.TrimRight(string(data[0x0144:0x0145+1]), "\x00")
	cm.SGBFlag = data[0x0146]
	cm.CartridgeType = data[0x0147]
	cm.ROMSize = 32 * KiB << uint64(data[0x0148])
	switch data[0x0149] {
	case 0x00:
		cm.RAMSize = 0 * KiB
	case 0x01:
		cm.RAMSize = 2 * KiB
	case 0x02:
		cm.RAMSize = 8 * KiB
	case 0x03:
		cm.RAMSize = 32 * KiB
	case 0x04:
		cm.RAMSize = 128 * KiB
	case 0x05:
		cm.RAMSize = 64 * KiB
	}
	cm.DestinationCode = data[0x014A]
	cm.OldLicenseeCode = data[0x014B]
	cm.MaskROMVersionNumber = data[0x014C]
	cm.HeaderChecksum = data[0x014D]
	cm.GlobalChecksum = data[0x014E : 0x014F+1]
}

func (cm CartridgeInfo) String() string {
	return fmt.Sprintf(`Title: %q
ManufacturerCode: %q
CGBFlag: 0x%X
NewLicenseeCode: %q
SGBFlag: 0x%X
CartridgeType: 0x%X
ROMSize: %v
RAMSize: %v
DestinationCode: 0x%X
OldLicenseeCode: 0x%X
MaskROMVersionNumber: 0x%X
HeaderChecksum: 0x%X
GlobalChecksum: 0x%X`,
		cm.Title,
		cm.ManufacturerCode,
		cm.CGBFlag,
		cm.NewLicenseeCode,
		cm.SGBFlag,
		cm.CartridgeType,
		cm.ROMSize,
		cm.RAMSize,
		cm.DestinationCode,
		cm.OldLicenseeCode,
		cm.MaskROMVersionNumber,
		cm.HeaderChecksum,
		cm.GlobalChecksum,
	)
}
