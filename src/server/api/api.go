// API implements worker and manager API endpoints for Scaffold functionality
package api

import (
	"bytes"
	"encoding/json"
	"errors"
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

// Check if the Scaffold node is healthy
func Healthy(c *gin.Context) {
	if health.IsHealthy {
		c.JSON(http.StatusOK, gin.H{"version": constants.VERSION})
		return
	}

	c.Status(http.StatusServiceUnavailable)
}

// Check if the Scaffold node is ready
func Ready(c *gin.Context) {
	if health.IsReady {
		c.Status(http.StatusOK)
		return
	}

	c.Status(http.StatusServiceUnavailable)
}

// Check if a worker node is available
// This corresponds to no containers currently running
func Available(c *gin.Context) {
	if health.IsAvailable {
		c.Status(http.StatusOK)
		return
	}

	c.Status(http.StatusServiceUnavailable)
}

/*~~~~~~~~ CASCADE ~~~~~~~~*/

// Create a cascade from a JSON object
func CreateCascade(ctx *gin.Context) {
	var c cascade.Cascade
	if err := ctx.ShouldBindJSON(&c); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if c.Groups != nil {
		if !validateUserGroup(ctx, c.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusUnauthorized)
		}
	}

	err := cascade.CreateCascade(&c)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created"})
}

// Delete a cascade by its name
func DeleteCascadeByName(ctx *gin.Context) {
	name := ctx.Param("name")

	err := cascade.DeleteCascadeByName(name)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

// Get all cascade objects that the Scaffold instance knows about
func GetAllCascades(ctx *gin.Context) {
	cascades, err := cascade.GetAllCascades()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNoContent, gin.H{"cascades": []interface{}{}})
		}
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	// Need to copy each cascade from pointer to value since pointers are returned
	// weirdly (I think at least)
	cascadesOut := make([]cascade.Cascade, 0)
	for _, c := range cascades {
		if c.Groups != nil {
			if validateUserGroup(ctx, c.Groups) {
				cascadesOut = append(cascadesOut, *c)
			}
			continue
		}
		cascadesOut = append(cascadesOut, *c)
	}

	ctx.JSON(http.StatusOK, gin.H{"cascades": cascadesOut})
}

// Get a cascade by its name
func GetCascadeByName(ctx *gin.Context) {
	name := ctx.Param("name")

	c, err := cascade.GetCascadeByName(name)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, *c)
}

// Update a Cascade by name with a JSON object
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

// Create a datastore by a JSON object
func CreateDataStore(ctx *gin.Context) {
	var d datastore.DataStore
	if err := ctx.ShouldBindJSON(&d); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	c, err := cascade.GetCascadeByName(d.Name)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
	}
	if c.Groups != nil {
		if !validateUserGroup(ctx, c.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusUnauthorized)
		}
	}

	err = datastore.CreateDataStore(&d)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created"})
}

// Delete a datastore by name
func DeleteDataStoreByName(ctx *gin.Context) {
	name := ctx.Param("name")

	err := datastore.DeleteDataStoreByName(name)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

// Get all datastores a Scaffold instance knows about
func GetAllDataStores(ctx *gin.Context) {
	datastores, err := datastore.GetAllDataStores()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNoContent, gin.H{"datastores": []interface{}{}})
		}
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	// Need to copy each cascade from pointer to value since pointers are returned
	// weirdly (I think at least)
	datastoresOut := make([]datastore.DataStore, 0)
	for _, d := range datastores {
		c, err := cascade.GetCascadeByName(d.Name)
		if err != nil {
			continue
		}
		if c.Groups != nil {
			if validateUserGroup(ctx, c.Groups) {
				datastoresOut = append(datastoresOut, *d)
			}
			continue
		}
		datastoresOut = append(datastoresOut, *d)
	}

	ctx.JSON(http.StatusOK, gin.H{"datastores": datastoresOut})
}

// Get a datastore by name
func GetDataStoreByName(ctx *gin.Context) {
	name := ctx.Param("name")

	d, err := datastore.GetDataStoreByName(name)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, *d)
}

// Update a datastore by name from a JSON object
func UpdateDataStoreByName(ctx *gin.Context) {
	name := ctx.Param("name")

	var d datastore.DataStore
	if err := ctx.ShouldBindJSON(&d); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	// Need to copy over cascade inputs since some weirdness happens when updating the
	// datastore
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

// Create a state from a JSON object
func CreateState(ctx *gin.Context) {
	var s state.State
	if err := ctx.ShouldBindJSON(&s); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	c, err := cascade.GetCascadeByName(s.Cascade)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
	}
	if c.Groups != nil {
		if !validateUserGroup(ctx, c.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusUnauthorized)
		}
	}

	err = state.CreateState(&s)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created"})
}

