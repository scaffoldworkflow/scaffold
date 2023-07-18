package apply

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"scaffold/client/objects"

	"gopkg.in/yaml.v3"
)

func sendPost(uri, path string, data map[string]interface{}) (map[string]interface{}, error) {
	if data == nil {
		data = make(map[string]interface{})
	}
	var obj map[string]interface{}
	requestURL := fmt.Sprintf("%v/%v", uri, path)
	json_data, err := json.Marshal(data)
	if err != nil {
		log.Printf("Encountered error: %v", err)
		return nil, err
	}
	resp, err := http.Post(requestURL, "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		log.Printf("Encountered error: %v", err)
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, errors.New(fmt.Sprintf("Request failed with status code %v", resp.StatusCode))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Encountered error: %v", err)
		return nil, err
	}
	err = json.Unmarshal([]byte(body), &obj)
	if err != nil {
		log.Printf("Encountered error: %v", err)
		return nil, err
	}
	return obj, nil
}

func DoApply(host, port, fileName string) {
	uri := fmt.Sprintf("http://%s:%s", host, port)

	var yamlData map[string]interface{}

	fileData, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(fileData, &yamlData)
	if err != nil {
		panic(err)
	}

	for _, pipelineInterface := range yamlData["pipelines"].([]interface{}) {
		pipeline := objects.Pipeline{}
		pipelineMap := pipelineInterface.(map[string]interface{})
		pipelineYaml, _ := yaml.Marshal(&pipelineMap)
		yaml.Unmarshal(pipelineYaml, &pipeline)
		for _, jobInterface := range pipelineMap["jobs"].([]interface{}) {
			job := objects.Job{}
			jobMap := jobInterface.(map[string]interface{})
			jobYaml, _ := yaml.Marshal(&jobMap)
			yaml.Unmarshal(jobYaml, &job)
			for _, stepInterface := range jobMap["steps"].([]interface{}) {
				step := objects.Step{}
				stepMap := stepInterface.(map[string]interface{})
				stepYaml, _ := yaml.Marshal(&stepMap)
				yaml.Unmarshal(stepYaml, &step)
				stepJsonBytes, _ := json.Marshal(&step)
				var stepJson map[string]interface{}
				yaml.Unmarshal(stepJsonBytes, &stepJson)
				obj, err := sendPost(uri, "v1/step", stepJson)
				if err != nil {
					panic(err)
				}
				id := obj["id"].(string)

				job.Steps = append(job.Steps, id)
			}
			jobJsonBytes, _ := json.Marshal(&job)
			var jobJson map[string]interface{}
			yaml.Unmarshal(jobJsonBytes, &jobJson)
			obj, err := sendPost(uri, "v1/job", jobJson)
			if err != nil {
				panic(err)
			}
			id := obj["id"].(string)

			pipeline.Jobs = append(pipeline.Jobs, id)
		}
		pipelineJsonBytes, _ := json.Marshal(&pipeline)
		var pipelineJson map[string]interface{}
		yaml.Unmarshal(pipelineJsonBytes, &pipelineJson)
		_, err := sendPost(uri, "v1/pipeline", pipelineJson)
		if err != nil {
			panic(err)
		}
	}

}
