package gosock

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
)

func TestMaskData(t *testing.T) {

	masked := MaskData([]byte("hello"), []byte("12321"))
	original := MaskData(masked, []byte("12321"))
	fmt.Println(string(masked), string(original))

}

func TestFitFrame(t *testing.T) {

	frame := make([]byte, 11)
	firstByte := uint8(0b10000001)
	secondByte := uint8(0b10000101)
	thirdByte := uint32(12345)
	msg := []byte("hello")

	frame[0] = firstByte
	frame[1] = secondByte
	binary.BigEndian.PutUint32(frame[2:], thirdByte)

	buf := bytes.NewBuffer(frame)
	buf.Write(msg)

	fmt.Println(buf.Bytes())
}
