package get

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"scaffold/client/auth"
	"scaffold/client/utils"
	"strings"

	logger "github.com/jfcarter2358/go-logger"
)

func DoGet(profile, object, context string) {
	logger.Debugf("", "Getting objects")
	p := auth.ReadProfile(profile)
	uri := fmt.Sprintf("%s://%s:%s", p.Protocol, p.Host, p.Port)

	logger.Debugf("", "Checking if object is valid")
	objects := []string{"project", "user", "runbook", "alert", "release", "history"}

	parts := strings.Split(object, "/")

	if !utils.Contains(objects, parts[0]) {
		logger.Fatalf("", "Invalid object type passed: '%s'. Valid object types are %v", object, objects)
	}

	data := getJSON(p, uri, object)
	if len(parts) == 2 {
		logger.Debugf("", "data is individual objects, editing JSON response")
		data = []byte(fmt.Sprintf(`[%s]`, string(data)))
	}
	logger.Debugf("", "JSON response: %s", string(data))

	// switch parts[0] {
	// case "project":
	// 	listWorkflows(data)
	// case "user":
	// 	listStates(data, context)
	// case "runbook":
	// 	listTasks(data, context)
	// case "alert":
	// 	listDataStores(data)
	// case "release":
	// 	listFiles(data, context)
	// case "history":
	// 	listUsers(data)
	// }
}

func getJSON(p auth.ProfileObj, uri, object string) []byte {
	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("%s/api/v1/%s", uri, object)
	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", p.APIToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Fatalf("", "Encountered error: %s", err.Error())
	}
	if resp.StatusCode >= 400 {
		logger.Fatalf("", "Got status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Fatalf("", "Error reading body: %s", err.Error())
	}
	resp.Body.Close()

	return body
}

// func listWorkflows(data []byte) {
// 	var workflows []map[string]interface{}

// 	err := json.Unmarshal(data, &workflows)
// 	if err != nil {
// 		logger.Fatalf("", "Unable to marshal workflows JSON: %s", err.Error())
// 	}

// 	w := tabwriter.NewWriter(os.Stdout, 8, 1, 1, ' ', 0)
// 	fmt.Fprintln(w, "NAME \tVERSION \tGROUPS \tCREATED \tUPDATED \t")
// 	for _, c := range workflows {
// 		name := c["name"].(string)
// 		version := c["version"].(string)
// 		groupList := c["groups"].([]interface{})
// 		groups := []string{}
// 		for _, g := range groupList {
// 			groups = append(groups, g.(string))
// 		}
// 		created := c["created"].(string)
// 		updated := c["updated"].(string)
// 		fmt.Fprintf(w, "%s \t%s \t%s \t%s \t%s \n", name, version, groups, created, updated)
// 	}
// 	w.Flush()
// }

// func listDataStores(data []byte) {
// 	var datastores []map[string]interface{}

// 	err := json.Unmarshal(data, &datastores)
// 	if err != nil {
// 		logger.Fatalf("", "Unable to marshal datastores JSON: %s", err.Error())
// 	}

// 	w := tabwriter.NewWriter(os.Stdout, 8, 1, 1, ' ', 0)
// 	fmt.Fprintln(w, "NAME \tCREATED \tUPDATED \t")
// 	for _, d := range datastores {
// 		name := d["name"].(string)
// 		created := d["created"].(string)
// 		updated := d["updated"].(string)
// 		fmt.Fprintf(w, "%s \t%s \t%s \n", name, created, updated)
// 	}
// 	w.Flush()
// }

// func listStates(data []byte, context string) {
// 	var states []map[string]interface{}

// 	err := json.Unmarshal(data, &states)
// 	if err != nil {
// 		logger.Fatalf("", "Unable to marshal states JSON: %s", err.Error())
// 	}

