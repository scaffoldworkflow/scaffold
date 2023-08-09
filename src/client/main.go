package main

import (
	"fmt"
	"os"
	"scaffold/client/cascade"
	"scaffold/client/config"
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

	applyCommand := parser.NewCommand("apply", "Create or update a cascade")
	applyProfile := applyCommand.String("p", "profile", &argparse.Options{Help: "Profile to use to connect to Scaffold instance", Default: "default"})
	applyFile := applyCommand.String("f", "file", &argparse.Options{Required: true, Help: "Scaffold manifest to apply"})

	deleteCommand := parser.NewCommand("delete", "Delete an existing cascade")
	deleteProfile := deleteCommand.String("p", "profile", &argparse.Options{Help: "Profile to use to connect to Scaffold instance", Default: "default"})
	deleteName := deleteCommand.String("n", "name", &argparse.Options{Required: true, Help: "Name of the cascade to remove"})

	execCommand := parser.NewCommand("exec", "Exec into a scaffold container")
	execProfile := execCommand.String("p", "profile", &argparse.Options{Help: "Profile to use to connect to Scaffold instance", Default: "default"})

	configCommand := parser.NewCommand("configure", "Configure credentials for a Scaffold instance")
	configHost := configCommand.String("", "host", &argparse.Options{Help: "Hostname for Scaffold instance", Default: "localhost"})
	configPort := configCommand.String("", "port", &argparse.Options{Help: "Port for Scaffold instance", Default: "2997"})
	configWSPort := configCommand.String("", "ws-port", &argparse.Options{Help: "Websocket port for Scaffold instance", Default: "8080"})
	configProtocol := configCommand.String("", "protocol", &argparse.Options{Help: "Protocol to use to connect to Scaffold instance", Default: "http"})
	configProfile := configCommand.String("p", "profile", &argparse.Options{Help: "Name for the profile to configure", Default: "default"})
	configUsername := configCommand.String("", "username", &argparse.Options{Required: true, Help: "Username to use to connect to Scaffold instance"})
	configPassword := configCommand.String("", "password", &argparse.Options{Required: true, Help: "Password to use to connect to Scaffold instance"})

	// Parse input
	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Print(parser.Usage(err))
	}

	if applyCommand.Happened() {
		cascade.DoApply(*applyProfile, *applyFile)
	}

	if deleteCommand.Happened() {
		cascade.DoDelete(*deleteProfile, *deleteName)
	}

	if execCommand.Happened() {
		exec.DoExec(*execProfile)
	}

	if configCommand.Happened() {
		config.DoConfig(*configHost, *configPort, *configProtocol, *configWSPort, *configProfile, *configUsername, *configPassword)
	}
}
