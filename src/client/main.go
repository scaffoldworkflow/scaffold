package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"scaffold/client/apply"
	"scaffold/client/config"
	"scaffold/client/constants"
	"scaffold/client/delete"
	"scaffold/client/get"
	"scaffold/client/version"

	"github.com/akamensky/argparse"

	logger "github.com/jfcarter2358/go-logger"
)

/*
NOTE:
This code is pretty awful, I really need to clean it up significantly
*/
func main() {
	// config.LoadConfig()
	// logger.SetLevel(config.Config.LogLevel)
	logger.SetLevel(constants.LOG_LEVEL_DEBUG)

	parser := argparse.NewParser("scaffold", "Scaffold infrastructure management CLI")

	applyCommand := parser.NewCommand("apply", "Create or update a Scaffold object")
	// applyObject := applyCommand.StringPositional(&argparse.Options{Required: true, Help: fmt.Sprintf("Scaffold object type to create. Valid object types are '%s'", strings.Join(apply.ValidObjects, "', '"))})
	applyProfile := applyCommand.String("p", "profile", &argparse.Options{Help: "Profile to use to connect to Scaffold instance", Default: "default"})
	applyFile := applyCommand.String("f", "file", &argparse.Options{Required: true, Help: "Scaffold manifest to apply"})
	applyLogLevel := applyCommand.Selector("l", "log-level", []string{"NONE", "FATAL", "SUCCESS", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}, &argparse.Options{Help: "Log level to use. Valid options are 'NONE', 'FATAL', 'SUCCESS', 'ERROR', 'WARN', 'INFO', 'DEBUG', 'TRACE'. Defaults to 'ERROR'", Default: "ERROR"})

	deleteCommand := parser.NewCommand("delete", "Delete an existing Scaffold object")
	deleteObject := deleteCommand.StringPositional(&argparse.Options{Required: true, Help: "Scaffold object to get. Can be of format '<object type>', or '<object type>/<object name>'. Valid object types are 'datastore', 'file', 'runbook', 'state', 'task', 'user', and 'workflow'"})
	// deleteContext := deleteCommand.String("c", "context", &argparse.Options{Help: "Workflow context to use. If not set the value in your config file will be pulled", Default: ""})
	deleteFilter := deleteCommand.StringList("f", "filter", &argparse.Options{Help: "Filter to use for deletion, e.g. '-f name=foo', can be repeated for multiple filters"})
	deleteProfile := deleteCommand.String("p", "profile", &argparse.Options{Help: "Profile to use to connect to Scaffold instance", Default: "default"})
	deleteLogLevel := deleteCommand.Selector("l", "log-level", []string{"NONE", "FATAL", "SUCCESS", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}, &argparse.Options{Help: "Log level to use. Valid options are 'NONE', 'FATAL', 'SUCCESS', 'ERROR', 'WARN', 'INFO', 'DEBUG', 'TRACE'. Defaults to 'ERROR'", Default: "ERROR"})

	getCommand := parser.NewCommand("get", "Get Scaffold objects")
	getObject := getCommand.StringPositional(&argparse.Options{Required: true, Help: "Scaffold object to get. Can be of format '<object type>', or '<object type>/<object name>'. Valid object types are 'datastore', 'file', 'runbook', 'state', 'task', 'user', and 'workflow'"})
	getContext := getCommand.String("c", "context", &argparse.Options{Help: "Workflow context to use. If not set the value in your config file will be pulled", Default: ""})
	getProfile := getCommand.String("p", "profile", &argparse.Options{Help: "Profile to use to connect to Scaffold instance", Default: "default"})
	getLogLevel := getCommand.Selector("l", "log-level", []string{"NONE", "FATAL", "SUCCESS", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}, &argparse.Options{Help: "Log level to use. Valid options are 'NONE', 'FATAL', 'SUCCESS', 'ERROR', 'WARN', 'INFO', 'DEBUG', 'TRACE'. Defaults to 'ERROR'", Default: "ERROR"})

	// describeCommand := parser.NewCommand("describe", "Describe a Scaffold object")
	// describeObject := describeCommand.StringPositional(&argparse.Options{Required: true, Help: "Scaffold object to describe. Must be of format '<object type>/<object name>'. Valid object types are 'datastore', 'file', 'runbook', 'state', 'task', 'user', and 'workflow'"})
	// describeProfile := describeCommand.String("p", "profile", &argparse.Options{Help: "Profile to use to connect to Scaffold instance", Default: "default"})
	// describeContext := describeCommand.String("c", "context", &argparse.Options{Help: "Workflow context to use. If not set the value in your config file will be pulled", Default: ""})
	// describeFormat := describeCommand.Selector("o", "output", []string{"yaml", "json"}, &argparse.Options{Help: "Output format to print. Valid options are 'yaml' and 'json'. Defaults to 'yaml'", Default: "yaml"})
	// describeLogLevel := describeCommand.Selector("l", "log-level", []string{"NONE", "FATAL", "SUCCESS", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}, &argparse.Options{Help: "Log level to use. Valid options are 'NONE', 'FATAL', 'SUCCESS', 'ERROR', 'WARN', 'INFO', 'DEBUG', 'TRACE'. Defaults to 'ERROR'", Default: "ERROR"})

	configCommand := parser.NewCommand("configure", "Configure credentials for a Scaffold instance")
	configHost := configCommand.String("", "host", &argparse.Options{Help: "Hostname for Scaffold instance", Default: "localhost"})
	configPort := configCommand.String("", "port", &argparse.Options{Help: "Port for Scaffold instance", Default: "2997"})
	configWSPort := configCommand.String("", "ws-port", &argparse.Options{Help: "Websocket port for Scaffold instance", Default: "8080"})
	configProtocol := configCommand.String("", "protocol", &argparse.Options{Help: "Protocol to use to connect to Scaffold instance", Default: "http"})
	configProfile := configCommand.String("p", "profile", &argparse.Options{Help: "Name for the profile to configure", Default: "default"})
	configUsername := configCommand.String("", "username", &argparse.Options{Required: true, Help: "Username to use to connect to Scaffold instance"})
	configPassword := configCommand.String("", "password", &argparse.Options{Required: true, Help: "Password to use to connect to Scaffold instance"})
	configLogLevel := configCommand.Selector("l", "log-level", []string{"NONE", "FATAL", "SUCCESS", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}, &argparse.Options{Help: "Log level to use. Valid options are 'NONE', 'FATAL', 'SUCCESS', 'ERROR', 'WARN', 'INFO', 'DEBUG', 'TRACE'. Defaults to 'ERROR'", Default: "ERROR"})
	configSkipVerify := configCommand.Flag("", "skip-verify", &argparse.Options{Help: "Should SSL certificates not be verified on connection"})

	versionCommand := parser.NewCommand("version", "Get Scaffold versions")

	localCommand := versionCommand.NewCommand("local", "Get local Scaffold CLI version")
	localLogLevel := localCommand.Selector("l", "log-level", []string{"NONE", "FATAL", "SUCCESS", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}, &argparse.Options{Help: "Log level to use. Valid options are 'NONE', 'FATAL', 'SUCCESS', 'ERROR', 'WARN', 'INFO', 'DEBUG', 'TRACE'. Defaults to 'ERROR'", Default: "ERROR"})

	remoteCommand := versionCommand.NewCommand("remote", "Get remote Scaffold manager version")
	remoteProfile := remoteCommand.String("p", "profile", &argparse.Options{Help: "Profile to use to connect to Scaffold instance", Default: "default"})
	remoteLogLevel := remoteCommand.Selector("l", "log-level", []string{"NONE", "FATAL", "SUCCESS", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}, &argparse.Options{Help: "Log level to use. Valid options are 'NONE', 'FATAL', 'SUCCESS', 'ERROR', 'WARN', 'INFO', 'DEBUG', 'TRACE'. Defaults to 'ERROR'", Default: "ERROR"})

	// triggerCommand := parser.NewCommand("trigger", "Trigger a workflow task")
	// triggerProfile := triggerCommand.String("p", "profile", &argparse.Options{Help: "Profile to use to connect to Scaffold instance", Default: "default"})
	// triggerTask := triggerCommand.String("t", "task", &argparse.Options{Required: true, Help: "Task to trigger"})
	// triggerContext := triggerCommand.String("c", "context", &argparse.Options{Help: "Workflow context to use. If not set the value in your config file will be pulled", Default: ""})
	// triggerFollow := triggerCommand.Flag("f", "follow", &argparse.Options{Help: "Should the run status be tailed out"})
	// triggerData := triggerCommand.String("d", "data", &argparse.Options{Help: "JSON data to include with trigger", Default: ""})
	// triggerLogLevel := triggerCommand.Selector("l", "log-level", []string{"NONE", "FATAL", "SUCCESS", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}, &argparse.Options{Help: "Log level to use. Valid options are 'NONE', 'FATAL', 'SUCCESS', 'ERROR', 'WARN', 'INFO', 'DEBUG', 'TRACE'. Defaults to 'ERROR'", Default: "INFO"})

	// Parse input
	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Print(parser.Usage(err))
	}

	if applyCommand.Happened() {
		logger.SetLevel(*applyLogLevel)
		apply.DoApply(*applyProfile, *applyFile)
		os.Exit(0)
	}

	if deleteCommand.Happened() {
		logger.SetLevel(*deleteLogLevel)
		delete.DoDelete(*deleteProfile, *deleteObject, *deleteFilter)
		os.Exit(0)
	}

	if configCommand.Happened() {
		logger.SetLevel(*configLogLevel)
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: *configSkipVerify}
		config.DoConfig(*configHost, *configPort, *configProtocol, *configWSPort, *configProfile, *configUsername, *configPassword, *configSkipVerify)
		os.Exit(0)
	}

	if getCommand.Happened() {
		logger.SetLevel(*getLogLevel)
		get.DoGet(*getProfile, *getObject, *getContext)
		os.Exit(0)
	}

	// if describeCommand.Happened() {
	// 	logger.SetLevel(*describeLogLevel)
	// 	describe.DoDescribe(*describeProfile, *describeObject, *describeContext, *describeFormat)
	// 	os.Exit(0)
	// }

	if localCommand.Happened() {
		logger.SetLevel(*localLogLevel)
		version.DoLocal()
		os.Exit(0)
	}

	if remoteCommand.Happened() {
		logger.SetLevel(*remoteLogLevel)
		version.DoRemote(*remoteProfile)
		os.Exit(0)
	}

	// if triggerCommand.Happened() {
	// 	logger.SetLevel(*triggerLogLevel)
	// 	trigger.DoTrigger(*triggerProfile, *triggerTask, *triggerData, *triggerContext, *triggerFollow)
	// 	os.Exit(0)
	// }
}
