package auth

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"scaffold/client/logger"

	"gopkg.in/yaml.v3"
)

type ProfileObj struct {
	Protocol string
	Host     string
	Port     string
	WSPort   string
	APIToken string
}

func WriteProfile(profile, protocol, host, port, wsPort, apiToken string) {
	usr, _ := user.Current()
	configDir := fmt.Sprintf("%s/.scaffold", usr.HomeDir)
	configFile := fmt.Sprintf("%s/config", configDir)

	profileData := make(map[string]ProfileObj)

	if _, err := os.Stat(configDir); err != nil {
		logger.Debugf("", "Config directory not present at %s, creating", configDir)
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			logger.Fatalf("", "Encountered error creating conifg directory at %s: %s", configDir, err.Error())
		}
		logger.Debug("", "Config directory created successfully")
	}

	if _, err := os.Stat(configFile); err == nil {
		logger.Debugf("", "Config file exists at %s", configFile)
		fileData, err := os.ReadFile(configFile)
		if err != nil {
			logger.Fatalf("", "Encountered error reading profile YAML file: %s", err.Error())
		}
		logger.Debugf("", "Config file read successfully")
		err = yaml.Unmarshal(fileData, &profileData)
		if err != nil {
			logger.Fatalf("", "Encountered error unmarshalling profile YAML: %s", err.Error())
		}
		logger.Debugf("", "Config unmarshalled successfully")
	}

	profileData[profile] = ProfileObj{
		Protocol: protocol,
		Host:     host,
		Port:     port,
		WSPort:   wsPort,
		APIToken: apiToken,
	}

	yamlData, err := yaml.Marshal(profileData)

	err = ioutil.WriteFile(configFile, yamlData, 0644)
	if err != nil {
		logger.Fatalf("", "Error writing config file: %s", err.Error())
	}
	logger.Successf("", "Successfully added profile %s to config file", profile)
}

func ReadProfile(profile string) ProfileObj {
	usr, _ := user.Current()
	configDir := fmt.Sprintf("%s/.scaffold", usr.HomeDir)
	configFile := fmt.Sprintf("%s/config", configDir)

	profileData := make(map[string]ProfileObj)

	if _, err := os.Stat(configDir); err != nil {
		logger.Fatalf("", "Config directory not present at %s", configDir)
	}

	if _, err := os.Stat(configFile); err != nil {
		logger.Fatalf("", "No config file exists at %s", configFile)
	}
	fileData, err := os.ReadFile(configFile)
	if err != nil {
		logger.Fatalf("", "Encountered error reading profile YAML file: %s", err.Error())
	}
	logger.Debugf("", "Config file read successfully")
	err = yaml.Unmarshal(fileData, &profileData)
	if err != nil {
		logger.Fatalf("", "Encountered error unmarshalling profile YAML: %s", err.Error())
	}
	logger.Debugf("", "Config unmarshalled successfully")

	if val, ok := profileData[profile]; ok {
		logger.Debugf("", "Found profile %s in config", profile)
		return val
	}
	logger.Fatalf("", "No profile found with name %s", profile)
	return ProfileObj{}
}
