package gb

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

type rom []uint8

func (r rom) read(addr uint16) uint8 {
	return r[int(addr)%cap(r)]
}

type cartridge struct {
	rom    rom
	mapper mapper

	info CartridgeInfo
}

func newCartridge(r io.Reader) (*cartridge, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("unable to read rom: %w", err)
	}

	ret := &cartridge{
		rom:  rom(data),
		info: CartridgeInfo{},
	}

	ret.info.parseFrom(data)

	return ret, nil
}

func (c *cartridge) translateRead(addr uint16) uint16 {
	return c.mapper.translateRead(addr)
}

func (c *cartridge) translateWrite(addr uint16) uint16 {
	return c.mapper.translateWrite(addr)
}

func (c *cartridge) read(addr uint16) uint8 {
	return 0 // todo
}

func (c *cartridge) write(addr uint16, v uint8) {
	// todo
}

type mapper interface {
	translateRead(addr uint16) uint16
	translateWrite(addr uint16) uint16
}

type CartridgeInfo struct {
	Title                string
	ManufacturerCode     string
	CGBFlag              uint8
	NewLicenseeCode      string
	SGBFlag              uint8
	CartridgeType        uint8
	ROMSize              uint8
	RAMSize              uint8
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
	cm.ROMSize = data[0x0148]
	cm.RAMSize = data[0x0149]
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
ROMSize: 0x%X
RAMSize: 0x%X
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
