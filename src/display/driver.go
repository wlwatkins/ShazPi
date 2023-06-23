// Package epd provides driver for Waveshare's E-paper e-ink display
package display // import "go.riyazali.net/epd"

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"time"
)

// Display resolution
const EPD_WIDTH = 296
const EPD_HEIGHT = 122
const (
	DRIVER_OUTPUT_CONTROL                = 0x01
	BOOSTER_SOFT_START_CONTROL           = 0x0C
	GATE_SCAN_START_POSITION             = 0x0F
	DEEP_SLEEP_MODE                      = 0x10
	DATA_ENTRY_MODE_SETTING              = 0x11
	SW_RESET                             = 0x12
	TEMPERATURE_SENSOR_CONTROL           = 0x1A
	MASTER_ACTIVATION                    = 0x20
	DISPLAY_UPDATE_CONTROL_1             = 0x21
	DISPLAY_UPDATE_CONTROL_2             = 0x22
	WRITE_RAM                            = 0x24
	WRITE_VCOM_REGISTER                  = 0x2C
	WRITE_LUT_REGISTER                   = 0x32
	SET_DUMMY_LINE_PERIOD                = 0x3A
	SET_GATE_TIME                        = 0x3B
	BORDER_WAVEFORM_CONTROL              = 0x3C
	SET_RAM_X_ADDRESS_START_END_POSITION = 0x44
	SET_RAM_Y_ADDRESS_START_END_POSITION = 0x45
	SET_RAM_X_ADDRESS_COUNTER            = 0x4E
	SET_RAM_Y_ADDRESS_COUNTER            = 0x4F
	TERMINATE_FRAME_READ_WRITE           = 0xFF

	NO_ROTATION  Rotation = 0
	ROTATION_90  Rotation = 1 // 90 degrees clock-wise rotation
	ROTATION_180 Rotation = 2
	ROTATION_270 Rotation = 3
)

// ErrInvalidImageSize is returned if the given image bounds doesn't fit into display bounds
var ErrInvalidImageSize = errors.New("invalid image size")

// LookupTable defines a type holding the instruction lookup table
// This lookup table is used by the device when performing refreshes
type Mode uint8

type Config struct {
	Width        int16 // Width is the display resolution
	Height       int16
	LogicalWidth int16    // LogicalWidth must be a multiple of 8 and same size or bigger than Width
	Rotation     Rotation // Rotation is clock-wise
}

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
var partialUpdate = []uint8{
	0x0, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x80, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x40, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0A, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1,
	0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x0, 0x0, 0x0,
	0x22, 0x17, 0x41, 0xB0, 0x32, 0x36,
}

// partialUpdate is a lookup table used whilst in partial update mode
var fullUpdate = []uint8{
	0x80, 0x66, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x0, 0x0, 0x0,
	0x10, 0x66, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x20, 0x0, 0x0, 0x0,
	0x80, 0x66, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x0, 0x0, 0x0,
	0x10, 0x66, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x20, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x14, 0x8, 0x0, 0x0, 0x0, 0x0, 0x2,
	0xA, 0xA, 0x0, 0xA, 0xA, 0x0, 0x1,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x14, 0x8, 0x0, 0x1, 0x0, 0x0, 0x1,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x0, 0x0, 0x0,
	0x22, 0x17, 0x41, 0x0, 0x32, 0x36,
}

type Rotation uint8

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
	lut  []uint8

	// SPI transmitter
	transmit Transmit

	logicalWidth int16
	width        int16
	height       int16
	buffer       []uint8
	bufferLength uint32
	rotation     Rotation
}

