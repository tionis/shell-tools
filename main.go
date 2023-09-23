package main

import (
	"errors"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"golang.design/x/clipboard"
	"io"
	"log"
	"os"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	//homeDir, err := os.UserHomeDir()
	//if err != nil {
	//	log.Println("failed to get home dir: ", err)
	//	return
	//}
	//hostName, err := os.Hostname()
	//if err != nil {
	//	log.Println("failed to get hostname: ", err)
	//	return
	//}

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "clipboard",
				Aliases: []string{"clip"},
				Usage:   "generic clipboard handling",
				Subcommands: []*cli.Command{
					{
						Name:    "copy",
						Aliases: []string{"c"},
						Usage:   "copy to clipboard",
						Action: func(c *cli.Context) error {
							all, err := io.ReadAll(os.Stdin)
							if err != nil {
								return err
							}
							clipboard.Write(clipboard.FmtText, all)
							return nil
						},
					},
					{
						Name:    "paste",
						Aliases: []string{"p"},
						Usage:   "paste from clipboard",
						Action: func(c *cli.Context) error {
							result := clipboard.Read(clipboard.FmtText)
							_, err := os.Stdout.Write(result)
							if err != nil {
								return err
							}
							return nil
						},
					},
				},
			},
			{
				Name:    "util",
				Aliases: []string{"u"},
				Usage:   "some general utils",
				Subcommands: []*cli.Command{
					{
						Name:    "uuid",
						Aliases: []string{"u"},
						Usage:   "generate uuid",
						Action: func(c *cli.Context) error {
							uuid := uuid.New().String()
							_, err := os.Stdout.Write([]byte(uuid))
							if err != nil {
								return err
							}
							return nil
						},
					},
					{
						Name:    "hostname",
						Aliases: []string{"h"},
						Usage:   "get hostname",
						Action: func(c *cli.Context) error {
							hostname, err := os.Hostname()
							if err != nil {
								return err
							}
							_, err = os.Stdout.Write([]byte(hostname))
							if err != nil {
								return err
							}
							return nil
						},
					},
					{
						Name:    "home",
						Aliases: []string{"h"},
						Usage:   "get home dir",
						Action: func(c *cli.Context) error {
							homeDir, err := os.UserHomeDir()
							if err != nil {
								return err
							}
							_, err = os.Stdout.Write([]byte(homeDir))
							if err != nil {
								return err
							}
							return nil
						},
					},
					{
						Name:    "sponge",
						Aliases: []string{"s"},
						Usage:   "soak all input from stdin and write to file/stdin",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "file",
								Aliases: []string{"f"},
								Usage:   "file to write to",
							},
						},
						Action: func(c *cli.Context) error {
							all, err := io.ReadAll(os.Stdin)
							if err != nil {
								return err
							}
							if c.String("file") != "" {
								err = os.WriteFile(c.String("file"), all, 0644)
							} else {
								_, err = os.Stdout.Write(all)
							}
							return err
						},
					},
					{
						Name:    "ts",
						Aliases: []string{"t"},
						Usage:   "add timestamps to lines from stdin",
						Action: func(c *cli.Context) error {
							// TODO
							return errors.New("not implemented")
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("failed to run app: %v+", err)
	}
}
