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
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	_ "embed"

	"github.com/creack/pty"
	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"
)

// run command
var command string = getenv("SHELL", "bash")

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

var ptmx *os.File
var execCmd *exec.Cmd

func run(ws *websocket.Conn) {

	vars := mux.Vars(ws.Request())
	fmt.Printf("cascade: %s\n", vars["cascade"])
	fmt.Printf("run: %s\n", vars["run"])
	fmt.Printf("version: %s\n", vars["version"])

	// parts := strings.Split(name, ".")
	// runDir := fmt.Sprintf("/tmp/run/%s/%s/%s", parts[0], parts[1], parts[2])
	// containerName := fmt.Sprintf("%s-%s-%s", parts[0], parts[1], parts[2])
	// podmanCommand := fmt.Sprintf("podman commit %s %s/%s && ", containerName, c.User.Username, containerName)
	// podmanCommand += "podman run --privileged --security-opt label=disabled -it "
	// podmanCommand += fmt.Sprintf("--mount type=bind,src=%s,dst=/tmp/run ", runDir)
	// podmanCommand += fmt.Sprintf("%s/%s ", c.User.Username, containerName)
	// podmanCommand += "sh"

	defer ws.Close()

	wsconn := &wsConn{
		conn: ws,
	}

	if ptmx == nil {
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

		c := filter(strings.Split(command, " "))
		if len(c) > 1 {
			//nolint
			execCmd = exec.Command(c[0], c[1:]...)
		} else {
			//nolint
			execCmd = exec.Command(c[0])
		}

		ptmx, err = pty.StartWithSize(execCmd, winsize)
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

				if err := pty.Setsize(ptmx, winsize); err != nil {
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

			fmt.Printf("message data: %s", data)

			_, err := ptmx.WriteString(data)
			if err != nil {
				log.Println("failed to write data to ptmx:", err)
				return
			}
		}
	}()

	_, _ = io.Copy(wsconn, ptmx)
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

func StartWSServer(args []string) {
	if len(args) > 0 {
		command = args[0]
	}

	var serverErr error
	// r := http.NewServeMux()
	r := mux.NewRouter()
	// mux.Handle("/ws", websocket.Handler(run))
	r.HandleFunc("/ws/{cascade}/{run}/{version}",
		func(w http.ResponseWriter, req *http.Request) {
			s := websocket.Server{Handler: websocket.Handler(run)}
			s.ServeHTTP(w, req)
		})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Config.WSPort),
		Handler: r,
	}

	go func() {
		log.Println("running command: " + command)
		log.Printf("running http://0.0.0.0:%d\n", config.Config.WSPort)

		if serverErr := server.ListenAndServe(); serverErr != nil {
			log.Println(serverErr)
		}
		fmt.Printf("Server exited!")
	}()

	// check process state
	go func() {
		ticker := time.NewTicker(time.Duration(checkProcInterval) * time.Second)
		for range ticker.C {
			if execCmd != nil {
				state, err := execCmd.Process.Wait()
				if err != nil {
					return
				}

				if state.ExitCode() != -1 {
					ptmx.Close()
					ptmx = nil
					execCmd = nil
				}
			}
		}
	}()

	// wait for run server
	time.Sleep(time.Duration(waitTime) * time.Microsecond)

	if serverErr == nil {
		fmt.Println("Server ready")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit

	if ptmx != nil {
		_ = ptmx.Close()
	}
	if execCmd != nil {
		_ = execCmd.Process.Kill()
		_, _ = execCmd.Process.Wait()
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
