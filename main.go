package main

import (
	"log"
	"os"
	"sort"

	"github.com/mblode/gonotes/commands"

)

func main() {
	response := commands.Execute(os.Args[1:])

	if response.Err != nil {
		if response.IsUserError() {
			response.Cmd.Println("")
			response.Cmd.Println(response.Cmd.UsageString())
		}
		os.Exit(-1)
	}
}
