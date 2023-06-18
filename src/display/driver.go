// Package epd provides driver for Waveshare's E-paper e-ink display
package display // import "go.riyazali.net/epd"

import (
	"errors"
	"image"
	"image/color"
	"math"
	"time"
)

// Display resolution
const EPD_WIDTH = 296
const EPD_HEIGHT = 122

// EPD2IN9 commands
const DRIVER_OUTPUT_CONTROL = 0x01
const BOOSTER_SOFT_START_CONTROL = 0x0C
const GATE_SCAN_START_POSITION = 0x0F
const DEEP_SLEEP_MODE = 0x10
const DATA_ENTRY_MODE_SETTING = 0x11
const SW_RESET = 0x12
const TEMPERATURE_SENSOR_CONTROL = 0x1A
const MASTER_ACTIVATION = 0x20
const DISPLAY_UPDATE_CONTROL_1 = 0x21
const DISPLAY_UPDATE_CONTROL_2 = 0x22
const WRITE_RAM = 0x24
const WRITE_VCOM_REGISTER = 0x2C
const WRITE_LUT_REGISTER = 0x32
const SET_DUMMY_LINE_PERIOD = 0x3A
const SET_GATE_TIME = 0x3B
const BORDER_WAVEFORM_CONTROL = 0x3C
const SET_RAM_X_ADDRESS_START_END_POSITION = 0x44
const SET_RAM_Y_ADDRESS_START_END_POSITION = 0x45
const SET_RAM_X_ADDRESS_COUNTER = 0x4E
const SET_RAM_Y_ADDRESS_COUNTER = 0x4F
const TERMINATE_FRAME_READ_WRITE = 0xFF

// ErrInvalidImageSize is returned if the given image bounds doesn't fit into display bounds
var ErrInvalidImageSize = errors.New("invalid image size")

// LookupTable defines a type holding the instruction lookup table
// This lookup table is used by the device when performing refreshes
type Mode uint8

// WriteablePin is a GPIO pin through which the driver can write digital data
type WriteablePin interface {
	// High sets the pins output to digital high
	High()

	// Low sets the pins output to digital low
	Low()
}

// ReadablePin is a GPIO pin through which the driver can read digital data
type ReadablePin interface {
	// Read reads from the pin and return the data as a byte
	Read() uint8
}

// Transmit is a function that sends the data payload across to the device via the SPI line
type Transmit func(data ...byte)

const (
	FullUpdate Mode = iota
	PartialUpdate
)

// fullUpdate is a lookup table used whilst in full update mode
var fullUpdate = []byte{
	0x02, 0x02, 0x01, 0x11, 0x12, 0x12, 0x22, 0x22,
	0x66, 0x69, 0x69, 0x59, 0x58, 0x99, 0x99, 0x88,
	0x00, 0x00, 0x00, 0x00, 0xF8, 0xB4, 0x13, 0x51,
	0x35, 0x51, 0x51, 0x19, 0x01, 0x00,
}

