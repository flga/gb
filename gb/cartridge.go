package gb

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

type rom []uint8

func (r rom) read(addr uint64) uint8 {
	return r[int(addr)%len(r)]
}

type Cartridge struct {
	CartridgeInfo
	mbc       mbc
	savWriter io.WriteCloser
}

func NewCartridge(r io.Reader) (*Cartridge, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("unable to read rom: %w", err)
	}

	info := parseCartridgeHeader(data)
	mbcFunc, ok := mbcs[info.CartridgeType]
	if !ok {
		return nil, fmt.Errorf("unsupported mapper %02x", info.CartridgeType)
	}

	mbc := mbcFunc(rom(data), info)

	fmt.Printf("mapper: %T\n", mbc)
	ret := &Cartridge{
		CartridgeInfo: info,
		mbc:           mbc,
	}

	return ret, nil
}

func (c *Cartridge) clock(gb *GameBoy) {
	c.mbc.clock(gb)
}

func (c *Cartridge) read(addr uint16) uint8 {
	return c.mbc.read(addr)
}

func (c *Cartridge) write(addr uint16, v uint8) {
	c.mbc.write(addr, v)
}

func (c *Cartridge) Saveable() bool {
	return c.mbc.saveable()
}

func (c *Cartridge) loadSave(d []byte) {
	if !c.Saveable() {
		return
	}

	c.mbc.loadSave(d)
}

func (c *Cartridge) save() error {
	if !c.Saveable() {
		return nil
	}

	data := c.mbc.save()

	if _, err := c.savWriter.Write(data); err != nil {
		return err
	}

	return c.savWriter.Close()
}

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

func parseCartridgeHeader(data []byte) CartridgeInfo {
	var cm CartridgeInfo

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

	return cm
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
