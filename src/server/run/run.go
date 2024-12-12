package run

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/datastore"
	"scaffold/server/filestore"
	"scaffold/server/msg"
	"scaffold/server/rabbitmq"
	"scaffold/server/state"
	"scaffold/server/task"
	"scaffold/server/utils"
	"strings"
	"time"

	logger "github.com/jfcarter2358/go-logger"
)

var runError error
var killed = false

type Run struct {
	Name    string            `json:"name" yaml:"name"`
	Task    task.Task         `json:"task" yaml:"task"`
	State   state.State       `json:"state" yaml:"state"`
	Number  int               `json:"number" yaml:"number"`
	Groups  []string          `json:"groups" yaml:"groups"`
	Worker  string            `json:"worker" yaml:"worker"`
	PID     int               `json:"pid" yaml:"pid"`
	Context map[string]string `json:"context" yaml:"context"`
	RunID   string            `json:"run_id" yaml:"run_id"`
}

type RunContext struct {
	Run         *Run
	DataStore   *datastore.DataStore
	RunDir      string
	ScriptPath  string
	EnvInPath   string
	EnvOutPath  string
	DisplayPath string
}

func setErrorStatus(r *Run, output string) {
	r.PID = 0
	r.State.PID = 0
	r.State.Output = output
	r.State.Status = constants.STATE_STATUS_ERROR
	currentTime := time.Now().UTC()
	r.State.Finished = currentTime.Format("2006-01-02T15:04:05Z")
}

func runCmd(cmd *exec.Cmd) {
	runError = cmd.Run()
}

func updateRunState(r *Run, send bool) error {
	r.State.PID = r.PID
	m := msg.RunMsg{
		Task:     r.Task.Name,
		Workflow: r.Task.Workflow,
		Status:   r.State.Status,
		Context:  r.Context,
		State:    r.State,
		RunID:    r.RunID,
	}
	logger.Debugf("", "Updating run state for %v", m)
	if err := state.UpdateStateRunByNames(r.State.Workflow, r.State.Task, r.State); err != nil {
		logger.Errorf("", "Cannot update state run: %s %s %v %s", r.Task.Workflow, r.Task.Name, r.State, err.Error())
		return err
	}
	if send {
		return rabbitmq.WorkerPublish(m)
	}
	return nil
}

func nukeDir(path string) {
	logger.Debugf("", "Removing directory %s", path)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			// file does not exist
		} else {
			os.RemoveAll(path)
		}
	}
}

func setupRun(rc *RunContext) (bool, error) {
	rc.Run.State.Status = constants.STATE_STATUS_RUNNING
	currentTime := time.Now().UTC()
	rc.Run.State.Started = currentTime.Format("2006-01-02T15:04:05Z")
	if err := state.UpdateStateKilledByNames(rc.Run.Task.Workflow, rc.Run.Task.Name, false); err != nil {
		logger.Infof("", "Cannot update state killed: %s %s %s", rc.Run.Task.Workflow, rc.Run.Task.Name, err.Error())
		return false, err
	}
	if err := updateRunState(rc.Run, false); err != nil {
		return false, err
	}

	rc.RunDir = fmt.Sprintf("/tmp/run/%s/%s/%d", rc.Run.State.Workflow, rc.Run.State.Task, rc.Run.Number)
	rc.ScriptPath = rc.RunDir + "/.run.sh"
	rc.EnvInPath = rc.RunDir + "/.envin"
	rc.EnvOutPath = rc.RunDir + ".envout"
	rc.DisplayPath = rc.RunDir + "/.display"

	// Setup run directory
	nukeDir(rc.RunDir)
	err := os.MkdirAll(rc.RunDir, 0755)
	if err != nil {
		logger.Errorf("", "Error creating run directory %s", err.Error())
		setErrorStatus(rc.Run, err.Error())
		if err := updateRunState(rc.Run, true); err != nil {
			return false, err
		}
		return false, err
	}

	// Setup datastore
	rc.DataStore, err = datastore.GetDataStoreByWorkflow(rc.Run.State.Workflow)
	if err != nil {
		logger.Errorf("", "Cannot get datastore %s", rc.Run.State.Workflow)
		setErrorStatus(rc.Run, err.Error())
		if err := updateRunState(rc.Run, true); err != nil {
			return false, err
		}
		return false, err
	}

	return false, nil
}

