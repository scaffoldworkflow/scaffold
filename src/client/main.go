package main

import (
	"fmt"
	"os"
	"scaffold/client/apply"
	"scaffold/client/exec"
	"scaffold/client/trigger"

	"github.com/akamensky/argparse"
)

func main() {
	parser := argparse.NewParser("scaffold", "Scaffold infrastructure management client")

	applyCommand := parser.NewCommand("apply", "Apply a Scaffold manifest file against a Scaffold instance")
	applyHost := applyCommand.String("H", "host", &argparse.Options{Help: "Hostname for Scaffold instance", Default: "localhost"})
	applyPort := applyCommand.String("p", "port", &argparse.Options{Help: "Port for Scaffold instance", Default: "2997"})
	applyFile := applyCommand.String("f", "file", &argparse.Options{Required: true, Help: "Scaffold manifest to apply"})

	triggerCommand := parser.NewCommand("trigger", "Trigger a Scaffold manifest")
	triggerHost := triggerCommand.String("H", "host", &argparse.Options{Help: "Hostname for Scaffold instance", Default: "localhost"})
	triggerPort := triggerCommand.String("p", "port", &argparse.Options{Help: "Port for Scaffold instance", Default: "2997"})
	triggerName := triggerCommand.String("n", "name", &argparse.Options{Required: true, Help: "Scaffold manifest to apply"})

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
		apply.DoApply(*applyHost, *applyPort, *applyFile)
	}

	if triggerCommand.Happened() {
		trigger.DoTrigger(*triggerHost, *triggerPort, *triggerName)
	}

	if execCommand.Happened() {
		exec.DoExec(*execHost, *execPort, *execWSPort)
	}
}
