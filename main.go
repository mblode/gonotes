package main

import (
	"flag"
	"fmt"

	"github.com/mblode/gonotes/copy"
	"github.com/mblode/gonotes/server"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	numberOfDocs := 0
	homeDir := copy.HomeDir()
	srcDir := homeDir + "/Library/Mobile Documents/27N4MQEA55~pro~writer/Documents/"
	destDir := homeDir + "/Google Drive/Backups/Notes"

	// srcDir := homeDir + "/Downloads/Notes"
	// destDir := homeDir + "/Downloads/Notes2"

	src := flag.String("src", srcDir, "The folder where the files should be copied from")
	dest := flag.String("dest", destDir, "The folder that the files should be copied to")
	flag.Parse()

	err := copy.Directory(*src, *dest, &numberOfDocs)
	check(err)

	fmt.Println(numberOfDocs)

	err = server.Process(*dest)
	check(err)
}
