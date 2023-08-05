package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"

	"scaffold/server/auth"
	"scaffold/server/cascade"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/container"
	"scaffold/server/datastore"
	"scaffold/server/filestore"
	"scaffold/server/health"
	"scaffold/server/input"
	"scaffold/server/logger"
	"scaffold/server/manager"
	"scaffold/server/run"
	"scaffold/server/state"
	"scaffold/server/task"
	"scaffold/server/user"
	"scaffold/server/utils"
	"scaffold/server/worker"
)

func Healthy(c *gin.Context) {
	if health.IsHealthy {
		c.Status(http.StatusOK)
		return
	}

	c.Status(http.StatusServiceUnavailable)
}

func Ready(c *gin.Context) {
	if health.IsReady {
		c.Status(http.StatusOK)
		return
	}

	c.Status(http.StatusServiceUnavailable)
}

func Available(c *gin.Context) {
	if health.IsAvailable {
		c.Status(http.StatusOK)
		return
	}

	c.Status(http.StatusServiceUnavailable)
}

/*~~~~~~~~ CASCADE ~~~~~~~~*/

func CreateCascade(ctx *gin.Context) {
	var c cascade.Cascade
	if err := ctx.ShouldBindJSON(&c); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	err := cascade.CreateCascade(&c)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created"})
}

func DeleteCascadeByName(ctx *gin.Context) {
	name := ctx.Param("name")

	err := cascade.DeleteCascadeByName(name)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func GetAllCascades(ctx *gin.Context) {
	cascades, err := cascade.GetAllCascades()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNoContent, gin.H{"cascades": []interface{}{}})
		}
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	cascadesOut := make([]cascade.Cascade, len(cascades))
	for idx, c := range cascades {
		cascadesOut[idx] = *c
	}

	ctx.JSON(http.StatusOK, gin.H{"cascades": cascadesOut})
}

func GetCascadeByName(ctx *gin.Context) {
	name := ctx.Param("name")

	c, err := cascade.GetCascadeByName(name)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, *c)
}

func UpdateCascadeByName(ctx *gin.Context) {
	name := ctx.Param("name")

	var c cascade.Cascade
	if err := ctx.ShouldBindJSON(&c); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	err := cascade.UpdateCascadeByName(name, &c)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

/*~~~~~~~~ DATASTORE ~~~~~~~~*/

func CreateDataStore(ctx *gin.Context) {
	var d datastore.DataStore
	if err := ctx.ShouldBindJSON(&d); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	err := datastore.CreateDataStore(&d)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created"})
}

func DeleteDataStoreByName(ctx *gin.Context) {
	name := ctx.Param("name")

	err := datastore.DeleteDataStoreByName(name)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func GetAllDataStores(ctx *gin.Context) {
	datastores, err := datastore.GetAllDataStores()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNoContent, gin.H{"datastores": []interface{}{}})
		}
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	datastoresOut := make([]datastore.DataStore, len(datastores))
	for idx, d := range datastores {
		datastoresOut[idx] = *d
	}

	ctx.JSON(http.StatusOK, gin.H{"datastores": datastoresOut})
}

func GetDataStoreByName(ctx *gin.Context) {
	name := ctx.Param("name")

	d, err := datastore.GetDataStoreByName(name)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, *d)
}

