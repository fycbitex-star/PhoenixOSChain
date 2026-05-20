package vm

import "encoding/binary"

func Push(value uint64) []byte {
	out := make([]byte, 9)
	out[0] = byte(OpPush)
	binary.BigEndian.PutUint64(out[1:], value)
	return out
}

func Program(parts ...[]byte) []byte {
	var out []byte
	for _, part := range parts {
		out = append(out, part...)
	}
	return out
}

func Op(op OpCode) []byte {
	return []byte{byte(op)}
}
