package main

import (

	// "shazammini/src/microphone"

	"log"
	"shazammini/src/display"
	"shazammini/src/structs"
	"time"

	"gobot.io/x/gobot"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	master := gobot.NewMaster()

	commCahnnels := structs.CommChannels{
		PlayChannel:   make(chan bool),
		RecordChannel: make(chan time.Duration),
		FetchAPI:      make(chan bool),
	}

	// mic := microphone.Microphone(&commCahnnels)
	// com := commands.Commands(&commCahnnels)
	// api := api.Api(&commCahnnels)
	dis := display.Screen(&commCahnnels)

	// master.AddRobot(api)
	// master.AddRobot(com)
	// master.AddRobot(mic)
	master.AddRobot(dis)

	master.Start()
}
