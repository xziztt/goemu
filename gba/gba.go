package gba

import "goemu/gba/cpu"

type GBA struct {
	CPU  cpu.CPU
	Save bool
}
