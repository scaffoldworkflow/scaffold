package user

import (
	"fmt"
	"scaffold/server/config"
	"scaffold/server/constants"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"scaffold/server/mongodb"
)

type User struct {
	Username          string `json:"username" bson:"username"`
	Password          string `json:"password" bson:"password"`
	GivenName         string `json:"given_name" bson:"given_name"`
	FamilyName        string `json:"family_name" bson:"family_name"`
	Email             string `json:"email" bson:"email"`
	ResetToken        string `json:"reset_token" bson:"reset_token"`
	ResetTokenCreated string `json:"reset_token_created" bson:"reset_token_created"`
	Created           string `json:"created" bson:"created"`
	Updated           string `json:"updated" bson:"updated"`
	LoginToken        string `json:"login_token" bson:"login_token"`
}

func CreateUser(u *User) error {
	currentTime := time.Now().UTC()
	u.Created = currentTime.Format("2006-01-02T15:04:05Z")
	u.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	if _, err := GetUserByUsername(u.Username); err == nil {
		return fmt.Errorf("user already exists with username %s", u.Username)
	}

	password, err := HashAndSalt([]byte(u.Password))
	if err != nil {
		return err
	}

	u.Password = password

	_, err = mongodb.Collections[constants.MONGODB_USER_COLLECTION_NAME].InsertOne(mongodb.Ctx, u)
	return err
}

func DeleteUserByUsername(username string) error {
	filter := bson.M{"username": username}

	collection := mongodb.Collections[constants.MONGODB_USER_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount != 1 {
		return fmt.Errorf("no user found with username %s", username)
	}

	return nil

}

func GetAllUsers() ([]*User, error) {
	filter := bson.D{{}}

	users, err := FilterUsers(filter)

	return users, err
}

func GetUserByUsername(username string) (*User, error) {
	filter := bson.M{"username": username}

	users, err := FilterUsers(filter)

	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("no user found with username %s", username)
	}

	if len(users) > 1 {
		return nil, fmt.Errorf("multiple users found with username %s", username)
	}

	return users[0], nil
}

func GetUserByEmail(email string) (*User, error) {
	filter := bson.M{"email": email}

	users, err := FilterUsers(filter)

	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("no user found with email %s", email)
	}

	if len(users) > 1 {
		return nil, fmt.Errorf("multiple users found with email %s", email)
	}

	return users[0], nil
}

func GetUserByLoginToken(loginToken string) (*User, error) {
	if loginToken == "" {
		return nil, fmt.Errorf("invalid login token")
	}

	filter := bson.M{"login_token": loginToken}

	users, err := FilterUsers(filter)

	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("no user found with login token %s", loginToken)
	}

	if len(users) > 1 {
		return nil, fmt.Errorf("multiple users found with login token %s", loginToken)
	}

	return users[0], nil
}

func GetUserByResetToken(resetToken string) (*User, error) {
	filter := bson.M{"reset_token": resetToken}

	users, err := FilterUsers(filter)

	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("no user found with reset_token %s", resetToken)
	}

	if len(users) > 1 {
		return nil, fmt.Errorf("multiple users found with reset_token %s", resetToken)
	}

	return users[0], nil
}

func UpdateUserByUsername(username string, u *User) error {
	filter := bson.M{"username": username}

	currentTime := time.Now().UTC()
	u.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	collection := mongodb.Collections[constants.MONGODB_USER_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.ReplaceOne(ctx, filter, u)

	if err != nil {
		return err
	}

	if result.ModifiedCount != 1 {
		return fmt.Errorf("no user found with username %s", username)
	}

	return nil
}

func FilterUsers(filter interface{}) ([]*User, error) {
	// A slice of tasks for storing the decoded documents
	var users []*User

	collection := mongodb.Collections[constants.MONGODB_USER_COLLECTION_NAME]
	ctx := mongodb.Ctx

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return users, err
	}

	for cur.Next(ctx) {
		var u User
		err := cur.Decode(&u)
		if err != nil {
			return users, err
		}

		users = append(users, &u)
	}

	if err := cur.Err(); err != nil {
		return users, err
	}

	// once exhausted, close the cursor
	cur.Close(ctx)

	if len(users) == 0 {
		return users, mongo.ErrNoDocuments
	}

	return users, nil
}

func VerifyAdmin() error {
	user, _ := GetUserByUsername(config.Config.Admin.Username)

	if user != nil {
		return nil
	}

	u := &User{
		Username:          config.Config.Admin.Username,
		Password:          config.Config.Admin.Password,
		GivenName:         "admin",
		FamilyName:        "admin",
		Email:             config.Config.Admin.Email,
		ResetToken:        "",
		ResetTokenCreated: "",
	}

	err := CreateUser(u)

	return err
}

func HashAndSalt(pwd []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		return "", nil
	}
	return string(hash), nil
}

func VerifyUser(username, password string) (bool, error) {
	u, err := GetUserByUsername(username)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return false, err
	}
	return true, nil
}
