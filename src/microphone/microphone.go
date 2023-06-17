package microphone

// import (
// 	"fmt"
// 	"io"
// 	"os"
// 	"shazammini/src/structs"
// 	"time"

// 	"gobot.io/x/gobot"

// 	"github.com/gen2brain/malgo"
// )

// func play(ctx *malgo.AllocatedContext, deviceConfig malgo.DeviceConfig, playbackCallbacks malgo.DeviceCallbacks) {
// 	device, err := malgo.InitDevice(ctx.Context, deviceConfig, playbackCallbacks)
// 	if err != nil {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}

// 	err = device.Start()
// 	if err != nil {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}

// 	timer1 := time.NewTimer(10 * time.Second)
// 	<-timer1.C
// 	fmt.Println("Stop playing")

// 	device.Uninit()
// }

// func record(ctx *malgo.AllocatedContext, deviceConfig malgo.DeviceConfig, captureCallbacks malgo.DeviceCallbacks) {
// 	fmt.Println("Starting recording")

// 	filePath := "output.wav"

// 	// Open the file for writing
// 	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
// 	if err != nil {
// 		fmt.Println("Error opening file:", err)
// 		return
// 	}
// 	defer file.Close()
// 	w := io.MultiWriter(os.Stdout, file)
// 	abortChan := make(chan error)
// 	defer close(abortChan)
// 	aborted := false

// 	deviceCallbacks := malgo.DeviceCallbacks{
// 		Data: func(outputSamples, inputSamples []byte, frameCount uint32) {
// 			if aborted {
// 				return
// 			}

// 			_, err := w.Write(inputSamples)
// 			if err != nil {
// 				aborted = true
// 				abortChan <- err
// 			}

// 		},
// 	}

// 	device, err := malgo.InitDevice(ctx.Context, deviceConfig, deviceCallbacks) // captureCallbacks)
// 	if err != nil {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}

// 	err = device.Start()
// 	if err != nil {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}

// 	timer1 := time.NewTimer(10 * time.Second)
// 	<-timer1.C

// 	fmt.Println("Stop recording")

// 	device.Uninit()

// }

// func run(commChannels *structs.CommChannels) {
// 	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
// 		fmt.Printf("LOG : %v", message)
// 	})

// 	if err != nil {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}
// 	defer func() {
// 		_ = ctx.Uninit()
// 		ctx.Free()
// 	}()

// 	deviceConfig := malgo.DefaultDeviceConfig(malgo.Duplex)
// 	deviceConfig.Capture.Format = malgo.FormatS16
// 	deviceConfig.Capture.Channels = 1
// 	deviceConfig.Playback.Format = malgo.FormatS16
// 	deviceConfig.Playback.Channels = 1
// 	deviceConfig.SampleRate = 44100
// 	deviceConfig.Alsa.NoMMap = 1
// 	var playbackSampleCount uint32
// 	var capturedSampleCount uint32
// 	pCapturedSamples := make([]byte, 0)

// 	sizeInBytes := uint32(malgo.SampleSizeInBytes(deviceConfig.Capture.Format))
// 	onRecvFrames := func(pSample2, pSample []byte, framecount uint32) {

// 		sampleCount := framecount * deviceConfig.Capture.Channels * sizeInBytes

// 		newCapturedSampleCount := capturedSampleCount + sampleCount

// 		pCapturedSamples = append(pCapturedSamples, pSample...)

// 		capturedSampleCount = newCapturedSampleCount

// 	}

// 	captureCallbacks := malgo.DeviceCallbacks{
// 		Data: onRecvFrames,
// 	}

// 	onSendFrames := func(pSample, nil []byte, framecount uint32) {
// 		samplesToRead := framecount * deviceConfig.Playback.Channels * sizeInBytes
// 		if samplesToRead > capturedSampleCount-playbackSampleCount {
// 			samplesToRead = capturedSampleCount - playbackSampleCount
// 		}

// 		copy(pSample, pCapturedSamples[playbackSampleCount:playbackSampleCount+samplesToRead])

// 		playbackSampleCount += samplesToRead

// 		if playbackSampleCount == uint32(len(pCapturedSamples)) {
// 			playbackSampleCount = 0
// 		}
// 	}

// 	playbackCallbacks := malgo.DeviceCallbacks{
// 		Data: onSendFrames,
// 	}

// 	for {
// 		select {
// 		case <-commChannels.RecordChannel:
// 			record(ctx, deviceConfig, captureCallbacks)
// 		case <-commChannels.PlayChannel:
// 			play(ctx, deviceConfig, playbackCallbacks)
// 		}
// 	}

// }

// func Microphone(commChannels *structs.CommChannels) *gobot.Robot {
// 	work := func() {
// 		run(commChannels)
// 	}

// 	robot := gobot.NewRobot("microphone",
// 		work,
// 	)

// 	return robot

// }
