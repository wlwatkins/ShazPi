package microphone

import (
	"fmt"
	"shazammini/src/structs"

	"github.com/gordonklaus/portaudio"
	"gobot.io/x/gobot"
)

const sampleRate = 44100
const seconds = 1

func record() {
	portaudio.Initialize()
	defer portaudio.Terminate()
	buffer := make([]float32, sampleRate*seconds)
	stream, err := portaudio.OpenDefaultStream(1, 0, sampleRate, len(buffer), func(in []float32) {
		for i := range buffer {
			buffer[i] = in[i]
		}
	})
	if err != nil {
		panic(err)
	}
	stream.Start()
	defer stream.Close()
}

func run(commChannels *structs.CommChannels) {

	for {
		select {
		case <-commChannels.RecordChannel:
			record()
		case <-commChannels.PlayChannel:
			fmt.Println("Not implemented")
		}
	}

}

func Microphone(commChannels *structs.CommChannels) *gobot.Robot {
	work := func() {
		run(commChannels)
	}

	robot := gobot.NewRobot("microphone",
		work,
	)

	return robot

}
