package main

import (
	"fmt"
	"os"
	"scaffold/client/cascade"
	"scaffold/client/constants"
	"scaffold/client/exec"
	"scaffold/client/logger"

	"github.com/akamensky/argparse"
)

func main() {
	// config.LoadConfig()
	// logger.SetLevel(config.Config.LogLevel)
	logger.SetLevel(constants.LOG_LEVEL_DEBUG)

	parser := argparse.NewParser("scaffold", "Scaffold infrastructure management client")

	cascadeCommand := parser.NewCommand("cascade", "Manage scaffold cascades")

	applyCommand := cascadeCommand.NewCommand("apply", "Create or update a cascade")
	applyHost := applyCommand.String("H", "host", &argparse.Options{Help: "Hostname for Scaffold instance", Default: "localhost"})
	applyPort := applyCommand.String("p", "port", &argparse.Options{Help: "Port for Scaffold instance", Default: "2997"})
	applyFile := applyCommand.String("f", "file", &argparse.Options{Required: true, Help: "Scaffold manifest to apply"})

	deleteCommand := cascadeCommand.NewCommand("delete", "Delete an existing cascade")
	deleteHost := deleteCommand.String("H", "host", &argparse.Options{Help: "Hostname for Scaffold instance", Default: "localhost"})
	deletePort := deleteCommand.String("p", "port", &argparse.Options{Help: "Port for Scaffold instance", Default: "2997"})
	deleteName := deleteCommand.String("n", "name", &argparse.Options{Required: true, Help: "Name of the cascade to remove"})

	execCommand := parser.NewCommand("exec", "Exec into a scaffold container")
	execHost := execCommand.String("H", "host", &argparse.Options{Help: "Hostname for Scaffold instance", Default: "localhost"})
	execPort := execCommand.String("p", "port", &argparse.Options{Help: "Port for Scaffold instance", Default: "2997"})
	execWSPort := execCommand.String("w", "ws-port", &argparse.Options{Help: "Websocket port for Scaffold instance", Default: "8080"})

	// Parse input
	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Print(parser.Usage(err))
	}

	if applyCommand.Happened() {
		cascade.DoApply(*applyHost, *applyPort, *applyFile)
	}

	if deleteCommand.Happened() {
		cascade.DoDelete(*deleteHost, *deletePort, *deleteName)
	}

	if execCommand.Happened() {
		exec.DoExec(*execHost, *execPort, *execWSPort)
	}
}
