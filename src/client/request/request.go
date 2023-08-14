package request

import (
	"crypto/tls"
	"net/http"
)

func CreateClient(protocol string, skipVerify bool) *http.Client {
	if protocol == "https" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
		}
		c := &http.Client{Transport: tr}
		return c
	}
	return &http.Client{}
}
