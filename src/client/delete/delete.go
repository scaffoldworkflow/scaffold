package delete

import (
	"fmt"
	"net/http"
	"scaffold/client/auth"
	"scaffold/client/constants"
	"scaffold/client/logger"
	"scaffold/client/utils"
	"strings"
)

func DoDelete(profile, object, context string) {
	if context == constants.ALL_CONTEXTS {
		logger.Fatalf("", "%s is not allowed for delete actions", constants.ALL_CONTEXTS)
	}

	logger.Debugf("", "Getting objects")
	p := auth.ReadProfile(profile)
	uri := fmt.Sprintf("%s://%s:%s", p.Protocol, p.Host, p.Port)

	logger.Debugf("", "Checking if object is valid")
	objects := []string{"cascade", "datastore", "state", "task", "file", "user", "input"}

	parts := strings.Split(object, "/")

	if !utils.Contains(objects, parts[0]) {
		logger.Fatalf("", "Invalid object type passed: '%s'. Valid object types are %v", object, objects)
	}

	if len(parts) == 1 {
		logger.Fatalf("", "Object passed in need to be of format '<object type>/<object name>")
	}

	if parts[0] != "cascade" && parts[0] != "datastore" && parts[0] != "user" {
		if context == "" {
			context = p.Cascade
		}
		object = fmt.Sprintf("%s/%s/%s", parts[0], context, parts[1])
	}

	err := doDelete(p, uri, object)
	if err != nil {
		logger.Fatalf("", "Error deleting object %s: %s", object, err.Error())
	}
}

func doDelete(p auth.ProfileObj, uri, object string) error {
	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("%s/api/v1/%s", uri, object)
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
	return nil
}
