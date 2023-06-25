package main

import (

	// "shazammini/src/microphone"

	"log"
	"shazammini/src/api"
	"shazammini/src/commands"
	"shazammini/src/display"
	"shazammini/src/microphone"
	"shazammini/src/structs"
	"time"

	"gobot.io/x/gobot"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	master := gobot.NewMaster()

	commCahnnels := structs.CommChannels{
		PlayChannel:     make(chan bool),
		RecordChannel:   make(chan time.Duration),
		FetchAPI:        make(chan bool),
		DisplayResult:   make(chan structs.Track),
		DisplayRecord:   make(chan bool),
		DisplayThinking: make(chan bool),
	}

	dis := display.Screen(&commCahnnels)
	mic := microphone.Microphone(&commCahnnels)
	com := commands.Commands(&commCahnnels)
	api := api.Api(&commCahnnels)

	master.AddRobot(dis)
	master.AddRobot(api)
	master.AddRobot(com)
	master.AddRobot(mic)

	master.Start()
}