func UpdateDataStoreByName(ctx *gin.Context) {
	name := ctx.Param("name")

	var d datastore.DataStore
	if err := ctx.ShouldBindJSON(&d); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	inputs := []input.Input{}
	if config.Config.Node.Type == constants.NODE_TYPE_MANAGER {
		c, err := cascade.GetCascadeByName(name)
		if err != nil {
			utils.Error(err, ctx, http.StatusInternalServerError)
			return
		}
		inputs = c.Inputs
	}

	err := datastore.UpdateDataStoreByName(name, &d, inputs)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

/*~~~~~~~~ STATE ~~~~~~~~*/

func CreateState(ctx *gin.Context) {
	var s state.State
	if err := ctx.ShouldBindJSON(&s); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	err := state.CreateState(&s)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created"})
}

func DeleteStateByNames(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	tn := ctx.Param("task")

	err := state.DeleteStateByNames(cn, tn)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func DeleteStatesByCascade(ctx *gin.Context) {
	cn := ctx.Param("cascade")

	err := state.DeleteStatesByCascade(cn)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func GetAllStates(ctx *gin.Context) {
	states, err := state.GetAllStates()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNoContent, gin.H{"states": []interface{}{}})
		}
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	statesOut := make([]state.State, len(states))
	for idx, s := range states {
		statesOut[idx] = *s
	}

	ctx.JSON(http.StatusOK, gin.H{"states": statesOut})
}

func GetStateByNames(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	tn := ctx.Param("task")

	s, err := state.GetStateByNames(cn, tn)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, *s)
}

func GetStatesByCascade(ctx *gin.Context) {
	cn := ctx.Param("cascade")

	s, err := state.GetStatesByCascade(cn)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, s)
}

func UpdateStateByNames(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	tn := ctx.Param("task")

	var s state.State
	if err := ctx.ShouldBindJSON(&s); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	err := state.UpdateStateByNames(cn, tn, &s)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

/*~~~~~~~~ INPUT ~~~~~~~~*/

func CreateInput(ctx *gin.Context) {
	var i input.Input
	if err := ctx.ShouldBindJSON(&i); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	err := input.CreateInput(&i)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created"})
}

func DeleteInputByNames(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	n := ctx.Param("name")

	err := input.DeleteInputByNames(cn, n)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func DeleteInputsByCascade(ctx *gin.Context) {
	cn := ctx.Param("cascade")

	err := input.DeleteInputsByCascade(cn)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func GetAllInputs(ctx *gin.Context) {
	inputs, err := input.GetAllInputs()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNoContent, gin.H{"inputs": []interface{}{}})
		}
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	inputsOut := make([]input.Input, len(inputs))
	for idx, i := range inputs {
		inputsOut[idx] = *i
	}

	ctx.JSON(http.StatusOK, gin.H{"inputs": inputsOut})
}

func GetInputByNames(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	n := ctx.Param("name")

	i, err := input.GetInputByNames(cn, n)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, *i)
}

func GetInputsByCascade(ctx *gin.Context) {
	cn := ctx.Param("cascade")

	i, err := input.GetInputsByCascade(cn)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, i)
}

func UpdateInputByNames(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	n := ctx.Param("name")

	var i input.Input
	if err := ctx.ShouldBindJSON(&i); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	err := input.UpdateInputByNames(cn, n, &i)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func UpdateInputDependenciesByName(ctx *gin.Context) {
	name := ctx.Param("cascade")

	var changed []string
	if err := ctx.ShouldBindJSON(&changed); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	err := manager.InputChangeStateChange(name, changed)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

/*~~~~~~~~ TASK ~~~~~~~~*/

func CreateTask(ctx *gin.Context) {
	var t task.Task
	if err := ctx.ShouldBindJSON(&t); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	err := task.CreateTask(&t)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created"})
}

func DeleteTaskByNames(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	tn := ctx.Param("task")

	err := task.DeleteTaskByNames(cn, tn)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func DeleteTasksByCascade(ctx *gin.Context) {
	cn := ctx.Param("cascade")

	err := task.DeleteTasksByCascade(cn)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func GetAllTasks(ctx *gin.Context) {
	tasks, err := task.GetAllTasks()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNoContent, gin.H{"tasks": []interface{}{}})
		}
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	tasksOut := make([]task.Task, len(tasks))
	for idx, t := range tasks {
		tasksOut[idx] = *t
	}

	ctx.JSON(http.StatusOK, gin.H{"tasks": tasksOut})
}

func GetTaskByNames(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	tn := ctx.Param("task")

	t, err := task.GetTaskByNames(cn, tn)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, *t)
}

func GetTasksByCascade(ctx *gin.Context) {
	cn := ctx.Param("cascade")

	t, err := task.GetTasksByCascade(cn)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, t)
}

