package display

import (
	"fmt"
	"image/color"
	"log"
	"net"
	"os"
	"shazammini/src/structs"
	"strings"

	"github.com/fogleman/gg"
	"github.com/stianeikeland/go-rpio/v4"
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

	rpio.Pin(RST_PIN).Mode(rpio.Output)
	rpio.Pin(DC_PIN).Mode(rpio.Output)
	rpio.Pin(CS_PIN).Mode(rpio.Output)
	rpio.Pin(BUSY_PIN).Mode(rpio.Input)
	rpio.Pin(PWR_PIN).Mode(rpio.Output)
	rpio.Pin(PWR_PIN).High()
	fmt.Println("Init done")
}

type Display struct {
	epd    *EPD
	img    *gg.Context
	assets Assets
}

func (d *Display) Initialise() {

	d.epd = New(rpio.Pin(RST_PIN), rpio.Pin(DC_PIN), rpio.Pin(CS_PIN), ReadablePinPatch{rpio.Pin(BUSY_PIN)}, rpio.SpiTransmit)
	config := Config{Rotation: ROTATION_90}
	d.epd.Configure(config)

	d.img = gg.NewContext(int(d.epd.height), int(d.epd.width))
	d.img.Translate(-float64(d.epd.width/2), 0)
	d.img.RotateAbout(PI/2, float64(d.epd.height/2), float64(d.epd.width)/2)
	d.img.SetColor(color.White)
	d.img.Clear()
}

func (d *Display) Welcome() {
	if err := d.img.LoadFontFace("/home/pi/dev/static/8-BIT_WONDER.TTF", 18); err != nil {
		panic(err)
	}

	d.img.SetColor(color.Black)
	d.img.DrawRectangle(0, 0, float64(d.epd.height), float64(d.epd.width))
	d.img.Fill()

	d.img.SetColor(color.White)
	d.img.DrawStringAnchored("ShazPi", float64(d.epd.height)/2, float64(d.epd.width)/2, 0.5, 0.5)
	d.img.Stroke()
	d.epd.Draw(d.img)
	d.img.SetColor(color.White)
	d.img.Clear()
}

func (d *Display) loadAssets() {
	d.assets.LoadAssets()
}

func (d *Display) DrawPNG(e *EPDPNG) {
	d.img.SetColor(color.Black)
	d.img.DrawImageAnchored(e.png, e.coord.X, e.coord.Y, 0.5, 0.5)
	d.img.Fill()
}

// func (d *Display) draw() {
// 	// d.epd.Sleep()
// 	// d.epd.Clear(color.White)
// 	if e := d.epd.Draw(d.img.Image()); e != nil {
// 		fmt.Printf("failed to draw: %v\n", e)
// 		d.epd.Clear(color.White)
// 	}

// 	// d.epd.Sleep()
// }

func (d *Display) CheckConnection() {
	if err := d.img.LoadFontFace("/home/pi/dev/static/8-BIT_WONDER.TTF", 10); err != nil {
		panic(err)
	}
	byNameInterface, _ := net.InterfaceByName("eth0")
	fmt.Println(byNameInterface)
	d.img.SetColor(color.Black)
	if strings.Contains(byNameInterface.Flags.String(), "up") {
		d.img.DrawStringAnchored("Ethernet", 230, 10, 1, 0.5)
		d.DrawPNG(&d.assets.WifiOn)

	} else if Connected() {
		d.img.DrawStringAnchored("Nokia 8110 4G", 230, 10, 1, 0.5)
		d.DrawPNG(&d.assets.WifiOn)
	} else {
		d.img.DrawStringAnchored("No internet", 230, 10, 1, 0.5)
		d.DrawPNG(&d.assets.WifiOff)
	}
	d.img.Fill()
}

func (d *Display) drawRectangle(s string, coord Coordonates) {

	if err := d.img.LoadFontFace("/home/pi/dev/static/8-BIT_WONDER.TTF", 10); err != nil {
		panic(err)
	}

	d.img.SetColor(color.Black)
	fmt.Printf("float64(d.epd.width)/2 %d\n", float64(d.epd.width)/2)
	d.img.DrawStringAnchored(s, 0, float64(d.epd.height)/2, 0.5, 0.5)
	// d.img.Fill()
	d.img.Stroke()
	d.epd.Draw(d.img)
}

