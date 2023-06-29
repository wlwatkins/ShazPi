package commands

import (
	"fmt"
	"log"
	"time"

	"github.com/d2r2/go-i2c"
)

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

type GT_Development struct {
	Touch           int
	TouchpointFlag  int
	TouchCount      int
	Touchkeytrackid [5]int
	X               [5]int
	Y               [5]int
	S               [5]int
}

type GT1151 struct {
	rst  WriteablePin
	DC   WriteablePin
	CS   WriteablePin
	BUSY ReadablePin
	TRST Transmit
	INT  int
	i2c  *i2c.I2C
}

func New() GT1151 {
	// Create new connection to I2C bus on 2 line with address 0x27
	i2c, err := i2c.NewI2C(0x27, 2)
	if err != nil {
		log.Fatal(err)
	}
	// Free I2C connection on exit
	defer i2c.Close()
	gt := GT1151{
		i2c: i2c,
	}
	return gt
}

func (gt *GT1151) GT_Reset() {

	gt.rst.High()
	time.Sleep(100 * time.Millisecond)
	gt.rst.Low()
	time.Sleep(100 * time.Millisecond)
	gt.rst.High()
	time.Sleep(100 * time.Millisecond)

}

func (gt *GT1151) GT_Write(Reg, Data int) {
	i2cWriteByte(Reg, Data) // Replace with your i2cWriteByte implementation
	_, err := gt.i2c.WriteRegU8([]byte{0x1, 0xF3})
	if err != nil {
		log.Fatal(err)
	}

}

func (gt *GT1151) GT_Read(Reg, length int) []byte {
	buffer := make([]byte, length)
	_, err := gt.i2c.ReadBytes(buffer)
	if err != nil {
		log.Fatal(err)
	}
	return buffer
}

func (gt *GT1151) GT_ReadVersion() {
	buf := gt.GT_Read(0x8140, 4)
	fmt.Println(buf)
}

func (gt *GT1151) GT_Init() {
	gt.GT_Reset()
	gt.GT_ReadVersion()
}

func (gt *GT1151) GT_Scan(GT_Dev, GT_Old *GT_Development) {
	buf := make([]int, 0)
	mask := 0x00

	if GT_Dev.Touch == 1 {
		GT_Dev.Touch = 0
		buf = gt.GT_Read(0x814E, 1)

		if buf[0]&0x80 == 0x00 {
			gt.GT_Write(0x814E, mask)
			delayMs(10) // Replace with your delay implementation
		} else {
			GT_Dev.TouchpointFlag = buf[0] & 0x80
			GT_Dev.TouchCount = buf[0] & 0x0f

			if GT_Dev.TouchCount > 5 || GT_Dev.TouchCount < 1 {
				gt.GT_Write(0x814E, mask)
				return
			}

			buf = gt.GT_Read(0x814F, GT_Dev.TouchCount*8)
			gt.GT_Write(0x814E, mask)

			GT_Old.X[0] = GT_Dev.X[0]
			GT_Old.Y[0] = GT_Dev.Y[0]
			GT_Old.S[0] = GT_Dev.S[0]

			for i := 0; i < GT_Dev.TouchCount; i++ {
				GT_Dev.Touchkeytrackid[i] = buf[0+8*i]
				GT_Dev.X[i] = (buf[2+8*i] << 8) + buf[1+8*i]
				GT_Dev.Y[i] = (buf[4+8*i] << 8) + buf[3+8*i]
				GT_Dev.S[i] = (buf[6+8*i] << 8) + buf[5+8*i]
			}

			fmt.Println(GT_Dev.X[0], GT_Dev.Y[0], GT_Dev.S[0])
		}
	}
}

// Helper functions, replace with your own implementations
func digitalRead(pin int) int {
	// Implementation goes here
	return 0
}

func digitalWrite(pin, value int) {
	// Implementation goes here
}

func delayMs(ms int) {
	// Implementation goes here
}

func i2cWriteByte(reg, data int) {
	// Implementation goes here
}

func i2cReadByte(reg, length int) []int {
	// Implementation goes here
	return make([]int, length)
}
