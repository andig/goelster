package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/brutella/can"
	"github.com/urfave/cli"

	. "github.com/andig/goelster"
)

type Command int

const (
	dump Command = iota
	scan
	read
	write
)

var command Command
var device string
var sender uint16
var receiver uint16
var register uint16
var value uint16
var numeric float64

func main() {
	app := cli.NewApp()
	app.HideVersion = true
	app.Name = "goelster"
	app.Usage = "CAN bus interface for Elster/Kromschröder devices"

	// fmt.Println(cli.AppHelpTemplate)
	cli.AppHelpTemplate = `
NAME:
   {{.Name}}{{if .Usage}} - {{.Usage}}{{end}}

USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .Version}}{{if not .HideVersion}}

VERSION:
   {{.Version}}{{end}}{{end}}{{if .Description}}

DESCRIPTION:
   {{.Description}}{{end}}{{if len .Authors}}

AUTHOR{{with $length := len .Authors}}{{if ne 1 $length}}S{{end}}{{end}}:
   {{range $index, $author := .Authors}}{{if $index}}
   {{end}}{{$author}}{{end}}{{end}}{{if .VisibleCommands}}

OPTIONS:
   {{range $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}

EXAMPLES:

	dump traffic:    goelster slcan0
	scan device:     goelster slcan0 680 180
	read register:   goelster slcan0 680 180.0013
	write register:  goelster slcan0 680 180.0013.01a4
	numeric write:   goelster slcan0 680 180.0013 42.1 (NOT IMPLEMENTED YET)
{{if .Copyright}}
COPYRIGHT:
   {{.Copyright}}{{end}}
`
	cli.CommandHelpTemplate = cli.AppHelpTemplate
	cli.SubcommandHelpTemplate = cli.AppHelpTemplate

	app.UsageText = `goelster [options] [can device] [sender id] [receiver id][.register][.raw value] [numeric value]`

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "verbose mode",
		},
	}

	app.Action = func(c *cli.Context) error {
		if c.NArg() < 1 || c.NArg() > 3 {
			cli.ShowCommandHelp(c, "")
			return nil
		}

		command = dump
		device = c.Args().Get(0)

		if c.NArg() > 1 {
			if c.NArg() != 3 {
				cli.ShowCommandHelp(c, "")
				return nil
			}

			RawLog = c.Bool("verbose")

			command = scan
			s, err := strconv.ParseUint(c.Args().Get(1), 16, 16)
			if err != nil {
				fmt.Printf("Could not parse sender id %s", c.Args().Get(1))
				return nil
			} else {
				sender = uint16(s)
			}

			a := strings.Split(c.Args().Get(2), ".")
			if len(a) > 3 {
				cli.ShowCommandHelp(c, "")
				return nil
			}

			rcvr, err := strconv.ParseUint(a[0], 16, 16)
			if err != nil {
				fmt.Printf("Could not parse hex receiver id '%s'", a[0])
				return nil
			} else {
				receiver = uint16(rcvr)
			}

			if len(a) > 1 {
				command = read
				reg, err := strconv.ParseUint(a[1], 16, 16)
				if err != nil {
					fmt.Printf("Could not parse hex register id '%s'", a[1])
					return nil
				} else {
					register = uint16(reg)
				}
			}

			if len(a) > 2 {
				command = write
				val, err := strconv.ParseUint(a[2], 16, 16)
				if err != nil {
					fmt.Printf("Could not parse hex value '%s'", a[2])
					return nil
				} else {
					value = uint16(val)
				}
			}
		}

		bus, err := can.NewBusForInterfaceWithName(device)
		if err != nil {
			log.Fatal(err)
		}

		quit := make(chan os.Signal)
		signal.Notify(quit, os.Interrupt)
		signal.Notify(quit, os.Kill)

		go func() {
			select {
			case <-quit:
				bus.Disconnect()
				os.Exit(1)
			}
		}()

		switch command {
		case dump:
			CanDump(bus)
		case scan:
			CanScan(bus, sender, receiver)
		case read:
			CanRead(bus, sender, receiver, register)
		case write:
			CanWrite(bus, sender, receiver, register, value)
		}

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