// 	w := tabwriter.NewWriter(os.Stdout, 8, 1, 1, ' ', 0)
// 	fmt.Fprintln(w, "TASK \tWORKFLOW \tSTATUS \tSTARTED \tFINISHED \t")
// 	for _, s := range states {
// 		workflow := s["workflow"].(string)
// 		status := s["status"].(string)
// 		task := s["task"].(string)
// 		started := s["started"].(string)
// 		finished := s["finished"].(string)
// 		if workflow == context || context == constants.ALL_CONTEXTS {
// 			fmt.Fprintf(w, "%s \t%s \t%s \t%s \t%s \n", task, workflow, status, started, finished)
// 		}
// 	}
// 	w.Flush()
// }

// func listTasks(data []byte, context string) {
// 	var tasks []map[string]interface{}

// 	err := json.Unmarshal(data, &tasks)
// 	if err != nil {
// 		logger.Fatalf("", "Unable to marshal tasks JSON: %s", err.Error())
// 	}

// 	logger.Debugf("", "Data: %s", string(data))
// 	logger.Debugf("", "Task data: %v", tasks)

// 	w := tabwriter.NewWriter(os.Stdout, 8, 1, 1, ' ', 0)
// 	fmt.Fprintln(w, "NAME \tWORKFLOW \tIMAGE \tRUN NUMBER \tUPDATED \t")
// 	for _, t := range tasks {
// 		workflow := t["workflow"].(string)
// 		name := t["name"].(string)
// 		image := t["image"].(string)
// 		runNumber := int(t["run_number"].(float64))
// 		updated := t["updated"].(string)
// 		if workflow == context || context == constants.ALL_CONTEXTS {
// 			fmt.Fprintf(w, "%s \t%s \t%s \t%d \t%s \n", name, workflow, image, runNumber, updated)
// 		}
// 	}
// 	w.Flush()
// }

// func listFiles(data []byte, context string) {
// 	var files []map[string]interface{}

// 	err := json.Unmarshal(data, &files)
// 	if err != nil {
// 		logger.Fatalf("", "Unable to marshal file JSON: %s", err.Error())
// 	}

// 	w := tabwriter.NewWriter(os.Stdout, 8, 1, 1, ' ', 0)
// 	fmt.Fprintln(w, "NAME \tWORKFLOW \tUPDATED \t")
// 	for _, f := range files {
// 		workflow := f["workflow"].(string)
// 		name := f["name"].(string)
// 		updated := f["modified"].(string)
// 		if workflow == context || context == constants.ALL_CONTEXTS {
// 			fmt.Fprintf(w, "%s \t%s \t%s \n", name, workflow, updated)
// 		}
// 	}
// 	w.Flush()
// }

// func listUsers(data []byte) {
// 	var users []map[string]interface{}

// 	err := json.Unmarshal(data, &users)
// 	if err != nil {
// 		logger.Fatalf("", "Unable to marshal users JSON: %s", err.Error())
// 	}

// 	w := tabwriter.NewWriter(os.Stdout, 8, 1, 1, ' ', 0)
// 	fmt.Fprintln(w, "USERNAME \tGIVEN NAME \tFAMILY NAME \tEMAIL \tGROUPS \tROLES \tCREATED \tUPDATED \t")
// 	for _, u := range users {
// 		username := u["username"].(string)
// 		givenName := u["given_name"].(string)
// 		familyName := u["family_name"].(string)
// 		groupList := u["groups"].([]interface{})
// 		groups := []string{}
// 		for _, g := range groupList {
// 			groups = append(groups, g.(string))
// 		}
// 		roleList := u["roles"].([]interface{})
// 		roles := []string{}
// 		for _, r := range roleList {
// 			roles = append(roles, r.(string))
// 		}
// 		created := u["created"].(string)
// 		updated := u["updated"].(string)
// 		email := u["email"].(string)
// 		fmt.Fprintf(w, "%s \t%s \t%s \t%s \t%s \t%s \t%s \t%s \n", username, givenName, familyName, email, groups, roles, created, updated)
// 	}
// 	w.Flush()
// }

// func listInputs(data []byte, context string) {
// 	var inputs []map[string]interface{}

// 	err := json.Unmarshal(data, &inputs)
// 	if err != nil {
// 		logger.Fatalf("", "Unable to marshal input JSON: %s", err.Error())
// 	}