func UpdateTaskByNames(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	tn := ctx.Param("task")

	var t task.Task
	if err := ctx.ShouldBindJSON(&t); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	err := task.UpdateTaskByNames(cn, tn, &t)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

/*~~~~~~~~ USER ~~~~~~~~*/

func CreateUser(ctx *gin.Context) {
	var u user.User
	if err := ctx.ShouldBindJSON(&u); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	err := user.CreateUser(&u)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created"})
}

func DeleteUserByUsername(ctx *gin.Context) {
	username := ctx.Param("username")

	err := user.DeleteUserByUsername(username)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func GetAllUsers(ctx *gin.Context) {
	users, err := user.GetAllUsers()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNoContent, gin.H{"users": []interface{}{}})
		}
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	usersOut := make([]user.User, len(users))
	for idx, u := range users {
		usersOut[idx] = *u
	}

	ctx.JSON(http.StatusOK, gin.H{"users": usersOut})
}

func GetUserByUsername(ctx *gin.Context) {
	username := ctx.Param("username")

	u, err := user.GetUserByUsername(username)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, *u)
}

func UpdateUserByUsername(ctx *gin.Context) {
	username := ctx.Param("username")

	var u user.User
	if err := ctx.ShouldBindJSON(&u); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	uu, err := user.GetUserByUsername(username)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}
	uu.GivenName = u.GivenName
	uu.FamilyName = u.FamilyName
	uu.Email = u.Email
	uu.Groups = u.Groups
	uu.Roles = u.Roles

	if uu.Password != u.Password {
		uu.Password, err = user.HashAndSalt([]byte(u.Password))
		if err != nil {
			utils.Error(err, ctx, http.StatusInternalServerError)
			return
		}
	}

	err = user.UpdateUserByUsername(username, uu)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func GenerateAPIToken(ctx *gin.Context) {
	username := ctx.Param("username")
	name := ctx.Param("name")

	token, err := user.GenerateAPIToken(username, name)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"token": token})
}

