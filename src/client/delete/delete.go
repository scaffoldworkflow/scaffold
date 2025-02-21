package delete

import (
	"fmt"
	"net/http"
	"scaffold/client/auth"
	"scaffold/client/utils"
	"strings"

	logger "github.com/jfcarter2358/go-logger"
)

var ValidObjects = []string{"alert", "release", "project", "runbook", "user"}

func DoDelete(profile, object string, filter []string) {
	// if context == constants.ALL_CONTEXTS {
	// 	logger.Fatalf("", "%s is not allowed for delete actions", constants.ALL_CONTEXTS)
	// }

	if len(filter) == 0 {
		logger.Fatalf("", "No filter passed to delete command")
	}

	logger.Debugf("", "Reading profile")
	p := auth.ReadProfile(profile)
	uri := fmt.Sprintf("%s://%s:%s", p.Protocol, p.Host, p.Port)

	logger.Debugf("", "Checking if object is valid")

	if !utils.Contains(ValidObjects, object) {
		logger.Fatalf("", "Invalid object type passed: '%s'. Valid object types are '%s'", object, strings.Join(ValidObjects, "', '"))
	}

	err := doDelete(p, uri, object, filter)
	if err != nil {
		logger.Fatalf("", "Error deleting object %s: %s", object, err.Error())
	}
}

func doDelete(p auth.ProfileObj, uri, object string, filter []string) error {
	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("%s/api/v1/%s?%s", uri, object, strings.Join(filter, "&"))

	req, _ := http.NewRequest("DELETE", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", p.APIToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Fatalf("", "Encountered error: %s", err.Error())
		return err
	}
	if resp.StatusCode >= 400 {
		logger.Fatalf("", "Got status code %d", resp.StatusCode)
		return fmt.Errorf("got status code %d on %s delete", resp.StatusCode, object)
	}
	logger.Successf("", "%s successfully deleted", object)
	return nil
}
