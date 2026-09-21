package bios

import (
	"goemu/config"
	"os"
)

type BIOS struct {
	Bios [16 * 1024]byte // load in the 16kb GBA bios
	Ptr  uint32          // pointer to the current instruction in the bios
}

func (b *BIOS) LoadBios(bios []byte) {
	biosPath := config.NewConfig().BiosPath
	biosFile, err := os.ReadFile(biosPath)
	if err != nil {
		panic("Failed to load BIOS file: " + err.Error())
	}
	copy(b.Bios[:], biosFile)
}
