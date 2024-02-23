package api

import (
	"errors"
	"fmt"
	"net/http"
	"scaffold/server/cascade"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/datastore"
	"scaffold/server/input"
	"scaffold/server/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

//	@summary					Create a datastore
//	@description				Create a datastore from a JSON object
//	@tags						manager
//	@tags						datastore
//	@accept						json
//	@produce					json
//	@Param						datastore	body		datastore.DataStore	true	"DataStore Data"
//	@success					201			{object}	object
//	@failure					500			{object}	object
//	@failure					401			{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/datastore [post]
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
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusForbidden)
		}
	}

	err = datastore.CreateDataStore(&d)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created"})
}

//	@summary					Delete a datastore
//	@description				Delete a datastore by its name
//	@tags						manager
//	@tags						datastore
//	@produce					json
//	@success					200	{object}	object
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/datastore/{cascade_name} [delete]
func DeleteDataStoreByCascade(ctx *gin.Context) {
	name := ctx.Param("name")

	err := datastore.DeleteDataStoreByCascade(name)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

//	@summary					Get all datastores
//	@description				Get all datastores
//	@tags						manager
//	@tags						datastore
//	@produce					json
//	@success					200	{array}		datastore.DataStore
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/datastore [get]
func GetAllDataStores(ctx *gin.Context) {
	datastores, err := datastore.GetAllDataStores()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNoContent, []interface{}{})
			return
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

	ctx.JSON(http.StatusOK, datastoresOut)
}

//	@summary					Get a datastore
//	@description				Get a datastore by its name
//	@tags						manager
//	@tags						datastore
//	@produce					json
//	@success					200	{object}	datastore.DataStore
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/datastore/{cascade_name} [get]
func GetDataStoreByName(ctx *gin.Context) {
	name := ctx.Param("name")

	d, err := datastore.GetDataStoreByCascade(name)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	if d == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Datastore %s does not exist", name)})
		return
	}

	ctx.JSON(http.StatusOK, *d)
}

//	@summary					Update a datastore
//	@description				Update a datastore from a JSON object
//	@tags						manager
//	@tags						datastore
//	@accept						json
//	@produce					json
//	@Param						datastore	body		datastore.DataStore	true	"DataStore Data"
//	@success					201			{object}	object
//	@failure					500			{object}	object
//	@failure					401			{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/datastore/{cascade_name} [put]
func UpdateDataStoreByCascade(ctx *gin.Context) {
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

	err := datastore.UpdateDataStoreByCascade(name, &d, inputs)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}
