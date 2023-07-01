// Package epd provides driver for Waveshare's E-paper e-ink display https://github.com/waveshareteam/Touch_e-Paper_HAT/blob/main/python/lib/TP_lib/epd2in13_V3.py
package display // import "go.riyazali.net/epd"

import (
	"errors"
	"image/color"
	"math"
	"shazammini/src/io"
	"time"

	"github.com/fogleman/gg"
)

// Display resolution
const EPD_WIDTH = 122
const EPD_HEIGHT = 250
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
	SET_ANALOG_BLOCK_CONTROL             = 0x74
	SET_DIGITAL_BLOCK_CONTROL            = 0x7E

	ROTATION_0   Rotation = 0
	ROTATION_90  Rotation = 1 // 90 degrees clock-wise rotation
	ROTATION_180 Rotation = 2
	ROTATION_270 Rotation = 3
)

// ErrInvalidImageSize is returned if the given image bounds doesn't fit into display bounds
var ErrInvalidImageSize = errors.New("invalid image size")

type Rotation uint8

type Config struct {
	Width        int16 // Width is the display resolution
	Height       int16
	LogicalWidth int16    // LogicalWidth must be a multiple of 8 and same size or bigger than Width
	Rotation     Rotation // Rotation is clock-wise
}

// Transmit is a function that sends the data payload across to the device via the SPI line
type Transmit func(data ...byte)

// fullUpdate is a lookup table used whilst in full update mode
var fullUpdateLut = [159]uint8{
	0x80, 0x4A, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x40, 0x4A, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x80, 0x4A, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x40, 0x4A, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0xF, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0xF, 0x0, 0x0, 0xF, 0x0, 0x0, 0x2,
	0xF, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x0, 0x0, 0x0,
	0x22, 0x17, 0x41, 0x0, 0x32, 0x36,
}

// partialUpdate is a lookup table used whilst in partial update mode

var partialUpdateLut = [159]uint8{
	0x0, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x80, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x40, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x10, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
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
	0x22, 0x17, 0x41, 0x00, 0x32, 0x36,
}

// EPD defines the base type for the e-paper display driver
type EPD struct {
	// pins used by this driver
	rst  io.WriteablePin // for reset signal
	dc   io.WriteablePin // for data/command select signal; D=HIGH C=LOW
	cs   io.WriteablePin // for chip select signal; this pin is active low
	busy io.ReadablePin  // for reading in busy signal

	// SPI transmitter
	transmit Transmit

	logicalWidth int16
	width        int16
	height       int16
	buffer       []uint8
	bufferLength uint32
	rotation     Rotation

	Update Update
}

// New creates a new EPD device driver
func New(rst, dc, cs io.WriteablePin, busy io.ReadablePin, transmit Transmit) *EPD {
	return &EPD{
		rst:      rst,
		dc:       dc,
		cs:       cs,
		busy:     busy,
		Update:   FullUpdate,
		transmit: transmit,
	}
}

// reset resets the display back to defaults
func (epd *EPD) Reset() {
	epd.rst.High()
	time.Sleep(20 * time.Millisecond)
	epd.rst.Low()
	time.Sleep(2 * time.Millisecond)
	epd.rst.High()
	time.Sleep(20 * time.Millisecond)
}

// command transmits single byte of command instruction over the SPI line
func (epd *EPD) sendCommand(c byte) {
	epd.dc.Low()
	epd.cs.Low()
	epd.transmit(c)
	epd.cs.High()
}

// data transmits single byte of data payload over SPI line
func (epd *EPD) sendData(d byte) {
	epd.dc.High()
	epd.cs.Low()
	epd.transmit(d)
	epd.cs.High()
}

// idle reads from busy line and waits for the device to get into idle state
func (epd *EPD) ReadBusy() {
	for epd.busy.Read() == 0x01 {
		time.Sleep(10 * time.Millisecond)
	}
}

// turnOnDisplay activates the display and renders the image that's there in the device's RAM
func (epd *EPD) turnOnDisplay() {
	epd.sendCommand(DISPLAY_UPDATE_CONTROL_2)
	epd.sendData(0xC7)
	epd.sendCommand(MASTER_ACTIVATION)
	epd.ReadBusy()
}

// turnOnDisplay activates the display and renders the image that's there in the device's RAM
func (epd *EPD) turnOnDisplayPartial() {
	epd.sendCommand(DISPLAY_UPDATE_CONTROL_2)
	epd.sendData(0x0C)
	epd.sendCommand(MASTER_ACTIVATION)
	// epd.ReadBusy()
}

