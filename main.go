package main

import (
	"flag"
	"github.com/mblode/gonotes/copy"
	"github.com/mblode/gonotes/server"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	homeDir := copy.HomeDir()
	srcDir := homeDir + "/Library/Mobile Documents/27N4MQEA55~pro~writer/Documents/"
	destDir := homeDir + "/Google Drive/Backups/Notes"

	// srcDir := homeDir + "/Google Drive/Backups/Notes"
	// destDir := homeDir + "/Google Drive/Backups/Notes2"

	src := flag.String("src", srcDir, "The folder where the files should be copied from")
	dest := flag.String("dest", destDir, "The folder that the files should be copied to")
	flag.Parse()

	err := copy.Directory(*src, *dest)
	check(err)

	err = server.Process(*dest)
	check(err)
}