// partialUpdate is a lookup table used whilst in partial update mode
var partialUpdate = []byte{ //20 bytes
	0x10, 0x18, 0x18, 0x08, 0x18, 0x18, 0x08, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x13, 0x14, 0x44, 0x12,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

// EPD defines the base type for the e-paper display driver
type EPD struct {
	// dimensions of the display
	Height int
	Width  int

	// pins used by this driver
	rst  WriteablePin // for reset signal
	dc   WriteablePin // for data/command select signal; D=HIGH C=LOW
	cs   WriteablePin // for chip select signal; this pin is active low
	busy ReadablePin  // for reading in busy signal

	lut []byte

	// SPI transmitter
	transmit Transmit
}

// New creates a new EPD device driver
func New(rst, dc, cs WriteablePin, busy ReadablePin, transmit Transmit) *EPD {
	return &EPD{EPD_HEIGHT, EPD_WIDTH, rst, dc, cs, busy, fullUpdate, transmit}
}

// reset resets the display back to defaults
func (epd *EPD) reset() {
	epd.rst.Low()
	time.Sleep(5 * time.Millisecond)
	epd.rst.High()
	time.Sleep(200 * time.Millisecond)
}

// command transmits single byte of command instruction over the SPI line
func (epd *EPD) command(c byte) {
	epd.dc.Low()
	epd.cs.Low()
	epd.transmit(c)
	epd.cs.High()
}

// data transmits single byte of data payload over SPI line
func (epd *EPD) data(d byte) {
	epd.dc.High()
	epd.cs.Low()
	epd.transmit(d)
	epd.cs.High()
}

// idle reads from busy line and waits for the device to get into idle state
func (epd *EPD) idle() {
	for epd.busy.Read() == 0x1 {
		time.Sleep(100 * time.Millisecond)
	}
}

// mode sets the device's mode (based on the LookupTable)
// The device can either be in FullUpdate mode where the whole display is updated each time an image is rendered
// or in PartialUpdate mode where only the changed section is updated (and it doesn't cause any flicker)
//
// Waveshare recommends doing full update of the display at least once per-day to prevent ghost image problems
func (epd *EPD) Mode(mode Mode) {
	epd.reset()

	// command+data below is taken from the python sample driver

	// DRIVER_OUTPUT_CONTROL
	epd.command(DRIVER_OUTPUT_CONTROL)
	epd.data(byte((epd.Height - 1) & 0xFF))
	epd.data(byte(((epd.Height - 1) >> 8) & 0xFF))
	epd.data(0x00)

	// BOOSTER_SOFT_START_CONTROL
	epd.command(BOOSTER_SOFT_START_CONTROL)
	epd.data(0xD7)
	epd.data(0xD6)
	epd.data(0x9D)

	// WRITE_VCOM_REGISTER
	epd.command(WRITE_VCOM_REGISTER)
	epd.data(0xA8)

	// SET_DUMMY_LINE_PERIOD
	epd.command(SET_DUMMY_LINE_PERIOD)
	epd.data(0x1A)

	// SET_GATE_TIME
	epd.command(SET_GATE_TIME)
	epd.data(0x08)

	// DATA_ENTRY_MODE_SETTING
	epd.command(DATA_ENTRY_MODE_SETTING)
	epd.data(0x03)

	// WRITE_LUT_REGISTER
	epd.command(WRITE_LUT_REGISTER)

	for _, b := range epd.lut {
		epd.data(b)
	}
}

// Sleep puts the device into "deep sleep" mode where it draws zero (0) current
//
// Waveshare recommends putting the device in "deep sleep" mode (or disconnect from power)
// if doesn't need updating/refreshing.
func (epd *EPD) Sleep() {
	epd.command(DEEP_SLEEP_MODE)
	epd.idle()
}

// turnOnDisplay activates the display and renders the image that's there in the device's RAM
func (epd *EPD) turnOnDisplay() {
	// epd.command(0x22)
	// epd.data(0xC7)
	// epd.command(0x20)
	epd.command(DISPLAY_UPDATE_CONTROL_2)
	epd.data(0xC4)
	epd.command(MASTER_ACTIVATION)
	epd.command(TERMINATE_FRAME_READ_WRITE)
	epd.idle()
}

// window sets the window plane used by device when drawing the image in the buffer
func (epd *EPD) window(x_start, x_end byte, y_start, y_end uint16) {
	epd.command(SET_RAM_X_ADDRESS_START_END_POSITION)
	epd.data((x_start >> 3) & 0xFF)
	epd.data((x_end >> 3) & 0xFF)

	epd.command(SET_RAM_Y_ADDRESS_START_END_POSITION)
	epd.data(byte(y_start & 0xFF))
	epd.data(byte((y_start >> 8) & 0xFF))
	epd.data(byte(y_end & 0xFF))
	epd.data(byte((y_end >> 8) & 0xFF))
}

// cursor sets the cursor position in the device window frame
func (epd *EPD) cursor(x uint8, y uint16) {
	epd.command(SET_RAM_X_ADDRESS_COUNTER)
	epd.data((x >> 3) & 0xFF)

	epd.command(SET_RAM_Y_ADDRESS_COUNTER)
	epd.data(byte(y & 0xFF))
	epd.data(byte((y >> 8) & 0xFF))

	epd.idle()
}

// Clear clears the display and paints the whole display into c color
func (epd *EPD) Clear(c color.Color) {
	var img = image.White
	if c != color.White {
		img = image.Black // anything other than white is treated as black
	}
	_ = epd.Draw(img)
}

// Draw renders the given image onto the display
func (epd *EPD) Draw(img image.Image) error {
	// var isvertical = img.Bounds().Size().X == epd.Width && img.Bounds().Size().Y == epd.Height
	// var _, uniform = img.(*image.Uniform) // special case for uniform images which have infinite bound
	// if !uniform && !isvertical {
	// 	return ErrInvalidImageSize
	// }

	// epd.window(0, byte(epd.Width-1), 0, uint16(epd.Height-1))
	// for i := 0; i < epd.Height; i++ {
	// 	epd.cursor(0, uint16(i))
	// 	epd.command(0x24) // WRITE_RAM
	// 	for j := 0; j < epd.Width; j += 8 {
	// 		// this loop converts individual pixels into a single byte
	// 		// 8-pixels at a time and then sends that byte to render
	// 		var b = 0xFF
	// 		for px := 0; px < 8; px++ {
	// 			var pixel = img.At(j+px, i)
	// 			if isdark(pixel.RGBA()) {
	// 				b &= ^(0x80 >> (px % 8))
	// 			}
	// 		}
	// 		epd.data(byte(b))
	// 	}
	// }
	// epd.turnOnDisplay()
	// return nil

	epd.window(0, byte(epd.Width-1), 0, uint16(epd.Height-1))
	epd.cursor(0, 0)
	epd.command(WRITE_RAM)
	var b byte = 0x00
	for j := 0; j < epd.Height; j++ {
		for i := 0; i < epd.Width; i++ {
			var pixel = img.At(j, i)
			if isdark(pixel.RGBA()) {
				b |= 0x80 >> (i % 8)
			}
			if i%8 == 7 {
				epd.data(b)
				b = 0x00
			}
		}

	}

	return nil

}

// isdark is a utility method which returns true if the pixel color is considered dark else false
// this function is taken from https://git.io/JviWg
func isdark(r, g, b, _ uint32) bool {
	return math.Sqrt(
		0.299*math.Pow(float64(r), 2)+
			0.587*math.Pow(float64(g), 2)+
			0.114*math.Pow(float64(b), 2)) <= 130
}
