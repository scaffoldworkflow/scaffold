package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func SendPost(uri, path string, data map[string]interface{}) (map[string]interface{}, error) {
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
		return nil, fmt.Errorf("request failed with status code %v", resp.StatusCode)
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

func SendGet(uri, path string) (map[string]interface{}, error) {
	var objs []map[string]interface{}
	requestURL := fmt.Sprintf("%v/%v", uri, path)
	resp, err := http.Get(requestURL)
	if err != nil {
		log.Printf("Encountered error: %v", err)
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed with status code %v", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Encountered error: %v", err)
		return nil, err
	}
	err = json.Unmarshal([]byte(body), &objs)
	if err != nil {
		log.Printf("Encountered error: %v", err)
		return nil, err
	}
	if len(objs) == 0 {
		return nil, errors.New("no pipeline found")
	}
	return objs[0], nil
}
