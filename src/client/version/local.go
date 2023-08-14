package version

import (
	"fmt"
	"scaffold/client/constants"
)

func DoLocal() {
	fmt.Printf("Scaffold CLI Version: %s\n", constants.VERSION)
}