// turnOnDisplay activates the display and renders the image that's there in the device's RAM
func (epd *EPD) turnOnDisplayPartialWait() {
	epd.sendCommand(DISPLAY_UPDATE_CONTROL_2)
	epd.sendData(0x0C)
	epd.sendCommand(MASTER_ACTIVATION)
	epd.ReadBusy()
}

type Update int64

const (
	FullUpdate    Update = 0
	PartialUpdate Update = 1
)

/*
function : Set lut
parameter:

	lut : lut data
*/
func (epd *EPD) Lut(lut [159]uint8) {

	epd.sendCommand(0x32)
	for _, v := range lut {
		epd.sendData(v)
	}
	epd.ReadBusy()
}

/*
function : Send lut data and configuration
parameter:

	lut : lut data
*/
func (epd *EPD) setLut(lut [159]uint8) {

	epd.Lut(lut)
	epd.sendCommand(0x3f)
	epd.sendData(lut[153])
	epd.sendCommand(0x03) // gate voltage
	epd.sendData(lut[154])
	epd.sendCommand(0x04)  // source voltage
	epd.sendData(lut[155]) // VSH
	epd.sendData(lut[156]) // VSH2
	epd.sendData(lut[157]) // VSL
	epd.sendCommand(0x2c)  // VCOM
	epd.sendData(lut[158])
}

/*
function : Setting the display window
setWindow sets the area of the display that will be updated
parameter:

	xstart : X-axis starting position
	ystart : Y-axis starting position
	xend : End position of X-axis
	yend : End position of Y-axis
*/
func (epd *EPD) setWindow(x0 int16, y0 int16, x1 int16, y1 int16) {
	epd.sendCommand(SET_RAM_X_ADDRESS_START_END_POSITION)
	epd.sendData(uint8((x0 >> 3) & 0xFF))
	epd.sendData(uint8((x1 >> 3) & 0xFF))
	epd.sendCommand(SET_RAM_Y_ADDRESS_START_END_POSITION)
	epd.sendData(uint8(y0 & 0xFF))
	epd.sendData(uint8((y0 >> 8) & 0xFF))
	epd.sendData(uint8(y1 & 0xFF))
	epd.sendData(uint8((y1 >> 8) & 0xFF))
}

// setCursor moves the internal pointer to the speficied coordinates
func (epd *EPD) setCursor(x int16, y int16) {
	epd.sendCommand(SET_RAM_X_ADDRESS_COUNTER)
	epd.sendData(uint8(x & 0xFF))
	epd.sendCommand(SET_RAM_Y_ADDRESS_COUNTER)
	epd.sendData(uint8(y & 0xFF))
	epd.sendData(uint8((y >> 8) & 0xFF))
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
		epd.logicalWidth = EPD_WIDTH
	}
	if cfg.Width != 0 {
		epd.width = cfg.Width
	} else {
		epd.width = EPD_WIDTH
	}
	if cfg.Height != 0 {
		epd.height = cfg.Height
	} else {
		epd.height = EPD_HEIGHT
	}
	epd.rotation = cfg.Rotation
	epd.bufferLength = (uint32(epd.logicalWidth) * uint32(epd.height)) / 8
	epd.buffer = make([]uint8, epd.bufferLength)
	for i := uint32(0); i < epd.bufferLength; i++ {
		epd.buffer[i] = 0xFF
	}

	if epd.Update == FullUpdate {
		epd.Reset()

		epd.ReadBusy()
		epd.sendCommand(SW_RESET)
		epd.ReadBusy()

		// epd.sendCommand(SET_ANALOG_BLOCK_CONTROL) //set analog block control
		// epd.sendData(0x54)
		// epd.sendCommand(SET_DIGITAL_BLOCK_CONTROL) //set digital block control
		// epd.sendData(SET_GATE_TIME)

		epd.sendCommand(DRIVER_OUTPUT_CONTROL) //Driver output control
		epd.sendData(0xF9)
		epd.sendData(0x00)
		epd.sendData(0x00)

		epd.sendCommand(DATA_ENTRY_MODE_SETTING) //data entry mode
		epd.sendData(0x03)

		epd.setWindow(0, 0, epd.width-1, epd.height-1)
		epd.setCursor(0, 0)

		// epd.sendCommand(SET_RAM_X_ADDRESS_START_END_POSITION) //set Ram-X address start/end position
		// epd.sendData(0x00)
		// epd.sendData(GATE_SCAN_START_POSITION) //0x0C-->(15+1)*8=128

		// epd.sendCommand(SET_RAM_Y_ADDRESS_START_END_POSITION) //set Ram-Y address start/end position
		// epd.sendData(0xF9)                                    //0xF9-->(249+1)=250
		// epd.sendData(0x00)
		// epd.sendData(0x00)
		// epd.sendData(0x00)

		epd.sendCommand(BORDER_WAVEFORM_CONTROL) //BorderWavefrom
		epd.sendData(0x05)

		epd.sendCommand(DISPLAY_UPDATE_CONTROL_1) //VCOM Voltage
		epd.sendData(0x00)                        //
		epd.sendData(0x80)                        //

		epd.sendCommand(0x18) //VCOM Voltage
		epd.sendData(0x80)    //

		epd.ReadBusy()

		epd.setLut(fullUpdateLut)

	} else {

		epd.rst.Low()
		time.Sleep(1 * time.Millisecond)
		epd.rst.High()
		
		epd.setLut(partialUpdateLut)

		epd.sendCommand(0x37)
		epd.sendData(0x00)
		epd.sendData(0x00)
		epd.sendData(0x00)
		epd.sendData(0x00)
		epd.sendData(0x00)
		epd.sendData(0x40)
		epd.sendData(0x00)
		epd.sendData(0x00)
		epd.sendData(0x00)
		epd.sendData(0x00)

		epd.sendCommand(BORDER_WAVEFORM_CONTROL)
		epd.sendData(0x80)

		epd.sendCommand(DISPLAY_UPDATE_CONTROL_2)
		epd.sendData(0xC0)
		epd.sendCommand(MASTER_ACTIVATION)
		epd.ReadBusy()

		epd.setWindow(0, 0, epd.width-1, epd.height-1)
		epd.setCursor(0, 0)
	}

}

