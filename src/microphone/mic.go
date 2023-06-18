package microphone

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"shazammini/src/structs"

	"github.com/gen2brain/malgo"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"gobot.io/x/gobot"
)

func byteSliceToInt16Slice(data []byte) []int16 {
	var result []int16
	reader := bytes.NewReader(data)
	for reader.Len() > 0 {
		var value int16
		if err := binary.Read(reader, binary.LittleEndian, &value); err != nil {
			log.Fatal(err)
		}
		result = append(result, value)
	}
	return result
}

func int16SliceToIntSlice(data []int16) []int {
	result := make([]int, len(data))
	for i, value := range data {
		result[i] = int(value)
	}
	return result
}

func run(commChannels *structs.CommChannels) {
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 1
	deviceConfig.SampleRate = 44100
	deviceConfig.Alsa.NoMMap = 1

	context, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		log.Println(message)
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = context.Uninit()
		context.Free()
	}()

	var capturedAudio []int16

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: func(outputSamples, inputSamples []byte, frameCount uint32) {
			capturedAudio = append(capturedAudio, byteSliceToInt16Slice(inputSamples)...)
		},
	}

	device, err := malgo.InitDevice(context.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		log.Fatal(err)
	}
	defer device.Uninit()

	err = device.Start()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Press Enter to stop recording...")
	fmt.Scanln()

	err = device.Stop()
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create("output.wav")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	e := wav.NewEncoder(f, int(deviceConfig.SampleRate), int(deviceConfig.Capture.Channels), int(deviceConfig.Capture.Format), 1)
	buf := &audio.IntBuffer{Data: int16SliceToIntSlice(capturedAudio), Format: &audio.Format{
		NumChannels: int(deviceConfig.Capture.Channels),
		SampleRate:  int(deviceConfig.SampleRate),
	}}
	if err := e.Write(buf); err != nil {
		log.Fatal(err)
	}
	if err := e.Close(); err != nil {
		log.Fatal(err)
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
