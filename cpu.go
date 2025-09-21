package main

import (
	"fmt"
	"math/rand"
	"slices"
)

type Operation struct {
	opcodeHex uint16
	opcode    OPCODE
	x         byte
	y         byte
	n         byte
	nn        uint8
	nnn       uint16
}

type OPCODE int

const (
	OP_NONE            OPCODE = iota //0000
	OP_CLEAR                         //00E0
	OP_JMP                           //1NNN
	OP_SUBROUTINE                    //2NNN
	OP_RET                           //00EE
	OP_EQUAL                         //3XNN
	OP_NEQUAL                        //4XNN
	OP_REG_EQUAL                     //5XY0
	OP_REG_NEQUAL                    //9XY0
	OP_REG_SET                       //6XNN
	OP_REG_ADD                       //7XNN
	OP_REG_SET_REG                   //8XY0
	OP_OR                            //8XY1
	OP_AND                           //8XY2
	OP_XOR                           //8XY3
	OP_ADD_EQUAL                     //8XY4
	OP_SUB                           //8XY5
	OP_SUB_INV                       //8XY7
	OP_RSHIFT                        //8XY6
	OP_LSHIFT                        //8XYE
	OP_SET_IDX                       //ANNN
	OP_JMP_OFF                       //BNNN
	OP_RANDOM                        //CXNN
	OP_DISPLAY                       //DXYN
	OP_KEY_PRESSED                   //EX9E
	OP_KEY_NOT_PRESSED               //EXA1
	OP_GET_DTIMER                    //FX07
	OP_SET_DTIMER                    //FX15
	OP_SET_STIMER                    //FX18
	OP_ADD_IDX                       //FX1E
	OP_GET_KEY                       //FX0A
	OP_FONT                          //FX29
	OP_BCD                           //FX33
	OP_STORE_MEM                     //FX55
	OP_LOAD_MEM                      //FX65
)

type Cpu struct {
	PC         uint16
	I          uint16
	stack      []uint16
	delayTimer *uint8
	soundTimer *uint8
	regs       [0x10]uint8

	display *Display

	memory *Memory
}

func (c *Cpu) String() string {
	return fmt.Sprintf("CPU STATE:\nPC: 0x%X\nI: 0x%X\nStack: %v\nDelay Timer: %d\nSound Timer: %d\nRegisters: %v\n",
		c.PC, c.I, c.stack, *c.delayTimer, *c.soundTimer, c.regs)
}

func (op *Operation) String() string {
	return fmt.Sprintf("0x%X, OPERATION %s, x: 0x%X, y: 0x%X, n: 0x%X, nn: 0x%X, nnn: 0x%X\n",
		op.opcodeHex, op.opcode, op.x, op.y, op.n, op.nn, op.nnn)
}

func InitCpu() Cpu {
	memory := InitMemory()
	memory.CreateFont()
	display := NewDisplay()
	return Cpu{0x200, 0, []uint16{}, &display.delayTimer, &display.soundTimer, [0x10]uint8{0}, display, &memory}
}

func (c *Cpu) ReadMemoryByte(addr uint16) (byte, error) {
	return c.memory.ReadMemoryByte(addr)
}

func (c *Cpu) WriteMemoryByte(addr uint16, b byte) (byte, error) {
	return c.memory.WriteMemoryByte(addr, b)
}

func (c *Cpu) IncrementPCByTwo() {
	c.PC += 2
}

func (c *Cpu) LoadROM(rom []byte) {
	//ROM usually start at address 0x200
	offset := 0x200
	for i := range len(rom) {
		c.WriteMemoryByte(uint16(offset+i), rom[i])
	}
}

func (c *Cpu) Fetch() uint16 {
	opCode_high, _ := c.ReadMemoryByte(c.PC)
	opCode_low, _ := c.ReadMemoryByte(c.PC + 1)
	c.IncrementPCByTwo()
	return uint16(opCode_high)<<8 | uint16(opCode_low)

}

