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

// type CheckRun struct {
// 	Enabled bool           `json:"disabled" yaml:"disabled"`
// 	Name    string         `json:"name" yaml:"name"`
// 	Task    task.TaskCheck `json:"task" yaml:"task"`
// 	State   state.State    `json:"state" yaml:"state"`
// }

type Run struct {
	Name    string            `json:"name" yaml:"name"`
	Task    task.Task         `json:"task" yaml:"task"`
	State   state.State       `json:"state" yaml:"state"`
	Number  int               `json:"number" yaml:"number"`
	Groups  []string          `json:"groups" yaml:"groups"`
	Worker  string            `json:"worker" yaml:"worker"`
	PID     int               `json:"pid" yaml:"pid"`
	Context map[string]string `json:"context" yaml:"context"`
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
		Task:    r.Task.Name,
		Cascade: r.Task.Cascade,
		Status:  r.State.Status,
		Context: r.Context,
	}
	logger.Debugf("", "Updating run state for %v", m)
	if err := state.UpdateStateRunByNames(r.State.Cascade, r.State.Task, r.State); err != nil {
		logger.Errorf("", "Cannot update state run: %s %s %v %s", r.Task.Cascade, r.Task.Name, r.State, err.Error())
		return err
	}
	if send {
		return rabbitmq.WorkerPublish(m)
	}
	return nil
}

