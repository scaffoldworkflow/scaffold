package run

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"scaffold/server/constants"
	"scaffold/server/datastore"
	"scaffold/server/filestore"
	"scaffold/server/state"
	"scaffold/server/task"
	"strings"
	"time"
)

var runError error

type Run struct {
	Name  string      `json:"name"`
	Task  task.Task   `json:"task"`
	State state.State `json:"state"`
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

func StartRun(r *Run) error {
	r.State.Status = constants.STATE_STATUS_RUNNING
	currentTime := time.Now().UTC()
	r.State.Started = currentTime.Format("2006-01-02T15:04:05Z")

	cName := strings.Split(r.Name, ".")[0]
	ds, err := datastore.GetDataStoreByName(cName)
	if err != nil {
		return err
	}
	for _, name := range r.Task.Store.Env {
		val := os.Getenv(name)
		ds.Env[name] = val
	}

	containerName := strings.Replace(r.Name, ".", "-", -1)

	runDir := fmt.Sprintf("/tmp/run-%s", containerName)
	err = os.MkdirAll(runDir, 0755)
	if err != nil {
		setErrorStatus(r, err.Error())
		return err
	}

	scriptPath := runDir + "/.run.sh"
	envInPath := runDir + "/.envin"

	envInput := ""
	for key, val := range r.Task.Inputs {
		encoded := base64.StdEncoding.EncodeToString([]byte(ds.Env[val]))
		envInput += fmt.Sprintf("%s;%s\n", key, encoded)
	}
	for _, key := range r.Task.Load.Env {
		encoded := base64.StdEncoding.EncodeToString([]byte(ds.Env[key]))
		envInput += fmt.Sprintf("%s;%s\n", key, encoded)
	}

	envOutput := ""
	for key := range r.Task.Outputs {
		envOutput += fmt.Sprintf("echo \"%s;$(echo \"${%s}\" | base64)\" >> /tmp/run/.envout\n", key, key)
	}
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
		setErrorStatus(r, err.Error())
		return err
	}

	// Write out envin script
	data = []byte(envInput)
	err = os.WriteFile(envInPath, data, 0777)
	if err != nil {
		setErrorStatus(r, err.Error())
		return err
	}

	for _, name := range r.Task.Load.File {
		err := filestore.GetFile(name, fmt.Sprintf("%s/%s", runDir, name))
		if err != nil {
			setErrorStatus(r, err.Error())
			return err
		}
	}

	// Clean up any possible artifacts
	if err := exec.Command("/bin/sh", "-c", fmt.Sprintf("podman kill %s", containerName)).Run(); err != nil {
		fmt.Printf("No running container with name %s exists, skipping kill\n", containerName)
	}
	if err := exec.Command("/bin/sh", "-c", fmt.Sprintf("podman rm %s", containerName)).Run(); err != nil {
		fmt.Printf("No running container with name %s exists, skipping removal\n", containerName)
	}

	podmanCommand := "podman run --privileged -d --security-opt label=disabled "
	podmanCommand += fmt.Sprintf("--name %s ", containerName)
	podmanCommand += fmt.Sprintf("--mount type=bind,src=%s,dst=/tmp/run ", runDir)
	podmanCommand += r.Task.Image
	podmanCommand += " bash -c /tmp/run/.run.sh"

	fmt.Printf("command: %s", podmanCommand)

	cmd := exec.Command("/bin/sh", "-c", podmanCommand)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	go runCmd(cmd)

	output, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("podman ps -a --filter \"name=%s\" --format \"{{.Status}}\"", containerName)).CombinedOutput()
	if err != nil {
		fmt.Printf("Encountered error: %v\n", err.Error())
		fmt.Printf("STDOUT: %s\n", string(output))

		setErrorStatus(r, string(output))
		return err
	}

	var podmanOutput string
	erroredOut := false
	for !strings.HasPrefix(string(output), "Exited") {
		if string(output) == "" {
			podmanOutput = outb.String() + "\n\n" + errb.String()
			r.State.Output = podmanOutput
			if runError != nil {
				setErrorStatus(r, string(runError.Error()))
				erroredOut = true
				break
			}
		} else {
			logs, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("podman logs %s", containerName)).CombinedOutput()
			if err != nil {
				r.State.Output = fmt.Sprintf("%s\n\n--------------------------------\n\n%s", podmanOutput, string(err.Error()))
			} else {
				r.State.Output = fmt.Sprintf("%s\n\n--------------------------------\n\n%s", podmanOutput, string(logs))
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
			filestore.UploadFile(fmt.Sprintf("%s/%s", runDir, name), name)
		}

		envOutPath := fmt.Sprintf("%s/.envout", runDir)
		var dat []byte
		if _, err := os.Stat(envOutPath); err == nil {
			dat, err = os.ReadFile(envOutPath)
			if err != nil {
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
		for env, name := range r.Task.Outputs {
			ds.Env[name] = envVarMap[env]
		}

		if err := datastore.UpdateDataStoreByName(cName, ds); err != nil {
			setErrorStatus(r, err.Error())
		}

		currentTime = time.Now().UTC()
		r.State.Finished = currentTime.Format("2006-01-02T15:04:05Z")
		if returnCode == "0" {
			r.State.Status = constants.STATE_STATUS_SUCCESS
		} else {
			r.State.Status = constants.STATE_STATUS_ERROR
		}
	}
	return nil
}
