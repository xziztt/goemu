package cpu

import (
	"fmt"
	"slices"
)

type Bus struct {
	memory Memory
}

func GetMemoryRegion(m uint32) (MemoryRegions, error) {
	switch {
	case m >= 0x00 && m < 0x02:
		return BiosRegion, nil
	case m >= 0x02 && m < 0x03:
		return EWRAMRegion, nil
	case m >= 0x03 && m < 0x04:
		return IWRAMRegion, nil
	case m >= 0x04 && m < 0x05:
		return IORegion, nil
	case m >= 0x05 && m < 0x06:
		return PaletteRAMRegion, nil
	case m >= 0x06 && m < 0x07:
		return VRAMRegion, nil
	case m >= 0x07 && m < 0x08:
		return OAMRegion, nil
	case m >= 0x08 && m < 0x0E:
		return GamePakROM0Region, nil
	default:
		return 0, fmt.Errorf("invalid memory region") // Invalid region
	}
}

// Each memory region has different access times, calculate cycle count based on region and size
func CalculateCycleCount(memoryRegion MemoryRegions, size int) (int, error) {
	switch memoryRegion {
	case BiosRegion:
		return 1, nil
	case EWRAMRegion:
		return 1, nil
	case IWRAMRegion:
		if slices.Contains([]int{8, 16}, size) {
			return 3, nil
		} else {
			return 6, nil
		}
	case IORegion:
		return 1, nil
	case OAMRegion:
		return 1, nil
	case PaletteRAMRegion:
		if slices.Contains([]int{8, 16}, size) {
			return 1, nil
		} else {
			return 2, nil
		}
	case VRAMRegion:
		if slices.Contains([]int{8, 16}, size) {
			return 1, nil
		} else {
			return 2, nil
		}
	case GamePakROM0Region, GamePakROM1Region, GamePakROM2Region:
		if slices.Contains([]int{8, 16}, size) {
			return 5, nil
		} else {
			return 8, nil
		}
	case GamePakSaveRegion:
		return 5, nil
	}
	return 0, nil // Invalid region
}

//	00000000-00003FFF   BIOS - System ROM         (16 KBytes)
//	00004000-01FFFFFF   Not used
//	02000000-0203FFFF   WRAM - On-board Work RAM  (256 KBytes) 2 Wait
//	02040000-02FFFFFF   Not used
//	03000000-03007FFF   WRAM - On-chip Work RAM   (32 KBytes)
//	03008000-03FFFFFF   Not used
//	04000000-040003FE   I/O Registers
//	04000400-04FFFFFF   Not used
//
// 05000000-050003FF   BG/OBJ Palette RAM       (1 KB)
// 05000400-05FFFFFF   Not used (mirrored)
//
// 06000000-06017FFF   VRAM - Video RAM         (96 KB)
// 06018000-06FFFFFF   Not used (partially mirrored)
//
// 07000000-070003FF   OAM - OBJ Attributes     (1 KB)
// 07000400-07FFFFFF   Not used (mirrored)
//
// 08000000-09FFFFFF   Game Pak ROM - Wait State 0
// 0A000000-0BFFFFFF   Game Pak ROM - Wait State 1
// 0C000000-0DFFFFFF   Game Pak ROM - Wait State 2
//
// 0E000000-0E00FFFF   Game Pak SRAM            (up to 64 KB)
// 0E010000-0FFFFFFF   SRAM region mirrors / unused
func (b *Bus) Read8(addr uint32) uint32 {
	// We can use the first 8 bits to check the region
	switch addr >> 24 {
	case 0x00:
		return uint32(b.memory.Bios[addr])
	case 0x02:
		return uint32(b.memory.IWRam[addr])
	case 0x03:
		return uint32(b.memory.EWRam[addr])
	case 0x04:
		return uint32(b.memory.IORegion[addr])
	case 0x05:
		return uint32(b.memory.PaletteRAMRegion[addr])
	case 0x06:
		return uint32(b.memory.VRAMRegion[addr])
	case 0x07:
		return uint32(b.memory.OAMRegion[addr])
	case 0x08, 0x09:
		return uint32(b.memory.GamePakROM0Region[addr])
	case 0x0A, 0x0B:
		return uint32(b.memory.GamePakROM1Region[addr])
	case 0x0C, 0x0D:
		return uint32(b.memory.GamePakROM2Region[addr])
	case 0xE, 0xF:
		return uint32(b.memory.GamePakSaveRegion[addr])
	default:
		return 0
	}
}
