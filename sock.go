package gosock

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
)

// Represents a websocket client.
type App struct {
	// The port the websocket server is running on.
	Port uint16
	// The url of the resouce
	Uri string
	// The context
	ctx *context.Context
	// The TCP connection.
	Conn net.Conn
}

// Create a new websocket client.
func NewApp(uri string, port uint16, ctx *context.Context) *App {
	return &App{
		Port: port,
		Uri:  uri,
		ctx:  ctx,
	}
}

func (a *App) InitHandshake() {
	// Ref:
	/*
	   GET /chat HTTP/1.1
	   Host: server.example.com
	   Upgrade: websocket
	   Connection: Upgrade
	   Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
	   Origin: http://example.com
	   Sec-WebSocket-Protocol: chat, superchat
	   Sec-WebSocket-Version: 13
	*/

	handShakeReq := strings.Builder{}
	handShakeReq.WriteString("GET / HTTP/1.1\r\n")
	handShakeReq.WriteString("Host: " + a.Uri + "\r\n")
	handShakeReq.WriteString("Upgrade: websocket\r\n")
	handShakeReq.WriteString("Connection: Upgrade\r\n")
	handShakeReq.WriteString("Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n")
	handShakeReq.WriteString("Sec-WebSocket-Version: 13\r\n")
	handShakeReq.WriteString("\r\n")

	req := handShakeReq.String()

	uri := fmt.Sprintf("%s:%d", a.Uri, a.Port)

	conf := &tls.Config{}

	conn, err := tls.Dial("tcp", uri, conf)
	if err != nil {
		panic(err)
	}

	n, err := conn.Write([]byte(req))
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("[LOG]: Written %d bytes\n", n)

	buf := make([]byte, 1024)

	n, err = conn.Read(buf)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("[LOG]: Read %d bytes\n", n)
	a.Conn = conn
	fmt.Println(string(buf))

	// check if handshake is successful.

}

func (a *App) WriteMessage(message []byte, opcode Opcode) (int, error) {
	frame, err := SendMessage(string(message), opcode)
	if err != nil {
		return 0, err
	}
	return a.Conn.Write(frame)
}

func (a *App) ReadMessage() ([]byte, error) {
	buf := make([]byte, 1024)

	_, err := a.Conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[2:], nil
}

func (a *App) Close() error {

	frame, err := SendMessage("", Close)
	if err != nil {
		panic(err)
	}
	_, err = a.Conn.Write(frame)

	if err != nil {
		return err
	}

	buf := make([]byte, 1024)
	a.Conn.Read(buf)
	if buf[0] != 0x88 {
		return fmt.Errorf("failed to close: expected %d, got %d", 0b10001000, buf[0])
	}
	a.Conn.Close()
	return nil
}

func SendMessage(message string, messageType Opcode) ([]byte, error) {
	msgLength := len(message)
	if msgLength > 125 {
		return nil, errors.New("payload length above 125 bytes not yet supported")
	}

	frame := make([]byte, 6+msgLength)
	firstByte := GenerateOpcode(true, messageType)

	// secondByte := uint8(0b10000101)

	lengthAndMask := uint8(0)
	lengthAndMask |= 1 << 7

	fromBit := 6
	lengthInBits := fmt.Sprintf("%7b", msgLength)

	for _, v := range lengthInBits {
		if v == '1' {
			lengthAndMask |= 1 << fromBit
		}
		fromBit--
	}

	msg := []byte(message)

	maskKey := []byte("12321")

	maskKeyh := binary.LittleEndian.Uint32(maskKey)
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, maskKeyh)

	msg = MaskData(msg, maskKey)

	frame[0] = firstByte
	frame[1] = lengthAndMask
	binary.LittleEndian.PutUint32(frame[2:], maskKeyh)

	from := 6
	for _, v := range msg {
		frame[from] = v
		from++
	}

	return frame, nil
}

func MaskData(data []byte, key []byte) []byte {
	/*
	   	MASKING ALGO:

	      Octet i of the transformed data ("transformed-octet-i") is the XOR of
	      octet i of the original data ("original-octet-i") with octet at index
	      i modulo 4 of the masking key ("masking-key-octet-j"):

	        j                   = i MOD 4
	        transformed-octet-i = original-octet-i XOR masking-key-octet-j
	*/
	transformed := make([]byte, len(data))
	for i, _ := range data {
		transformed[i] = data[i] ^ key[i%4]
	}
	return transformed
}
