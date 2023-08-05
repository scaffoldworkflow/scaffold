package socket

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"scaffold/server/container"
	"scaffold/server/user"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var Manager = ClientManager{
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
}

//Client management
type ClientManager struct {
	//The client map stores and manages all long connection clients, online is TRUE, and those who are not there are FALSE
	clients map[*Client]bool
	//Web side MESSAGE we use Broadcast to receive, and finally distribute it to all clients
	broadcast chan []byte
	//Newly created long connection client
	register chan *Client
	//Newly canceled long connection client
	unregister chan *Client
}

//Client
type Client struct {
	//User ID
	ID string
	//Connected socket
	Socket *websocket.Conn
	//Message
	Send          chan []byte
	Authenticated bool
	Container     container.Container
	User          user.User
}

//Will formatting Message into JSON
// type Message struct {
// 	//Message Struct
// 	Sender    string `json:"sender,omitempty"`
// 	Recipient string `json:"recipient,omitempty"`
// 	Content   string `json:"content,omitempty"`
// 	ServerIP  string `json:"serverIp,omitempty"`
// 	SenderIP  string `json:"senderIp,omitempty"`
// }

type Message struct {
	Content string `json:"content"`
	Status  int    `json:"status"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 1024 * 1024,
	WriteBufferSize: 1024 * 1024 * 1024,
	//Solving cross-domain problems
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (Manager *ClientManager) Start() {
	for {
		select {
		//If there is a new connection access, pass the connection to conn through the channel
		case conn := <-Manager.register:
			conn.Container = container.Container{}
			//Set the client connection to true
			Manager.clients[conn] = true
			//Format the message of returning to the successful connection JSON
			jsonMessage, _ := json.Marshal(&Message{Content: "Connection established", Status: 202})
			//Call the client's send method and send messages
			Manager.Send(jsonMessage, conn)
			//If the connection is disconnected
		case conn := <-Manager.unregister:
			//Determine the state of the connection, if it is true, turn off Send and delete the value of connecting client
			if _, ok := Manager.clients[conn]; ok {
				close(conn.Send)
				delete(Manager.clients, conn)
				conn.Container.Stop()
				jsonMessage, _ := json.Marshal(&Message{Content: "Connection terminated ", Status: 406})
				Manager.Send(jsonMessage, conn)
			}
			//broadcast
		case message := <-Manager.broadcast:
			//Traversing the client that has been connected, send the message to them
			for conn := range Manager.clients {
				select {
				case conn.Send <- message:
				default:
					conn.Container.Stop()
					close(conn.Send)
					delete(Manager.clients, conn)
				}
			}
		}
	}
}

//Define the send method of client management
func (Manager *ClientManager) Send(message []byte, ignore *Client) {
	for conn := range Manager.clients {
		//Send messages not to the shielded connection
		if conn != ignore {
			conn.Send <- message
		}
	}
}

//Define the read method of the client structure
func (c *Client) Read() {
	defer func() {
		Manager.unregister <- c
		c.Container.Stop()
		_ = c.Socket.Close()
	}()

	for {
		//Read message
		_, inbound, err := c.Socket.ReadMessage()
		//If there is an error message, cancel this connection and then close it
		if err != nil {
			c.Container.Stop()
			Manager.unregister <- c
			_ = c.Socket.Close()
			break
		}

		var inboundData Message
		json.Unmarshal([]byte(inbound), &inboundData)

		// outbound := []byte{}

		outboundData := Message{}

		if !c.Authenticated {
			// Use "status teapot" for auth challenge
			if inboundData.Status == 418 {
				usr, _ := user.GetUserByAPIToken(inboundData.Content)
				if usr != nil {
					outboundData.Content = "authenticated"
					outboundData.Status = 202
					c.Authenticated = true
					c.User = *usr
				} else {
					outboundData.Content = "access denied"
					outboundData.Status = 401
				}
			}
		} else {
			fmt.Println("Authenticated!")
			switch inboundData.Status {
			case 100:
				if c.Container.Name == "" {
					fmt.Println("Getting all containers")
					data := map[int]string{}
					for idx, name := range container.LastRun {
						data[idx] = name
					}

					outputBytes, _ := json.Marshal(data)
					output := string(outputBytes)

					outboundData.Content = output
					outboundData.Status = 300
				}
			case 200:
				fmt.Println("Got read code")
				if c.Container.Error != "" {
					outboundMap := map[string]string{"stdout": c.Container.Error}
					output, _ := json.Marshal(outboundMap)

					outboundData.Content = string(output)
					outboundData.Status = 500
				} else {
					c.Container.OutputReady = false
					stdout := c.Container.Output
					c.Container.Output = ""
					c.Container.OutputReady = true
					status := 200
					if c.Container.Error != "" {
						status = 500
					}

					fmt.Printf("Get stdout '%s' and status '%d'\n", stdout, status)

					outboundMap := map[string]string{"stdout": stdout}
					output, _ := json.Marshal(outboundMap)

					outboundData.Content = string(output)
					outboundData.Status = status
				}
			case 201:
				data := map[string]string{}
				json.Unmarshal([]byte(inboundData.Content), &data)
				// outString, status := c.Container.Write(data["stdin"])

				c.Container.Input = data["stdin"]
				c.Container.InputReady = true

				if c.Container.Error != "" {
					outboundMap := map[string]string{"stdout": c.Container.Error}
					output, _ := json.Marshal(outboundMap)

					outboundData.Content = string(output)
					outboundData.Status = 500
				} else {
					c.Container.OutputReady = false
					stdout := c.Container.Output
					c.Container.Output = ""
					c.Container.OutputReady = true
					status := 200
					if c.Container.Error != "" {
						status = 500
					}

					outboundMap := map[string]string{"stdout": stdout}
					output, _ := json.Marshal(outboundMap)

					outboundData.Content = string(output)
					outboundData.Status = status
				}
			case 302:
				name := inboundData.Content
				cn, err := container.SetupContainer(name)
				if err != nil {
					outboundData.Content = err.Error()
					outboundData.Status = 500
				} else {
					cn.User = c.User
					cn.OutputReady = true
					c.Container = cn
					outboundData.Status = 200
					outboundData.Content = fmt.Sprintf(`{"stdout":"Connection established to container %s"}`, cn.Name)
				}
				go c.Container.ExecContainer(name)
			}
		}

		outbound, _ := json.Marshal(outboundData)
		Manager.broadcast <- []byte(outbound)
	}
}

func (c *Client) Write() {
	defer func() {
		_ = c.Socket.Close()
	}()

	for {
		select {
		//Read the message from send
		case message, ok := <-c.Send:
			//If there is no message
			if !ok {
				_ = c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			//Write it if there is news and send it to the web side
			_ = c.Socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func WebsocketHandler(ctx *gin.Context) {
	//Upgrade the HTTP protocol to the websocket protocol
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	//Every connection will open a new client, client.id generates through UUID to ensure that each time it is different
	client := &Client{ID: uuid.New().String(), Socket: conn, Send: make(chan []byte), Authenticated: false, Container: container.Container{Stdin: nil, Stdout: nil, Name: ""}}
	//Register a new link
	Manager.register <- client

	//Start the message to collect the news from the web side
	go client.Read()
	//Start the corporation to return the message to the web side
	go client.Write()
}

func HealthHandler(res http.ResponseWriter, _ *http.Request) {
	_, _ = res.Write([]byte("ok"))
}

func LocalIp() string {
	address, _ := net.InterfaceAddrs()
	var ip = "localhost"
	for _, address := range address {
		if ipAddress, ok := address.(*net.IPNet); ok && !ipAddress.IP.IsLoopback() {
			if ipAddress.IP.To4() != nil {
				ip = ipAddress.IP.String()
			}
		}
	}
	return ip
}
