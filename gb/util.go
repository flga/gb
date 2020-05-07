package gb

import "fmt"

type size int

const (
	Byte size = 1
	KiB       = 1024 * Byte
	MiB       = 1024 * KiB
)

func (s size) String() string {
	if s > MiB {
		return fmt.Sprintf("%dMiB", s/MiB)
	}
	if s > KiB {
		return fmt.Sprintf("%dKiB", s/KiB)
	}
	return fmt.Sprintf("%d", s)
}
