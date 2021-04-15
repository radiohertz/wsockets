package gosock

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
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

// Start the websocket hanshake process.
// FIXME:
// Will return an error if it fails.
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

// Send a message to the websocket server.
// The opcode decides the actual type of the message.
func (a *App) WriteMessage(message []byte, opcode Opcode) (int, error) {
	frame, err := SendMessage(string(message), opcode)
	if err != nil {
		return 0, err
	}
	return a.Conn.Write(frame)
}

// Read a message from the websocket server.
// Will return a buffer of data on success.
func (a *App) ReadMessage() (*Opcode, []byte, error) {
	buf := make([]byte, 1024)

	_, err := a.Conn.Read(buf)
	if err != nil {
		return nil, nil, err
	}

	opCode := uint8(buf[0])
	maskedInt := ^(1 << 7)
	mask := uint8(maskedInt)
	opCode &= mask

	op := Opcode(opCode)

	return &op, buf[2:], nil
}

// Write a struct with json tags to the websocket.
func (a *App) WriteJSON(v interface{}) error {
	jsonData, err := json.Marshal(v)
	if err != nil {
		return err
	}
	data, err := SendMessage(string(jsonData), TextFrame)
	if err != nil {
		return err
	}
	_, err = a.Conn.Write(data)
	return err
}

// Read from a struct and encode it in a json tagged struct.
func (a *App) ReadJSON(v interface{}) error {
	buf := make([]byte, 1024)

	//FIXME: check for errors.
	a.Conn.Read(buf)
	reqData := buf[2:]
	return json.NewDecoder(bytes.NewReader(reqData)).Decode(v)
}

// Pings the websocket server.
func (a *App) Ping() {

}

// Send a close frame to the websocket server.
// This also closes the websocket and underlying TCP connection.
func (a *App) Close() error {

	frame, err := SendMessage("", CloseFrame)
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

// Prepares the message and other required parts and create []byte thar represents a websocket frame.
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

// Masks the payload data using a specific key.
// Key is supposed to be random and will be sent in the websocket frame.
// Only clients are required to frame the data they send.
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
