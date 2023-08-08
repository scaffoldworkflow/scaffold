package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"scaffold/server/config"
	"scaffold/server/logger"
	"sort"
	"syscall"
	"time"
	"unicode/utf8"

	_ "embed"

	"github.com/creack/pty"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"
)

// wait time for server start
var waitTime = 500
var checkProcInterval = 5

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

type Connection struct {
	PTMX    *os.File
	ExecCmd *exec.Cmd
	Image   string
	Done    bool
	Name    string
}

var connections []Connection

func (c *Connection) run(ws *websocket.Conn) {

	vars := mux.Vars(ws.Request())

	cascade := vars["cascade"]
	run := vars["run"]
	version := vars["version"]

	id := uuid.New().String()

	// parts := strings.Split(name, ".")
	runDir := fmt.Sprintf("/tmp/run/%s/%s/%s", cascade, run, version)
	containerName := fmt.Sprintf("%s-%s-%s", cascade, run, version)
	c.Image = containerName
	c.Name = containerName

	commitCommand := fmt.Sprintf("podman commit %s %s/%s", containerName, id, containerName)

	out, err := exec.Command("bash", "-c", commitCommand).CombinedOutput()
	logger.Debugf("", "podman commit: %s", string(out))
	if err != nil {
		logger.Error("", err.Error())
	}

	podmanCommand := "podman run --rm --privileged --security-opt label=disabled -it "
	podmanCommand += fmt.Sprintf("--mount type=bind,src=%s,dst=/tmp/run ", runDir)
	podmanCommand += fmt.Sprintf("%s/%s ", id, containerName)
	podmanCommand += "sh"

	logger.Debugf("", "Podman websocket command: %s", podmanCommand)

	c.Name = containerName

	defer ws.Close()

	wsconn := &wsConn{
		conn: ws,
	}

	if c.PTMX == nil {
		var msg Message
		if err := json.NewDecoder(ws).Decode(&msg); err != nil {
			log.Println("failed to decode message:", err)
			return
		}

		rows, cols, err := windowSize(msg.Data)
		if err != nil {
			_, _ = ws.Write([]byte(fmt.Sprintf("%s\r\n", err)))
			return
		}
		winsize := &pty.Winsize{
			Rows: rows,
			Cols: cols,
		}

		pc := []string{"bash", "-c", podmanCommand}
		c.ExecCmd = exec.Command(pc[0], pc[1:]...)

		c.PTMX, err = pty.StartWithSize(c.ExecCmd, winsize)
		if err != nil {
			log.Println("failed to create pty", err)
			return
		}
	}

	// write data to process
	go func() {
		for {
			var msg Message
			if err := json.NewDecoder(ws).Decode(&msg); err != nil {
				fmt.Printf("Error decoding JSON: %s\n", err.Error())
				return
			}

			if msg.Event == EventClose {
				log.Println("close websocket")
				ws.Close()
				return
			}

			if msg.Event == EventResize {
				log.Println("do resize")
				rows, cols, err := windowSize(msg.Data)
				if err != nil {
					log.Println(err)
					return
				}

				winsize := &pty.Winsize{
					Rows: rows,
					Cols: cols,
				}

				if err := pty.Setsize(c.PTMX, winsize); err != nil {
					log.Println("failed to set window size:", err)
					return
				}
				fmt.Println("resize done")
				continue
			}

			data, ok := msg.Data.(string)
			if !ok {
				log.Println("invalid message data:", data)
				return
			}

			_, err := c.PTMX.WriteString(data)
			if err != nil {
				log.Println("failed to write data to ptmx:", err)
				return
			}
		}
	}()

	go c.ExecuteCommand()

	_, _ = io.Copy(wsconn, c.PTMX)

}

type wsConn struct {
	conn *websocket.Conn
	buf  []byte
}

// Checking and buffering `b`
// If `b` is invalid UTF-8, it would be buffered
// if buffer is valid UTF-8, it would write to connection
func (ws *wsConn) Write(b []byte) (i int, err error) {
	if !utf8.Valid(b) {
		buflen := len(ws.buf)
		blen := len(b)
		ws.buf = append(ws.buf, b...)[:buflen+blen]
		if utf8.Valid(ws.buf) {
			_, e := ws.conn.Write(ws.buf)
			ws.buf = ws.buf[:0]
			return blen, e
		}
		return blen, nil
	}

	if len(ws.buf) > 0 {
		n, err := ws.conn.Write(ws.buf)
		ws.buf = ws.buf[:0]
		if err != nil {
			return n, err
		}
	}
	n, e := ws.conn.Write(b)
	return n, e
}