func setupEnvLoad(rc *RunContext) (bool, error) {
	envInput := ""
	for key, val := range rc.Run.Task.Inputs {
		dsVal, ok := rc.DataStore.Env[val]
		var encoded string
		if ok {
			encoded = base64.StdEncoding.EncodeToString([]byte(dsVal))
			envInput += fmt.Sprintf("%s;%s\n", key, encoded)
			continue
		}
		logger.Warnf("", "Input value missing for %s", val)
	}
	for key, val := range rc.Run.Context {
		encoded := base64.StdEncoding.EncodeToString([]byte(val))
		envInput += fmt.Sprintf("%s;%s\n", key, encoded)
		continue
	}
	for key, val := range rc.Run.Task.Env {
		encoded := base64.StdEncoding.EncodeToString([]byte(val))
		envInput += fmt.Sprintf("%s;%s\n", key, encoded)
	}

	// Write out envin script
	data := []byte(envInput)
	err := os.WriteFile(rc.EnvInPath, data, 0777)
	if err != nil {
		logger.Errorf("", "Error writing envin file %s", err.Error())
		setErrorStatus(rc.Run, err.Error())
		if err := updateRunState(rc.Run, true); err != nil {
			return false, err
		}
		return false, err
	}

	return false, nil
}

func setupRunScript(rc *RunContext) (bool, error) {
	envOutput := ""
	for _, key := range rc.Run.Task.Store.Env {
		envOutput += fmt.Sprintf("echo \"%s;$(echo \"${%s}\" | base64)\" >> %s\n", key, key, rc.EnvOutPath)
	}

	runScript := fmt.Sprintf(`
	# load ENV var

	while read -r line; do
		name=${line%%;*}
		value=${line#*;}
		export ${name}="$(echo "${value}" | base64 -d)"
	done < %s
	
	# run command
	%s
	
	# save ENV vars
	%s
	`, rc.EnvInPath, rc.Run.Task.Run, envOutput)

	// Write out our run script
	data := []byte(runScript)
	err := os.WriteFile(rc.ScriptPath, data, 0777)
	if err != nil {
		logger.Errorf("", "Error writing run file %s", err.Error())
		setErrorStatus(rc.Run, err.Error())
		if err := updateRunState(rc.Run, true); err != nil {
			return false, err
		}
		return false, err
	}

	return false, nil
}

func loadFiles(rc *RunContext) (bool, error) {
	for _, name := range rc.Run.Task.Load.File {
		err := filestore.GetFile(fmt.Sprintf("%s/%s", rc.Run.State.Workflow, name), fmt.Sprintf("%s/%s", rc.RunDir, name))
		if err != nil {
			logger.Errorf("", "Error getting file %s", err.Error())
			setErrorStatus(rc.Run, err.Error())
			if err := updateRunState(rc.Run, true); err != nil {
				return false, err
			}
			return false, err
		}
	}

	return false, nil
}

func checkDisplay(rc *RunContext) (bool, error) {
	logger.Tracef("", "Checking for display at %s", rc.DisplayPath)
	if _, err := os.Stat(rc.DisplayPath); err == nil {
		logger.Tracef("", "Display path is present")
		if data, err := os.ReadFile(rc.DisplayPath); err == nil {
			logger.Tracef("", "Read display file")
			var obj []map[string]interface{}
			if err := json.Unmarshal(data, &obj); err != nil {
				logger.Errorf("", "Error unmarshalling display JSON: %v", err)
			} else {
				logger.Tracef("", "Updating display object")
				rc.Run.State.Display = obj
			}
		} else {
			logger.Tracef("", "Display read error: %s", err.Error())
		}
	} else {
		logger.Tracef("", "Display stat error: %s", err.Error())
	}
	if err := updateRunState(rc.Run, false); err != nil {
		logger.Errorf("", "Error updating run: %s", err.Error())
		nukeDir(rc.RunDir)
		return false, err
	}
	return false, nil
}

func storeFiles(rc *RunContext) {
	for _, name := range rc.Run.Task.Store.File {
		filePath := fmt.Sprintf("%s/%s", rc.RunDir, name)
		if _, err := os.Stat(filePath); err == nil {
			err := filestore.UploadFile(filePath, fmt.Sprintf("%s/%s", rc.Run.Task.Workflow, name))
			if err != nil {
				logger.Errorf("", "Error uploading file %s: %s\n", fmt.Sprintf("%s/%s", rc.Run.Task.Workflow, name), err.Error())
			}
			rc.DataStore.Files = append(rc.DataStore.Files, name)
			rc.DataStore.Files = utils.RemoveDuplicateValues(rc.DataStore.Files)
		}
	}
}