func StartContainerRun(r *Run) (bool, error) {
	r.State.Status = constants.STATE_STATUS_RUNNING
	currentTime := time.Now().UTC()
	r.State.Started = currentTime.Format("2006-01-02T15:04:05Z")
	if err := state.UpdateStateKilledByNames(r.Task.Cascade, r.Task.Name, false); err != nil {
		logger.Infof("", "Cannot update state killed: %s %s %s", r.Task.Cascade, r.Task.Name, err.Error())
		return false, err
	}
	if err := updateRunState(r, false); err != nil {
		return false, err
	}

	cName := r.State.Cascade
	tName := r.State.Task

	logger.Debugf("", "Getting datastore by name %s", cName)
	ds, err := datastore.GetDataStoreByCascade(cName)
	if err != nil {
		logger.Errorf("", "Cannot get datastore %s", cName)
		setErrorStatus(r, err.Error())
		if err := updateRunState(r, true); err != nil {
			return false, err
		}
		return false, err
	}

	// containerName := strings.Replace(r.Name, ".", "-", -1)
	containerName := fmt.Sprintf("%s-%s-%d", cName, tName, r.Number)

	runDir := fmt.Sprintf("/tmp/run/%s/%s/%d", cName, tName, r.Number)

	if _, err := os.Stat(runDir); err != nil {
		if os.IsNotExist(err) {
			// file does not exist
		} else {
			os.RemoveAll(runDir)
		}
	}

	err = os.MkdirAll(runDir, 0755)
	if err != nil {
		logger.Errorf("", "Error creating run directory %s", err.Error())
		setErrorStatus(r, err.Error())
		if err := updateRunState(r, true); err != nil {
			return false, err
		}
		return false, err
	}

	scriptPath := runDir + "/.run.sh"
	envInPath := runDir + "/.envin"
	displayPath := runDir + "/.display"

	envInput := ""
	for key, val := range r.Task.Inputs {
		contextVal, ok := r.Context[val]
		var encoded string
		if ok {
			encoded = base64.StdEncoding.EncodeToString([]byte(contextVal))
		} else {
			encoded = base64.StdEncoding.EncodeToString([]byte(ds.Env[val]))
		}
		envInput += fmt.Sprintf("%s;%s\n", key, encoded)
	}
	for _, key := range r.Task.Load.Env {
		contextVal, ok := r.Context[key]
		var encoded string
		if ok {
			encoded = base64.StdEncoding.EncodeToString([]byte(contextVal))
		} else {
			encoded = base64.StdEncoding.EncodeToString([]byte(ds.Env[key]))
		}
		envInput += fmt.Sprintf("%s;%s\n", key, encoded)
	}
	for key, val := range r.Task.Env {
		encoded := base64.StdEncoding.EncodeToString([]byte(val))
		envInput += fmt.Sprintf("%s;%s\n", key, encoded)
	}

	envOutput := ""
	for _, key := range r.Task.Store.Env {
		envOutput += fmt.Sprintf("echo \"%s;$(echo \"${%s}\" | base64)\" >> /tmp/run/.envout\n", key, key)
	}

	runScript := fmt.Sprintf(`
	# load ENV var

	while read -r line; do
		name=${line%%;*}
		value=${line#*;}
		export ${name}="$(echo "${value}" | base64 -d)"
	done < /tmp/run/.envin
	
	# run command
	%s
	
	# save ENV vars
	%s
	`, r.Task.Run, envOutput)

	// Write out our run script
	data := []byte(runScript)
	err = os.WriteFile(scriptPath, data, 0777)
	if err != nil {
		logger.Errorf("", "Error writing run file %s", err.Error())
		setErrorStatus(r, err.Error())
		if err := updateRunState(r, true); err != nil {
			return false, err
		}
		return false, err
	}

	// Write out envin script
	data = []byte(envInput)
	err = os.WriteFile(envInPath, data, 0777)
	if err != nil {
		logger.Errorf("", "Error writing envin file %s", err.Error())
		setErrorStatus(r, err.Error())
		if err := updateRunState(r, true); err != nil {
			return false, err
		}
		return false, err
	}

	for _, name := range r.Task.Load.File {
		err := filestore.GetFile(fmt.Sprintf("%s/%s", cName, name), fmt.Sprintf("%s/%s", runDir, name))
		if err != nil {
			logger.Errorf("", "Error getting file %s", err.Error())
			setErrorStatus(r, err.Error())
			if err := updateRunState(r, true); err != nil {
				return false, err
			}
			return false, err
		}
	}

	// Clean up any possible artifacts
	if err := exec.Command("/bin/sh", "-c", fmt.Sprintf("podman kill %s", containerName)).Run(); err != nil {
		logger.Infof("", "No running container with name %s exists, skipping kill\n", containerName)
	}
	if err := exec.Command("/bin/sh", "-c", fmt.Sprintf("podman rm %s", containerName)).Run(); err != nil {
		logger.Infof("", "No running container with name %s exists, skipping removal\n", containerName)
	}

	if r.Task.ContainerLoginCommand != "" {
		logger.Debugf("", "Logging into registry with command %s", r.Task.ContainerLoginCommand)
		if err := exec.Command("/bin/sh", "-c", r.Task.ContainerLoginCommand).Run(); err != nil {
			logger.Infof("", "Cannot login to container registry: %s\n", err.Error())
		}
	}

	podmanCommand := fmt.Sprintf("podman run --rm --privileged -d %s --device /dev/net/tun:/dev/net/tun ", config.Config.PodmanOpts)

	podmanCommand += fmt.Sprintf("--name %s ", containerName)
	podmanCommand += fmt.Sprintf("--mount type=bind,src=%s,dst=/tmp/run ", runDir)
	for _, m := range r.Task.Load.Mounts {
		podmanCommand += fmt.Sprintf("--mount type=bind,src=%s,dst=%s ", m, m)
	}
	for _, e := range r.Task.Load.EnvPassthrough {
		podmanCommand += fmt.Sprintf("--env %s=\"${%s}\" ", e, e)
	}
	podmanCommand += r.Task.Image
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
		setErrorStatus(r, string(output))
		if err := updateRunState(r, true); err != nil {
			if _, err := os.Stat(runDir); err != nil {
				if os.IsNotExist(err) {
					// file does not exist
				} else {
					os.RemoveAll(runDir)
				}
			}
			return false, err
		}
		if _, err := os.Stat(runDir); err != nil {
			if os.IsNotExist(err) {
				// file does not exist
			} else {
				os.RemoveAll(runDir)
			}
		}
		return shouldRestart, err
	}

	var podmanOutput string
	erroredOut := false
	for !strings.HasPrefix(string(output), "Exited") {
		logger.Debugf("", "Checking for exit status: %s", string(output))
		if string(output) == "" {
			podmanOutput = outb.String() + "\n\n" + errb.String()
			r.State.Output = podmanOutput
			if runError != nil {
				logger.Errorf("", "Error running pod %s\n", runError.Error())
				setErrorStatus(r, fmt.Sprintf("%s :: %s", podmanOutput, string(runError.Error())))
				if err := updateRunState(r, true); err != nil {
					if _, err := os.Stat(runDir); err != nil {
						if os.IsNotExist(err) {
							// file does not exist
						} else {
							os.RemoveAll(runDir)
						}
					}
					return false, err
				}
				erroredOut = true
				break
			}
			// Load in display file if present and able
			logger.Debugf("", "Display path: %s", displayPath)
			if _, err = os.Stat(displayPath); err == nil {
				logger.Tracef("", "Display path is present")
				if data, err := os.ReadFile(displayPath); err == nil {
					logger.Tracef("", "Read display file")
					var obj []map[string]interface{}
					if err := json.Unmarshal(data, &obj); err != nil {
						logger.Errorf("", "Error unmarshalling display JSON: %v", err)
					} else {
						logger.Tracef("", "Updating display object")
						r.State.Display = obj
					}
				} else {
					logger.Errorf("", "read error: %s", err.Error())
				}
			} else {
				logger.Errorf("", "stat error: %s", err.Error())
			}
			if err := updateRunState(r, false); err != nil {
				if _, err := os.Stat(runDir); err != nil {
					if os.IsNotExist(err) {
						// file does not exist
					} else {
						os.RemoveAll(runDir)
					}
				}
				return false, err
			}
		} else {
			logs, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("podman logs %s", containerName)).CombinedOutput()
			if err != nil {
				r.State.Output = fmt.Sprintf("%s\n\n--------------------------------\n\n%s--------------------------------\n\n%s", podmanOutput, logs, string(err.Error()))
			} else {
				r.State.Output = fmt.Sprintf("%s\n\n--------------------------------\n\n%s", podmanOutput, string(logs))
			}
			// Load in display file if present and able
			logger.Tracef("", "Checking for display at %s", displayPath)
			if _, err = os.Stat(displayPath); err == nil {
				logger.Tracef("", "Display path is present")
				if data, err := os.ReadFile(displayPath); err == nil {
					logger.Tracef("", "Read display file")
					var obj []map[string]interface{}
					if err := json.Unmarshal(data, &obj); err != nil {
						logger.Errorf("", "Error unmarshalling display JSON: %v", err)
					} else {
						logger.Tracef("", "Updating display object")
						r.State.Display = obj
					}
				} else {
					logger.Errorf("", "Display read error: %s", err.Error())
				}
			} else {
				logger.Tracef("", "Display stat error: %s", err.Error())
			}
			if err := updateRunState(r, false); err != nil {
				if _, err := os.Stat(runDir); err != nil {
					if os.IsNotExist(err) {
						// file does not exist
					} else {
						os.RemoveAll(runDir)
					}
				}
				return false, err
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
			r.State.Output = fmt.Sprintf("%s\n\n--------------------------------\n\n%s", podmanOutput, string(err.Error()))
		} else {
			r.State.Output = fmt.Sprintf("%s\n\n--------------------------------\n\n%s", podmanOutput, string(logs))
		}

		for _, name := range r.Task.Store.File {
			filePath := fmt.Sprintf("%s/%s", runDir, name)
			if _, err = os.Stat(filePath); err == nil {
				err := filestore.UploadFile(filePath, fmt.Sprintf("%s/%s", r.Task.Cascade, name))
				if err != nil {
					logger.Errorf("", "ERROR UPLOADING %s: %s\n", fmt.Sprintf("%s/%s", r.Task.Cascade, name), err.Error())
				}
				ds.Files = append(ds.Files, name)
				ds.Files = utils.RemoveDuplicateValues(ds.Files)
			}
		}

		// Load in display file if present and able
		logger.Tracef("", "Checking for display at %s", displayPath)
		if _, err = os.Stat(displayPath); err == nil {
			logger.Tracef("", "Display path is present")
			if data, err := os.ReadFile(displayPath); err == nil {
				logger.Tracef("", "Read display file")
				var obj []map[string]interface{}
				if err := json.Unmarshal(data, &obj); err != nil {
					logger.Errorf("", "Error unmarshalling display JSON: %v", err)
				} else {
					logger.Tracef("", "Updating display object")
					r.State.Display = obj
				}
			} else {
				logger.Errorf("", "read error: %s", err.Error())
			}
		} else {
			logger.Tracef("", "stat error: %s", err.Error())
		}

		if err := updateRunState(r, false); err != nil {
			if _, err := os.Stat(runDir); err != nil {
				if os.IsNotExist(err) {
					// file does not exist
				} else {
					os.RemoveAll(runDir)
				}
			}
			return false, err
		}

		envOutPath := fmt.Sprintf("%s/.envout", runDir)
		var dat []byte
		if _, err := os.Stat(envOutPath); err == nil {
			dat, err = os.ReadFile(envOutPath)
			if err != nil {
				logger.Errorf("", "Error reading file %s\n", err.Error())
				setErrorStatus(r, err.Error())
				if err := updateRunState(r, true); err != nil {
					if _, err := os.Stat(runDir); err != nil {
						if os.IsNotExist(err) {
							// file does not exist
						} else {
							os.RemoveAll(runDir)
						}
					}
					return false, err
				}
			}
		}
		envVarList := strings.Split(string(dat), "\n")
		envVarMap := map[string]string{}

		for _, val := range envVarList {
			name, val, _ := strings.Cut(val, ";")
			decoded, _ := base64.StdEncoding.DecodeString(val)
			envVarMap[name] = string(decoded)
		}

		for _, name := range r.Task.Store.Env {
			r.Context[name] = envVarMap[name]
		}

		currentTime = time.Now().UTC()
		r.State.Finished = currentTime.Format("2006-01-02T15:04:05Z")
		if returnCode == "0" {
			r.State.Status = constants.STATE_STATUS_SUCCESS
		} else {
			r.State.Status = constants.STATE_STATUS_ERROR
		}

		if r.Task.ShouldRM {
			rmCommand := fmt.Sprintf("podman rm -f %s", containerName)
			out, err := exec.Command("bash", "-c", rmCommand).CombinedOutput()
			logger.Debugf("", "Podman rm: %s", string(out))
			r.State.Output += fmt.Sprintf("\n\n--------------------------------\n\n%s", string(out))
			if err != nil {
				logger.Error("", err.Error())
				r.State.Output += fmt.Sprintf("\n\n--------------------------------\n\n%s", err.Error())
			}
		}
	}
	if err := updateRunState(r, true); err != nil {
		if _, err := os.Stat(runDir); err != nil {
			if os.IsNotExist(err) {
				// file does not exist
			} else {
				os.RemoveAll(runDir)
			}
		}
		return false, err
	}
	if _, err := os.Stat(runDir); err != nil {
		if os.IsNotExist(err) {
			// file does not exist
		} else {
			os.RemoveAll(runDir)
		}
	}
	return false, err
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
			// m := msg.RunMsg{
			// 	Task:    tn,
			// 	Cascade: cn,
			// 	Status:  constants.STATE_STATUS_KILLED,
			// }
			s, err := state.GetStateByNames(cn, tn)
			if err != nil {
				logger.Errorf("", "Unable to get state for run %s.%s with error %s", cn, tn, err.Error())
				return err
			}
			s.Status = constants.STATE_STATUS_KILLED
			// logger.Debugf("", "Updating run state for %v", m)
			if err := state.UpdateStateByNames(cn, tn, s); err != nil {
				logger.Errorf("", "Unable to update state %s.%s: %s", cn, tn, err.Error())
				return err
			}
			// if err := bulwark.QueuePush(c, m); err != nil {
			// 	logger.Errorf("", "Unable to push killed message to queue: %s", err.Error())
			// 	return err
			// }
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