func run(commChannels *structs.CommChannels) {

	defer rpio.Close()

	display := Display{}

	display.Initialise()
	display.Welcome()
	// display.drawRectangle("B", Coordonates{X: 0, Y: 0})

	// display.loadAssets()
	// display.CheckConnection()
	// time.Sleep(2 * time.Second)
	// display.epd.Draw(display.img.Image())

	os.Exit(0)

	// for {
	// 	select {
	// 	case <-commChannels.RecordChannel:
	// 		fmt.Println("Hello")
	// 	case <-commChannels.PlayChannel:
	// 		fmt.Println("Bye")
	// 	}
	// }

	// // initialize the driver

	// img.SetColor(color.White)
	// img.Clear()
	// if err := img.LoadFontFace("/home/pi/dev/8-BIT_WONDER.TTF", 20); err != nil {
	// 	panic(err)
	// }

	// var cx, cy = float64(display.Height) / 2, float64(display.Width) / 2
	// var s1 = "go get"
	// var hs1, _ = img.MeasureString(s1)
	// var s2 = "I love my pioupiou"
	// var hs2, ws2 = img.MeasureString(s2)
	// fmt.Printf("width: %d, height %d, cx-(hs1/2) %f", display.Width, display.Height, (cx - (hs1 / 2)))

	// img.SetColor(color.White)
	// img.DrawRectangle(0, 0, float64(display.Height), float64(display.Width))
	// img.Fill()

	// _, _, _ = cy, hs2, ws2
	// // img.SetColor(color.Black)
	// // img.DrawString(s1, cx-(hs1/2), cy-ws2-8)
	// // img.Stroke()
	// img.SetColor(color.Black)
	// img.DrawString(s2, cx-(hs2/2)-20, cy)
	// img.Stroke()
	// // display.Clear(color.Black)

	// wifi_connected, err := gg.LoadPNG("wifi_connected.png")
	// if err != nil {
	// 	panic(err)
	// }

	// wifi_connected = resize.Resize(uint(float64(wifi_connected.Bounds().Dx())*0.04), 0, wifi_connected, resize.Lanczos2)

	// wifi_unconnected, err := gg.LoadPNG("wifi_unconnected.png")
	// if err != nil {
	// 	panic(err)
	// }

	// wifi_unconnected = resize.Resize(uint(float64(wifi_unconnected.Bounds().Dx())*0.04), 0, wifi_unconnected, resize.Lanczos2)

	// // w := wifi_connected.Bounds().Size().X
	// // h := wifi_connected.Bounds().Size().Y

	// if err := img.LoadFontFace("/home/pi/dev/8-BIT_WONDER.TTF", 10); err != nil {
	// 	panic(err)
	// }
	// byNameInterface, _ := net.InterfaceByName("eth0")
	// fmt.Println(byNameInterface)
	// img.SetColor(color.Black)
	// if strings.Contains(byNameInterface.Flags.String(), "up") {
	// 	img.DrawStringAnchored("Ethernet", 230, 10, 1, 0.5)
	// 	img.DrawImageAnchored(wifi_connected, 240, 10, 0.5, 0.5)

	// } else if Connected() {
	// 	img.DrawStringAnchored("Nokia 8110 4G", 230, 10, 1, 0.5)
	// 	img.DrawImageAnchored(wifi_connected, 240, 10, 0.5, 0.5)
	// } else {
	// 	img.DrawStringAnchored("No internet", 230, 10, 1, 0.5)
	// 	img.DrawImageAnchored(wifi_unconnected, 240, 10, 0.5, 0.5)

	// }
	// // img.DrawImage(wifi_connected, 10, 10)
	// img.Fill()

	// if e := display.Draw(img.Image()); e != nil {
	// 	fmt.Printf("failed to draw: %v\n", e)
	// 	display.Clear(color.White)
	// }
	// fmt.Println("Sleep")

	// display.Sleep()
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