func storeEnv(rc *RunContext) (bool, error) {
	logger.Tracef("", "Storing ENV values to context")
	var dat []byte
	if _, err := os.Stat(rc.EnvOutPath); err == nil {
		dat, err = os.ReadFile(rc.EnvOutPath)
		if err != nil {
			logger.Errorf("", "Error reading file %s\n", err.Error())
			setErrorStatus(rc.Run, err.Error())
			if err := updateRunState(rc.Run, true); err != nil {
				nukeDir(rc.RunDir)
				return false, err
			}
		}
	}
	envVarList := strings.Split(string(dat), "\n")
	logger.Tracef("", "Got env var list: %v", envVarList)
	envVarMap := map[string]string{}

	for _, val := range envVarList {
		name, val, _ := strings.Cut(val, ";")
		decoded, _ := base64.StdEncoding.DecodeString(val)
		envVarMap[name] = string(decoded)
		logger.Tracef("", "Got ENV var %s: %s", name, string(decoded))
	}

	for _, name := range rc.Run.Task.Store.Env {
		logger.Tracef("", "Storing %s to context", name)
		if rc.Run.Context == nil {
			rc.Run.Context = make(map[string]string)
		}
		rc.Run.Context[name] = envVarMap[name]
	}

	return false, nil
}

func setStatus(rc *RunContext, returnCodeString string, returnCodeInt int) {
	currentTime := time.Now().UTC()
	rc.Run.State.Finished = currentTime.Format("2006-01-02T15:04:05Z")
	if returnCodeString == "" {
		if returnCodeInt == 0 {
			rc.Run.State.Status = constants.STATE_STATUS_SUCCESS
		} else {
			rc.Run.State.Status = constants.STATE_STATUS_ERROR
		}
		return
	}
	if returnCodeString == "0" {
		rc.Run.State.Status = constants.STATE_STATUS_SUCCESS
	} else {
		rc.Run.State.Status = constants.STATE_STATUS_ERROR
	}
}