func RevokeAPIToken(ctx *gin.Context) {
	username := ctx.Param("username")
	name := ctx.Param("name")

	err := user.RevokeAPIToken(username, name)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func CreateRun(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	tn := ctx.Param("task")

	s, err := state.GetStateByNames(cn, tn)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	t, err := task.GetTaskByNames(cn, tn)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}
	t.RunNumber += 1

	previousName := fmt.Sprintf("SCAFFOLD_PREVIOUS-%s", t.Name)
	ps := *s
	ps.Task = previousName
	err = state.UpdateStateByNames(cn, previousName, &ps)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	s.Status = constants.STATE_STATUS_WAITING

	obj := run.Run{
		Name:   fmt.Sprintf("%s.%s.%d", cn, tn, t.RunNumber),
		Task:   *t,
		State:  *s,
		Number: t.RunNumber,
	}
	postBody, _ := json.Marshal(obj)
	postBodyBuffer := bytes.NewBuffer(postBody)

	n, err := getAvailableNode()
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("http://%s:%d/api/v1/trigger", n.Host, n.Port)
	req, _ := http.NewRequest("POST", requestURL, postBodyBuffer)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", config.Config.Node.PrimaryKey))
	resp, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode >= 400 {
		utils.Error(fmt.Errorf("received trigger status code %d", resp.StatusCode), ctx, resp.StatusCode)
	}
	if _, ok := manager.InProgress[cn]; !ok {
		manager.InProgress[cn] = map[string]string{fmt.Sprintf("%s.%d", tn, t.RunNumber): fmt.Sprintf("%s:%d", n.Host, n.Port)}
	} else {
		manager.InProgress[cn][fmt.Sprintf("%s.%d", tn, t.RunNumber)] = fmt.Sprintf("%s:%d", n.Host, n.Port)
	}

	if err := task.UpdateTaskByNames(cn, tn, t); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	cs, err := cascade.GetCascadeByName(cn)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}
	manager.SetDependsState(cs, tn)

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func CreateCheckRun(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	tn := ctx.Param("task")

	t, err := task.GetTaskByNames(cn, tn)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}
	t.Check.RunNumber += 1

	checkStateName := fmt.Sprintf("SCAFFOLD_CHECK-%s", tn)

	s, err := state.GetStateByNames(cn, checkStateName)
	if err != nil {
		logger.Infof("", "No existing state for %s/%s found", cn, checkStateName)
		s = &state.State{
			Task:     checkStateName,
			Cascade:  cn,
			Status:   constants.STATE_STATUS_WAITING,
			Started:  "",
			Finished: "",
			Output:   "",
		}
		if err := state.CreateState(s); err != nil {
			utils.Error(err, ctx, http.StatusInternalServerError)
			return
		}
	}

	obj := run.Run{
		Name: fmt.Sprintf("%s.%s.%d", cn, checkStateName, t.Check.RunNumber),
		Task: task.Task{
			Name:      checkStateName,
			Cascade:   t.Cascade,
			Verb:      "",
			DependsOn: []string{},
			Image:     t.Check.Image,
			Run:       t.Check.Run,
			Store:     t.Check.Store,
			Load:      t.Check.Load,
			Outputs:   t.Check.Outputs,
			Inputs:    t.Check.Inputs,
			Updated:   t.Check.Updated,
			ShouldRM:  true,
		},
		State:  *s,
		Number: t.RunNumber,
	}
	postBody, _ := json.Marshal(obj)
	postBodyBuffer := bytes.NewBuffer(postBody)

	n, err := getAvailableNode()
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("http://%s:%d/api/v1/trigger", n.Host, n.Port)
	req, _ := http.NewRequest("POST", requestURL, postBodyBuffer)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", config.Config.Node.PrimaryKey))
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Error("", err.Error())
	}
	if resp.StatusCode >= 400 {
		logger.Errorf("", "Received trigger status code %d", resp.StatusCode)
	}
	if _, ok := manager.InProgress[cn]; !ok {
		manager.InProgress[cn] = map[string]string{fmt.Sprintf("%s.%d", checkStateName, t.Check.RunNumber): fmt.Sprintf("%s:%d", n.Host, n.Port)}
	} else {
		manager.InProgress[cn][fmt.Sprintf("%s.%d", checkStateName, t.Check.RunNumber)] = fmt.Sprintf("%s:%d", n.Host, n.Port)
	}

	if err := task.UpdateTaskByNames(cn, tn, t); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func getAvailableNode() (*auth.NodeObject, error) {
	if len(auth.Nodes) == 0 {
		return nil, fmt.Errorf("no nodes to schedule runs on")
	}
	nodeIdx := auth.LastScheduledIdx + 1

	for idx, n := range auth.Nodes {
		queryURL := fmt.Sprintf("http://%s:%d/health/available", n.Host, n.Port)
		resp, err := http.Get(queryURL)
		if err != nil || resp.StatusCode >= 400 {
			continue
		}
		nodeIdx = idx
		break
	}
	if nodeIdx >= len(auth.Nodes) {
		nodeIdx = 0
	}
	auth.LastScheduledIdx = nodeIdx

	return &auth.Nodes[nodeIdx], nil
}

