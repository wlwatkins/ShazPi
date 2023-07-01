package commands

import (
	"fmt"
	"log"
	"shazammini/src/io"
	"shazammini/src/utils"
	"time"

	"golang.org/x/exp/io/i2c"
)

const Address = 0x14

type Development struct {
	Touch           int
	TouchpointFlag  uint16
	TouchCount      uint16
	Touchkeytrackid [5]uint16
	X               [5]uint16
	Y               [5]uint16
	S               [5]uint16
}

func (gt *Development) Init() {
	gt.Touch = 0
	gt.TouchpointFlag = 0
	gt.TouchCount = 0
	gt.Touchkeytrackid = [5]uint16{0, 1, 2, 3, 4}
	gt.X = [5]uint16{0, 1, 2, 3, 4}
	gt.Y = [5]uint16{0, 1, 2, 3, 4}
	gt.S = [5]uint16{0, 1, 2, 3, 4}
}

type GT1151 struct {
	i2c  *i2c.Device
	TRST io.WriteablePin
	INT  io.ReadablePin
}

func NewGT1151() GT1151 {
	// Create new connection to I2C bus on 2 line with address 0x27

	i2cCon, err := i2c.Open(&i2c.Devfs{Dev: "/dev/i2c-1"}, Address)
	if err != nil {
		panic(err)
	}
	gt := GT1151{}
	gt.TRST = io.GetWritePin(io.TRST_PIN)
	gt.INT = io.GetReadPin(io.INT_PIN)
	gt.i2c = i2cCon

	return gt
}

func (gt *GT1151) Kill() {
	gt.i2c.Close()
}

func (gt *GT1151) Reset() {
	gt.TRST.High()
	time.Sleep(100 * time.Millisecond)
	gt.TRST.Low()
	time.Sleep(100 * time.Millisecond)
	gt.TRST.High()
	time.Sleep(100 * time.Millisecond)
}

func (gt *GT1151) WriteData(Reg, Data uint16) {
	regBytes := []byte{byte((Reg >> 8) & 0xFF), byte(Reg & 0xFF)}
	dataBytes := []byte{byte((Data >> 8) & 0xFF), byte(Data & 0xFF)}

	err := gt.i2c.WriteReg(regBytes[0], append(regBytes[1:], dataBytes...))
	if err != nil {
		log.Fatal(err)
	}

}

func (gt *GT1151) Write(Reg uint16) {
	regBytes := []byte{byte((Reg >> 8) & 0xFF), byte(Reg & 0xFF)}

	err := gt.i2c.WriteReg(regBytes[0], regBytes[1:])
	if err != nil {
		log.Fatal(err)
	}
}

func (gt *GT1151) Read(Reg, length uint16) []uint16 {
	gt.Write(Reg)
	buffer := make([]byte, length)
	err := gt.i2c.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}

	return utils.ByteSliceToUint16Slice(buffer)
}

func (gt *GT1151) ReadVersion() {
	buffer := gt.Read(0x8140, 4)
	buf_byte := utils.Uint16SliceToByteSlice(buffer)
	fmt.Printf("Version: %s\n", string(buf_byte))
}

func (gt *GT1151) Scan(Dev, Old *Development) {
	var mask uint16 = 0x00

	if Dev.Touch == 1 {

		Dev.Touch = 0
		buf := gt.Read(0x814E, 1)

		if buf[0]&0x80 == 0x00 {
			gt.WriteData(0x814E, mask)
			time.Sleep(10 * time.Millisecond)
		} else {
			Dev.TouchpointFlag = buf[0] & 0x80
			Dev.TouchCount = buf[0] & 0x0F

			if Dev.TouchCount > 5 || Dev.TouchCount < 1 {
				gt.WriteData(0x814E, mask)
				return
			}
			buf = gt.Read(0x814F, Dev.TouchCount*8)
			gt.WriteData(0x814E, mask)

			Old.X[0] = Dev.X[0]
			Old.Y[0] = Dev.Y[0]
			Old.S[0] = Dev.S[0]

			for i := uint16(0); i < Dev.TouchCount; i++ {
				Dev.Touchkeytrackid[i] = buf[0+8*i]
				Dev.X[i] = (buf[2+8*i] << 8) + buf[1+8*i]
				Dev.Y[i] = (buf[4+8*i] << 8) + buf[3+8*i]
				Dev.S[i] = (buf[6+8*i] << 8) + buf[5+8*i]
			}

			fmt.Println(Dev.X[0], Dev.Y[0], Dev.S[0])
		}
	}
}
