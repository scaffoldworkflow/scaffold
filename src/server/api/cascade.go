package api

import (
	"errors"
	"net/http"
	"scaffold/server/cascade"
	"scaffold/server/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

//	@summary					Create a cascade
//	@description				Create a cascade from a JSON object
//	@tags						manager
//	@tags						cascade
//	@accept						json
//	@produce					json
//	@Param						cascade	body		cascade.Cascade	true	"Cascade Data"
//	@success					201		{object}	object
//	@failure					500		{object}	object
//	@failure					401		{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/cascade [post]
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

//	@summary					Delete a cascade
//	@description				Delete a cascade by its name
//	@tags						manager
//	@tags						cascade
//	@produce					json
//	@success					200	{object}	object
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/cascade/{cascade_name} [delete]
func DeleteCascadeByName(ctx *gin.Context) {
	name := ctx.Param("name")

	err := cascade.DeleteCascadeByName(name)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

//	@summary					Get all cascades
//	@description				Get all cascades
//	@tags						manager
//	@tags						cascade
//	@produce					json
//	@success					200	{array}		cascade.Cascade
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/cascade [get]
func GetAllCascades(ctx *gin.Context) {
	cascades, err := cascade.GetAllCascades()

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

	ctx.JSON(http.StatusOK, cascadesOut)
}

//	@summary					Get a cascade
//	@description				Get a cascade by its name
//	@tags						manager
//	@tags						cascade
//	@produce					json
//	@success					200	{object}	cascade.Cascade
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/cascade/{cascade_name} [get]
func GetCascadeByName(ctx *gin.Context) {
	name := ctx.Param("name")

	c, err := cascade.GetCascadeByName(name)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, *c)
}

//	@summary					Update a cascade
//	@description				Update a cascade from a JSON object
//	@tags						manager
//	@tags						cascade
//	@accept						json
//	@produce					json
//	@Param						cascade	body		cascade.Cascade	true	"Cascade Data"
//	@success					201		{object}	object
//	@failure					500		{object}	object
//	@failure					401		{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/cascade/{cascade_name} [put]
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
