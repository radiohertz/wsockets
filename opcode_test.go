package gosock

import (
	"testing"
)

func TestGenerateOpcode(t *testing.T) {

	opcode := GenerateOpcode(true, Text)
	if opcode != 129 {
		t.Error("Expected: 129, got: ", opcode)
	}
}
