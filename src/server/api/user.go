package api

import (
	"net/http"
	"scaffold/server/user"
	"scaffold/server/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

//	@summary					Create a user
//	@description				Create a user from a JSON object
//	@tags						manager
//	@tags						user
//	@accept						json
//	@produce					json
//	@Param						user	body		user.User	true	"User Data"
//	@success					201		{object}	object
//	@failure					500		{object}	object
//	@failure					401		{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/user [post]
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

//	@summary					Delete a user
//	@description				Delete a user by its username
//	@tags						manager
//	@tags						user
//	@produce					json
//	@success					200	{object}	object
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/user/{username} [delete]
func DeleteUserByUsername(ctx *gin.Context) {
	username := ctx.Param("username")

	err := user.DeleteUserByUsername(username)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
}

//	@summary					Get all users
//	@description				Get all users
//	@tags						manager
//	@tags						user
//	@produce					json
//	@success					200	{array}		user.User
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/user [get]
func GetAllUsers(ctx *gin.Context) {
	users, err := user.GetAllUsers()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNoContent, []interface{}{})
			return
		}
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	usersOut := make([]user.User, len(users))
	for idx, u := range users {
		usersOut[idx] = *u
	}

	ctx.JSON(http.StatusOK, usersOut)
}

//	@summary					Get a user
//	@description				Get a user by its username
//	@tags						manager
//	@tags						user
//	@produce					json
//	@success					200	{array}		user.User
//	@failure					500	{object}	object
//	@failure					401	{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/user/{username} [get]
func GetUserByUsername(ctx *gin.Context) {
	username := ctx.Param("username")

	u, err := user.GetUserByUsername(username)

	if err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, *u)
}

//	@summary					Update a user
//	@description				Update a user from a JSON object
//	@tags						manager
//	@tags						user
//	@accept						json
//	@produce					json
//	@Param						user	body		user.User	true	"User Data"
//	@success					201		{object}	object
//	@failure					500		{object}	object
//	@failure					401		{object}	object
//	@securityDefinitions.apiKey	token
//	@in							header
//	@name						Authorization
//	@security					X-Scaffold-API
//	@router						/api/v1/user/{user_name} [put]
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
