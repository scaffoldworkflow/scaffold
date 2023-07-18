// utils.go

package utils

import (
	"log"
	"os/exec"

	"github.com/gin-gonic/gin"
)

func Error(err error, c *gin.Context, statusCode int) {
	log.Printf("Encountered error: %v", err)
	c.JSON(statusCode, gin.H{"error": err.Error()})
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func RemoveDuplicateValues(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	// If the key(values of the slice) is not equal
	// to the already present value in new slice (list)
	// then we append it. else we jump on another element.
	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func InstallPackages(packages []string) error {
	out, err := exec.Command("apt-get", "update", "-y").CombinedOutput()
	if err != nil {
		log.Printf("apt-get update: %v", string(out))
		return err
	}
	for _, pkg := range packages {
		out, err := exec.Command("apt-get", "install", "-y", pkg).CombinedOutput()
		if err != nil {
			log.Printf("apt-get install %s: %s", pkg, string(out))
			return err
		}
	}
	return nil
}
