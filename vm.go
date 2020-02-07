package main

import (
	"fmt"
	"io/ioutil"

	"github.com/faiface/pixel/pixelgl"
	"github.com/gammazero/deque"
)

// chipVM holds all the register, memory, etc.
type chipVM struct {
	v            [16]uint8
	i            uint16
	sound, delay uint8
	stack        deque.Deque
	memory       [4096]uint8
	display      [32][64]int
	keys         map[pixelgl.Button]uint8
	pc           int
}

func newChipVM(file string) *chipVM {
	vm := new(chipVM)
	vm.sound = 0
	vm.delay = 0
	vm.setFont()
	vm.keys = map[pixelgl.Button]uint8{
		pixelgl.Key1: 0x0,
		pixelgl.Key2: 0x1,
		pixelgl.Key3: 0x2,
		pixelgl.Key4: 0x3,
		pixelgl.KeyQ: 0x4,
		pixelgl.KeyW: 0x5,
		pixelgl.KeyE: 0x6,
		pixelgl.KeyR: 0x7,
		pixelgl.KeyA: 0x8,
		pixelgl.KeyS: 0x9,
		pixelgl.KeyD: 0xA,
		pixelgl.KeyF: 0xB,
		pixelgl.KeyZ: 0xC,
		pixelgl.KeyX: 0xD,
		pixelgl.KeyC: 0xE,
		pixelgl.KeyV: 0xF,
	}
	rom, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	for addr := 0x200; addr < len(rom)+0x200; addr++ {
		vm.memory[addr] = rom[addr-0x200]
		fmt.Printf("addr:\t%x\topcode:\t%x\n", addr, vm.memory[addr])
	}
	return vm
}

func (vm *chipVM) setFont() {
	fontlist := [][]int{
		{0xF0, 0x90, 0x90, 0x90, 0xF0}, // 0
		{0x20, 0x60, 0x20, 0x20, 0x70}, // 1
		{0xF0, 0x10, 0xF0, 0x80, 0xF0}, // 2
		{0xF0, 0x10, 0xF0, 0x10, 0xF0}, // 3
		{0x90, 0x90, 0xF0, 0x10, 0x10}, // 4
		{0xF0, 0x80, 0xF0, 0x10, 0xF0}, // 5
		{0xF0, 0x80, 0xF0, 0x90, 0xF0}, // 6
		{0xF0, 0x10, 0x20, 0x40, 0x40}, // 7
		{0xF0, 0x90, 0xF0, 0x90, 0xF0}, // 8
		{0xF0, 0x90, 0xF0, 0x10, 0xF0}, // 9
		{0xF0, 0x90, 0xF0, 0x90, 0x90}, // a
		{0xE0, 0x90, 0xE0, 0x90, 0xE0}, // b
		{0xF0, 0x80, 0x80, 0x80, 0xF0}, // c
		{0xE0, 0x90, 0x90, 0x90, 0xE0}, // d
		{0xF0, 0x80, 0xF0, 0x80, 0xF0}, // e
		{0xF0, 0x80, 0xF0, 0x80, 0x80}, // f
	}
	i := 0
	for _, sprite := range fontlist {
		for _, v := range sprite {
			vm.memory[i] = uint8(v)
			i++
		}
	}
}