func (epd *EPD) Draw(img *gg.Context) error {
	// fmt.Println(img.Bounds())
	// for i := 0; i < int(img.Bounds().Max.Y); i++ {
	// 	for j := 0; j < int(img.Bounds().Max.X); j++ {

	// 		var pixel = img.At(i, j)
	// 		r, g, b, a := pixel.RGBA()
	// 		var c color.RGBA = color.RGBA{
	// 			R: uint8(r),
	// 			G: uint8(g),
	// 			B: uint8(b),
	// 			A: uint8(a),
	// 		}
	// 		epd.SetPixel(int16(i), int16(j), c)
	// 		// }
	// 	}
	// }

	epd.Display(img)
	return nil
}

// SetPixel modifies the internal buffer in a single pixel.
// The display have 2 colors: black and white
// We use RGBA(0,0,0, 255) as white (transparent)
// Anything else as black
func (epd *EPD) SetPixel(x int16, y int16, c color.RGBA) {
	x, y = epd.xy(x, y)
	if x < 0 || x >= epd.logicalWidth || y < 0 || y >= epd.height {
		return
	}
	byteIndex := (x + y*epd.logicalWidth) / 8

	for px := 0; px < 8; px++ {
		if c.R == 0 && c.G == 0 && c.B == 0 { // TRANSPARENT / WHITE
			epd.buffer[byteIndex] |= 0x80 >> uint8(x%8)
		} else { // WHITE / EMPTY
			epd.buffer[byteIndex] &= ^(0x80 >> (x % 8))
		}
	}
}

// Display sends the buffer to the screen.
func (epd *EPD) Display(img *gg.Context) error {
	// img.RotateAbout(PI/2, float64(epd.height/2), float64(epd.width)/2)
	// img.Translate(float64(epd.height/2), 0)
	// epd.setMemoryArea(0, 0, epd.logicalWidth-1, epd.height-1)
	// for j := int16(0); j < epd.height; j++ {
	// 	epd.setMemoryPointer(0, j)
	// 	epd.sendCommand(WRITE_RAM)
	// 	for i := 0; i < int(epd.logicalWidth); i++ {
	// 		// this loop converts individual pixels into a single byte
	// 		// 8-pixels at a time and then sends that byte to render
	// 		var b = 0xFF
	// 		it, jt := epd.xy(int16(i), int16(j))
	// 		_ = it
	// 		_ = jt
	// 		var pixel = img.At(int(it), int(jt))
	// 		for px := int16(0); px < 8; px++ {
	// 			if isdark(pixel.RGBA()) {
	// 				b &= ^(0x80 >> (px % 8))
	// 			}
	// 		}
	// 		epd.sendData(byte(b))
	// 	}
	// }

	epd.setWindow(0, 0, epd.logicalWidth-1, epd.height-1)
	for j := int16(0); j < epd.height; j++ {
		epd.setCursor(0, j)
		epd.sendCommand(WRITE_RAM)
		for i := int16(0); i < epd.logicalWidth; i += 8 {
			var b = 0xFF
			for px := 0; px < 8; px++ {
				i2, j2 := epd.xy(i, j)
				var pixel = img.Image().At(int(i2)+px, int(j2))
				if isdark(pixel.RGBA()) {
					b &= ^(0x80 >> (px % 8))
				}
			}
			epd.sendData(byte(b))
		}
	}

	epd.turnOnDisplay()
	return nil
}

