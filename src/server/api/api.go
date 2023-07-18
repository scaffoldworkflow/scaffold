package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"scaffold/server/cascade"
	"scaffold/server/datastore"
	"scaffold/server/state"
	"scaffold/server/user"
	"scaffold/server/utils"
)

var IsHealthy = false
var IsReady = false
var IsAvailable = false

func Healthy(c *gin.Context) {
	if IsHealthy {
		c.Status(http.StatusOK)
		return
	}

	c.Status(http.StatusServiceUnavailable)
}

func Ready(c *gin.Context) {
	if IsReady {
		c.Status(http.StatusOK)
		return
	}

	c.Status(http.StatusServiceUnavailable)
}

func Available(c *gin.Context) {
	if IsAvailable {
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

	ctx.Status(http.StatusCreated)
}

func DeleteCascadeByName(ctx *gin.Context) {
	name := ctx.Param("name")

	err := cascade.DeleteCascadeByName(name)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
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

	ctx.Status(http.StatusOK)
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

	ctx.Status(http.StatusCreated)
}

func DeleteDataStoreByName(ctx *gin.Context) {
	name := ctx.Param("name")

	err := datastore.DeleteDataStoreByName(name)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
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

	ctx.Status(http.StatusOK)
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

	ctx.Status(http.StatusCreated)
}

func DeleteStateByName(ctx *gin.Context) {
	name := ctx.Param("name")

	err := state.DeleteStateByName(name)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
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

func GetStateByName(ctx *gin.Context) {
	name := ctx.Param("name")

	s, err := state.GetStateByName(name)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, *s)
}

func UpdateStateByName(ctx *gin.Context) {
	name := ctx.Param("name")

	var s state.State
	if err := ctx.ShouldBindJSON(&s); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	err := state.UpdateStateByName(name, &s)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
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

	ctx.Status(http.StatusCreated)
}

func DeleteUserByUsername(ctx *gin.Context) {
	username := ctx.Param("username")

	err := user.DeleteUserByUsername(username)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
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

	ctx.Status(http.StatusOK)
}
