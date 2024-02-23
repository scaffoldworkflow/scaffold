package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"scaffold/client/apply"
	"scaffold/client/config"
	"scaffold/client/constants"
	"scaffold/client/context"
	"scaffold/client/delete"
	"scaffold/client/describe"
	"scaffold/client/exec"
	"scaffold/client/file"
	"scaffold/client/get"
	"scaffold/client/logger"
	"scaffold/client/version"

	"github.com/akamensky/argparse"
)

func main() {
	// config.LoadConfig()
	// logger.SetLevel(config.Config.LogLevel)
	logger.SetLevel(constants.LOG_LEVEL_DEBUG)

	parser := argparse.NewParser("scaffold", "Scaffold infrastructure management client")

	applyCommand := parser.NewCommand("apply", "Create or update a Scaffold object")
	applyObject := applyCommand.StringPositional(&argparse.Options{Required: true, Help: "Scaffold object type to create. Valid object types are 'cascade', 'datastore', 'task', 'state', 'file', and 'user"})
	applyContext := applyCommand.String("c", "context", &argparse.Options{Help: "Cascade context to use. If not set the value in your config file will be pulled", Default: ""})
	applyProfile := applyCommand.String("p", "profile", &argparse.Options{Help: "Profile to use to connect to Scaffold instance", Default: "default"})
	applyFile := applyCommand.String("f", "file", &argparse.Options{Required: true, Help: "Scaffold manifest to apply"})
	applyLogLevel := applyCommand.Selector("l", "log-level", []string{"NONE", "FATAL", "SUCCESS", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}, &argparse.Options{Help: "Log level to use. Valid options are 'NONE', 'FATAL', 'SUCCESS', 'ERROR', 'WARN', 'INFO', 'DEBUG', 'TRACE'. Defaults to 'ERROR'", Default: "ERROR"})

	deleteCommand := parser.NewCommand("delete", "Delete an existing Scaffold object")
	deleteObject := deleteCommand.StringPositional(&argparse.Options{Required: true, Help: "Scaffold object to get. Can be of format '<object type>', or '<object type>/<object name>'. Valid object types are 'cascade', 'datastore', 'task', 'state', 'file', and 'user"})
	deleteContext := deleteCommand.String("c", "context", &argparse.Options{Help: "Cascade context to use. If not set the value in your config file will be pulled", Default: ""})
	deleteProfile := deleteCommand.String("p", "profile", &argparse.Options{Help: "Profile to use to connect to Scaffold instance", Default: "default"})
	deleteLogLevel := deleteCommand.Selector("l", "log-level", []string{"NONE", "FATAL", "SUCCESS", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}, &argparse.Options{Help: "Log level to use. Valid options are 'NONE', 'FATAL', 'SUCCESS', 'ERROR', 'WARN', 'INFO', 'DEBUG', 'TRACE'. Defaults to 'ERROR'", Default: "ERROR"})

	execCommand := parser.NewCommand("exec", "Exec into a scaffold container")
	execProfile := execCommand.String("p", "profile", &argparse.Options{Help: "Profile to use to connect to Scaffold instance", Default: "default"})
	execLogLevel := execCommand.Selector("l", "log-level", []string{"NONE", "FATAL", "SUCCESS", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}, &argparse.Options{Help: "Log level to use. Valid options are 'NONE', 'FATAL', 'SUCCESS', 'ERROR', 'WARN', 'INFO', 'DEBUG', 'TRACE'. Defaults to 'ERROR'", Default: "ERROR"})

	getCommand := parser.NewCommand("get", "Get Scaffold objects")
	getObject := getCommand.StringPositional(&argparse.Options{Required: true, Help: "Scaffold object to get. Can be of format '<object type>', or '<object type>/<object name>'. Valid object types are 'cascade', 'datastore', 'task', 'state', 'file', and 'user"})
	getContext := getCommand.String("c", "context", &argparse.Options{Help: "Cascade context to use. If not set the value in your config file will be pulled", Default: ""})
	getProfile := getCommand.String("p", "profile", &argparse.Options{Help: "Profile to use to connect to Scaffold instance", Default: "default"})
	getLogLevel := getCommand.Selector("l", "log-level", []string{"NONE", "FATAL", "SUCCESS", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}, &argparse.Options{Help: "Log level to use. Valid options are 'NONE', 'FATAL', 'SUCCESS', 'ERROR', 'WARN', 'INFO', 'DEBUG', 'TRACE'. Defaults to 'ERROR'", Default: "ERROR"})

	describeCommand := parser.NewCommand("describe", "Describe a Scaffold object")
	describeObject := describeCommand.StringPositional(&argparse.Options{Required: true, Help: "Scaffold object to describe. Must be of format '<object type>/<object name>'. Valid object types are 'cascade', 'datastore', 'task', 'state', 'file', and 'user'"})
	describeProfile := describeCommand.String("p", "profile", &argparse.Options{Help: "Profile to use to connect to Scaffold instance", Default: "default"})
	describeContext := describeCommand.String("c", "context", &argparse.Options{Help: "Cascade context to use. If not set the value in your config file will be pulled", Default: ""})
	describeFormat := describeCommand.Selector("o", "output", []string{"yaml", "json"}, &argparse.Options{Help: "Output format to print. Valid options are 'yaml' and 'json'. Defaults to 'yaml'", Default: "yaml"})
	describeLogLevel := describeCommand.Selector("l", "log-level", []string{"NONE", "FATAL", "SUCCESS", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}, &argparse.Options{Help: "Log level to use. Valid options are 'NONE', 'FATAL', 'SUCCESS', 'ERROR', 'WARN', 'INFO', 'DEBUG', 'TRACE'. Defaults to 'ERROR'", Default: "ERROR"})

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

	contextCommand := parser.NewCommand("context", "Configure cascade context for a Scaffold instance")
	contextContext := contextCommand.StringPositional(&argparse.Options{Required: true, Help: "Scaffold cascade context to use"})
	contextProfile := contextCommand.String("p", "profile", &argparse.Options{Help: "Name for the profile to configure", Default: "default"})
	contextLogLevel := contextCommand.Selector("l", "log-level", []string{"NONE", "FATAL", "SUCCESS", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}, &argparse.Options{Help: "Log level to use. Valid options are 'NONE', 'FATAL', 'SUCCESS', 'ERROR', 'WARN', 'INFO', 'DEBUG', 'TRACE'. Defaults to 'ERROR'", Default: "ERROR"})

	fileCommand := parser.NewCommand("file", "Interact with filestore files")

	uploadCommand := fileCommand.NewCommand("upload", "Upload a file to a filestore")
	uploadProfile := uploadCommand.String("p", "profile", &argparse.Options{Help: "Profile to use to connect to Scaffold instance", Default: "default"})
	uploadFile := uploadCommand.String("f", "file", &argparse.Options{Required: true, Help: "Path to file to upload"})
	uploadCascade := uploadCommand.String("c", "cascade", &argparse.Options{Required: true, Help: "Cascade filestore to upload file to"})
	uploadLogLevel := uploadCommand.Selector("l", "log-level", []string{"NONE", "FATAL", "SUCCESS", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}, &argparse.Options{Help: "Log level to use. Valid options are 'NONE', 'FATAL', 'SUCCESS', 'ERROR', 'WARN', 'INFO', 'DEBUG', 'TRACE'. Defaults to 'ERROR'", Default: "ERROR"})

	downloadCommand := fileCommand.NewCommand("download", "Download a file from a filestore")
	downloadProfile := downloadCommand.String("p", "profile", &argparse.Options{Help: "Profile to use to connect to Scaffold instance", Default: "default"})
	downloadFile := downloadCommand.String("f", "file", &argparse.Options{Required: true, Help: "Path to file to download"})
	downloadCascade := downloadCommand.String("c", "cascade", &argparse.Options{Required: true, Help: "Cascade filestore to download file from"})
	downloadName := downloadCommand.String("n", "name", &argparse.Options{Required: true, Help: "Filename to download from cascade filestore"})
	downloadLogLevel := downloadCommand.Selector("l", "log-level", []string{"NONE", "FATAL", "SUCCESS", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}, &argparse.Options{Help: "Log level to use. Valid options are 'NONE', 'FATAL', 'SUCCESS', 'ERROR', 'WARN', 'INFO', 'DEBUG', 'TRACE'. Defaults to 'ERROR'", Default: "ERROR"})

	versionCommand := parser.NewCommand("version", "Get Scaffold versions")

	localCommand := versionCommand.NewCommand("local", "Get local Scaffold CLI version")
	localLogLevel := localCommand.Selector("l", "log-level", []string{"NONE", "FATAL", "SUCCESS", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}, &argparse.Options{Help: "Log level to use. Valid options are 'NONE', 'FATAL', 'SUCCESS', 'ERROR', 'WARN', 'INFO', 'DEBUG', 'TRACE'. Defaults to 'ERROR'", Default: "ERROR"})

	remoteCommand := versionCommand.NewCommand("remote", "Get remote Scaffold server version")
	remoteProfile := remoteCommand.String("p", "profile", &argparse.Options{Help: "Profile to use to connect to Scaffold instance", Default: "default"})
	remoteLogLevel := remoteCommand.Selector("l", "log-level", []string{"NONE", "FATAL", "SUCCESS", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}, &argparse.Options{Help: "Log level to use. Valid options are 'NONE', 'FATAL', 'SUCCESS', 'ERROR', 'WARN', 'INFO', 'DEBUG', 'TRACE'. Defaults to 'ERROR'", Default: "ERROR"})

	// Parse input
	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Print(parser.Usage(err))
	}

	if applyCommand.Happened() {
		logger.SetLevel(*applyLogLevel)
		apply.DoApply(*applyProfile, *applyObject, *applyContext, *applyFile)
		os.Exit(0)
	}

	if deleteCommand.Happened() {
		logger.SetLevel(*deleteLogLevel)
		delete.DoDelete(*deleteProfile, *deleteObject, *deleteContext)
		os.Exit(0)
	}

	if execCommand.Happened() {
		logger.SetLevel(*execLogLevel)
		exec.DoExec(*execProfile)
		os.Exit(0)
	}

	if configCommand.Happened() {
		logger.SetLevel(*configLogLevel)
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: *configSkipVerify}
		config.DoConfig(*configHost, *configPort, *configProtocol, *configWSPort, *configProfile, *configUsername, *configPassword, *configSkipVerify)
		os.Exit(0)
	}

	if contextCommand.Happened() {
		logger.SetLevel(*contextLogLevel)
		context.DoContext(*contextProfile, *contextContext)
		os.Exit(0)
	}

	if getCommand.Happened() {
		logger.SetLevel(*getLogLevel)
		get.DoGet(*getProfile, *getObject, *getContext)
		os.Exit(0)
	}

	if describeCommand.Happened() {
		logger.SetLevel(*describeLogLevel)
		describe.DoDescribe(*describeProfile, *describeObject, *describeContext, *describeFormat)
		os.Exit(0)
	}

	if uploadCommand.Happened() {
		logger.SetLevel(*uploadLogLevel)
		file.DoUpload(*uploadProfile, *uploadCascade, *uploadFile)
		os.Exit(0)
	}

	if downloadCommand.Happened() {
		logger.SetLevel(*downloadLogLevel)
		file.DoDownload(*downloadProfile, *downloadCascade, *downloadName, *downloadFile)
		os.Exit(0)
	}

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
}
