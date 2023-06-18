package main

import (

	// "shazammini/src/microphone"
	"shazammini/src/microphone"
	"shazammini/src/structs"

	"gobot.io/x/gobot"
)

func main() {
	master := gobot.NewMaster()

	commCahnnels := structs.CommChannels{
		PlayChannel:   make(chan bool),
		RecordChannel: make(chan bool),
	}

	mic := microphone.Microphone(&commCahnnels)
	// com := commands.Commands(&commCahnnels)
	// dis := display.Screen(&commCahnnels)
	// api := api.Api(&commCahnnels)

	// master.AddRobot(api)
	master.AddRobot(mic)
	// master.AddRobot(dis)
	// master.AddRobot(com)

	master.Start()
}
