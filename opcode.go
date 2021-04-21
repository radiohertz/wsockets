package gosock

// The opcodes used by websocket protocol.
type Opcode uint8

const (

	// denotes a continuation frame
	ContFrame Opcode = iota

	// denotes a text frame
	TextFrame

	// denotes a binary frame
	BinaryFrame

	// 3-7 reserved for further non-control frames.

	// denotes a connection close
	CloseFrame = iota + 5

	// denotes a ping
	PingFrame

	// denotes a pong
	PongFrame
)

func GenerateOpcode(fin bool, opcode Opcode) uint8 {
	finAndOpcode := uint8(opcode)

	if fin {
		finAndOpcode |= (1 << 7)
	}

	return finAndOpcode
}