// Sleep puts the device into "deep sleep" mode where it draws zero (0) current
//
// Waveshare recommends putting the device in "deep sleep" mode (or disconnect from power)
// if doesn't need updating/refreshing.
func (epd *EPD) Sleep() {
	epd.sendCommand(DEEP_SLEEP_MODE)
	epd.sendData(0x03)
}

// xy chages the coordinates according to the rotation
func (d *EPD) xy(x, y int16) (int16, int16) {
	switch d.rotation {
	case ROTATION_0:
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

// ClearBuffer sets the buffer to 0xFF (white)
func (epd *EPD) ClearBuffer() {
	for i := uint32(0); i < epd.bufferLength; i++ {
		epd.buffer[i] = 0xFF
	}
}

// ClearDisplay erases the device SRAM
func (epd *EPD) ClearDisplay() {
	epd.setWindow(0, 0, epd.logicalWidth-1, epd.height-1)
	epd.setCursor(0, 0)
	epd.sendCommand(WRITE_RAM)
	for i := uint32(0); i < epd.bufferLength; i++ {
		epd.sendData(0xFF)
	}
	// epd.Display()
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
// 	d.sendCommand(WRITE_RAM)
// 	for i := uint32(0); i < d.bufferLength; i++ {
// 		d.sendData(0xFF)
// 	}
// 	d.draw()
// }

// SetRotation changes the rotation (clock-wise) of the device
func (d *EPD) SetRotation(rotation Rotation) {
	d.rotation = rotation
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
// 	epd.sendCommand(0x01)
// 	epd.sendData(byte((epd.Height - 1) & 0xFF))
// 	epd.sendData(byte(((epd.Height - 1) >> 8) & 0xFF))
// 	epd.sendData(0x00)

// 	// BOOSTER_SOFT_START_CONTROL
// 	epd.sendCommand(0x0C)
// 	epd.sendData(0xD7)
// 	epd.sendData(0xD6)
// 	epd.sendData(0x9D)

// 	// WRITE_VCOM_REGISTER
// 	epd.sendCommand(0x2C)
// 	epd.sendData(0xA8)

// 	// SET_DUMMY_LINE_PERIOD
// 	epd.sendCommand(0x3A)
// 	epd.sendData(0x1A)

// 	// SET_GATE_TIME
// 	epd.sendCommand(0x3B)
// 	epd.sendData(0x08)

// 	// DATA_ENTRY_MODE_SETTING
// 	epd.sendCommand(0x11)
// 	epd.sendData(0x03)

// 	// WRITE_LUT_REGISTER
// 	epd.sendCommand(0x32)
// 	var lut = fullUpdate
// 	if mode == PartialUpdate {
// 		lut = partialUpdate
// 	}
// 	for _, b := range lut {
// 		epd.sendData(b)
// 	}
// }

// // Sleep puts the device into "deep sleep" mode where it draws zero (0) current
// //
// // Waveshare recommends putting the device in "deep sleep" mode (or disconnect from power)
// // if doesn't need updating/refreshing.
// func (epd *EPD) Sleep() {
// 	epd.sendCommand(0x10)
// 	epd.sendData(0x01)
// }

// // turnOnDisplay activates the display and renders the image that's there in the device's RAM
// func (epd *EPD) turnOnDisplay() {
// 	epd.sendCommand(0x22)
// 	epd.sendData(0xC4)
// 	epd.sendCommand(0x20)
// 	epd.sendCommand(0xFF)
// 	epd.idle()
// }

// // window sets the window plane used by device when drawing the image in the buffer
// func (epd *EPD) window(x0, x1 byte, y0, y1 uint16) {
// 	epd.sendCommand(0x44)
// 	epd.sendData((x0 >> 3) & 0xFF)
// 	epd.sendData((x1 >> 3) & 0xFF)

// 	epd.sendCommand(0x45)
// 	epd.sendData(byte(y0 & 0xFF))
// 	epd.sendData(byte((y0 >> 8) & 0xFF))
// 	epd.sendData(byte(y1 & 0xFF))
// 	epd.sendData(byte((y1 >> 8) & 0xFF))
// }

// // cursor sets the cursor position in the device window frame
// func (epd *EPD) cursor(x uint8, y uint16) {
// 	epd.sendCommand(0x4E)
// 	epd.sendData((x >> 3) & 0xFF)

// 	epd.sendCommand(0x4F)
// 	epd.sendData(byte(y & 0xFF))
// 	epd.sendData(byte((y >> 8) & 0xFF))

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
// 		epd.sendCommand(0x24) // WRITE_RAM
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
// 			epd.sendData(byte(b))
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
