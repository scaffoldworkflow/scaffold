package trigger

import (
	"fmt"
	"scaffold/client/utils"
)

func DoTrigger(host, port, pipelineName string) {
	uri := fmt.Sprintf("http://%s:%s", host, port)

	pipeline, err := utils.SendGet(uri, "v1/pipeline?filter=name%20=%20\""+pipelineName+"\"")

	if err != nil {
		panic(err)
	}

	data := map[string]interface{}{
		"pipeline": pipeline["id"].(string),
	}
	utils.SendPost(uri, "v1/run", data)
}
