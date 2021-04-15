package main

import (
	"fmt"

	"github.com/segfult/gosock"
)

func main() {

	app := gosock.NewApp("gateway.discord.gg", 443, nil)
	app.InitHandshake()

	msg, err := app.ReadMessage()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(msg))

}
