package main

import (
	"fmt"
	"log"
	"time"

	"github.com/segfult/gosock"
)

func main() {

	app := gosock.NewApp("gateway.discord.gg", 443, nil)
	app.InitHandshake()

	type GatewayMessages struct {
		Op int                    `json:"op"`
		D  map[string]interface{} `json:"d"`
	}

	init := &GatewayMessages{}

	err := app.ReadJson(init)
	if err != nil {
		panic(err)
	}

	interval, _ := init.D["heartbeat_interval"].(float64)

	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)

	go func() {

		type Beat struct {
			Op int `json:"op"`
			D  int `json:"d"`
		}

		for {
			<-ticker.C
			beat := &Beat{
				Op: 1,
				D:  0,
			}

			err := app.WriteJson(beat)
			if err != nil {
				panic(err)
			}
			log.Println("HEARTBEAT SENT")

			beatAck := &GatewayMessages{}
			err = app.ReadJson(beatAck)
			if err != nil {
				panic(err)
			}
			log.Println(beatAck)
			log.Println("HEARTBEAT ACK RECV")
		}

	}()

	for {
	}

	fmt.Println(interval)

}
