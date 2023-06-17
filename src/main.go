package main

import (
	"shazammini/src/commands"
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
	com := commands.Commands(&commCahnnels)
	// api := api.Api(&commCahnnels)

	// master.AddRobot(api)
	master.AddRobot(mic)
	master.AddRobot(com)

	master.Start()
}
