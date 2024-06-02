package user

import (
	"fmt"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"scaffold/server/mongodb"

	logger "github.com/jfcarter2358/go-logger"
)

type User struct {
	Username          string     `json:"username" bson:"username" yaml:"username"`
	Password          string     `json:"password" bson:"password" yaml:"password"`
	GivenName         string     `json:"given_name" bson:"given_name" yaml:"given_name"`
	FamilyName        string     `json:"family_name" bson:"family_name" yaml:"family_name"`
	Email             string     `json:"email" bson:"email" yaml:"email"`
	ResetToken        string     `json:"reset_token" bson:"reset_token" yaml:"reset_token"`
	ResetTokenCreated string     `json:"reset_token_created" bson:"reset_token_created" yaml:"reset_token_created"`
	Created           string     `json:"created" bson:"created" yaml:"created"`
	Updated           string     `json:"updated" bson:"updated" yaml:"updated"`
	LoginToken        string     `json:"login_token" bson:"login_token" yaml:"login_token"`
	APITokens         []APIToken `json:"api_tokens" bson:"api_tokens" yaml:"api_tokens"`
	Groups            []string   `json:"groups" bson:"groups" yaml:"groups"`
	Roles             []string   `json:"roles" bson:"roles" yaml:"roles"`
}

type APIToken struct {
	Name    string `json:"name" bson:"name" yaml:"name"`
	Token   string `json:"token" bson:"token" yaml:"token"`
	Created string `json:"created" bson:"created" yaml:"created"`
}

func CreateUser(u *User) error {
	currentTime := time.Now().UTC()
	u.Created = currentTime.Format("2006-01-02T15:04:05Z")
	u.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	uu, err := GetUserByUsername(u.Username)
	if err != nil {
		return fmt.Errorf("error getting users: %s", err.Error())
	}
	if uu != nil {
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

	allUsers, err := GetAllUsers()
	if err != nil {
		logger.Errorf("", "Unable to get all users")
		return nil, err
	}
	logger.Debugf("", "All users: %v", allUsers)

	logger.Debugf("", "Searching for username %s with filter %v", username, filter)

	users, err := FilterUsers(filter)

	logger.Debugf("", "Got users: %v", users)

	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, nil
	}

	if len(users) > 1 {
		return nil, fmt.Errorf("multiple users found with username %s", username)
	}

	return users[0], nil
}

func GetUserByAPIToken(apiToken string) (*User, error) {
	// filter := bson.M{"api_tokens": bson.M{"token": apiToken}}

	filter := bson.D{{}}
	users, err := FilterUsers(filter)

	if err != nil {
		return nil, err
	}

	for _, u := range users {
		for _, t := range u.APITokens {
			if err := bcrypt.CompareHashAndPassword([]byte(t.Token), []byte(apiToken)); err == nil {
				return u, nil
			}
		}
	}

	return nil, fmt.Errorf("no user found with api token %s", apiToken)
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
	/*
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
	*/

	filter := bson.D{{}}
	users, err := FilterUsers(filter)

	if err != nil {
		return nil, err
	}

	for _, u := range users {
		if err := bcrypt.CompareHashAndPassword([]byte(u.LoginToken), []byte(loginToken)); err == nil {
			return u, nil
		}
	}

	return nil, fmt.Errorf("no user found with login token %s", loginToken)
}

func GetUserByResetToken(resetToken string) (*User, error) {
	/*
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
	*/

	filter := bson.D{{}}
	users, err := FilterUsers(filter)

	if err != nil {
		return nil, err
	}

	for _, u := range users {
		if err := bcrypt.CompareHashAndPassword([]byte(u.ResetToken), []byte(resetToken)); err == nil {
			return u, nil
		}
	}

	return nil, fmt.Errorf("no user found with reset token %s", resetToken)
}

func GenerateAPIToken(username, name string) (string, error) {
	token := utils.GenerateToken(32)
	currentTime := time.Now().UTC()

	hashedToken, err := HashAndSalt([]byte(token))
	if err != nil {
		return "", err
	}

	apiToken := APIToken{
		Name:    name,
		Token:   hashedToken,
		Created: currentTime.Format("2006-01-02T15:04:05Z"),
	}

	u, err := GetUserByUsername(username)
	if err != nil {
		return "", err
	}

	u.APITokens = append(u.APITokens, apiToken)

	err = UpdateUserByUsername(username, u)
	return token, err
}

func RevokeAPIToken(username, name string) error {
	u, err := GetUserByUsername(username)
	if err != nil {
		return err
	}

	for idx, apiToken := range u.APITokens {
		if apiToken.Name == name {
			u.APITokens = append(u.APITokens[:idx], u.APITokens[idx+1:]...)
			break
		}
	}

	err = UpdateUserByUsername(username, u)
	return err
}

func UpdateUserByUsername(username string, u *User) error {
	filter := bson.M{"username": username}

	currentTime := time.Now().UTC()
	u.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	collection := mongodb.Collections[constants.MONGODB_USER_COLLECTION_NAME]
	ctx := mongodb.Ctx

	opts := options.Replace().SetUpsert(true)

	result, err := collection.ReplaceOne(ctx, filter, u, opts)

	if err != nil {
		return err
	}

	if result.ModifiedCount != 1 {
		return CreateUser(u)
		// return fmt.Errorf("no user found with username %s", username)
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
		logger.Tracef("", "Found user: %s", u.Username)

		users = append(users, &u)
	}

	if err := cur.Err(); err != nil {
		return users, err
	}

	// once exhausted, close the cursor
	cur.Close(ctx)

	return users, nil
}

func VerifyAdmin() error {
	user, _ := GetUserByUsername(config.Config.Admin.Username)

	logger.Debugf("", "No user found for admin")
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
		LoginToken:        "",
		APITokens:         []APIToken{},
		Groups:            []string{"admin"},
		Roles:             []string{"admin"},
	}

	logger.Infof("", "Creating admin user")

	err := CreateUser(u)

	if err != nil {
		logger.Errorf("", "Could not create admin user: %s", err.Error())
	}

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

	if u == nil {
		return false, fmt.Errorf("no user found that matches credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return false, err
	}
	return true, nil
}
