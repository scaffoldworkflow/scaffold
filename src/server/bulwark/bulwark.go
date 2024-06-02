package bulwark

import (
	"encoding/json"
	"fmt"
	"net/http"
	"scaffold/server/config"

	client "github.com/jfcarter2358/bulwarkmp/client"
	bconst "github.com/jfcarter2358/bulwarkmp/constants"
	logger "github.com/jfcarter2358/go-logger"
)

var ManagerClient *client.Client
var WorkerClient *client.Client
var BufferClient *client.Client

func RunManager(queueFunc func(string, string) error) {
	ManagerClient = &client.Client{}
	for {
		ManagerClient.New(fmt.Sprintf("%s/%s", bconst.VERSION_1, bconst.PROTOCOL_PLAIN), fmt.Sprintf("%s/%s", bconst.ENDPOINT_TYPE_QUEUE, config.Config.ManagerQueueName), nil, queueFunc)
		ManagerClient.Start(config.Config.LogLevel, config.Config.LogFormat, config.Config.BulwarkConnectionString)
	}
}

func RunWorker(queueFunc func(string, string) error) {
	WorkerClient = &client.Client{}
	for {
		WorkerClient.New(fmt.Sprintf("%s/%s", bconst.VERSION_1, bconst.PROTOCOL_PLAIN), fmt.Sprintf("%s/%s", bconst.ENDPOINT_TYPE_QUEUE, config.Config.WorkerQueueName), nil, queueFunc)
		WorkerClient.Start(config.Config.LogLevel, config.Config.LogFormat, config.Config.BulwarkConnectionString)
	}
}

func RunBuffer(bufferFunc func(string, string) error) {
	BufferClient = &client.Client{}
	BufferClient.New(fmt.Sprintf("%s/%s", bconst.VERSION_1, bconst.PROTOCOL_PLAIN), fmt.Sprintf("%s/%s", bconst.ENDPOINT_TYPE_BUFFER, config.Config.KillBufferName), bufferFunc, nil)
	BufferClient.Start(config.Config.LogLevel, config.Config.LogFormat, config.Config.BulwarkConnectionString)
}

func QueueCreate(name string) error {
	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("%s/api/v1/queue/%s", config.Config.BulwarkAPIConnectionString, name)
	req, _ := http.NewRequest("POST", requestURL, nil)
	req.Header.Set("X-Bulwark-API", config.Config.BulwarkSecretKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Errorf("", "Post request failed with error: %s", err.Error())
		return err
	}
	if resp.StatusCode >= 400 {
		logger.Errorf("", "Post request failed with status code %v", resp.StatusCode)
		return err
	}
	return nil
}

func QueueDelete(name string) error {
	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("%s/api/v1/queue/%s", config.Config.BulwarkAPIConnectionString, name)
	req, _ := http.NewRequest("DELETE", requestURL, nil)
	req.Header.Set("X-Bulwark-API", config.Config.BulwarkSecretKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Errorf("", "Delete request failed with error: %s", err.Error())
		return err
	}
	if resp.StatusCode >= 400 {
		logger.Errorf("", "Delete request failed with status code %v", resp.StatusCode)
		return err
	}
	return nil
}

func QueuePop(c *client.Client) {
	logger.Debugf("", "Pulling data")
	c.Pull()
}

func QueuePush(c *client.Client, data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	logger.Debugf("", "Pushing data")
	c.Push(bconst.CONTENT_TYPE_TEXT, string(bytes))
	return nil
}

func BufferCreate(name string) error {
	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("%s/api/v1/buffer/%s", config.Config.BulwarkAPIConnectionString, name)
	req, _ := http.NewRequest("POST", requestURL, nil)
	req.Header.Set("X-Bulwark-API", config.Config.BulwarkSecretKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Errorf("", "Post request failed with error: %s", err.Error())
		return err
	}
	if resp.StatusCode >= 400 {
		logger.Errorf("", "Post request failed with status code %v", resp.StatusCode)
		return err
	}
	return nil
}

func BufferDelete(name string) error {
	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("%s/api/v1/buffer/%s", config.Config.BulwarkAPIConnectionString, name)
	req, _ := http.NewRequest("DELETE", requestURL, nil)
	req.Header.Set("X-Bulwark-API", config.Config.BulwarkSecretKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Errorf("", "Delete request failed with error: %s", err.Error())
		return err
	}
	if resp.StatusCode >= 400 {
		logger.Errorf("", "Delete request failed with status code %v", resp.StatusCode)
		return err
	}
	return nil
}

func BufferGet(c *client.Client) {
	logger.Debugf("", "Pulling data")
	c.Pull()
}

func BufferSet(c *client.Client, data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	logger.Debugf("", "Pushing data")
	c.Push(bconst.CONTENT_TYPE_TEXT, string(bytes))
	return nil
}