// Delete a state by name
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

	statesOut := make([]state.State, 0)
	for _, s := range states {
		c, err := cascade.GetCascadeByName(s.Cascade)
		if err != nil {
			continue
		}
		if c.Groups != nil {
			if validateUserGroup(ctx, c.Groups) {
				statesOut = append(statesOut, *s)
			}
			continue
		}
		statesOut = append(statesOut, *s)
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

	c, err := cascade.GetCascadeByName(i.Cascade)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
	}
	if c.Groups != nil {
		if !validateUserGroup(ctx, c.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusUnauthorized)
		}
	}

	err = input.CreateInput(&i)

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

	inputsOut := make([]input.Input, 0)
	for _, i := range inputs {
		c, err := cascade.GetCascadeByName(i.Cascade)
		if err == nil {
			continue
		}
		if c.Groups != nil {
			if validateUserGroup(ctx, c.Groups) {
				inputsOut = append(inputsOut, *i)
			}
			continue
		}
		inputsOut = append(inputsOut, *i)
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

	c, err := cascade.GetCascadeByName(t.Cascade)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
	}
	if c.Groups != nil {
		if !validateUserGroup(ctx, c.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusUnauthorized)
		}
	}

	err = task.CreateTask(&t)

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

	tasksOut := make([]task.Task, 0)
	for _, t := range tasks {
		c, err := cascade.GetCascadeByName(t.Cascade)
		if err != nil {
			continue
		}
		if c.Groups != nil {
			if validateUserGroup(ctx, c.Groups) {
				tasksOut = append(tasksOut, *t)
			}
			continue
		}
		tasksOut = append(tasksOut, *t)
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

	c, err := cascade.GetCascadeByName(cn)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

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

	s.Number += 1

	s.Status = constants.STATE_STATUS_WAITING

	obj := run.Run{
		Name:   fmt.Sprintf("%s.%s.%d", cn, tn, t.RunNumber),
		Task:   *t,
		State:  *s,
		Number: t.RunNumber,
		Groups: c.Groups,
	}
	postBody, _ := json.Marshal(obj)
	postBodyBuffer := bytes.NewBuffer(postBody)

	n, err := getAvailableNode()
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	httpClient := http.Client{}
	requestURL := fmt.Sprintf("%s://%s:%d/api/v1/trigger", n.Protocol, n.Host, n.Port)
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
		manager.InProgress[cn] = map[string]string{fmt.Sprintf("%s.%d", tn, t.RunNumber): fmt.Sprintf("%s://%s:%d", n.Protocol, n.Host, n.Port)}
		manager.ToCheck[cn] = map[string]string{fmt.Sprintf("%s.%d", tn, t.RunNumber): fmt.Sprintf("%s://%s:%d", n.Protocol, n.Host, n.Port)}
	} else {
		manager.InProgress[cn][fmt.Sprintf("%s.%d", tn, t.RunNumber)] = fmt.Sprintf("%s://%s:%d", n.Protocol, n.Host, n.Port)
		manager.ToCheck[cn][fmt.Sprintf("%s.%d", tn, t.RunNumber)] = fmt.Sprintf("%s://%s:%d", n.Protocol, n.Host, n.Port)
	}

	if err := task.UpdateTaskByNames(cn, tn, t); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	// if err := state.UpdateStateByNames(cn, tn, s); err != nil {
	// 	utils.Error(err, ctx, http.StatusInternalServerError)
	// 	return
	// }

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

	c, err := cascade.GetCascadeByName(cn)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

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
			Number:   t.RunNumber,
			Display:  make([]map[string]interface{}, 0),
		}
		if err := state.CreateState(s); err != nil {
			utils.Error(err, ctx, http.StatusInternalServerError)
			return
		}
	}
	s.Number = t.RunNumber

	obj := run.Run{
		Name: fmt.Sprintf("%s.%s.%d", cn, checkStateName, t.Check.RunNumber),
		Task: task.Task{
			Name:        checkStateName,
			Cascade:     t.Cascade,
			Verb:        "",
			DependsOn:   task.TaskDependsOn{},
			Image:       t.Check.Image,
			Run:         t.Check.Run,
			Store:       t.Check.Store,
			Load:        t.Check.Load,
			Env:         t.Check.Env,
			Inputs:      t.Check.Inputs,
			Updated:     t.Check.Updated,
			AutoExecute: true,
			ShouldRM:    true,
			RunNumber:   t.RunNumber,
		},
		State:  *s,
		Number: t.RunNumber,
		Groups: c.Groups,
	}
	postBody, _ := json.Marshal(obj)
	postBodyBuffer := bytes.NewBuffer(postBody)

	n, err := getAvailableNode()
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	httpClient := http.Client{}
	requestURL := fmt.Sprintf("%s://%s:%d/api/v1/trigger", n.Protocol, n.Host, n.Port)
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
		manager.InProgress[cn] = map[string]string{fmt.Sprintf("%s.%d", checkStateName, t.Check.RunNumber): fmt.Sprintf("%s://%s:%d", n.Protocol, n.Host, n.Port)}
		manager.ToCheck[cn] = map[string]string{fmt.Sprintf("%s.%d", checkStateName, t.Check.RunNumber): fmt.Sprintf("%s://%s:%d", n.Protocol, n.Host, n.Port)}
	} else {
		manager.InProgress[cn][fmt.Sprintf("%s.%d", checkStateName, t.Check.RunNumber)] = fmt.Sprintf("%s://%s:%d", n.Protocol, n.Host, n.Port)
		manager.ToCheck[cn][fmt.Sprintf("%s.%d", checkStateName, t.Check.RunNumber)] = fmt.Sprintf("%s://%s:%d", n.Protocol, n.Host, n.Port)
	}

	if err := task.UpdateTaskByNames(cn, tn, t); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	// if err := state.UpdateStateByNames(cn, tn, s); err != nil {
	// 	utils.Error(err, ctx, http.StatusInternalServerError)
	// 	return
	// }

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func ManagerKillRun(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	tn := ctx.Param("task")
	nn := ctx.Param("number")

	key := fmt.Sprintf("%s.%s", tn, nn)

	uri := ""
	logger.Debugf("", "Looking for %s/%s.%s", cn, tn, nn)
	logger.Debugf("", "Kill in progress: %s", manager.InProgress)
	if _, ok := manager.InProgress[cn]; ok {
		if val, ok := manager.InProgress[cn][key]; ok {
			uri = val
		}
	}

	httpClient := http.Client{}
	requestURL := fmt.Sprintf("%s://%s/api/v1/kill/%s/%s/%s", uri, cn, tn, nn)
	req, _ := http.NewRequest("DELETE", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("X-Scaffold-API %s", config.Config.Node.PrimaryKey))
	resp, err := httpClient.Do(req)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
	}
	if resp.StatusCode >= 400 {
		utils.Error(fmt.Errorf("received kill status code %d", resp.StatusCode), ctx, resp.StatusCode)
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
		queryURL := fmt.Sprintf("%s://%s:%d/health/available", n.Protocol, n.Host, n.Port)
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
		httpClient := http.Client{}
		requestURL := fmt.Sprintf("%s://%s:%d/api/v1/available", n.Protocol, n.Host, n.Port)
		req, _ := http.NewRequest("GET", requestURL, nil)
		req.Header.Set("Authorization", ctx.Request.Header.Get("X-Scaffold-API"))
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
				available[fmt.Sprintf("%s:%d", n.Host, n.WSPort)] = data["containers"]
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

	if r.Groups != nil {
		if !validateUserGroup(ctx, r.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusUnauthorized)
			return
		}
	}

	names := strings.Split(r.Name, ".")

	r.State = state.State{
		Task:     names[1],
		Cascade:  names[0],
		Status:   constants.STATE_STATUS_WAITING,
		Started:  "",
		Finished: "",
		Output:   "",
		Number:   r.Number,
		Display:  make([]map[string]interface{}, 0),
	}

	logger.Debugf("", "Writing new run to queue %v", r)

	worker.RunQueue = append(worker.RunQueue, r)
}