func StartLocalRun(r *Run) (bool, error) {
	r.State.Status = constants.STATE_STATUS_RUNNING
	currentTime := time.Now().UTC()
	r.State.Started = currentTime.Format("2006-01-02T15:04:05Z")
	if err := state.UpdateStateKilledByNames(r.Task.Cascade, r.Task.Name, false); err != nil {
		logger.Infof("", "Cannot update state killed: %s %s %s", r.Task.Cascade, r.Task.Name, err.Error())
		return false, err
	}
	if err := updateRunState(r, false); err != nil {
		return false, err
	}

	cName := r.State.Cascade
	tName := r.State.Task

	logger.Debugf("", "Getting datastore by name %s", cName)
	ds, err := datastore.GetDataStoreByCascade(cName)
	if err != nil {
		logger.Errorf("", "Cannot get datastore %s", cName)
		setErrorStatus(r, err.Error())
		if err := updateRunState(r, true); err != nil {
			return false, err
		}
		return false, err
	}

	runDir := fmt.Sprintf("/tmp/run/%s/%s/%d", cName, tName, r.Number)

	if _, err := os.Stat(runDir); err != nil {
		if os.IsNotExist(err) {
			// file does not exist
		} else {
			os.RemoveAll(runDir)
		}
	}

	err = os.MkdirAll(runDir, 0755)
	if err != nil {
		logger.Errorf("", "Error creating run directory %s", err.Error())
		setErrorStatus(r, err.Error())
		if err := updateRunState(r, true); err != nil {
			return false, err
		}
		return false, err
	}

	scriptPath := runDir + "/.run.sh"
	envInPath := runDir + "/.envin"
	displayPath := runDir + "/.display"

	envInput := ""
	for key, val := range r.Task.Inputs {
		contextVal, ok := r.Context[val]
		var encoded string
		if ok {
			encoded = base64.StdEncoding.EncodeToString([]byte(contextVal))
		} else {
			encoded = base64.StdEncoding.EncodeToString([]byte(ds.Env[val]))
		}
		envInput += fmt.Sprintf("%s;%s\n", key, encoded)
	}
	for _, key := range r.Task.Load.Env {
		contextVal, ok := r.Context[key]
		var encoded string
		if ok {
			encoded = base64.StdEncoding.EncodeToString([]byte(contextVal))
		} else {
			encoded = base64.StdEncoding.EncodeToString([]byte(ds.Env[key]))
		}
		envInput += fmt.Sprintf("%s;%s\n", key, encoded)
	}
	for key, val := range r.Task.Env {
		encoded := base64.StdEncoding.EncodeToString([]byte(val))
		envInput += fmt.Sprintf("%s;%s\n", key, encoded)
	}

	envOutput := ""
	for _, key := range r.Task.Store.Env {
		envOutput += fmt.Sprintf("echo \"%s;$(echo \"${%s}\" | base64)\" >> %s/.envout\n", key, key, runDir)
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
	`, envInPath, r.Task.Run, envOutput)

	// Write out our run script
	data := []byte(runScript)
	err = os.WriteFile(scriptPath, data, 0777)
	if err != nil {
		logger.Errorf("", "Error writing run file %s", err.Error())
		setErrorStatus(r, err.Error())
		if err := updateRunState(r, true); err != nil {
			return false, err
		}
		return false, err
	}

	// Write out envin script
	data = []byte(envInput)
	err = os.WriteFile(envInPath, data, 0777)
	if err != nil {
		logger.Errorf("", "Error writing envin file %s", err.Error())
		setErrorStatus(r, err.Error())
		if err := updateRunState(r, true); err != nil {
			return false, err
		}
		return false, err
	}

	for _, name := range r.Task.Load.File {
		err := filestore.GetFile(fmt.Sprintf("%s/%s", cName, name), fmt.Sprintf("%s/%s", runDir, name))
		if err != nil {
			logger.Errorf("", "Error getting file %s", err.Error())
			setErrorStatus(r, err.Error())
			if err := updateRunState(r, true); err != nil {
				return false, err
			}
			return false, err
		}
	}

	if r.PID > 0 {
		if err := exec.Command("/bin/sh", "-c", fmt.Sprintf("kill %d", r.PID)).Run(); err != nil {
			logger.Infof("", "Cannot kill existing run: %s\n", err.Error())
		}
	}

	localCommand := fmt.Sprintf("cd %s && ./.run.sh", runDir)

	logger.Debugf("", "command: %s", localCommand)

	cmd := exec.Command("/bin/bash", "-c", localCommand)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	runError = cmd.Start()
	if err != nil {
		return false, runError
	}
	r.PID = cmd.Process.Pid
	if err := updateRunState(r, true); err != nil {
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
			setErrorStatus(r, fmt.Sprintf("Error running pod %s\n", runError.Error()))
			if err := updateRunState(r, true); err != nil {
				return false, err
			}
		}
		output := outb.String() + "\n\n" + errb.String()
		logger.Tracef("", "setting output 1 %s", output)
		r.State.Output = output

		logger.Tracef("", "Checking for display at %s", displayPath)
		if _, err = os.Stat(displayPath); err == nil {
			logger.Tracef("", "Display path is present")
			if data, err := os.ReadFile(displayPath); err == nil {
				logger.Tracef("", "Read display file")
				var obj []map[string]interface{}
				if err := json.Unmarshal(data, &obj); err != nil {
					logger.Errorf("", "Error unmarshalling display JSON: %v", err)
				} else {
					logger.Tracef("", "Updating display object")
					r.State.Display = obj
				}
			} else {
				logger.Tracef("", "Display read error: %s", err.Error())
			}
		} else {
			logger.Tracef("", "Display stat error: %s", err.Error())
		}
		if err := updateRunState(r, false); err != nil {
			logger.Errorf("", "Error updating run: %s", err.Error())
			return false, err
		}
		time.Sleep(500 * time.Millisecond)
	}

	output := outb.String() + "\n\n" + errb.String()
	logger.Tracef("", "setting output 2 %s", output)
	r.State.Output = output

	if err := updateRunState(r, false); err != nil {
		logger.Errorf("", "Error updating run: %s", err.Error())
		return false, err
	}

	returnCode := 0
	returnCode = ExitCode(runError)

	for _, name := range r.Task.Store.File {
		filePath := fmt.Sprintf("%s/%s", runDir, name)
		if _, err = os.Stat(filePath); err == nil {
			err := filestore.UploadFile(filePath, fmt.Sprintf("%s/%s", r.Task.Cascade, name))
			if err != nil {
				logger.Errorf("", "ERROR UPLOADING %s: %s\n", fmt.Sprintf("%s/%s", r.Task.Cascade, name), err.Error())
			}
			ds.Files = append(ds.Files, name)
			ds.Files = utils.RemoveDuplicateValues(ds.Files)
		}
	}

	// Load in display file if present and able
	logger.Tracef("", "Checking for display at %s", displayPath)
	if _, err = os.Stat(displayPath); err == nil {
		logger.Tracef("", "Display path is present")
		if data, err := os.ReadFile(displayPath); err == nil {
			logger.Tracef("", "Read display file")
			var obj []map[string]interface{}
			if err := json.Unmarshal(data, &obj); err != nil {
				logger.Errorf("", "Error unmarshalling display JSON: %v", err)
			} else {
				logger.Tracef("", "Updating display object")
				r.State.Display = obj
			}
		} else {
			logger.Errorf("", "read error: %s", err.Error())
		}
	} else {
		logger.Tracef("", "stat error: %s", err.Error())
	}

	if err := updateRunState(r, false); err != nil {
		return false, err
	}

	envOutPath := fmt.Sprintf("%s/.envout", runDir)
	var dat []byte
	if _, err := os.Stat(envOutPath); err == nil {
		dat, err = os.ReadFile(envOutPath)
		if err != nil {
			logger.Errorf("", "Error reading file %s\n", err.Error())
			setErrorStatus(r, err.Error())
			if err := updateRunState(r, true); err != nil {
				return false, err
			}
		}
	}
	envVarList := strings.Split(string(dat), "\n")
	envVarMap := map[string]string{}

	for _, val := range envVarList {
		name, val, _ := strings.Cut(val, ";")
		decoded, _ := base64.StdEncoding.DecodeString(val)
		envVarMap[name] = string(decoded)
	}

	for _, name := range r.Task.Store.Env {
		r.Context[name] = envVarMap[name]
	}

	currentTime = time.Now().UTC()
	r.State.Finished = currentTime.Format("2006-01-02T15:04:05Z")
	if returnCode == 0 {
		r.State.Status = constants.STATE_STATUS_SUCCESS
	} else {
		r.State.Status = constants.STATE_STATUS_ERROR
	}

	r.PID = 0

	err = updateRunState(r, true)
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
