package microphone

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"shazammini/src/structs"
	"time"

	"github.com/gen2brain/malgo"
	"github.com/youpy/go-wav"
	"gobot.io/x/gobot"
)

func formatToByInt(format malgo.FormatType) uint16 {
	switch format {
	case malgo.FormatS16:
		return 16
	case malgo.FormatS24:
		return 24
	case malgo.FormatS32:
		return 32
	default:
		return 16
	}
}

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

//	func int16SliceToByteSlice(data []int16) []byte {
//		buf := new(bytes.Buffer)
//		for _, value := range data {
//			if err := binary.Write(buf, binary.LittleEndian, value); err != nil {
//				log.Fatal(err)
//			}
//		}
//		return buf.Bytes()
//	}
func int16SliceToSampleSlice(data []int16) []wav.Sample {
	result := make([]wav.Sample, len(data))
	for i, value := range data {
		result[i] = wav.Sample{Values: [2]int{int(value) * 2, 0}}
	}
	return result
}

type microphone struct {
	ctx           *malgo.AllocatedContext
	deviceConfig  malgo.DeviceConfig
	device        *malgo.Device
	devicesList   []malgo.DeviceInfo
	capturedAudio []int16
}

func (m *microphone) Initialise() {
	context, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		log.Println(message)
	})
	if err != nil {
		log.Fatal(err)
	}

	m.ctx = context

	m.InitDevices()
}

func (m *microphone) Kill() {
	m.device.Uninit()

	err := m.ctx.Uninit()
	if err != nil {
		log.Fatal(err)
	}
	m.ctx.Free()

}

func (m *microphone) listDevices() {
	infos, err := m.ctx.Devices(malgo.Capture)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("-----------------------------")
	for _, d := range infos {
		fmt.Printf("ID %s\n", d.ID)
		fmt.Printf("IsDefault %d\n", d.IsDefault)
		fmt.Printf("Pointer %d\n", d.ID.Pointer())
		fmt.Printf("String %s\n", d.ID.String())
		fmt.Println("-----------------------------")
	}
	m.devicesList = infos
}

func (m *microphone) InitDevices() {
	defaultDevice := 0
	m.listDevices()
	for i, dev := range m.devicesList {
		if dev.IsDefault == 1 {
			defaultDevice = i
		}
	}
	fmt.Println(m.devicesList[defaultDevice].ID)

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: func(outputSamples, inputSamples []byte, frameCount uint32) {
			m.capturedAudio = append(m.capturedAudio, byteSliceToInt16Slice(inputSamples)...)
		},
	}
	m.deviceConfig = malgo.DeviceConfig{
		DeviceType: malgo.Capture,
		SampleRate: 44100,
		Periods:    4,
		Capture: malgo.SubConfig{
			DeviceID: m.devicesList[defaultDevice].ID.Pointer(),
			Format:   malgo.FormatS16,
			Channels: 1,
		},
		Alsa: malgo.AlsaDeviceConfig{
			NoMMap: 1,
		},
	}

	device, err := malgo.InitDevice(m.ctx.Context, m.deviceConfig, deviceCallbacks)
	if err != nil {
		log.Fatal(err)
	}

	m.device = device
}

func (m *microphone) StartRecord() {
	m.capturedAudio = []int16{}
	err := m.device.Start()
	if err != nil {
		log.Fatal(err)
	}
}
func (m *microphone) StopRecord() {
	err := m.device.Stop()
	if err != nil {
		log.Fatal(err)
	}
}

func (m *microphone) SaveToWAV() {

	if _, err := os.Stat("output.wav"); err == nil {
		// Delete file
		err := os.Remove("output.wav")
		if err != nil && !os.IsNotExist(err) {
			log.Fatal(err)
		}
	}

	f, err := os.Create("output.wav")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	waveWriter := wav.NewWriter(f,
		uint32(len(m.capturedAudio)),
		uint16(m.deviceConfig.Capture.Channels),
		m.deviceConfig.SampleRate,
		formatToByInt(m.deviceConfig.Capture.Format))
	if err := waveWriter.WriteSamples(int16SliceToSampleSlice(m.capturedAudio)); err != nil {
		log.Fatal(err)
	}
}

func run(commChannels *structs.CommChannels) {

	mic := microphone{}
	mic.Initialise()
	defer mic.Kill()

	for sleep := range commChannels.RecordChannel {
		mic.StartRecord()

		log.Println("recording...")
		time.Sleep(sleep)

		mic.StopRecord()
		log.Println("Saving to file..")
		mic.SaveToWAV()

		commChannels.FetchAPI <- true

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
