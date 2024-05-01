package api

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"scaffold/server/cascade"
	"scaffold/server/datastore"
	"scaffold/server/filestore"
	"scaffold/server/input"
	"scaffold/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

//	@summary					Download a file
//	@description				Download a file from a cascade
//	@tags						manager
//	@tags						file
//	@produce					application/text
//	@success					200
//	@failure					500
//	@failure					401
//	@failure					404
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/file/{cascade_name}/{file_name}/download [get]
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

// func ViewFile(ctx *gin.Context) {
// 	name := ctx.Param("name")
// 	fileName := ctx.Param("file")

// 	c, err := cascade.GetCascadeByName(name)
// 	if err != nil {
// 		utils.Error(err, ctx, http.StatusNotFound)
// 	}
// 	if c.Groups != nil {
// 		if !validateUserGroup(ctx, c.Groups) {
// 			utils.Error(errors.New("user is not part of required groups to access this resources"), ctx, http.StatusForbidden)
// 			return
// 		}
// 	}

// 	path := fmt.Sprintf("/tmp/%s", uuid.New().String())

// 	err = filestore.GetFile(fmt.Sprintf("%s/%s", name, fileName), path)
// 	if err != nil {
// 		utils.Error(err, ctx, http.StatusInternalServerError)
// 		return
// 	}
// 	data, err := os.ReadFile(path)
// 	if err != nil {
// 		utils.Error(err, ctx, http.StatusInternalServerError)
// 		return
// 	}
// 	if err := os.Remove(path); err != nil {
// 		utils.Error(err, ctx, http.StatusInternalServerError)
// 		return
// 	}

// 	ctx.Data(http.StatusOK, "application/text/plain", []byte(data))
// }

//	@summary					Upload a file
//	@description				Upload a file to a cascade
//	@tags						manager
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
//	@router						/api/v1/file/{datastore_name} [post]
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

	ds, err := datastore.GetDataStoreByCascade(name)
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ds.Files = append(ds.Files, fileName)
	ds.Files = utils.RemoveDuplicateValues(ds.Files)

	inputs := []input.Input{}

	if err := datastore.UpdateDataStoreByCascade(name, ds, inputs); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	// File saved successfully. Return proper result
	utils.DynamicAPIResponse(ctx, "/ui/files", http.StatusOK, gin.H{"message": "OK"})
}

//	@summary					Get files
//	@description				Get files by cascade
//	@tags						manager
//	@tags						file
//	@produce					json
//	@success					200	{array}	filestore.ObjectMetadata
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/file/{cascade_name} [get]
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

//	@summary					Get file
//	@description				Get file by cascade and name
//	@tags						manager
//	@tags						file
//	@produce					json
//	@success					200	{object}	filestore.ObjectMetadata
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/file/{cascade_name}/{file_name} [get]
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
		utils.Error(fmt.Errorf("file %s does not exist in datastore %s", file, name), ctx, http.StatusNotFound)
		return
	}

	if obj.Cascade != name {
		utils.Error(fmt.Errorf("cascade %s does not have file %s", name, file), ctx, http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, obj)
}

//	@summary					Get all files
//	@description				Get all files
//	@tags						manager
//	@tags						file
//	@produce					json
//	@success					200	{array}	filestore.ObjectMetadata
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/file [get]
func GetAllFiles(ctx *gin.Context) {
	objects, err := filestore.ListObjects()
	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	out := make([]filestore.ObjectMetadata, 0)

	for _, obj := range objects {
		c, err := cascade.GetCascadeByName(obj.Cascade)
		if err != nil {
			continue
		}
		if c.Groups != nil {
			if validateUserGroup(ctx, c.Groups) {
				out = append(out, obj)
			}
			continue
		}
		out = append(out, obj)
	}

	ctx.JSON(http.StatusOK, out)
}