func (vm *chipVM) fetchNextOpcode(win *pixelgl.Window) {
	var opcode int = int(vm.memory[vm.pc])<<0x8 | int(vm.memory[vm.pc+1])
	firstnibble := (opcode & 0xF000) >> 0xC
	x := (opcode & 0x0F00) >> 0x8
	y := (opcode & 0x00F0) >> 0x4
	vm.pc += 2
	switch firstnibble {
	case 0xF:
		{
			firsthalf := (opcode & 0xFF00) >> 0x8
			switch firsthalf {
			case 0x07:
				// Store the current value of the delay timer in register VX
				vm.v[x] = vm.delay
			case 0x0A:
				// Wait for a keypress and store the result in register VX
				pressed := false
				for k, v := range vm.keys {
					if win.Pressed(k) {
						vm.v[x] = v
						pressed = true
						break
					}
				}
				if !pressed {
					vm.pc -= 2
				}
			case 0x15:
				// Set the delay timer to the value of register VX
				vm.delay = vm.v[x]
			case 0x18:
				// Set the sound timer to the value of register VX
				vm.sound = vm.v[x]
			case 0x1E:
				// Add the value stored in register VX to register I
				vx := vm.v[x]
				vm.i += uint16(vx)
			case 0x29:
				// Set I to the memory address of the sprite data corresponding
				// to the hexadecimal digit stored in register VX
				addr := 5 * vm.v[x]
				vm.i = uint16(addr)
			case 0x33:
				// Store the binary-coded decimal equivalent of the value
				// stored in register VX at addresses I, I+1, and I+2
				vx := vm.v[x]
				vm.memory[vm.i] = vx / 100
				vm.memory[vm.i+1] = (vx % 100) / 10
				vm.memory[vm.i+2] = vx % 10
			case 0x55:
				// Store the values of registers V0 to VX
				// inclusive in memory starting at address I
				// I is set to I + X + 1 after operation
				for j := 0; j < 16; j++ {
					vm.memory[int(vm.i)+j] = vm.v[j]
				}
				vm.i = vm.i + uint16(x) + 1
			case 0x65:
				// Fill registers V0 to VX inclusive with the values stored
				// in memory starting at address I
				// I is set to I + X + 1 after operation
				for j := 0; j < 16; j++ {
					vm.v[j] = vm.memory[int(vm.i)+j]
				}
				vm.i = vm.i + uint16(x) + 1
			default:
			}
		}
	case 0x0:
		{
			lastthree := opcode & 0x0FFF
			switch lastthree {
			case 0x0E0:
				// clear the screen
				for y := 0; y < 32; y++ {
					for x := 0; x < 64; x++ {
						vm.display[y][x] = 0
					}
				}
			case 0x0EE:
				// return from a subroutine
				if v, ok := vm.stack.PopBack().(int); ok {
					vm.pc = v
				}
			default:
			}
		}
	case 0x8:
		{
			lastnibble := opcode & 0x000F
			switch lastnibble {
			case 0x0:
				// Store the value of register VY in register VX
				vm.v[x] = vm.v[y]
			case 0x1:
				// Set VX = VX | VY
				vm.v[x] |= vm.v[y]
			case 0x2:
				// Set VX = VX & VY
				vm.v[x] &= vm.v[y]
			case 0x3:
				// Set VX = VX ^ VY
				vm.v[x] ^= vm.v[y]
			case 0x4:
				// add value of VY to VX
				// vf = 1 if carry occurs else 0
				if vm.v[x] > vm.v[x]+vm.v[y] {
					vm.v[0xF] = 1
				} else {
					vm.v[0xF] = 0
				}
				vm.v[x] += vm.v[y]
			case 0x5:
				// subtract value of VY to VX
				// vf = 1 if borrow doesnt occurs else 0
				if vm.v[x] < vm.v[y] {
					vm.v[0xF] = 0
				} else {
					vm.v[0xF] = 1
				}
				vm.v[x] -= vm.v[y]
			case 0x6:
				// store value of VY shifted right one bit
				// in VY
				// VF = lsb prior to shift
				vm.v[0xF] = vm.v[y] & 0x1
				vm.v[x] = vm.v[y] >> 0x1
			case 0x7:
				// set VX = VY - VX
				// VF = 1 if borrow doesn't occur else 0
				if vm.v[x] > vm.v[y] {
					vm.v[0xF] = 0
				} else {
					vm.v[0xF] = 1
				}
				vm.v[x] = vm.v[y] - vm.v[x]
			case 0xE:
				// set VX = VY << 1
				// VF = msb prior to shift
				msb := (vm.v[y] & 0xF0) >> 0x7
				vm.v[0xF] = msb
				vm.v[x] = vm.v[y] << 1
			default:
			}
		}
	case 0xE:
		{
			lasthalf := opcode & 0x00FF
			x := (opcode & 0x0F00) >> 0x8
			switch lasthalf {
			case 0x9E:
				// Skip the next opcode
				// if key = VX is pressed
				for k, v := range vm.keys {
					if v == vm.v[x] {
						if win.Pressed(k) {
							vm.pc += 2
						}
					}
				}
			case 0xA1:
				// Skip the next opcode
				// if key = VX is not pressed
				for k, v := range vm.keys {
					if v == vm.v[x] {
						if !win.Pressed(k) {
							vm.pc += 2
						}
					}
				}
			default:
			}
		}
	case 0x1:
		// jump to addr NNN
		vm.pc = opcode & 0xFFF
	case 0x2:
		// execute subroutine at addr NNN
		vm.stack.PushBack(vm.pc)
		vm.pc = opcode & 0x0FFF
	case 0x3:
		// Skip next opcode if VX == NN
		if int(vm.v[x]) == (opcode & 0xFF) {
			vm.pc += 2
		}
	case 0x4:
		// Skip next opcode if VX != NN
		if int(vm.v[x]) != (opcode & 0xFF) {
			vm.pc += 2
		}
	case 0x5:
		// skip next opcode if VX == VY
		if opcode&0xF == 0 {
			if vm.v[x] == vm.v[y] {
				vm.pc += 2
			}
		} else {

		}
	case 0x6:
		// set VX = NN
		vm.v[x] = uint8(opcode & 0x00FF)
	case 0x7:
		// set VX = VY
		vm.v[x] = vm.v[y]
	case 0x9:
		// skip  next opcode if VX != VY
		if opcode&0xF == 0 {
			if vm.v[x] != vm.v[y] {
				vm.pc += 2
			}
		} else {
		}
	case 0xA:
		// set I = NNN:
		vm.i = uint16(opcode & 0x0FFF)
	default:
	}
}