// 	w := tabwriter.NewWriter(os.Stdout, 8, 1, 1, ' ', 0)
// 	fmt.Fprintln(w, "NAME \tWORKFLOW \tTYPE \tDEFAULT \t")
// 	for _, i := range inputs {
// 		workflow := i["workflow"].(string)
// 		name := i["name"].(string)
// 		inputType := i["type"].(string)
// 		inputDefault := i["default"].(string)
// 		if workflow == context || context == constants.ALL_CONTEXTS {
// 			fmt.Fprintf(w, "%s \t%s \t%s \t%s \n", name, workflow, inputType, inputDefault)
// 		}
// 	}
// 	w.Flush()
// }

// func listRunbooks(data []byte, context string) {
// 	var runbooks []map[string]interface{}

// 	err := json.Unmarshal(data, &runbooks)
// 	if err != nil {
// 		logger.Fatalf("", "Unable to marshal input JSON: %s", err.Error())
// 	}

// 	w := tabwriter.NewWriter(os.Stdout, 8, 1, 1, ' ', 0)
// 	fmt.Fprintln(w, "NAME \tID \tCATEGORY \tGROUPS \tCREATED \tUPDATED \t")
// 	for _, r := range runbooks {
// 		name := r["name"].(string)
// 		id := r["id"].(string)
// 		category := r["category"].(string)
// 		groupList := r["groups"].([]interface{})
// 		groups := []string{}
// 		for _, g := range groupList {
// 			groups = append(groups, g.(string))
// 		}
// 		created := r["created"].(string)
// 		updated := r["updated"].(string)
// 		fmt.Fprintf(w, "%s \t%s \t%s \t%s \t%s \t%s \n", name, id, category, groups, created, updated)
// 	}
// 	w.Flush()
// }

// func listMonitors(data []byte, context string) {
// 	var monitors []map[string]interface{}

// 	err := json.Unmarshal(data, &monitors)
// 	if err != nil {
// 		logger.Fatalf("", "Unable to marshal input JSON: %s", err.Error())
// 	}

// 	w := tabwriter.NewWriter(os.Stdout, 8, 1, 1, ' ', 0)
// 	fmt.Fprintln(w, "NAME \tID \tKIND \tSTATUS \tENABLED \tGROUPS \tCREATED \tUPDATED \t")
// 	for _, m := range monitors {
// 		name := m["name"].(string)
// 		id := m["id"].(string)
// 		kind := m["kind"].(string)
// 		groupList := m["groups"].([]interface{})
// 		groups := []string{}
// 		for _, g := range groupList {
// 			groups = append(groups, g.(string))
// 		}
// 		created := m["created"].(string)
// 		updated := m["updated"].(string)
// 		status := m["status"].(string)
// 		enabled := m["enabled"].(bool)
// 		fmt.Fprintf(w, "%s \t%s \t%s \t%s \t%v \t%s \t%s \t%s \n", name, id, kind, status, enabled, groups, created, updated)
// 	}
// 	w.Flush()
// }

// func listAlerts(data []byte, context string) {
// 	var alerts []map[string]interface{}

// 	err := json.Unmarshal(data, &alerts)
// 	if err != nil {
// 		logger.Fatalf("", "Unable to marshal input JSON: %s", err.Error())
// 	}

// 	w := tabwriter.NewWriter(os.Stdout, 8, 1, 1, ' ', 0)
// 	fmt.Fprintln(w, "NAME \tID \tWORKFLOW \tLANGUAGE \tCREATED \tUPDATED \t")
// 	for _, a := range alerts {
// 		name := a["name"].(string)
// 		id := a["id"].(string)
// 		workflow := a["workflow"].(string)
// 		language := a["language"].(string)
// 		created := a["created"].(string)
// 		updated := a["updated"].(string)
// 		fmt.Fprintf(w, "%s \t%s \t%s \t%s \t%s \t%s \n", name, id, workflow, language, created, updated)
// 	}
// 	w.Flush()
// }
