package get

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"scaffold/client/auth"
	"scaffold/client/logger"
	"scaffold/client/objects"
	"scaffold/client/utils"
	"strings"
	"text/tabwriter"
)

func DoGet(profile, object, context string) {
	logger.Debugf("", "Getting objects")
	p := auth.ReadProfile(profile)
	uri := fmt.Sprintf("%s://%s:%s", p.Protocol, p.Host, p.Port)

	logger.Debugf("", "Checking if object is valid")
	objects := []string{"cascade", "datastore", "state", "task"}

	parts := strings.Split(object, "/")

	if !utils.Contains(objects, parts[0]) {
		logger.Fatalf("Invalid object type passed: '%s'. Valid object types are 'cascade', 'datastore', 'state', 'task'", object)
	}

	logger.Debugf("", "Getting context")
	if context == "" {
		context = p.Cascade
	}
	if len(parts) == 2 {
		if parts[0] != "cascade" && parts[0] != "datastore" {
			object = fmt.Sprintf("%s/%s/%s", parts[0], context, parts[1])
		}
	}

	data := getJSON(p, uri, object)
	if len(parts) == 2 {
		logger.Debugf("", "data is individual objects, editing JSON response")
		data = []byte(fmt.Sprintf(`[%s]`, string(data)))
	}
	logger.Debugf("", "JSON response: %s", string(data))

	switch parts[0] {
	case "cascade":
		listCascades(data)
	case "state":
		listStates(data, context)
	case "task":
		listTasks(data, context)
	case "datastore":
		listDataStores(data)
	}
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

func listCascades(data []byte) {
	var cascades []objects.Cascade

	err := json.Unmarshal(data, &cascades)
	if err != nil {
		logger.Fatalf("", "Unable to marshal cascades JSON: %s", err.Error())
	}

	w := tabwriter.NewWriter(os.Stdout, 8, 1, 1, ' ', 0)
	fmt.Fprintln(w, "NAME \tVERSION \tGROUPS \tCREATED \tUPDATED \t")
	for _, c := range cascades {
		fmt.Fprintln(w, fmt.Sprintf("%s  \t%s \t%s \t%s \t%s ", c.Name, c.Version, strings.Join(c.Groups, ","), c.Created, c.Updated))
	}
	w.Flush()
}

func listDataStores(data []byte) {
	var datastores []objects.DataStore

	err := json.Unmarshal(data, &datastores)
	if err != nil {
		logger.Fatalf("", "Unable to marshal datastores JSON: %s", err.Error())
	}

	w := tabwriter.NewWriter(os.Stdout, 8, 1, 1, ' ', 0)
	fmt.Fprintln(w, "NAME \tCREATED \tUPDATED ")
	for _, d := range datastores {
		fmt.Fprintln(w, fmt.Sprintf("%s \t%s \t%s ", d.Name, d.Created, d.Updated))
	}
	w.Flush()
}

func listStates(data []byte, context string) {
	var states []objects.State

	err := json.Unmarshal(data, &states)
	if err != nil {
		logger.Fatalf("", "Unable to marshal states JSON: %s", err.Error())
	}

	w := tabwriter.NewWriter(os.Stdout, 8, 1, 1, ' ', 0)
	fmt.Fprintln(w, "TASK \tCASCADE \tSTATUS \tSTARTED \tFINISHED \t")
	for _, s := range states {
		if s.Cascade == context {
			fmt.Fprintln(w, fmt.Sprintf("%s \t%s \t%s \t%s \t%s ", s.Task, s.Cascade, s.Status, s.Started, s.Finished))
		}
	}
	w.Flush()
}

func listTasks(data []byte, context string) {
	var tasks []objects.Task

	err := json.Unmarshal(data, &tasks)
	if err != nil {
		logger.Fatalf("", "Unable to marshal tasks JSON: %s", err.Error())
	}

	logger.Debugf("", "Data: %s", string(data))
	logger.Debugf("", "Task data: %v", tasks)

	w := tabwriter.NewWriter(os.Stdout, 8, 1, 1, ' ', 0)
	fmt.Fprintln(w, "NAME \tCASCADE \tIMAGE \tRUN NUMBER \tUPDATED \t")
	for _, t := range tasks {
		if t.Cascade == context {
			fmt.Fprintln(w, fmt.Sprintf("%s \t%s \t%s \t%d \t%s ", t.Name, t.Cascade, t.Image, t.RunNumber, t.Updated))
		}
	}
	w.Flush()
}
