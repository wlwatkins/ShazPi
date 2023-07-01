package commands

import (
	"fmt"
	"shazammini/src/structs"

	"gobot.io/x/gobot"
)

func run(commChannels *structs.CommChannels) {

	gt := GT1151{}
	gt.New()
	defer gt.Kill()

	GT_Dev := Development{}
	GT_Old := Development{}

	GT_Dev.Init()
	GT_Old.Init()

	for {

		gt.Scan(&GT_Dev, &GT_Old)
		// fmt.Println(GT_Dev.X, GT_Dev.Y, GT_Dev.S)
		if GT_Old.X[0] == GT_Dev.X[0] && GT_Old.Y[0] == GT_Dev.Y[0] && GT_Old.S[0] == GT_Dev.S[0] {
			// time.Sleep(20 * time.Millisecond)
			continue
		}

		if GT_Dev.TouchpointFlag > 0 {
			GT_Dev.TouchpointFlag = 0
			fmt.Println(GT_Dev)
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
