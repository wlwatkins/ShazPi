package commands

import (
	"fmt"
	"log"
	"shazammini/src/structs"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
	"gobot.io/x/gobot"
)

func CheckTouch(gt *GT1151, dev *Development) {
	for {
		if gt.INT.Read() == rpio.Low {
			dev.Touch = 1
		} else {
			dev.Touch = 0
		}
	}
}

func run(commChannels *structs.CommChannels) {

	gt := NewGT1151()
	defer gt.Kill()

	GT_Dev := Development{}
	GT_Old := Development{}

	GT_Dev.Init()
	GT_Old.Init()

	gt.Reset()
	gt.ReadVersion()

	go CheckTouch(&gt, &GT_Dev)

	for {

		select {
		case <-commChannels.TouchEnabled:
			log.Println("Touch enabled")
			for {
				gt.Scan(&GT_Dev, &GT_Old)

				if GT_Old.X[0] == GT_Dev.X[0] && GT_Old.Y[0] == GT_Dev.Y[0] && GT_Old.S[0] == GT_Dev.S[0] {
					time.Sleep(20 * time.Millisecond)
					continue
				}
				fmt.Println(GT_Dev.X, GT_Old.X, GT_Dev.Y, GT_Old.Y, GT_Dev.S, GT_Old.S)

				commChannels.DisplayRecord <- true
				commChannels.RecordChannel <- time.Second * 5
				break

			}
		}
	}

}

func Commands(commChannels *structs.CommChannels) *gobot.Robot {
	work := func() {
		run(commChannels)
	}

	robot := gobot.NewRobot("commands",
		work,
	)

	return robot

}
