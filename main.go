package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/nullstone-io/module/config"
	"github.com/urfave/cli"
)

func main() {
	var dir string
	app := &cli.App{
		Commands: []cli.Command{
			{
				Name:        "manifest",
				Description: `This generates a nullstone module manifest from a directory of terraform files.`,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:        "dir",
						Usage:       "Directory to scan. By default, uses current directory.",
						Value:       ".",
						Destination: &dir,
					},
				},
				Action: func(c *cli.Context) error {
					files, err := config.ReadDir(dir)
					if err != nil {
						return err
					}
					cfg, err := config.Parse(files)
					if err != nil {
						return err
					}

					manifest := cfg.ToManifest()
					raw, _ := json.MarshalIndent(manifest, "", "  ")
					fmt.Println(string(raw))
					return nil
				},
			},
		},
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)
}
