package exec

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"scaffold/server/utils"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/buger/goterm"
	"github.com/gorilla/websocket"
)

// type Message struct {
// 	//Message Struct
// 	Sender    string `json:"sender,omitempty"`
// 	Recipient string `json:"recipient,omitempty"`
// 	Content   string `json:"content,omitempty"`
// 	ServerIP  string `json:"serverIp,omitempty"`
// 	SenderIP  string `json:"senderIp,omitempty"`
// }

var token = "MyCoolPrimaryKey12345"

type Event string

const (
	EventResize  Event = "resize"
	EventSendkey Event = "sendKey"
	EventClose   Event = "close"
)

type Message struct {
	Event Event       `json:"event"`
	Data  interface{} `json:"data"`
}

var messageReady = true
var message = []byte{}

func ChooseContainer(host, port string, optionMap map[string]string) {
	selected := -1
	shouldList := true
	fmt.Printf("Options map: %v\n", optionMap)
	for shouldList {
		fmt.Println("AVAILABLE CONTAINERS")
		fmt.Println("--------------------")
		idx := 1
		for node := range optionMap {
			fmt.Printf("(%d)  %s\n", idx, node)
			idx += 1
		}
		fmt.Printf("(%d)  Exit\n", idx)
		fmt.Print(": ")

		var input string
		fmt.Scanln(&input)
		if val, err := strconv.Atoi(input); err == nil {
			if val == len(optionMap)+1 {
				os.Exit(0)
			} else if val > len(optionMap)+1 {
				fmt.Printf("Invalid option: %d\n", val)
				fmt.Println()
			} else {
				selected = val
				break
			}
		}
	}

	selected -= 1

	keys := utils.Keys(optionMap)
	name := ""
	hostPort := ""
	for idx, key := range keys {
		if idx == selected {
			name = key
			hostPort = optionMap[key]
		}
	}
	connectionParts := strings.Split(hostPort, ":")
	nameParts := strings.Split(name, ".")

	// fmt.Printf("You chose %s on node %s\n", name, hostPort)
	// fmt.Printf("connectionParts: %v\n", connectionParts)
	// fmt.Printf("nameParts: %v\n", nameParts)
	ConnectWebsocket(host, port, connectionParts[0], connectionParts[1], nameParts[0], nameParts[1], nameParts[2])
}

func DoExec(host, port string) {
	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("http://%s:%s/api/v1/run/containers", host, port)
	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", token))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)

	if err != nil {
		panic(err)
	}
	var data map[string][]string
	if resp.StatusCode == http.StatusOK {
		//Read the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		json.Unmarshal(body, &data)
		resp.Body.Close()
	} else {
		panic(fmt.Errorf("status code is %d", resp.StatusCode))
	}

	optionMap := map[string]string{}
	for host, containers := range data {
		for _, container := range containers {
			optionMap[container] = host
		}
	}

	ChooseContainer(host, port, optionMap)
}

func getInput(c *websocket.Conn) {
	reader := bufio.NewReader(os.Stdin)
	for {
		var data string
		buf := make([]byte, 1024)

		for {
			n, err := reader.Read(buf)
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err)
			}
			if n > 0 {
				data += string(buf[:n])
				output := Message{Event: EventSendkey, Data: data}
				message, _ := json.Marshal(&output)
				err := c.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Println("write error: ", err)
					return
				}
				data = ""
			}
		}
	}
}

func ConnectWebsocket(proxyHost, proxyPort, host, port, cascade, run, version string) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	msg := Message{Event: EventResize, Data: map[string]int{"cols": goterm.Width(), "rows": goterm.Height()}}
	message, _ = json.Marshal(&msg)
	messageReady = true

	// u := url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:%s", host, port), Path: "/api/v1/exec"}
	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:8081", proxyHost), Path: fmt.Sprintf("/?host=%s&port=%s&cascade=%s&run=%s&version=%s", host, port, cascade, run, version)}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{"Authorization": []string{fmt.Sprintf("X-Scaffold-API %s", token)}})
	if err != nil {
		log.Fatal("dial error: ", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go getInput(c)

	go func() {
		defer close(done)
		for {
			_, data, err := c.ReadMessage()
			if err != nil {
				log.Println("read error: ", err)
				return
			}
			fmt.Print(string(data))
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if messageReady {
				err := c.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Println("write error: ", err)
					return
				}
				message = []byte{}
				messageReady = false
			}
		case <-interrupt:
			sig := <-interrupt
			if sig == syscall.SIGTERM {
				msg := Message{Event: EventSendkey, Data: string([]byte{'\003'})}
				message, _ = json.Marshal(&msg)
				err := c.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Println("control-c: ", err)
				}
				message = []byte{}
			} else {
				log.Println("interrupt -- exiting")

				// Cleanly close the connection by sending a close message and then
				// waiting (with timeout) for the server to close the connection.
				err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Println("write close error: ", err)
					return
				}
				select {
				case <-done:
				case <-time.After(time.Second):
				}
				return
			}
		}
	}
}