func GetAllContainers(ctx *gin.Context) {
	available := map[string][]string{}
	for _, n := range auth.Nodes {
		httpClient := &http.Client{}
		requestURL := fmt.Sprintf("http://%s:%d/api/v1/available", n.Host, n.Port)
		req, _ := http.NewRequest("GET", requestURL, nil)
		req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", config.Config.Node.PrimaryKey))
		req.Header.Set("Content-Type", "application/json")
		resp, err := httpClient.Do(req)

		if err != nil {
			logger.Errorf("", "Error getting available containers %s", err.Error())
			continue
		}
		if resp.StatusCode == http.StatusOK {
			//Read the response body
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Errorf("", "Error reading body: %s", err.Error())
				continue
			}
			var data map[string][]string
			json.Unmarshal(body, &data)

			if len(data["containers"]) > 0 {
				available[fmt.Sprintf("%s:%d", n.Host, n.Port)] = data["containers"]
			}
			resp.Body.Close()
		}
	}
	ctx.JSON(http.StatusOK, available)
}

/*
+----------------+
|   WORKER API   |
+----------------+
*/

func TriggerRun(ctx *gin.Context) {
	var r run.Run
	if err := ctx.ShouldBindJSON(&r); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	names := strings.Split(r.Name, ".")

	r.State = state.State{
		Task:     names[1],
		Cascade:  names[0],
		Status:   constants.STATE_STATUS_WAITING,
		Started:  "",
		Finished: "",
		Output:   "",
	}

	worker.RunQueue = append(worker.RunQueue, r)
}

func GetRunState(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	tn := ctx.Param("task")
	n := ctx.Param("number")

	runName := fmt.Sprintf("%s.%s.%s", cn, tn, n)
	if container.CurrentRun.Name == runName {
		logger.Debugf("", "Run %s is currently running", runName)
		ctx.JSON(http.StatusOK, gin.H{"state": container.CurrentRun.State})
		return
	}
	for _, r := range worker.RunQueue {
		if r.Name == runName {
			logger.Debugf("", "Run %s is waiting in queue", runName)
			ctx.JSON(http.StatusOK, gin.H{"state": r.State})
			return
		}
	}
	if r, ok := container.CompletedRuns[runName]; ok {
		logger.Debugf("", "Run %s is completed", runName)
		ctx.JSON(http.StatusOK, gin.H{"state": r.State})
		delete(container.CompletedRuns, runName)
		return
	}
	ctx.Status(http.StatusNotFound)
}

func DownloadFile(ctx *gin.Context) {
	name := ctx.Param("name")
	fileName := ctx.Param("file")

	path := fmt.Sprintf("/tmp/%s", uuid.New().String())

	err := filestore.GetFile(fmt.Sprintf("%s/%s", name, fileName), path)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}
	ctx.Header("Content-Disposition", "attachment; filename="+fileName)
	ctx.Header("Content-Type", "application/text/plain")
	ctx.Header("Accept-Length", fmt.Sprintf("%d", len(data)))
	ctx.Writer.Write([]byte(data))
	ctx.Status(http.StatusOK)

	if err := os.Remove(path); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
	}
}

func UploadFile(ctx *gin.Context) {
	name := ctx.Param("name")

	file, err := ctx.FormFile("file")
	fileName := file.Filename

	// The file cannot be received.
	if err != nil {
		utils.Error(err, ctx, http.StatusBadRequest)
		return
	}

	path := fmt.Sprintf("/tmp/%s", uuid.New().String())

	// The file is received, so let's save it
	if err := ctx.SaveUploadedFile(file, path); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if err := filestore.UploadFile(path, fmt.Sprintf("%s/%s", name, fileName)); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if err := os.Remove(path); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ds, err := datastore.GetDataStoreByName(name)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ds.Files = append(ds.Files, fileName)

	inputs := []input.Input{}

	if err := datastore.UpdateDataStoreByName(name, ds, inputs); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	// File saved successfully. Return proper result
	utils.DynamicAPIResponse(ctx, "/ui/files", http.StatusOK, gin.H{"message": "OK"})
}

func GetAvailableContainers(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"containers": container.LastRun})
}
