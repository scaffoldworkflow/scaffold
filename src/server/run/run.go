package run

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"scaffold/server/constants"
	"scaffold/server/datastore"
	"scaffold/server/filestore"
	"scaffold/server/input"
	"scaffold/server/logger"
	"scaffold/server/state"
	"scaffold/server/task"
	"strings"
	"time"
)

var runError error

// type CheckRun struct {
// 	Enabled bool           `json:"enabled"`
// 	Name    string         `json:"name"`
// 	Task    task.TaskCheck `json:"task"`
// 	State   state.State    `json:"state"`
// }

type Run struct {
	Name     string      `json:"name"`
	Task     task.Task   `json:"task"`
	State    state.State `json:"state"`
	Previous state.State `json:"previous"`
	Number   int         `json:"number"`
}

func setErrorStatus(r *Run, output string) {
	r.State.Output = output
	r.State.Status = constants.STATE_STATUS_ERROR
	currentTime := time.Now().UTC()
	r.State.Finished = currentTime.Format("2006-01-02T15:04:05Z")
}

func runCmd(cmd *exec.Cmd) {
	runError = cmd.Run()
}

func StartRun(r *Run) (bool, error) {
	r.State.Status = constants.STATE_STATUS_RUNNING
	currentTime := time.Now().UTC()
	r.State.Started = currentTime.Format("2006-01-02T15:04:05Z")

	cName := strings.Split(r.Name, ".")[0]
	logger.Debugf("", "Getting datastore by name %s", cName)
	ds, err := datastore.GetDataStoreByName(cName)
	if err != nil {
		logger.Errorf("", "Cannot get datastore %s", cName)
		return false, err
	}

	// containerName := strings.Replace(r.Name, ".", "-", -1)
	containerName := fmt.Sprintf("%s-%s-%d", r.State.Cascade, r.State.Task, r.Number)

	runDir := fmt.Sprintf("/tmp/run/%s/%s/%d", r.State.Cascade, r.State.Task, r.Number)
	err = os.MkdirAll(runDir, 0755)
	if err != nil {
		logger.Errorf("", "Error creating run directory %s", err.Error())
		setErrorStatus(r, err.Error())
		return false, err
	}

	scriptPath := runDir + "/.run.sh"
	envInPath := runDir + "/.envin"
	displayPath := runDir + "/.display"

	envInput := ""
	for key, val := range r.Task.Inputs {
		encoded := base64.StdEncoding.EncodeToString([]byte(ds.Env[val]))
		envInput += fmt.Sprintf("%s;%s\n", key, encoded)
	}
	for _, key := range r.Task.Load.Env {
		encoded := base64.StdEncoding.EncodeToString([]byte(ds.Env[key]))
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
		return false, err
	}

	// Write out envin script
	data = []byte(envInput)
	err = os.WriteFile(envInPath, data, 0777)
	if err != nil {
		logger.Errorf("", "Error writing envin file %s", err.Error())
		setErrorStatus(r, err.Error())
		return false, err
	}

	for _, name := range r.Task.Load.File {
		err := filestore.GetFile(fmt.Sprintf("%s/%s", r.Task.Cascade, name), fmt.Sprintf("%s/%s", runDir, name))
		if err != nil {
			logger.Errorf("", "Error getting file %s", err.Error())
			setErrorStatus(r, err.Error())
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

	podmanCommand := "podman run --privileged -d --security-opt label=disabled "

	podmanCommand += fmt.Sprintf("--name %s ", containerName)
	podmanCommand += fmt.Sprintf("--mount type=bind,src=%s,dst=/tmp/run ", runDir)
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
				setErrorStatus(r, string(runError.Error()))
				erroredOut = true
				break
			}
			// Load in display file if present and able
			if _, err := os.Stat(displayPath); err == nil {
				logger.Tracef("", "Display path is present")
				data, err := os.ReadFile(displayPath)
				if err == nil {
					logger.Tracef("", "Read display file")
					var obj []map[string]interface{}
					if err := json.Unmarshal(data, &obj); err != nil {
						logger.Errorf("", "Error unmarshalling display JSON: %v", err)
					} else {
						logger.Tracef("", "Updating display object")
						r.State.Display = obj
					}
				}
			}

		} else {
			logs, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("podman logs %s", containerName)).CombinedOutput()
			if err != nil {
				r.State.Output = fmt.Sprintf("%s\n\n--------------------------------\n\n%s--------------------------------\n\n%s", podmanOutput, logs, string(err.Error()))
			} else {
				r.State.Output = fmt.Sprintf("%s\n\n--------------------------------\n\n%s", podmanOutput, string(logs))
			}
			// Load in display file if present and able
			if _, err := os.Stat(displayPath); err == nil {
				logger.Tracef("", "Display path is present")
				data, err := os.ReadFile(displayPath)
				if err == nil {
					logger.Tracef("", "Read display file")
					var obj []map[string]interface{}
					if err := json.Unmarshal(data, &obj); err != nil {
						logger.Errorf("", "Error unmarshalling display JSON: %v", err)
					} else {
						logger.Tracef("", "Updating display object")
						r.State.Display = obj
					}
				}
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
			err := filestore.UploadFile(fmt.Sprintf("%s/%s", runDir, name), fmt.Sprintf("%s/%s", r.Task.Cascade, name))
			if err != nil {
				logger.Errorf("", "ERROR UPLOADING %s: %s\n", fmt.Sprintf("%s/%s", r.Task.Cascade, name), err.Error())
			}
			ds.Files = append(ds.Files, name)
		}

		envOutPath := fmt.Sprintf("%s/.envout", runDir)
		var dat []byte
		if _, err := os.Stat(envOutPath); err == nil {
			dat, err = os.ReadFile(envOutPath)
			if err != nil {
				logger.Errorf("", "Error reading file %s\n", err.Error())
				setErrorStatus(r, err.Error())
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
			ds.Env[name] = envVarMap[name]
		}

		inputs := []input.Input{}
		if err := datastore.UpdateDataStoreByName(cName, ds, inputs); err != nil {
			logger.Errorf("", "Error updating datastore %s\n", err.Error())
			setErrorStatus(r, err.Error())
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
	return false, err
}

func Kill(cn, tn, nn string) error {
	containerName := fmt.Sprintf("%s-%s-%s", cn, tn, nn)
	if err := exec.Command("/bin/sh", "-c", fmt.Sprintf("podman kill %s", containerName)).Run(); err != nil {
		logger.Infof("", "No running container with name %s exists, skipping kill\n", containerName)
		return err
	}
	return nil
}