func (c *Cpu) Decode(opcode uint16) Operation {
	var op OPCODE

	x := byte((opcode & 0x0F00) >> 8)
	y := byte((opcode & 0x00F0) >> 4)
	n := byte((opcode & 0x000F))
	nn := uint8((opcode & 0x00FF))
	nnn := uint16((opcode & 0x0FFF))

	switch opcode {
	case 0x00E0:
		op = OP_CLEAR
	case 0x00EE:
		op = OP_RET
	}
	switch (opcode & 0xF000) >> 12 {
	case 0x1:
		op = OP_JMP
	case 0x2:
		op = OP_SUBROUTINE
	case 0x3:
		op = OP_EQUAL
	case 0x4:
		op = OP_NEQUAL
	case 0x5:
		op = OP_REG_EQUAL
	case 0x6:
		op = OP_REG_SET
	case 0x7:
		op = OP_REG_ADD
	case 0x8:
		switch opcode & 0x000F {
		case 0x0:
			op = OP_REG_SET_REG
		case 0x1:
			op = OP_OR
		case 0x2:
			op = OP_AND
		case 0x3:
			op = OP_XOR
		case 0x4:
			op = OP_ADD_EQUAL
		case 0x5:
			op = OP_SUB
		case 0x6:
			op = OP_RSHIFT
		case 0x7:
			op = OP_SUB_INV
		case 0xE:
			op = OP_LSHIFT
		}
	case 0x9:
		op = OP_REG_NEQUAL
	case 0xA:
		op = OP_SET_IDX
	case 0xB:
		op = OP_JMP_OFF
	case 0xC:
		op = OP_RANDOM
	case 0xD:
		op = OP_DISPLAY
	case 0xE:
		switch opcode & 0x00FF {
		case 0x9E:
			op = OP_KEY_PRESSED
		case 0xA1:
			op = OP_KEY_NOT_PRESSED
		}
	case 0xF:
		switch opcode & 0x00FF {
		case 0x07:
			op = OP_GET_DTIMER
		case 0x15:
			op = OP_SET_DTIMER
		case 0x18:
			op = OP_SET_STIMER
		case 0x1E:
			op = OP_ADD_IDX
		case 0x0A:
			op = OP_GET_KEY
		case 0x29:
			op = OP_FONT
		case 0x33:
			op = OP_BCD
		case 0x55:
			op = OP_STORE_MEM
		case 0x65:
			op = OP_LOAD_MEM
		}
	}

	return Operation{
		opcode,
		op,
		x,
		y,
		n,
		nn,
		nnn,
	}
}

