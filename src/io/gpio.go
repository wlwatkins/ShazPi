package io

import (
	"fmt"
	"log"

	"github.com/stianeikeland/go-rpio/v4"
)

const RST_PIN = 17
const DC_PIN = 25
const CS_PIN = 8
const BUSY_PIN = 24
const PWR_PIN = 18
const TRST_PIN = 22
const INT_PIN = 27

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

type ReadablePinPatch struct {
	rpio.Pin
}

func (pin ReadablePinPatch) Read() uint8 {
	return uint8(pin.Pin.Read())
}

// Transmit is a function that sends the data payload across to the device via the SPI line
type Transmit func(data ...byte)

func New() {
	if err := rpio.Open(); err != nil {
		log.Fatalf("failed to start gpio: %v", err)
	}

	// Enable SPI on SPI0
	if err := rpio.SpiBegin(rpio.Spi0); err != nil {
		log.Fatalf("failed to enable SPI: %v", err)
	}

	// configure SPI settings
	rpio.SpiSpeed(10_000_000)
	rpio.SpiMode(0, 0)

	rpio.Pin(RST_PIN).Mode(rpio.Output)
	rpio.Pin(DC_PIN).Mode(rpio.Output)
	rpio.Pin(CS_PIN).Mode(rpio.Output)
	rpio.Pin(PWR_PIN).Mode(rpio.Output)
	rpio.Pin(BUSY_PIN).Mode(rpio.Input)
	rpio.Pin(TRST_PIN).Mode(rpio.Output)
	rpio.Pin(INT_PIN).Mode(rpio.Input)

	rpio.Pin(PWR_PIN).High()
	fmt.Println("Init done")
}

func Kill() {
	rpio.Pin(RST_PIN).Low()
	rpio.Pin(DC_PIN).Low()
	rpio.Pin(CS_PIN).Low()

	rpio.Pin(TRST_PIN).Low()
	rpio.Close()
}

func GetReadPin(pin int) ReadablePin {
	return ReadablePinPatch{rpio.Pin(pin)}
}

func GetWritePin(pin int) WriteablePin {
	return rpio.Pin(pin)
}
