package commands

import (
	"shazammini/src/structs"
	"time"

	"gobot.io/x/gobot"
)

func run(commChannels *structs.CommChannels) {

	gt := GT1151{
		TRST: 22,
		INT:  27,
	}
	gt.GT_Init()

	GT_Dev := GT_Development{}
	GT_Old := GT_Development{}

	for {
		gt.GT_Scan(&GT_Dev, &GT_Old)
		if GT_Old.X[0] == GT_Dev.X[0] && GT_Old.Y[0] == GT_Dev.Y[0] && GT_Old.S[0] == GT_Dev.S[0] {
			time.Sleep(20 * time.Millisecond)
			continue
		}

		if GT_Dev.TouchpointFlag {
			GT_Dev.TouchpointFlag = 0

		}

		// var input string
		// fmt.Print("Enter 'p' for play or 'r' for record: ")
		// _, err := fmt.Scan(&input)
		// if err != nil {
		// 	fmt.Println("Error reading input:", err)
		// 	continue
		// }
		// switch input {
		// case "p":
		// 	commChannels.PlayChannel <- true
		// case "r":
		// 	commChannels.DisplayRecord <- true
		// 	commChannels.RecordChannel <- time.Second * 5
		// case "q":
		// 	os.Exit(0)
		// }

		// if input == "p" || input == "r" || input == "s" {
		// 	fmt.Println("Valid input:", input)
		// } else {
		// 	fmt.Println("Invalid input. Please try again.")
		// }

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
