package database

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/warneki/spaced/server/auth"
	"github.com/warneki/spaced/server/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strings"
	"time"
)

type User struct {
	ID          *primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	Username    string              `bson:"username" json:"username"`
	DateCreated time.Time           `bson:"date_created" json:"date_created"`
	Hash        []byte              `bson:"hash" json:"-"`
	Clients     []string            `bson:"clients" json:"-"`
}

type registeringUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type registeringResult struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

const maxTokenLength = 500

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", config.OriginUrl)
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")

	decoder := json.NewDecoder(r.Body)

	var userData registeringUser
	err := decoder.Decode(&userData)
	if err != nil {
		panic(err)
	}

	for _, reserved := range config.ListOfReservedUsernames {
		// TODO: also check here for existing usernames
		if reserved == userData.Password {
			http.Error(w, "This username is not available to register", 409)
			return
		}
	}

	if len(userData.Password) < 8 || len(userData.Password) > 256 {
		http.Error(w, "Your password is too short or too long, minimum length is 8, maximum is 256", 422)
		return
	}

	var user User

	userCreated := time.Now()
	user.DateCreated = userCreated
	// TODO: add checks
	user.Username = strings.ToLower(userData.Username)

	hash, err := bcrypt.GenerateFromPassword([]byte(userData.Password), config.HashCost)
	if err != nil {
		http.Error(w, "Could not register user, please try again", 500)
		return
	}
	user.Hash = hash
	user.Clients = []string{"web"}

	result, err := insertNewUser(user)
	if err != nil {
		if _, ok := err.(mongo.WriteException); ok {
			http.Error(w, "This username is not available to register", 409)
			return
		}
		http.Error(w, "Could not register user, please try again", 500)
		return
	}

	claims := auth.GenerateJWT(user.Username, user.Clients[0])
	token, _ := auth.SignAndSerializeJWT(claims)

	_ = json.NewEncoder(w).Encode(registeringResult{
		User:  result,
		Token: token,
	})
}

func insertNewUser(user User) (User, error) {
	res, err := Users.InsertOne(context.Background(), user)
	if err != nil {
		log.Println("Got error when inserting new user: ", err)
		return User{}, err
	}

	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		user.ID = &oid
	}

	fmt.Printf("Inserted user: %+v \n", user)

	return user, nil
}

func VerifyUserWithToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	splitToken := strings.Split(token, "Bearer ")
	if len(splitToken) == 1 {
		json.NewEncoder(w).Encode(map[string]string{"error": "bad authorisation header"})
		return
	}
	token = splitToken[1]

	if token == "" || len(token) > maxTokenLength {
		json.NewEncoder(w).Encode(map[string]string{"error": "bad_token"})
	}
	user, err := getUserWithToken(token)

	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "could not verify"})
		return
	}
	json.NewEncoder(w).Encode(user)
}

func getUserWithToken(token string) (User, error) {
	claims, err := auth.VerifyJwt(token)
	if err != nil {
		return User{}, err
	}
	var user User
	err = Users.FindOne(context.Background(), bson.M{
		"username": claims.Subject,
	}).Decode(&user)
	if err != nil {
		return User{}, err
	}
	return user, nil
}
