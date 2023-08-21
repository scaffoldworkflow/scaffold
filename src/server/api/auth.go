package api

import (
	"net/http"
	"scaffold/server/user"
	"scaffold/server/utils"

	"github.com/gin-gonic/gin"
)

//	@summary					Generate API Token
//	@description				Generate an API token for a user
//	@tags						manager
//	@tags						user
//	@produce					json
//	@success					200	{array}		object
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/auth/token/{username}/{token_name} [post]
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

//	@summary					Revoke API Token
//	@description				Revoke an API token for a user
//	@tags						manager
//	@tags						user
//	@produce					json
//	@success					200	{array}		object
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/auth/token/{username}/{token_name} [delete]
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
