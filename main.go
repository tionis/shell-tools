package main

import (
	"context"
	"encoding/json"
	"errors"
	"filippo.io/age"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/urfave/cli/v2"
	"golang.design/x/clipboard"
	"io"
	"log"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
	"os"
	"os/exec"
	"strings"
)

type quickCommand struct {
	Description string `json:"description"`
	Command     string `json:"command"`
}

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
				Name:  "entr",
				Usage: "file watcher that executes a command on changes",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:    "path",
						Aliases: []string{"p"},
						Usage:   "a path to watch non-recursively",
					},
					&cli.StringFlag{
						Name:  "this",
						Usage: "just watch working dir, nothing else",
					},
					&cli.StringSliceFlag{
						Name:    "recursive",
						Aliases: []string{"r"},
						Usage:   "a path to watch recursively",
					},
				},
				UsageText: "entr [global options] command args",
				Action: func(c *cli.Context) error {
					watcher, err := fsnotify.NewWatcher()
					if err != nil {
						return fmt.Errorf("failed to create watcher: %w", err)
					}
					if c.String("this") != "" {
						for _, dir := range c.StringSlice("dir") {
							log.Printf("watching dir: %s\n", dir)
							err = watcher.Add(dir)
							if err != nil {
								return fmt.Errorf("failed to add dir to watcher: %w", err)
							}
						}
						for _, dir := range c.StringSlice("path") {
							log.Printf("watching path: %s\n", dir)
							// TODO watch recursively
							// this needs some plumping to work
							// needs to track watched paths and update that list on changes if necessary
							err = errors.New("recursive not implemented")
							if err != nil {
								return fmt.Errorf("failed to add path to watcher: %w", err)
							}
						}
					} else {
						workingDir, err := os.Getwd()
						if err != nil {
							return fmt.Errorf("failed to get working dir: %w", err)
						}
						log.Printf("watching working dir (%s)\n", workingDir)
						// TODO watch recursively
						err = watcher.Add(workingDir)
						if err != nil {
							return fmt.Errorf("failed to add dir to watcher: %w", err)
						}
					}
					command := c.Args().Slice()
					for {
						select {
						case event, ok := <-watcher.Events:
							if !ok {
								return errors.New("watcher closed")
							}
							// TODO implement cooldown period?
							log.Printf("event: %s\n", event.String())
							err := exec.Command(command[0], command[1:]...).Run()
							if err != nil {
								return err
							}
						case err, ok := <-watcher.Errors:
							if !ok {
								return errors.New("watcher closed")
							}
							log.Println("error:", err)
						}
					}
				},
			},
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
				Name:    "quick",
				Aliases: []string{"q"},
				Flags: []cli.Flag{
					&cli.PathFlag{
						Name:      "file",
						Aliases:   []string{"f"},
						Usage:     "file to read from",
						TakesFile: true,
					},
				},
				Usage: "quick-commands, small ui to quickly execute some commands\n" +
					"reads json config from stdin (or file) with the keys as command" +
					" names and the values as a tuple of description and command",
				Action: func(c *cli.Context) error {
					var all []byte
					var err error
					if c.String("file") != "" {
						all, err = os.ReadFile(c.String("file"))
						if err != nil {
							return fmt.Errorf("failed to read file: %w", err)
						}
					} else {
						all, err = io.ReadAll(os.Stdin)
						if err != nil {
							return fmt.Errorf("failed to read stdin: %w", err)
						}
					}
					var quickConfig map[string]quickCommand
					err = json.Unmarshal(all, &quickConfig)
					if err != nil {
						return fmt.Errorf("failed to unmarshal json: %w", err)
					}
					commandKeys := keys(quickConfig)
					selected, err := fuzzyfinder.Find(commandKeys, func(i int) string {
						return commandKeys[i]
					}, fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
						return quickConfig[commandKeys[i]].Description
					}))
					if err != nil {
						if err == fuzzyfinder.ErrAbort {
							return nil
						}
						return fmt.Errorf("failed to find command: %w", err)
					}
					quickCommand := quickConfig[commandKeys[selected]]
					file, _ := syntax.NewParser().Parse(strings.NewReader(quickCommand.Command), commandKeys[selected])
					runner, _ := interp.New(
						interp.Env(expand.ListEnviron(os.Environ()...)),
						interp.StdIO(nil, os.Stdout, os.Stdout),
					)
					err = runner.Run(context.TODO(), file)
					if err != nil {
						return fmt.Errorf("failed to run command: %w", err)
					}
					return nil
				},
			},
			{
				Name:    "pass",
				Aliases: []string{"p"},
				Usage:   "password manager",
				Subcommands: []*cli.Command{
					{
						Name: "decrypt",
						Flags: []cli.Flag{
							&cli.PathFlag{
								Name:      "file",
								Aliases:   []string{"f"},
								Usage:     "file to decrypt",
								TakesFile: true,
								Required:  true,
							},
						},
						Usage: "decrypt file",
						Action: func(c *cli.Context) error {
							file, err := os.Open(c.String("file"))
							if err != nil {
								return err
							}
							defer func(file *os.File) {
								err := file.Close()
								if err != nil {
									log.Printf("failed to close file: %v+", err)
								}
							}(file)
							idReader := strings.NewReader("AGE-PLUGIN-YUBIKEY-1JAAZ6QVZ0RQLA0GD5ZEPL\n" +
								"AGE-PLUGIN-YUBIKEY-1ZJRRUQVZE9DJFCC3UPNJD\n" +
								"AGE-PLUGIN-YUBIKEY-1ZWRRUQVZAJX2Q4C5LJRTD") // TODO load identities from file
							identities, err := age.ParseIdentities(idReader)
							if err != nil {
								return err
							}
							decrypt, err := age.Decrypt(file, identities...)
							if err != nil {
								return err
							}
							all, err := io.ReadAll(decrypt)
							if err != nil {
								return err
							}
							_, err = os.Stdout.Write(all)
							if err != nil {
								return err
							}
							return nil
						},
					},
					{
						Name:    "ui",
						Aliases: []string{"u"},
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:    "fuzzy",
								Aliases: []string{"f"},
								Usage:   "fuzzy UI", // NOTE this does nothing for now
								Value:   true,
							},
						},
						Usage: "start small fuzzy UI mostly to copy passwords out",
						Action: func(c *cli.Context) error {
							slice := []string{"one", "two", "three", os.Getenv("EDITOR")}
							find, err := fuzzyfinder.Find(slice, func(i int) string {
								return slice[i]
							})
							if err != nil {
								return err
							}
							fmt.Println(find)
							return errors.New("not implemented")
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
							tmpUUID := uuid.New().String()
							_, err := os.Stdout.Write([]byte(tmpUUID))
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
					// TODO add following commands:
					// - base64
					// - ifne
					// - git-root
					// - git-skm
					// - git interactive sparse clone
					// - git interactive sparse checkout
					// - combine
					// - chronic
					// - sponge
					// - gron and other json processing tools
					// - ssh-proxy (for huproxy)
					// - ts
					// - xargs like
					// - vidir
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("failed to run app: %v+", err)
	}
}

func keys(m map[string]quickCommand) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}
