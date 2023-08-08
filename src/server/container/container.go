package container

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"scaffold/server/logger"
	"scaffold/server/run"
	"scaffold/server/user"
	"strings"
	"time"
)

var LastRun []string
var LastImage []string
var CompletedRuns map[string]run.Run
var CurrentRun run.Run
var CurrentName string
var MaxAllowed = 10

type Container struct {
	// Stdin  *os.File
	Stdin       io.WriteCloser
	Stdout      io.ReadCloser
	Name        string
	Offset      int
	User        user.User
	Cmd         *exec.Cmd
	Error       string
	InputReady  bool
	Input       string
	OutputReady bool
	Output      string
}

func InitContainers() {
	LastRun = make([]string, 0)
	LastImage = make([]string, 0)
	CurrentName = ""
}

func PruneContainers() {
	for {
		for len(LastRun) > MaxAllowed {
			toDestroy := ""
			toDestroy, LastRun = LastRun[0], LastRun[1:]
			LastImage = LastImage[1:]

			parts := strings.Split(toDestroy, ".")

			containerName := fmt.Sprintf("%s-%s", parts[0], parts[1])
			if containerName != CurrentName {
				logs, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("podman rm %s", containerName)).CombinedOutput()
				if err != nil {
					logger.Errorf("", "Remove error string: %s", err.Error())
				}
				logger.Debugf("", "Remove output: %s", logs)
				runDir := fmt.Sprintf("/tmp/run/%s/%s/%s", parts[0], parts[1], parts[2])
				if err := os.RemoveAll(runDir); err != nil {
					logger.Errorf("", "Delete error string: %s", err.Error())
				}
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func (c *Container) Stop() error {
	c.Stdin.Close()
	if err := c.Cmd.Process.Kill(); err != nil {
		return err
	}
	return nil
}

func SetupContainer(name string) (Container, error) {
	parts := strings.Split(name, ".")
	dataDir := fmt.Sprintf("/tmp/data/%s/%s/%s", parts[0], parts[1], parts[2])
	containerName := fmt.Sprintf("%s-%s-%s", parts[0], parts[1], parts[2])

	os.MkdirAll(dataDir, 0755)

	c := Container{Name: containerName, Offset: 0}

	// stdinPath := fmt.Sprintf("%s/stdin", dataDir)
	// stdin, err := os.Create(stdinPath)
	// if err != nil {
	// 	return Container{}, err
	// }
	// c.Stdin = stdin

	stdoutPath := fmt.Sprintf("%s/stdout", dataDir)
	stdout, err := os.Create(stdoutPath)
	if err != nil {
		return Container{}, err
	}
	c.Stdout = stdout

	return c, nil
}

func (c *Container) ExecContainer(name string) {
	parts := strings.Split(name, ".")
	runDir := fmt.Sprintf("/tmp/run/%s/%s/%s", parts[0], parts[1], parts[2])
	containerName := fmt.Sprintf("%s-%s-%s", parts[0], parts[1], parts[2])
	podmanCommand := fmt.Sprintf("podman commit %s %s/%s && ", containerName, c.User.Username, containerName)
	podmanCommand += "podman run --privileged --security-opt label=disabled -it "
	podmanCommand += fmt.Sprintf("--mount type=bind,src=%s,dst=/tmp/run ", runDir)
	podmanCommand += fmt.Sprintf("%s/%s ", c.User.Username, containerName)
	podmanCommand += "sh"
	logger.Debugf("", "command: %s", podmanCommand)
	c.Cmd = exec.Command("/bin/sh", "-c", podmanCommand)
	// cmd.Stdin = c.Stdin
	var err error
	c.Stdin, err = c.Cmd.StdinPipe()
	if err != nil {
		c.Error = err.Error()
		return
	}
	// defer stdin.Close()
	// c.Cmd.Stdout = c.Stdout
	// c.Cmd.Stderr = c.Stdout

	c.Stdout, err = c.Cmd.StdoutPipe()
	if err != nil {
		c.Error = err.Error()
		return
	}

	c.Cmd.Stderr = c.Cmd.Stdout

	go c.Cmd.Run()

	for {
		if c.InputReady {
			if _, errWrite := io.WriteString(c.Stdin, fmt.Sprintf("%s\n", c.Input)); errWrite != nil {
				c.Error = errWrite.Error()
				break
			}
			c.InputReady = false
		}
		if c.OutputReady {
			data, errRead := io.ReadAll(c.Stdout)
			if err != nil {
				c.Error = errRead.Error()
				break
			}
			c.Output += string(data)
		}
	}
	c.Error = err.Error()
}

func (c *Container) Write(data string) (string, int) {
	if _, err := io.WriteString(c.Stdin, fmt.Sprintf("%s\n", data)); err != nil {
		return err.Error(), 500
	}
	return "", 200
}

func (c *Container) Read() (string, int) {
	stdout := ""
	// c.Stdout.Seek(int64(c.Offset), 0)
	length := 0
	buf := make([]byte, 1024)
	for {
		n, err := c.Stdout.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			// c.Stdout.Seek(0, 0)
			return err.Error(), 500
		}
		if n > 0 {
			stdout += string(buf[:n])
			length += n
		}
	}
	// c.Stdout.Seek(0, 0)
	c.Offset = c.Offset + length

	return stdout, 200
}

// func sliceIndex(list []string, val string) int {
// 	for i := 0; i < len(list); i++ {
// 		if list[i] == val {
// 			return i
// 		}
// 	}
// 	return -1
// }
