package main

import (
	"fmt"

	"github.com/segfult/gosock"
)

func main() {

	app := gosock.NewApp("echo.websocket.org", 80, nil)
	//app := gosock.NewApp("localhost", 6000, nil)
	app.InitHandshake()
	app.WriteMessage([]byte("Hello"))
	data, err := app.ReadMessage()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))

	err = app.Close()
	if err != nil {
		panic(err)
	}

}