func (c *Cpu) Execute(operation Operation) {
	//fmt.Printf("0x%X: %s x:0x%X y:0x%X n:0x%X nn:0x%X nnn:0x%X\n",
	//	c.PC, operation.opcode, operation.x, operation.y, operation.n, operation.nn, operation.nnn)
	switch operation.opcode {
	case OP_NONE:
	case OP_CLEAR:
		c.display.ClearScreen()
	case OP_JMP:
		c.PC = operation.nnn
	case OP_SUBROUTINE:
		c.stack = append(c.stack, c.PC)
		c.PC = operation.nnn
	case OP_RET:
		c.PC, c.stack = c.stack[len(c.stack)-1], c.stack[:len(c.stack)-1]
	case OP_EQUAL:
		if c.regs[operation.x] == operation.nn {
			c.IncrementPCByTwo()
		}
	case OP_NEQUAL:
		if c.regs[operation.x] != operation.nn {
			c.IncrementPCByTwo()
		}
	case OP_REG_EQUAL:
		if c.regs[operation.x] == c.regs[operation.y] {
			c.IncrementPCByTwo()
		}
	case OP_REG_NEQUAL:
		if c.regs[operation.x] != c.regs[operation.y] {
			c.IncrementPCByTwo()
		}
	case OP_REG_SET:
		c.regs[operation.x] = operation.nn
	case OP_REG_ADD:
		c.regs[operation.x] += operation.nn
	case OP_REG_SET_REG:
		c.regs[operation.x] = c.regs[operation.y]
	case OP_OR:
		c.regs[operation.x] = c.regs[operation.x] | c.regs[operation.y]
		c.regs[0xF] = 0
	case OP_AND:
		c.regs[operation.x] = c.regs[operation.x] & c.regs[operation.y]
		c.regs[0xF] = 0
	case OP_XOR:
		c.regs[operation.x] = c.regs[operation.x] ^ c.regs[operation.y]
		c.regs[0xF] = 0
	case OP_ADD_EQUAL:
		var flag byte
		if 255 < uint16(c.regs[operation.x])+uint16(c.regs[operation.y]) {
			flag = 1
		} else {
			flag = 0
		}
		c.regs[operation.x] += c.regs[operation.y]
		c.regs[0xF] = flag
	case OP_SUB:
		var flag byte
		if c.regs[operation.x] >= c.regs[operation.y] {
			flag = 1
		} else {
			flag = 0
		}
		c.regs[operation.x] = c.regs[operation.x] - c.regs[operation.y]
		c.regs[0xF] = flag
	case OP_SUB_INV:
		var flag byte
		if c.regs[operation.x] <= c.regs[operation.y] {
			flag = 1
		} else {
			flag = 0
		}
		c.regs[operation.x] = c.regs[operation.y] - c.regs[operation.x]
		c.regs[0xF] = flag
	case OP_RSHIFT:
		flag := c.regs[operation.y] & 0x1
		c.regs[operation.x] = c.regs[operation.y]
		c.regs[operation.x] >>= 1
		c.regs[0xF] = flag
	case OP_LSHIFT:
		flag := (c.regs[operation.y] & (0x1 << 7)) >> 7
		c.regs[operation.x] = c.regs[operation.y]
		c.regs[operation.x] <<= 1
		c.regs[0xF] = flag
	case OP_SET_IDX:
		c.I = operation.nnn
	case OP_JMP_OFF:
		c.PC = operation.nnn + uint16(c.regs[0])
	case OP_RANDOM:
		random := rand.Intn(255)
		c.regs[operation.x] = uint8(random) & operation.nn
	case OP_DISPLAY:
		c.regs[0xF] = 0
		xCoord := c.regs[operation.x]
		yCoord := c.regs[operation.y]
		for n := range operation.n {
			if yCoord+n > 31 && yCoord < 32 {
				fmt.Printf("DrawSprite Y clipped\n")
				break
			}
			if spriteData, err := c.ReadMemoryByte(c.I + uint16(n)); err == nil {
				c.display.DrawSprite(c, spriteData, xCoord, yCoord+n)
			}
		}
	case OP_KEY_PRESSED:
		if len(c.display.keys) == 0 || c.regs[operation.x] > 0xF {
			return
		}
		if c.display.keys[c.regs[operation.x]] {
			c.IncrementPCByTwo()
		}
	case OP_KEY_NOT_PRESSED:
		if c.regs[operation.x] > 0xF {
			return
		}
		if !c.display.keys[c.regs[operation.x]] {
			c.IncrementPCByTwo()
		}
	case OP_ADD_IDX:
		c.I += uint16(c.regs[operation.x])
	case OP_GET_DTIMER:
		c.regs[operation.x] = c.display.delayTimer
	case OP_SET_DTIMER:
		c.display.delayTimer = c.regs[operation.x]
	case OP_SET_STIMER:
		c.display.soundTimer = c.regs[operation.x]
		//fmt.Printf("setting soundTimer: %d\n", c.regs[operation.x])
	case OP_GET_KEY:
		for {
			gotKey := false
			for i := range 0xF {
				if c.display.keys[i] {
					gotKey = true
					break
				}
			}
			if gotKey {
				break
			}
		}
	case OP_FONT:
		c.I = uint16(c.regs[operation.x]) * 5
	case OP_BCD:
		num := c.regs[operation.x]
		digits := make([]byte, 3)
		for i := range 3 {
			digits[i] = num % 10
			num /= 10
		}
		slices.Reverse(digits)
		for i := range 3 {
			c.WriteMemoryByte(c.I+uint16(i), digits[i])
		}
	case OP_STORE_MEM:
		for i := range operation.x + 1 {
			c.WriteMemoryByte(c.I, c.regs[i])
			c.I++
		}
	case OP_LOAD_MEM:
		for i := range operation.x + 1 {
			c.regs[i], _ = c.ReadMemoryByte(c.I)
			c.I++
		}
	}
}

func (c *Cpu) CpuCycle() {
	opcode := c.Fetch()
	op := c.Decode(opcode)
	c.Execute(op)
	//fmt.Printf("%s", c.String())
	//fmt.Printf("%s", op.String())
}
