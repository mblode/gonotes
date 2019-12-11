package main

import (
	"log"
	"os"
	"sort"

	"github.com/mblode/gonotes/copy"
	"github.com/mblode/gonotes/server"
	"github.com/urfave/cli/v2"
)

func main() {
	homeDir := copy.HomeDir()

	app := &cli.App{
		Name:    "Go Notes",
		Usage:   "Copy notes and create a web based file viewer",
		Version: "v1.0.0",
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Matthew Blode",
				Email: "m@blode.co",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "src",
				Aliases: []string{"i"},
				Value:   homeDir + "/Library/Mobile Documents/27N4MQEA55~pro~writer/Documents/",
				Usage:   "Input source directory from `FOLDER`",
			},
			&cli.StringFlag{
				Name:    "dest",
				Aliases: []string{"o"},
				Value:   homeDir + "/Google Drive/Backups/Notes",
				Usage:   "Output destination directory from `FOLDER`",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "copy",
				Aliases: []string{"c"},
				Usage:   "Copy folder to a new location",
				Action: func(c *cli.Context) error {
					err := copy.Process(c.String("src"), c.String("dest"))
					if err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:    "serve",
				Aliases: []string{"s"},
				Usage:   "Open a web based file viewer",
				Action: func(c *cli.Context) error {
					err := copy.Process(c.String("src"), c.String("dest"))
					if err != nil {
						return err
					}

					err = server.Process(c.String("dest"), c.String("port"))
					if err != nil {
						return err
					}

					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Value:   "3000",
					},
				},
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
