package gb

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNewGameBoy(t *testing.T) {
	f, err := os.Open(filepath.Join("../testdata", "flappyboy.gb"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	gb := New()
	if err := gb.InsertCartridge(f); err != nil {
		t.Fatal(err)
	}

	fmt.Println(gb.CartridgeInfo())
}