func StartContainerRun(rr *Run) (bool, error) {
	rc := &RunContext{
		Run: rr,
	}

	if shouldRestart, err := setupRun(rc); err != nil {
		return shouldRestart, err
	}

	if shouldRestart, err := setupEnvLoad(rc); err != nil {
		return shouldRestart, err
	}

	if shouldRestart, err := setupRunScript(rc); err != nil {
		return shouldRestart, err
	}

	if shouldRestart, err := loadFiles(rc); err != nil {
		return shouldRestart, err
	}

	containerName := fmt.Sprintf("%s-%s-%d", rc.Run.State.Workflow, rc.Run.State.Task, rc.Run.Number)

	// Clean up any possible artifacts
	if err := exec.Command("/bin/sh", "-c", fmt.Sprintf("podman kill %s", containerName)).Run(); err != nil {
		logger.Infof("", "No running container with name %s exists, skipping kill\n", containerName)
	}
	if err := exec.Command("/bin/sh", "-c", fmt.Sprintf("podman rm %s", containerName)).Run(); err != nil {
		logger.Infof("", "No running container with name %s exists, skipping removal\n", containerName)
	}

	if rc.Run.Task.ContainerLoginCommand != "" {
		logger.Debugf("", "Logging into registry with command %s", rc.Run.Task.ContainerLoginCommand)
		if err := exec.Command("/bin/sh", "-c", rc.Run.Task.ContainerLoginCommand).Run(); err != nil {
			logger.Infof("", "Cannot login to container registry: %s\n", err.Error())
		}
	}

	podmanCommand := fmt.Sprintf("podman run --rm --privileged -d %s --device /dev/net/tun:/dev/net/tun ", config.Config.PodmanOpts)
	podmanCommand += fmt.Sprintf("--name %s ", containerName)
	podmanCommand += fmt.Sprintf("--mount type=bind,src=%s,dst=/tmp/run ", rc.RunDir)
	for _, m := range rc.Run.Task.Load.Mounts {
		podmanCommand += fmt.Sprintf("--mount type=bind,src=%s,dst=%s ", m, m)
	}
	for _, e := range rc.Run.Task.Load.EnvPassthrough {
		podmanCommand += fmt.Sprintf("--env %s=\"${%s}\" ", e, e)
	}
	podmanCommand += rc.Run.Task.Image
	podmanCommand += " bash -c /tmp/run/.run.sh"

	logger.Debugf("", "command: %s", podmanCommand)

	cmd := exec.Command("/bin/sh", "-c", podmanCommand)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	go runCmd(cmd)

	output, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("podman ps -a --filter \"name=%s\" --format \"{{.Status}}\"", containerName)).CombinedOutput()
	if err != nil {
		logger.Errorf("", "Encountered error: %v\n", err.Error())
		logger.Debugf("", "STDOUT: %s\n", string(output))

		shouldRestart := false
		if strings.Contains(string(output), "no space left on device") {
			shouldRestart = true
			logs, err := exec.Command("/bin/sh", "-c", "podman system prune -a -f").CombinedOutput()
			if err != nil {
				logger.Errorf("", "Prune error string: %s", err.Error())
			}
			logger.Debugf("", "Prune output: %s", logs)
		}
		setErrorStatus(rc.Run, string(output))
		if err := updateRunState(rc.Run, true); err != nil {
			nukeDir(rc.RunDir)
			return false, err
		}
		nukeDir(rc.RunDir)
		return shouldRestart, err
	}

	var podmanOutput string
	erroredOut := false
	for !strings.HasPrefix(string(output), "Exited") {
		logger.Debugf("", "Checking for exit status: %s", string(output))
		if string(output) == "" {
			podmanOutput = outb.String() + "\n\n" + errb.String()
			rc.Run.State.Output = podmanOutput
			if runError != nil {
				logger.Errorf("", "Error running pod %s\n", runError.Error())
				setErrorStatus(rc.Run, fmt.Sprintf("%s :: %s", podmanOutput, string(runError.Error())))
				if err := updateRunState(rc.Run, true); err != nil {
					nukeDir(rc.RunDir)
					return false, err
				}
				erroredOut = true
				break
			}
			// Load in display file if present and able
			if shouldRestart, err := checkDisplay(rc); err != nil {
				return shouldRestart, err
			}
		} else {
			logs, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("podman logs %s", containerName)).CombinedOutput()
			if err != nil {
				rc.Run.State.Output = fmt.Sprintf("%s\n\n--------------------------------\n\n%s--------------------------------\n\n%s", podmanOutput, logs, string(err.Error()))
			} else {
				rc.Run.State.Output = fmt.Sprintf("%s\n\n--------------------------------\n\n%s", podmanOutput, string(logs))
			}
			// Load in display file if present and able
			if shouldRestart, err := checkDisplay(rc); err != nil {
				return shouldRestart, err
			}
		}
		time.Sleep(500 * time.Millisecond)
		output, _ = exec.Command("/bin/sh", "-c", fmt.Sprintf("podman ps -a --filter \"name=%s\" --format \"{{.Status}}\"", containerName)).CombinedOutput()
	}

	if !erroredOut {
		openParenIdx := strings.Index(string(output), "(")
		closeParenIdx := strings.Index(string(output), ")")
		returnCode := string(output)[openParenIdx+1 : closeParenIdx]

		logs, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("podman logs %s", containerName)).CombinedOutput()
		if err != nil {
			rc.Run.State.Output = fmt.Sprintf("%s\n\n--------------------------------\n\n%s", podmanOutput, string(err.Error()))
		} else {
			rc.Run.State.Output = fmt.Sprintf("%s\n\n--------------------------------\n\n%s", podmanOutput, string(logs))
		}

		storeFiles(rc)

		if shouldRestart, err := checkDisplay(rc); err != nil {
			return shouldRestart, err
		}

		if shouldRestart, err := storeEnv(rc); err != nil {
			return shouldRestart, err
		}

		setStatus(rc, returnCode, 0)

		if rc.Run.Task.ShouldRM {
			rmCommand := fmt.Sprintf("podman rm -f %s", containerName)
			out, err := exec.Command("bash", "-c", rmCommand).CombinedOutput()
			logger.Debugf("", "Podman rm: %s", string(out))
			rc.Run.State.Output += fmt.Sprintf("\n\n--------------------------------\n\n%s", string(out))
			if err != nil {
				logger.Error("", err.Error())
				rc.Run.State.Output += fmt.Sprintf("\n\n--------------------------------\n\n%s", err.Error())
			}
		}
	}

	nukeDir(rc.RunDir)
	if err := updateRunState(rc.Run, true); err != nil {
		return false, err
	}
	return false, nil
}

func ContainerKill(cn, tn string) error {
	prefix := fmt.Sprintf("%s.%s", cn, tn)
	logger.Debugf("", "Trying to kill %s", prefix)
	output, err := exec.Command("/bin/sh", "-c", "podman ps -a --format \"{{.Names}}\"").CombinedOutput()
	if err != nil {
		logger.Infof("", "Unable to list running containers: %s | %s", err, string(output))
		return err
	}
	logger.Tracef("", "Container output: %s", string(output))
	lines := strings.Split(string(output), "\n")

	logger.Tracef("", "Checking prefix %s", prefix)
	for _, containerName := range lines {
		logger.Tracef("", "Checking container %s", containerName)
		if strings.HasPrefix(containerName, prefix) {
			logger.Infof("", "Killing container %s", containerName)
			if output, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("podman kill %s", containerName)).CombinedOutput(); err != nil {
				logger.Infof("", "Cannot kill container with name %s with output %s", containerName, output)
				return err
			}
			s, err := state.GetStateByNames(cn, tn)
			if err != nil {
				logger.Errorf("", "Unable to get state for run %s.%s with error %s", cn, tn, err.Error())
				return err
			}
			s.Status = constants.STATE_STATUS_KILLED
			if err := state.UpdateStateByNames(cn, tn, s); err != nil {
				logger.Errorf("", "Unable to update state %s.%s: %s", cn, tn, err.Error())
				return err
			}
		}
	}
	return nil
}

