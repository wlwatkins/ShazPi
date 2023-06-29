package display

import (
	"fmt"
	"image/color"
	"log"
	"net"
	"shazammini/src/structs"
	"time"

	"github.com/fogleman/gg"
	"github.com/stianeikeland/go-rpio/v4"
	wifiname "github.com/yelinaung/wifi-name"
	"gobot.io/x/gobot"
)

type ReadablePinPatch struct {
	rpio.Pin
}

func (pin ReadablePinPatch) Read() uint8 {
	return uint8(pin.Pin.Read())
}

const RST_PIN = 17
const DC_PIN = 25
const CS_PIN = 8
const BUSY_PIN = 24
const PWR_PIN = 18
const PI = 3.1416

func init() {
	//start the GPIO controller
	if err := rpio.Open(); err != nil {
		log.Fatalf("failed to start gpio: %v", err)
	}

	// Enable SPI on SPI0
	if err := rpio.SpiBegin(rpio.Spi0); err != nil {
		log.Fatalf("failed to enable SPI: %v", err)
	}

	// configure SPI settings
	rpio.SpiSpeed(4_000_000)
	rpio.SpiMode(0, 0)

	rpio.

	rpio.Pin(RST_PIN).Mode(rpio.Output)
	rpio.Pin(DC_PIN).Mode(rpio.Output)
	rpio.Pin(CS_PIN).Mode(rpio.Output)
	rpio.Pin(BUSY_PIN).Mode(rpio.Input)
	rpio.Pin(PWR_PIN).Mode(rpio.Output)
	rpio.Pin(PWR_PIN).High()
	fmt.Println("Init done")
}

type Display struct {
	epd       *EPD
	img       *gg.Context
	width     float64
	height    float64
	assets    Assets
	connected bool
}

func (d *Display) Initialise() {

	d.epd = New(rpio.Pin(RST_PIN), rpio.Pin(DC_PIN), rpio.Pin(CS_PIN), ReadablePinPatch{rpio.Pin(BUSY_PIN)}, rpio.SpiTransmit)
	config := Config{Rotation: ROTATION_0}
	d.epd.Configure(config)
	d.width = float64(d.epd.height)
	d.height = float64(d.epd.width)

	d.img = gg.NewContext(int(d.height), int(d.width))
	d.img.Translate(float64(int(-(d.height/2)*1.1)), float64(int((d.height/2)*1.1)))
	d.img.RotateAbout(PI/2+PI, float64(int(d.width/2)), float64(int(d.height/2)))
	d.img.SetColor(color.White)
	d.img.Clear()
}

func (d *Display) Print(str string, font float64, c Coordonates) float64 {
	if err := d.img.LoadFontFace("/home/pi/dev/static/Inter-Black.ttf", font); err != nil {
		panic(err)
	}

	d.img.SetColor(color.Black)
	lines := d.img.WordWrap(str, d.width)
	fullHeight := len(lines) * int(font)
	for i, l := range lines {
		d.img.DrawStringAnchored(l, c.X, c.Y+(float64(i)*font), c.OX, c.OY)
		d.img.Stroke()
	}
	return float64(fullHeight)
}

func (d *Display) Version() {
	d.Print("v0.1", 15, Coordonates{X: 10, Y: 10, OX: 0, OY: 0.5})
}

func (d *Display) Welcome() {
	d.Clear()
	d.Print("ShazPi", 30, Coordonates{X: d.width / 2, Y: d.height / 2, OX: 0.5, OY: 0.5})
	d.Print("Loading", 20, Coordonates{X: d.width / 2, Y: (d.height / 2) + 20, OX: 0.5, OY: 1})
	d.Print("connecting...", 15, Coordonates{X: d.width, Y: 10, OX: 1, OY: 0.5})
	d.Version()

	d.epd.Draw(d.img)

}

func (d *Display) loadAssets() {
	d.assets.LoadAssets(d)
}

func (d *Display) DrawPNG(e *EPDPNG) {
	d.img.SetColor(color.Black)
	d.img.DrawImageAnchored(e.png, int(e.coord.X), int(e.coord.Y), 0.5, 0.5)
	d.img.Fill()
}

func (d *Display) CheckConnection() {

	byNameInterface, _ := net.InterfaceByName("eth0")
	// if strings.Contains(byNameInterface.Flags.String(), "up") {
	if byNameInterface != nil {
		d.Print("Ethernet", 15, Coordonates{X: d.width, Y: 10, OX: 1, OY: 0.5})
		d.DrawPNG(&d.assets.WifiOn)
		d.connected = true
	} else if Connected() {
		d.Print(wifiname.WifiName(), 15, Coordonates{X: d.width - 25, Y: 10, OX: 1, OY: 0.5})
		d.DrawPNG(&d.assets.WifiOn)
		d.connected = true
	} else {
		d.Print("No internet", 15, Coordonates{X: d.width - 25, Y: 10, OX: 1, OY: 0.5})
		d.DrawPNG(&d.assets.WifiOff)
		d.connected = false
	}
}

func (d *Display) DrawWithDecoration() {
	d.CheckConnection()
	d.Version()
	d.epd.Draw(d.img)
}

func (d *Display) Clear() {
	d.img.SetColor(color.White)
	d.img.Clear()
}

func (d *Display) Idle() {
	d.Clear()
	d.Print("Ready to play music", 30, Coordonates{X: d.width / 2, Y: d.height / 2, OX: 0.5, OY: 0.5})
	d.DrawWithDecoration()
}

func (d *Display) Recording() {
	d.Clear()
	d.Print("Recording", 30, Coordonates{X: d.width / 2, Y: d.height / 2, OX: 0.5, OY: 0.5})
	d.DrawWithDecoration()
}

func (d *Display) Thinking() {
	d.Clear()
	d.Print("Thinking", 30, Coordonates{X: d.width / 2, Y: d.height / 2, OX: 0.5, OY: 0.5})
	d.DrawWithDecoration()
}

func (d *Display) Result(trackName, artistName string) {
	d.Clear()
	offset := d.Print(trackName, 30, Coordonates{X: d.width / 2, Y: 40, OX: 0.5, OY: 0.5})
	d.Print(artistName, 25, Coordonates{X: d.width / 2, Y: (40) + offset, OX: 0.5, OY: 1})
	d.DrawWithDecoration()
}

func (d *Display) TryConnect() {
	d.Clear()
	d.Print("Looking for WiFi", 30, Coordonates{X: d.width / 2, Y: d.height / 2, OX: 0.5, OY: 0.5})
	d.DrawWithDecoration()

}

func run(commChannels *structs.CommChannels) {

	defer rpio.Close()

	display := Display{}

	display.Initialise()
	display.Welcome()

	display.loadAssets()

	for {
		if display.connected {
			display.Idle()
			select {
			case <-commChannels.DisplayRecord:
				display.Recording()
			case <-commChannels.DisplayThinking:
				display.Thinking()
			case track := <-commChannels.DisplayResult:
				log.Println(track.Artists)
				// artist := "Unknown"
				// if len(track.Artists) > 1 {
				// 	artist = track.Artists[0].Name
				// }
				display.Result(track.Title, track.Subtitle)

				// case <-time.After(5 * time.Second):
				// 	display.CheckConnection()
			}
		} else {
			display.TryConnect()
			time.Sleep(5 * time.Second)
		}
	}
}

func Screen(commChannels *structs.CommChannels) *gobot.Robot {
	work := func() {
		run(commChannels)
	}

	robot := gobot.NewRobot("display",
		work,
	)

	return robot

}
