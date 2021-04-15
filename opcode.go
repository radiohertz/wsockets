package gosock

import "fmt"

// The opcodes used by websocket protocol.
type Opcode uint8

const (

	// denotes a continuation frame
	Cont Opcode = iota

	// denotes a text frame
	Text

	// denotes a binary frame
	Binary

	// 3-7 reserved for further non-control frames.

	// denotes a connection close
	Close = iota + 5

	// denotes a ping
	Ping

	// denotes a pong
	Pong
)

func GenerateOpcode(fin bool, opcode Opcode) uint8 {
	finAndOpcode := uint8(0b10000000)

	binRep := fmt.Sprintf("%4b", opcode)

	fromBit := 3
	for _, v := range binRep {
		if v == '1' {
			finAndOpcode |= 1 << fromBit
		}
		fromBit--
	}
	return finAndOpcode
}
