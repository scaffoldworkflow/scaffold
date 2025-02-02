// utils.go

package utils

import (
	"log"
	"math/rand"
	"os/exec"
	"strconv"

	logger "github.com/jfcarter2358/go-logger"

	"github.com/gin-gonic/gin"
)

func Error(err error, c *gin.Context, statusCode int) {
	logger.Error("", err.Error())
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

func ContainsInt(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Remove(l []string, item string) []string {
	for i, other := range l {
		if other == item {
			return append(l[:i], l[i+1:]...)
		}
	}
	return l
}

func Keys(m map[string]string) []string {
	keys := make([]string, len(m))

	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
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

func GenerateToken(length int) string {
	token := strconv.FormatInt(rand.Int63(), 16) + strconv.FormatInt(rand.Int63(), 16) + strconv.FormatInt(rand.Int63(), 16) + strconv.FormatInt(rand.Int63(), 16) + strconv.FormatInt(rand.Int63(), 16) + strconv.FormatInt(rand.Int63(), 16) + strconv.FormatInt(rand.Int63(), 16) + strconv.FormatInt(rand.Int63(), 16)

	token = token[:length]
	return token
}

func DynamicAPIResponse(ctx *gin.Context, redirect string, status int, response gin.H) {
	_, err := ctx.Cookie("scaffold_token")
	if err == nil {
		logger.Debugf("", "Redirecting to %s", redirect)
		ctx.Redirect(302, redirect)
		return
	}
	ctx.JSON(status, response)
}

func MergeDict(a, b map[string]string) map[string]string {
	if a == nil {
		a = make(map[string]string)
	}
	if b == nil {
		return a
	}
	for key, val := range b {
		a[key] = val
	}
	return a
}