func (c *Connection) ExecuteCommand() {
	state, err := c.ExecCmd.Process.Wait()
	if err != nil {
		return
	}

	logger.Debugf("", "Run exit code: %d", state.ExitCode())

	rmCommand := fmt.Sprintf("podman rm -f %s", c.Name)
	out, err := exec.Command("bash", "-c", rmCommand).CombinedOutput()
	logger.Debugf("", "Podman rm: %s", string(out))
	if err != nil {
		logger.Error("", err.Error())
	}

	rmiCommand := fmt.Sprintf("podman rmi -f %s", c.Image)
	out, err = exec.Command("bash", "-c", rmiCommand).CombinedOutput()
	logger.Debugf("", "Podman rmi: %s", string(out))
	if err != nil {
		logger.Error("", err.Error())
	}

	if state.ExitCode() != -1 {
		c.PTMX.Close()
		c.PTMX = nil
		c.ExecCmd = nil
	}

	c.Done = true
}

func PruneConnections() {
	for {
		toRemove := []int{}
		for idx, c := range connections {
			logger.Debugf("", "Runing container %s", c.Name)
			if c.Done {
				toRemove = append(toRemove, idx)
			}
		}

		sort.Sort(sort.Reverse(sort.IntSlice(toRemove)))

		for _, idx := range toRemove {
			connections = append(connections[:idx], connections[idx+1:]...)
		}
	}
}

func StartWSServer() {

	var serverErr error
	// r := http.NewServeMux()
	r := mux.NewRouter()
	// mux.Handle("/ws", websocket.Handler(run))
	r.HandleFunc("/ws/{cascade}/{run}/{version}",
		func(w http.ResponseWriter, req *http.Request) {
			c := Connection{}
			s := websocket.Server{Handler: websocket.Handler(c.run)}
			connections = append(connections, c)
			s.ServeHTTP(w, req)
		})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Config.WSPort),
		Handler: r,
	}

	go func() {
		log.Printf("running http://0.0.0.0:%d\n", config.Config.WSPort)

		if serverErr := server.ListenAndServe(); serverErr != nil {
			log.Println(serverErr)
		}
		fmt.Printf("Server exited!")
	}()

	// check process state
	// go func() {
	// 	ticker := time.NewTicker(time.Duration(checkProcInterval) * time.Second)
	// 	for range ticker.C {
	// 		if execCmd != nil {
	// 			state, err := execCmd.Process.Wait()
	// 			exec.Command("bash", "-c", rmiCommand).CombinedOutput()
	// 			if err != nil {
	// 				return
	// 			}

	// 			if state.ExitCode() != -1 {
	// 				ptmx.Close()
	// 				ptmx = nil
	// 				execCmd = nil
	// 			}
	// 		}
	// 	}
	// }()

	// wait for run server
	time.Sleep(time.Duration(waitTime) * time.Microsecond)

	if serverErr == nil {
		fmt.Println("Server ready")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit

	go PruneConnections()

	for _, c := range connections {
		if c.PTMX != nil {
			_ = c.PTMX.Close()
		}
		if c.ExecCmd != nil {
			_ = c.ExecCmd.Process.Kill()
			_, _ = c.ExecCmd.Process.Wait()
		}
	}
	if err := server.Shutdown(context.Background()); err != nil {
		log.Println("failed to shutdown server", err)
	}
	fmt.Println("Server has been shut down")
}

func filter(ss []string) []string {
	rs := []string{}

	for _, s := range ss {
		if s == "" {
			continue
		}
		rs = append(rs, s)
	}

	return rs
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v != "" {
		return v
	}
	return def
}

func windowSize(msg interface{}) (rows, cols uint16, err error) {
	data, ok := msg.(map[string]interface{})
	if !ok {
		return 0, 0, fmt.Errorf("invalid message: %#+v", msg)
	}

	rows = uint16(data["rows"].(float64))
	cols = uint16(data["cols"].(float64))

	return
}
