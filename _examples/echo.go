package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/segfult/gosock"
)

func main() {

	ep := "echo.websocket.org"
	app := gosock.NewApp(ep, 80, nil)
	//app := gosock.NewApp("localhost", 6000, nil)
	app.InitHandshake()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("[DATA to send]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSuffix(input, "\n")
		if input == "exit0" {
			break
		}
		app.WriteMessage([]byte(input), gosock.Text)
		log.Println("[SENT]: ", input)
		data, err := app.ReadMessage()
		if err != nil {
			panic(err)
		}
		log.Printf("[RECV from %s]: %s\n", ep, string(data))
	}

	err := app.Close()
	if err != nil {
		panic(err)
	}
}
