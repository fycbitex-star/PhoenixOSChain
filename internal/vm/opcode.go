package vm

type OpCode byte

const (
	OpStop   OpCode = 0x00
	OpPush   OpCode = 0x01
	OpAdd    OpCode = 0x02
	OpSub    OpCode = 0x03
	OpStore  OpCode = 0x04
	OpLoad   OpCode = 0x05
	OpLog    OpCode = 0x06
	OpReturn OpCode = 0x07
	OpRevert OpCode = 0x08
)

var gasTable = map[OpCode]uint64{
	OpStop:   0,
	OpPush:   2,
	OpAdd:    3,
	OpSub:    3,
	OpStore:  20,
	OpLoad:   8,
	OpLog:    10,
	OpReturn: 0,
	OpRevert: 0,
}