func ExitCode(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	// No error
	return 0
}

func StartLocalRun(rr *Run) (bool, error) {
	rc := &RunContext{
		Run: rr,
	}

	if shouldRestart, err := setupRun(rc); err != nil {
		return shouldRestart, err
	}

	if shouldRestart, err := setupEnvLoad(rc); err != nil {
		return shouldRestart, err
	}

	if shouldRestart, err := setupRunScript(rc); err != nil {
		return shouldRestart, err
	}

	if shouldRestart, err := loadFiles(rc); err != nil {
		return shouldRestart, err
	}

	if rc.Run.PID > 0 {
		if err := exec.Command("/bin/sh", "-c", fmt.Sprintf("kill %d", rc.Run.PID)).Run(); err != nil {
			logger.Infof("", "Cannot kill existing run: %s\n", err.Error())
		}
	}

	localCommand := fmt.Sprintf("cd %s && ./.run.sh", rc.RunDir)

	logger.Debugf("", "command: %s", localCommand)

	cmd := exec.Command("/bin/bash", "-c", localCommand)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	runError = cmd.Start()
	rc.Run.PID = cmd.Process.Pid
	if err := updateRunState(rc.Run, true); err != nil {
		return false, err
	}

	killed = false

	go func() {
		runError = cmd.Wait()
		killed = true
	}()

	for !killed {
		if runError != nil {
			logger.Errorf("", "Error running pod %s\n", runError.Error())
			setErrorStatus(rc.Run, fmt.Sprintf("Error running pod %s\n", runError.Error()))
			if err := updateRunState(rc.Run, true); err != nil {
				return false, err
			}
		}
		output := outb.String() + "\n\n" + errb.String()
		logger.Tracef("", "setting output 1 %s", output)
		rc.Run.State.Output = output

		// Load in display file if present and able
		if shouldRestart, err := checkDisplay(rc); err != nil {
			return shouldRestart, err
		}

		if shouldRestart, err := storeEnv(rc); err != nil {
			return shouldRestart, err
		}

		time.Sleep(500 * time.Millisecond)
	}

	output := outb.String() + "\n\n" + errb.String()
	logger.Tracef("", "setting output 2 %s", output)
	rc.Run.State.Output = output

	if err := updateRunState(rc.Run, false); err != nil {
		logger.Errorf("", "Error updating run: %s", err.Error())
		return false, err
	}

	returnCode := 0
	returnCode = ExitCode(runError)

	storeFiles(rc)

	// Load in display file if present and able
	if shouldRestart, err := checkDisplay(rc); err != nil {
		return shouldRestart, err
	}

	if shouldRestart, err := storeEnv(rc); err != nil {
		return shouldRestart, err
	}

	setStatus(rc, "", returnCode)

	rc.Run.PID = 0

	err := updateRunState(rc.Run, true)
	return false, err
}

func LocalKill(cn, tn string) error {
	s, err := state.GetStateByNames(cn, tn)
	if err != nil {
		return err
	}
	logger.Infof("", "Killing run %s/%s with PID %d", cn, tn, s.PID)
	if s.PID > 0 {
		if out, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("sudo kill %d", s.PID)).CombinedOutput(); err != nil {
			logger.Errorf("", "Cannot kill existing run: %s\n", err.Error())
			if out != nil {
				logger.Tracef("", "Kill output: %s", string(out))
			}
			s.Status = constants.STATE_STATUS_KILLED
			if err := state.UpdateStateByNames(cn, tn, s); err != nil {
				logger.Errorf("", "Unable to update state %s.%s: %s", cn, tn, err.Error())
				return err
			}
			return err
		}
		s.Status = constants.STATE_STATUS_KILLED
		if err := state.UpdateStateByNames(cn, tn, s); err != nil {
			logger.Errorf("", "Unable to update state %s.%s: %s", cn, tn, err.Error())
			return err
		}
	}
	return nil
}
