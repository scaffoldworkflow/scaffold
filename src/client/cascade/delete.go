package cascade

import (
	"fmt"
	"scaffold/client/logger"
	"scaffold/client/utils"
)

func DoDelete(host, port, cascadeName string) {
	uri := fmt.Sprintf("http://%s:%s", host, port)

	// Check to see if a cascade already exists
	path := fmt.Sprintf("api/v1/cascade/%s", cascadeName)
	_, err := utils.SendDelete(uri, path)
	if err != nil {
		logger.Fatalf("", "Encountered error deleting cascade: %s", err.Error())
	}
	logger.Successf("", "Successfully deleted cascade!")
}