// reset resets the display back to defaults
func (epd *EPD) Reset() {
	epd.rst.High()
	time.Sleep(50 * time.Millisecond)
	epd.rst.Low()
	time.Sleep(2 * time.Millisecond)
	epd.rst.High()
	time.Sleep(50 * time.Millisecond)
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

// // data transmits single byte of data payload over SPI line
// func (epd *EPD) data2(d byte) {
// 	epd.dc.High()
// 	epd.cs.Low()
// 	epd.transmit2(d)
// 	epd.cs.High()
// }

// idle reads from busy line and waits for the device to get into idle state
func (epd *EPD) ReadBusy() {
	for epd.busy.Read() == 0x01 {
		time.Sleep(10 * time.Millisecond)
	}
}

// turnOnDisplay activates the display and renders the image that's there in the device's RAM
func (epd *EPD) turnOnDisplay() {
	epd.command(DISPLAY_UPDATE_CONTROL_2)
	epd.data(0xC7)
	epd.command(MASTER_ACTIVATION)
	epd.ReadBusy()
}

// turnOnDisplay activates the display and renders the image that's there in the device's RAM
func (epd *EPD) turnOnDisplayPartial() {
	epd.command(DISPLAY_UPDATE_CONTROL_2)
	epd.data(0x0F)
	epd.command(MASTER_ACTIVATION)
	epd.ReadBusy()
}

// SetLUT sets the look up tables for full or partial updates
func (epd *EPD) sendLUT() {
	epd.command(WRITE_LUT_REGISTER)
	// for i := 0; i < 153; i++ {
	for _, value := range epd.lut {
		epd.data(value)
	}
	epd.ReadBusy()

}

// SetLUT sets the look up tables for full or partial updates
func (epd *EPD) SetLUT() {
	epd.sendLUT()
	epd.command(0x3f)
	fmt.Println("fullUpdate %v, partialUpdate %v", len(fullUpdate), len(partialUpdate))
	epd.data(epd.lut[153])
	epd.command(0x03) // gate voltage
	epd.data(epd.lut[154])
	epd.command(0x04)      // source voltage
	epd.data(epd.lut[155]) // VSH
	epd.data(epd.lut[156]) // VSH2
	epd.data(epd.lut[157]) // VSL
	epd.command(0x2c)      // VCOM
	epd.data(epd.lut[158])
}

// window sets the window plane used by device when drawing the image in the buffer
func (epd *EPD) SetWindow(x_start, y_start, x_end, y_end int16) {
	epd.command(SET_RAM_X_ADDRESS_START_END_POSITION)
	epd.data(uint8((x_start >> 3) & 0xFF))
	epd.data(uint8((x_end >> 3) & 0xFF))

	epd.command(SET_RAM_Y_ADDRESS_START_END_POSITION)
	epd.data(byte(y_start & 0xFF))
	epd.data(byte((y_start >> 8) & 0xFF))
	epd.data(byte(y_end & 0xFF))
	epd.data(byte((y_end >> 8) & 0xFF))
}

// cursor sets the cursor position in the device window frame
func (epd *EPD) SetCursor(x uint8, y uint16) {
	epd.command(SET_RAM_X_ADDRESS_COUNTER)
	epd.data((x >> 3) & 0xFF)

	epd.command(SET_RAM_Y_ADDRESS_COUNTER)
	epd.data(byte(y & 0xFF))
	epd.data(byte((y >> 8) & 0xFF))
}

// New creates a new EPD device driver
func New(rst, dc, cs WriteablePin, busy ReadablePin, transmit Transmit) *EPD {
	return &EPD{
		Height:   EPD_HEIGHT,
		Width:    EPD_WIDTH,
		rst:      rst,
		dc:       dc,
		cs:       cs,
		busy:     busy,
		lut:      fullUpdate,
		transmit: transmit,
	}
}

// mode sets the device's mode (based on the LookupTable)
// The device can either be in FullUpdate mode where the whole display is updated each time an image is rendered
// or in PartialUpdate mode where only the changed section is updated (and it doesn't cause any flicker)
//
// Waveshare recommends doing full update of the display at least once per-day to prevent ghost image problems
func (epd *EPD) Configure(cfg Config) {
	if cfg.LogicalWidth != 0 {
		epd.logicalWidth = cfg.LogicalWidth
	} else {
		epd.logicalWidth = 128
	}
	if cfg.Width != 0 {
		epd.width = cfg.Width
	} else {
		epd.width = 128
	}
	if cfg.Height != 0 {
		epd.height = cfg.Height
	} else {
		epd.height = 296
	}
	epd.rotation = cfg.Rotation
	epd.bufferLength = (uint32(epd.logicalWidth) * uint32(epd.height)) / 8
	epd.buffer = make([]uint8, epd.bufferLength)
	for i := uint32(0); i < epd.bufferLength; i++ {
		epd.buffer[i] = 0xFF
	}

	epd.cs.Low()
	epd.dc.Low()
	epd.rst.Low()
	epd.Reset()

	// command+data below is taken from the python sample driver
	epd.ReadBusy()
	epd.command(SW_RESET)
	epd.ReadBusy()

	// SET_GATE_TIME
	epd.command(DRIVER_OUTPUT_CONTROL)
	epd.data(0x27)
	epd.data(0x01)
	epd.data(0x00)

	// DATA_ENTRY_MODE_SETTING
	epd.command(DATA_ENTRY_MODE_SETTING)
	epd.data(0x03)

	epd.SetWindow(0, 0, int16(epd.width)-1, int16(epd.height)-1)

	epd.command(DISPLAY_UPDATE_CONTROL_1)
	epd.data(0x00)
	epd.data(0x80)

	epd.SetCursor(0, 0)
	epd.ReadBusy()

	// WRITE_LUT_REGISTER
	epd.lut = fullUpdate
	epd.SetLUT()
}

func (epd *EPD) getBuffer(img image.Image) error {

	// x, y = d.xy(x, y)
	// if x < 0 || x >= d.logicalWidth || y < 0 || y >= d.height {
	// 	return
	// }
	// byteIndex := (int32(x) + int32(y)*int32(d.logicalWidth)) / 8
	// if c.R == 0 && c.G == 0 && c.B == 0 { // TRANSPARENT / WHITE
	// 	epd.buffer[byteIndex] |= 0x80 >> uint8(x%8)
	// } else { // WHITE / EMPTY
	// 	d.buffer[byteIndex] &^= 0x80 >> uint8(x%8)
	// }

	// var isvertical = img.Bounds().Size().X == epd.Width && img.Bounds().Size().Y == epd.Height
	var _, uniform = img.(*image.Uniform) // special case for uniform images which have infinite bound
	if !uniform {
		return ErrInvalidImageSize
	}

	for i := 0; i < int(epd.height); i++ {
		for j := 0; j < int(epd.width); j += 8 {
			r, g, b, a := img.At(j, i).RGBA()
			px := color.RGBA{
				R: uint8(r),
				G: uint8(g),
				B: uint8(b),
				A: uint8(a),
			}
			epd.SetPixel(int16(j), int16(i), px)
		}
	}

	return nil

}

// Sleep puts the device into "deep sleep" mode where it draws zero (0) current
//
// Waveshare recommends putting the device in "deep sleep" mode (or disconnect from power)
// if doesn't need updating/refreshing.
func (epd *EPD) Sleep() {
	epd.command(DEEP_SLEEP_MODE)
	epd.data(0x01)
}

// xy chages the coordinates according to the rotation
func (d *EPD) xy(x, y int16) (int16, int16) {
	switch d.rotation {
	case NO_ROTATION:
		return x, y
	case ROTATION_90:
		return d.width - y - 1, x
	case ROTATION_180:
		return d.width - x - 1, d.height - y - 1
	case ROTATION_270:
		return y, d.height - x - 1
	}
	return x, y
}

// SetPixel modifies the internal buffer in a single pixel.
// The display have 2 colors: black and white
// We use RGBA(0,0,0, 255) as white (transparent)
// Anything else as black
func (d *EPD) SetPixel(x int16, y int16, c color.RGBA) {
	x, y = d.xy(x, y)
	if x < 0 || x >= d.logicalWidth || y < 0 || y >= d.height {
		return
	}
	byteIndex := (int32(x) + int32(y)*int32(d.logicalWidth)) / 8
	if c.R == 0 && c.G == 0 && c.B == 0 { // TRANSPARENT / WHITE
		d.buffer[byteIndex] |= 0x80 >> uint8(x%8)
	} else { // WHITE / EMPTY
		d.buffer[byteIndex] &^= 0x80 >> uint8(x%8)
	}
}

// Clear clears the display and paints the whole display into c color
func (epd *EPD) Clear(c color.Color) {
	// epd.setMemoryArea(0, 0, epd.logicalWidth-1, epd.height-1)
	// epd.setMemoryPointer(0, 0)
	epd.command(WRITE_RAM)
	for i := uint32(0); i < epd.bufferLength; i++ {
		epd.data(0xFF)
	}
	epd.turnOnDisplay()
}

// Size returns the current size of the display.
func (d *EPD) Size() (w, h int16) {
	if d.rotation == ROTATION_90 || d.rotation == ROTATION_270 {
		return d.height, d.logicalWidth
	}
	return d.logicalWidth, d.height
}

// // ClearDisplay erases the device SRAM
// func (d *EPD) ClearDisplay() {
// 	d.setMemoryArea(0, 0, d.logicalWidth-1, d.height-1)
// 	d.setMemoryPointer(0, 0)
// 	d.command(WRITE_RAM)
// 	for i := uint32(0); i < d.bufferLength; i++ {
// 		d.data(0xFF)
// 	}
// 	d.draw()
// }

// SetRotation changes the rotation (clock-wise) of the device
func (d *EPD) SetRotation(rotation Rotation) {
	d.rotation = rotation
}

// Draw renders the given image onto the display
func (epd *EPD) Draw(img image.Image) error {
	epd.getBuffer(img)
	for _, value := range epd.buffer {
		epd.data(value)
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

// // Package epd provides driver for Waveshare's E-paper e-ink display
// package epd  // import "go.riyazali.net/epd"

// import (
// 	"errors"
// 	"image"
// 	"image/color"
// 	"math"
// 	"time"
// )

// // ErrInvalidImageSize is returned if the given image bounds doesn't fit into display bounds
// var ErrInvalidImageSize = errors.New("invalid image size")

// // LookupTable defines a type holding the instruction lookup table
// // This lookup table is used by the device when performing refreshes
// type Mode uint8

// // WriteablePin is a GPIO pin through which the driver can write digital data
// type WriteablePin interface {
// 	// High sets the pins output to digital high
// 	High()

// 	// Low sets the pins output to digital low
// 	Low()
// }

// // ReadablePin is a GPIO pin through which the driver can read digital data
// type ReadablePin interface {
// 	// Read reads from the pin and return the data as a byte
// 	Read() uint8
// }

// // Transmit is a function that sends the data payload across to the device via the SPI line
// type Transmit func(data ...byte)

// const (
// 	FullUpdate Mode = iota
// 	PartialUpdate
// )

// // fullUpdate is a lookup table used whilst in full update mode
// var fullUpdate = []byte{
// 	0x50, 0xAA, 0x55, 0xAA, 0x11, 0x00,
// 	0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// 	0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// 	0x00, 0x00, 0xFF, 0xFF, 0x1F, 0x00,
// 	0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// }

// // partialUpdate is a lookup table used whilst in partial update mode
// var partialUpdate = []byte{
// 	0x10, 0x18, 0x18, 0x08, 0x18, 0x18,
// 	0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
// 	0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// 	0x00, 0x00, 0x13, 0x14, 0x44, 0x12,
// 	0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// }

// // EPD defines the base type for the e-paper display driver
// type EPD struct {
// 	// dimensions of the display
// 	Height int
// 	Width  int

// 	// pins used by this driver
// 	rst  WriteablePin // for reset signal
// 	dc   WriteablePin // for data/command select signal; D=HIGH C=LOW
// 	cs   WriteablePin // for chip select signal; this pin is active low
// 	busy ReadablePin  // for reading in busy signal

// 	// SPI transmitter
// 	transmit Transmit
// }

// // New creates a new EPD device driver
// func New(rst, dc, cs WriteablePin, busy ReadablePin, transmit Transmit) *EPD {
// 	return &EPD{296, 128, rst, dc, cs, busy, transmit}
// }

// // reset resets the display back to defaults
// func (epd *EPD) reset() {
// 	epd.rst.High()
// 	time.Sleep(200 * time.Millisecond)
// 	epd.rst.Low()
// 	time.Sleep(10 * time.Millisecond)
// 	epd.rst.High()
// 	time.Sleep(200 * time.Millisecond)
// }

// // command transmits single byte of command instruction over the SPI line
// func (epd *EPD) command(c byte) {
// 	epd.dc.Low()
// 	epd.cs.Low()
// 	epd.transmit(c)
// 	epd.cs.High()
// }

// // data transmits single byte of data payload over SPI line
// func (epd *EPD) data(d byte) {
// 	epd.dc.High()
// 	epd.cs.Low()
// 	epd.transmit(d)
// 	epd.cs.High()
// }

// // idle reads from busy line and waits for the device to get into idle state
// func (epd *EPD) idle() {
// 	for epd.busy.Read() == 0x1 {
// 		time.Sleep(200 * time.Millisecond)
// 	}
// }

// // mode sets the device's mode (based on the LookupTable)
// // The device can either be in FullUpdate mode where the whole display is updated each time an image is rendered
// // or in PartialUpdate mode where only the changed section is updated (and it doesn't cause any flicker)
// //
// // Waveshare recommends doing full update of the display at least once per-day to prevent ghost image problems
// func (epd *EPD) Mode(mode Mode) {
// 	epd.reset()

// 	// command+data below is taken from the python sample driver

// 	// DRIVER_OUTPUT_CONTROL
// 	epd.command(0x01)
// 	epd.data(byte((epd.Height - 1) & 0xFF))
// 	epd.data(byte(((epd.Height - 1) >> 8) & 0xFF))
// 	epd.data(0x00)

// 	// BOOSTER_SOFT_START_CONTROL
// 	epd.command(0x0C)
// 	epd.data(0xD7)
// 	epd.data(0xD6)
// 	epd.data(0x9D)

// 	// WRITE_VCOM_REGISTER
// 	epd.command(0x2C)
// 	epd.data(0xA8)

// 	// SET_DUMMY_LINE_PERIOD
// 	epd.command(0x3A)
// 	epd.data(0x1A)

// 	// SET_GATE_TIME
// 	epd.command(0x3B)
// 	epd.data(0x08)

// 	// DATA_ENTRY_MODE_SETTING
// 	epd.command(0x11)
// 	epd.data(0x03)

// 	// WRITE_LUT_REGISTER
// 	epd.command(0x32)
// 	var lut = fullUpdate
// 	if mode == PartialUpdate {
// 		lut = partialUpdate
// 	}
// 	for _, b := range lut {
// 		epd.data(b)
// 	}
// }

// // Sleep puts the device into "deep sleep" mode where it draws zero (0) current
// //
// // Waveshare recommends putting the device in "deep sleep" mode (or disconnect from power)
// // if doesn't need updating/refreshing.
// func (epd *EPD) Sleep() {
// 	epd.command(0x10)
// 	epd.data(0x01)
// }

// // turnOnDisplay activates the display and renders the image that's there in the device's RAM
// func (epd *EPD) turnOnDisplay() {
// 	epd.command(0x22)
// 	epd.data(0xC4)
// 	epd.command(0x20)
// 	epd.command(0xFF)
// 	epd.idle()
// }

// // window sets the window plane used by device when drawing the image in the buffer
// func (epd *EPD) window(x0, x1 byte, y0, y1 uint16) {
// 	epd.command(0x44)
// 	epd.data((x0 >> 3) & 0xFF)
// 	epd.data((x1 >> 3) & 0xFF)

// 	epd.command(0x45)
// 	epd.data(byte(y0 & 0xFF))
// 	epd.data(byte((y0 >> 8) & 0xFF))
// 	epd.data(byte(y1 & 0xFF))
// 	epd.data(byte((y1 >> 8) & 0xFF))
// }

// // cursor sets the cursor position in the device window frame
// func (epd *EPD) cursor(x uint8, y uint16) {
// 	epd.command(0x4E)
// 	epd.data((x >> 3) & 0xFF)

// 	epd.command(0x4F)
// 	epd.data(byte(y & 0xFF))
// 	epd.data(byte((y >> 8) & 0xFF))

// 	epd.idle()
// }

// // Clear clears the display and paints the whole display into c color
// func (epd *EPD) Clear(c color.Color) {
// 	var img = image.White
// 	if c != color.White {
// 		img = image.Black // anything other than white is treated as black
// 	}
// 	_ = epd.Draw(img)
// }

// // Draw renders the given image onto the display
// func (epd *EPD) Draw(img image.Image) error {
// 	var isvertical = img.Bounds().Size().X == epd.Width && img.Bounds().Size().Y == epd.Height
// 	var _, uniform = img.(*image.Uniform) // special case for uniform images which have infinite bound
// 	if !uniform && !isvertical {
// 		return ErrInvalidImageSize
// 	}

// 	epd.window(0, byte(epd.Width-1), 0, uint16(epd.Height-1))
// 	for i := 0; i < epd.Height; i++ {
// 		epd.cursor(0, uint16(i))
// 		epd.command(0x24) // WRITE_RAM
// 		for j := 0; j < epd.Width; j += 8 {
// 			// this loop converts individual pixels into a single byte
// 			// 8-pixels at a time and then sends that byte to render
// 			var b = 0xFF
// 			for px := 0; px < 8; px++ {
// 				var pixel = img.At(j+px, i)
// 				if isdark(pixel.RGBA()) {
// 					b &= ^(0x80 >> (px % 8))
// 				}
// 			}
// 			epd.data(byte(b))
// 		}
// 	}
// 	epd.turnOnDisplay()
// 	return nil
// }

// // isdark is a utility method which returns true if the pixel color is considered dark else false
// // this function is taken from https://git.io/JviWg
// func isdark(r, g, b, _ uint32) bool {
// 	return math.Sqrt(
// 		0.299*math.Pow(float64(r), 2)+
// 			0.587*math.Pow(float64(g), 2)+
// 			0.114*math.Pow(float64(b), 2)) <= 130
// }
