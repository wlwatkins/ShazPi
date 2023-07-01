package commands

import (
	"fmt"
	"log"
	"shazammini/src/io"
	"shazammini/src/utils"
	"time"

	"github.com/d2r2/go-i2c"
)

const Address = 0x14

type Development struct {
	Touch           int
	TouchpointFlag  int
	TouchCount      int
	Touchkeytrackid [5]int
	X               [5]int
	Y               [5]int
	S               [5]int
}

func (gt *Development) Init() {
	gt.Touch = 0
	gt.TouchpointFlag = 0
	gt.TouchCount = 0
	gt.Touchkeytrackid = [5]int{0, 1, 2, 3, 4}
	gt.X = [5]int{0, 1, 2, 3, 4}
	gt.Y = [5]int{0, 1, 2, 3, 4}
	gt.S = [5]int{0, 1, 2, 3, 4}
}

type GT1151 struct {
	i2c  *i2c.I2C
	TRST io.WriteablePin
	INT  io.ReadablePin
}

func (gt *GT1151) New() {
	// Create new connection to I2C bus on 2 line with address 0x27

	i2c, err := i2c.NewI2C(Address, 1)
	if err != nil {
		log.Fatal(err)
	}
	gt.i2c = i2c
	gt.TRST = io.GetWritePin(io.TRST_PIN)
	gt.INT = io.GetReadPin(io.INT_PIN)

	gt.Reset()
	gt.ReadVersion()
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

func (gt *GT1151) WriteData(Reg, Data int) {
	err := gt.i2c.WriteRegU8(uint8((Reg>>8)&0xFF), uint8((Reg&0xFF)|((Data&0xFF)<<8)))
	if err != nil {
		log.Fatal(err)
	}
}

func (gt *GT1151) Write(Reg int) {
	err := gt.i2c.WriteRegU8(uint8((Reg>>8)&0xFF), uint8(Reg&0xFF))
	if err != nil {
		log.Fatal(err)
	}
}

func (gt *GT1151) Read(Reg, length int) []int {
	gt.Write(Reg)
	buffer := make([]byte, length)
	_, err := gt.i2c.ReadBytes(buffer)
	if err != nil {
		log.Fatal(err)
	}

	return utils.ByteSliceToIntSlice(buffer)
}

func (gt *GT1151) ReadVersion() {
	buffer := gt.Read(0x8140, 4)
	buf_byte := utils.IntSliceToByteSlice(buffer)
	fmt.Printf("Version: %s\n", string(buf_byte))
}

func (gt *GT1151) Scan(Dev, Old *Development) {
	buf := make([]int, 0)
	mask := 0x00

	// if rpio.Pin(io.INT_PIN).Read() == 0 {
	// 	Dev.Touch = 1
	// } else {
	// 	Dev.Touch = 0
	// }
	fmt.Println(gt.INT.Read())
	// if Dev.Touch == 1 {
	if false {
		Dev.Touch = 0
		buf = gt.Read(0x814E, 1)
		if buf[0]&0x80 == 0x00 {
			gt.WriteData(0x814E, mask)
			time.Sleep(10 * time.Millisecond)
		} else {
			Dev.TouchpointFlag = buf[0] & 0x80
			Dev.TouchCount = buf[0] & 0x0f

			if Dev.TouchCount > 5 || Dev.TouchCount < 1 {
				gt.WriteData(0x814E, mask)
				return
			}

			buf = gt.Read(0x814F, Dev.TouchCount*8)
			gt.WriteData(0x814E, mask)

			Old.X[0] = Dev.X[0]
			Old.Y[0] = Dev.Y[0]
			Old.S[0] = Dev.S[0]

			for i := 0; i < Dev.TouchCount; i++ {
				Dev.Touchkeytrackid[i] = buf[0+8*i]
				Dev.X[i] = (buf[2+8*i] << 8) + buf[1+8*i]
				Dev.Y[i] = (buf[4+8*i] << 8) + buf[3+8*i]
				Dev.S[i] = (buf[6+8*i] << 8) + buf[5+8*i]
			}

			// fmt.Println(Dev.X[0], Dev.Y[0], Dev.S[0])
		}
	}
}