func KillRun(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	tn := ctx.Param("task")
	nn := ctx.Param("number")

	c, err := cascade.GetCascadeByName(cn)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
	}
	if c.Groups != nil {
		if !validateUserGroup(ctx, c.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusUnauthorized)
		}
	}

	if err := run.Kill(cn, tn, nn); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func GetRunState(ctx *gin.Context) {
	cn := ctx.Param("cascade")
	tn := ctx.Param("task")
	n := ctx.Param("number")

	c, err := cascade.GetCascadeByName(cn)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
	}
	if c.Groups != nil {
		if !validateUserGroup(ctx, c.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusUnauthorized)
		}
	}

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

	c, err := cascade.GetCascadeByName(name)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
	}
	if c.Groups != nil {
		if !validateUserGroup(ctx, c.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusUnauthorized)
		}
	}

	path := fmt.Sprintf("/tmp/%s", uuid.New().String())

	err = filestore.GetFile(fmt.Sprintf("%s/%s", name, fileName), path)
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

	c, err := cascade.GetCascadeByName(name)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
	}
	if c.Groups != nil {
		if !validateUserGroup(ctx, c.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusUnauthorized)
		}
	}

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
	output := []string{}

	for idx, groups := range container.LastGroups {
		if validateUserGroup(ctx, groups) {
			output = append(output, container.LastRun[idx])
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"containers": container.LastRun})
}

// HELPERS

func validateUserGroup(ctx *gin.Context, groups []string) bool {
	var token string
	var err error
	var usr *user.User

	logger.Infof("", "Validating user against groups: %v", groups)

	if len(groups) == 0 {
		return true
	}

	// Check if we have an auth header
	authString := ctx.Request.Header.Get("Authorization")
	if authString == "" {
		// Check if the request is coming from a logged in UI user
		token, err = ctx.Cookie("scaffold_token")
		if err != nil {
			return false
		}
		usr, _ = user.GetUserByLoginToken(token)
		if usr == nil {
			return false
		}
	} else {
		token = strings.Split(authString, " ")[1]
	}

	// Is the request coming from a node itself?
	if token == config.Config.Node.PrimaryKey {
		return true
	}

	// Get the user via the information
	usr, _ = user.GetUserByAPIToken(token)
	if usr == nil {
		return false
	}

	if utils.Contains(usr.Groups, "admin") {
		return true
	}
	for _, group := range groups {
		if utils.Contains(usr.Groups, group) {
			return true
		}
	}

	return false
}
