package commands

import (
	"fmt"
	"os"
	"shazammini/src/structs"

	"gobot.io/x/gobot"
)

func run(commChannels *structs.CommChannels) {

	for {
		var input string
		fmt.Print("Enter 'p' for play or 'r' for record: ")
		_, err := fmt.Scan(&input)
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}
		switch input {
		case "p":
			commChannels.PlayChannel <- true
		case "r":
			commChannels.RecordChannel <- true
		case "q":
			os.Exit(0)
		}

		if input == "p" || input == "r" {
			fmt.Println("Valid input:", input)
		} else {
			fmt.Println("Invalid input. Please try again.")
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
