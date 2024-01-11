package api

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"scaffold/server/cascade"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/datastore"
	"scaffold/server/filestore"
	"scaffold/server/input"
	"scaffold/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
//	@router						/api/v1/datastore/{datastore_name} [delete]
func DeleteDataStoreByName(ctx *gin.Context) {
	name := ctx.Param("name")

	err := datastore.DeleteDataStoreByName(name)

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
//	@router						/api/v1/datastore/{datastore_name} [get]
func GetDataStoreByName(ctx *gin.Context) {
	name := ctx.Param("name")

	d, err := datastore.GetDataStoreByName(name)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
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
//	@router						/api/v1/datastore/{datastore_name} [put]
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

//	@summary					Download a file
//	@description				Download a file from a datastore
//	@tags						manager
//	@tags						datastore
//	@tags						file
//	@produce					application/text/plain
//	@success					200
//	@failure					500
//	@failure					401
//	@failure					404
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/datastore/file/{datastore_name}/{file_name} [get]
func DownloadFile(ctx *gin.Context) {
	name := ctx.Param("name")
	fileName := ctx.Param("file")

	c, err := cascade.GetCascadeByName(name)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
	}
	if c.Groups != nil {
		if !validateUserGroup(ctx, c.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusForbidden)
			return
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
	if err := os.Remove(path); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}
	ctx.Header("Content-Disposition", "attachment; filename="+fileName)
	ctx.Header("Content-Type", "application/text/plain")
	ctx.Header("Accept-Length", fmt.Sprintf("%d", data))
	ctx.Writer.Write(data)
	ctx.Status(http.StatusOK)
}

func ViewFile(ctx *gin.Context) {
	name := ctx.Param("name")
	fileName := ctx.Param("file")

	c, err := cascade.GetCascadeByName(name)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
	}
	if c.Groups != nil {
		if !validateUserGroup(ctx, c.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusForbidden)
			return
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
	if err := os.Remove(path); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.Data(http.StatusOK, "application/text/plain", []byte(data))
}

//	@summary					Upload a file
//	@description				Upload a file to a datastore
//	@tags						manager
//	@tags						datastore
//	@tags						file
//	@accept						multipart/form-data
//	@produce					json
//	@success					200
//	@failure					500
//	@failure					400
//	@failure					401
//	@failure					404
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/datastore/file/{datastore_name} [post]
func UploadFile(ctx *gin.Context) {
	name := ctx.Param("name")

	c, err := cascade.GetCascadeByName(name)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
	}
	if c.Groups != nil {
		if !validateUserGroup(ctx, c.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusForbidden)
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
	ds.Files = utils.RemoveDuplicateValues(ds.Files)

	inputs := []input.Input{}

	if err := datastore.UpdateDataStoreByName(name, ds, inputs); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	// File saved successfully. Return proper result
	utils.DynamicAPIResponse(ctx, "/ui/files", http.StatusOK, gin.H{"message": "OK"})
}

func GetFilesByCascade(ctx *gin.Context) {
	name := ctx.Param("name")

	c, err := cascade.GetCascadeByName(name)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
	}
	if c.Groups != nil {
		if !validateUserGroup(ctx, c.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusForbidden)
			return
		}
	}

	objects, err := filestore.ListObjects()
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	out := make([]filestore.ObjectMetadata, 0)

	for _, obj := range objects {
		if obj.Cascade == name {
			out = append(out, obj)
		}
	}

	ctx.JSON(http.StatusOK, out)
}

func GetFileByNames(ctx *gin.Context) {
	name := ctx.Param("name")
	file := ctx.Param("file")

	c, err := cascade.GetCascadeByName(name)
	if err != nil {
		utils.Error(err, ctx, http.StatusNotFound)
	}
	if c.Groups != nil {
		if !validateUserGroup(ctx, c.Groups) {
			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusForbidden)
			return
		}
	}

	objects, err := filestore.ListObjects()
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	obj, ok := objects[file]
	if !ok {
		utils.Error(fmt.Errorf("file %s does not exist in datastore %s", file, name), ctx, http.StatusInternalServerError)
		return
	}

	if obj.Cascade != name {
		utils.Error(fmt.Errorf("cascade %s does not have file %s", name, file), ctx, http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, obj)
}
