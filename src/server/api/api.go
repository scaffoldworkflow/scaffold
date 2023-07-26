package api

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	"scaffold/server/datastore"
	"scaffold/server/filestore"
	"scaffold/server/health"
	"scaffold/server/input"
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

	err := datastore.UpdateDataStoreByName(name, &d)
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

	for _, ts := range t {
		fmt.Printf("Task: %v\n", *ts)
	}

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

	obj := run.Run{
		Name:  fmt.Sprintf("%s.%s", cn, tn),
		Task:  *t,
		State: *s,
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
		panic(fmt.Sprintf("Received trigger status code %d", resp.StatusCode))
	}
	if _, ok := manager.InProgress[cn]; !ok {
		manager.InProgress[cn] = map[string]string{tn: fmt.Sprintf("%s:%d", n.Host, n.Port)}
	} else {
		manager.InProgress[cn][tn] = fmt.Sprintf("%s:%d", n.Host, n.Port)
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

	runName := fmt.Sprintf("%s.%s", cn, tn)
	if worker.CurrentRun.Name == runName {
		ctx.JSON(http.StatusOK, gin.H{"state": worker.CurrentRun.State})
	}
	for _, r := range worker.RunQueue {
		if r.Name == runName {
			ctx.JSON(http.StatusOK, gin.H{"state": r.State})
			return
		}
	}
	for idx, r := range worker.CompletedRuns {
		if r.Name == runName {
			ctx.JSON(http.StatusOK, gin.H{"state": r.State})
			if len(worker.CompletedRuns) > 1 {
				worker.CompletedRuns = append(worker.CompletedRuns[:idx], worker.RunQueue[idx+1:]...)
			} else {
				worker.CompletedRuns = []run.Run{}
			}
			return
		}
	}
	ctx.Status(http.StatusNotFound)
}

func DownloadFile(ctx *gin.Context) {
	name := ctx.Param("name")

	path := fmt.Sprintf("/tmp/%s", uuid.New().String())

	err := filestore.GetFile(name, path)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}
	ctx.Header("Content-Disposition", "attachment; filename="+name)
	ctx.Header("Content-Type", "application/text/plain")
	ctx.Header("Accept-Length", fmt.Sprintf("%d", len(data)))
	ctx.Writer.Write([]byte(data))
	ctx.JSON(http.StatusOK, gin.H{
		"message": "OK",
	})

	if err := os.Remove(path); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
	}
}

func UploadFile(ctx *gin.Context) {
	name := ctx.Param("name")

	file, err := ctx.FormFile("file")

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

	if err := filestore.UploadFile(path, name); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if err := os.Remove(path); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	// File saved successfully. Return proper result
	ctx.JSON(http.StatusOK, gin.H{
		"message": "OK",
	})
}
