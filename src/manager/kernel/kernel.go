package kernel

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"scaffold/manager/config"
	"strings"
	"time"

	"github.com/google/uuid"
	logger "github.com/jfcarter2358/go-logger"
)

var Kernels = make(map[string]*Kernel)

//go:embed kernel.py
var pythonKernel string

type Kernel struct {
	ID                string
	Shell             *exec.Cmd
	Stdout            string
	PreviousStdout    string
	RunID             string
	PreviousRunID     string
	Stdin             io.WriteCloser
	StdoutLen         int
	PreviousStdoutLen int
	Channel           chan (string)
	Requirements      string
	Ready             bool
	Running           bool
	Created           string
	Updated           string
	Errored           bool
	Killed            bool
}

func (k *Kernel) Spawn() error {
	id := uuid.New().String()

	k.ID = id
	k.Shell = nil
	k.Channel = make(chan (string))

	currentTime := time.Now().UTC()
	k.Created = currentTime.Format("2006-01-02T15:04:05Z")
	k.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	logger.Debugf("", "Setup kernel arguments")

	if err := k.setupVenv(); err != nil {
		logger.Errorf("", "Error setting up kernel venv")
		return err
	}

	go k.runProcess()

	k.Ready = true

	return nil
}

func (k *Kernel) Kill() error {
	k.Ready = false

	return k.Shell.Process.Kill()
}

func (k *Kernel) setupVenv() error {
	workDir := fmt.Sprintf("%s/%s", config.Config.RunbookRunDir, k.ID)
	os.MkdirAll(config.Config.RunbookVenvDir, os.ModePerm)
	os.MkdirAll(workDir, os.ModePerm)

	command := fmt.Sprintf("python -m venv %s/%s", config.Config.RunbookVenvDir, k.ID)
	cmd := exec.Command(
		"/bin/bash",
		"-c",
		command,
	)
	cmd.Dir = workDir

	logger.Debugf("", "Setting up venv")
	if out, err := cmd.CombinedOutput(); err != nil {
		logger.Error("", string(out))
		return err
	}
	logger.Debugf("", "Writing kernel file")
	if err := os.WriteFile(fmt.Sprintf("%s/main.py", workDir), []byte(pythonKernel), 0644); err != nil {
		return err
	}
	if k.Requirements != "" {
		logger.Debugf("", "Writing requirements file")
		if err := os.WriteFile(fmt.Sprintf("%s/requirements.txt", workDir), []byte(k.Requirements), 0644); err != nil {
			return err
		}
		logger.Debugf("", "Installing requirements")
		command := "pip install -r requirements.txt"
		cmd := exec.Command(
			"/bin/bash",
			"-c",
			command,
		)
		cmd.Dir = workDir
		if out, err := cmd.CombinedOutput(); err != nil {
			logger.Error("", string(out))
			return err
		}
	}
	return nil
}

func (k *Kernel) GetOutput() string {
	logger.Tracef("", "Stdout length: %d", k.StdoutLen)
	logger.Tracef("", "Stdout: %s", k.Stdout)
	data := k.Stdout[k.StdoutLen:]
	data = strings.TrimSuffix(data, "\n")
	data = strings.TrimSuffix(data, "marathon::execute_done")
	return data
}

func (k *Kernel) runProcess() {
	k.Running = false
	currentTime := time.Now().UTC()
	k.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	var err error
	workDir := fmt.Sprintf("%s/%s", config.Config.RunbookRunDir, k.ID)

	k.Shell = exec.Command(
		"/bin/bash",
		"-c",
		fmt.Sprintf("source %s/%s/bin/activate; python main.py", config.Config.RunbookVenvDir, k.ID),
	)
	k.Shell.Dir = workDir

	logger.Debugf("", "Created execution command")

	k.Stdin, err = k.Shell.StdinPipe()
	if err != nil {
		logger.Errorf("", "Error getting cmd stdin pipe: %s", err.Error())
		k.Channel <- "marathon::setup_error"
		return
	}
	stdout := bytes.Buffer{}
	k.Shell.Stdout = &stdout
	k.Shell.Stderr = &stdout

	logger.Debugf("", "buffers setup")

	k.Shell.Start()

	for {
		script := <-k.Channel
		k.Running = true
		k.Errored = false
		k.Killed = false
		k.StdoutLen = len(k.Stdout)
		logger.Debugf("", "Got script: %s", script)
		if script == "marathon::kernel_kill" {
			logger.Debugf("", "Got kernel kill")
			k.Running = false
			k.Killed = true
			break
		}
		if strings.HasPrefix(script, "marathon::lang_python") {
			script = fmt.Sprintf("%s\n\nprint('marathon::execute_done')\nmarathon::script_end\n", script)
		} else if strings.HasPrefix(script, "marathon::lang_bash") {
			script = fmt.Sprintf("%s\n\necho 'marathon::execute_done'\nmarathon::script_end\n", script)
		}
		logger.Debugf("", "Running script %s", script)

		stdout.Reset()
		_, err := io.WriteString(k.Stdin, script)
		if err != nil {
			logger.Errorf("", "Error writing to stdin: %s", err.Error())
			k.Running = false
			k.Errored = true
			// k.Channel <- "marathon::write_error"
			return
		}
		// temp := stdout.String()
		// logger.Debugf("", "Got temp of %s", temp)
		// if temp != "<nil>" {
		// 	k.Stdout = temp
		// }
		k.Stdout = ""
		for !strings.HasSuffix(strings.TrimSuffix(k.Stdout, "\n"), "marathon::execute_done") {
			logger.Debug("", "Execution not finished")
			logger.Debugf("", "Stdout: %s", k.Stdout)
			if strings.HasSuffix(strings.TrimSuffix(k.Stdout, "\n"), "marathon::exception") {
				logger.Errorf("", "Encountered exception in kernel execution")
				k.Running = false
				k.Errored = true
				// k.Channel <- "marathon::exception"
				return
			}
			temp := stdout.String()
			logger.Debugf("", "Got temp of %s", temp)
			if temp != "<nil>" {
				k.Stdout += temp
			}
			time.Sleep(1000 * time.Millisecond)
		}

		logger.Debug("", "Finsihed!")
		logger.Debugf("", "Stdout: %s", k.Stdout)
		// logger.Debugf("", "Returning execute_done")
		k.Running = false
		// k.Channel <- "marathon::execute_done"
	}
	k.Running = false
	// k.Channel <- "marathon::kernel_shutdown"
}
