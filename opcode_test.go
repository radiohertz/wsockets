package gosock

import (
	"testing"
)

func TestGenerateOpcode(t *testing.T) {

	opcode := GenerateOpcode(true, TextFrame)
	if opcode != 129 {
		t.Error("Expected: 129, got: ", opcode)
	}
}
