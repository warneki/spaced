package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pascaldekloe/jwt"
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

	user.DateCreated = time.Now()
	// TODO: add checks
	// do not allow whitespaces and special chars except dashes and _
	user.Username = strings.ToLower(userData.Username)

	hash, err := bcrypt.GenerateFromPassword([]byte(userData.Password), config.HashCost)
	if err != nil {
		http.Error(w, "Could not register user, please try again", 500)
		return
	}
	user.Hash = hash

	// client is web TODO: add android
	claims, clientName := auth.GenerateJWT(user.Username, "web")
	token, _ := auth.SignAndSerializeJWT(claims)

	user.Clients = []string{clientName}

	result, err := insertNewUser(user)
	if err != nil {
		if _, ok := err.(mongo.WriteException); ok {
			http.Error(w, "This username is not available to register", 409)
			return
		}
		http.Error(w, "Could not register user, please try again", 500)
		return
	}

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

	user, _, err := verifyRequest(r, w)
	if err {
		return
	}
	json.NewEncoder(w).Encode(user)
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	// todo: add login by email
	w.Header().Set("Content-Type", "application/json")

	// get login data from request
	var userData registeringUser
	err := json.NewDecoder(r.Body).Decode(&userData)
	if err != nil {
		panic(err)
	}
	var user User
	user.Username = strings.ToLower(userData.Username)

	// check that queried user exists
	err = Users.FindOne(context.Background(), bson.M{
		"username": user.Username,
	}).Decode(&user)

	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "username or password are incorrect"})
		return
	}

	// check password
	err = bcrypt.CompareHashAndPassword(user.Hash, []byte(userData.Password))
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "username or password are incorrect"})
		return
	}

	// generate token
	claims, clientName := auth.GenerateJWT(user.Username, "web")
	token, _ := auth.SignAndSerializeJWT(claims)

	// update client name in db
	user.Clients = append(user.Clients, clientName)
	_, err = Users.UpdateOne(
		context.Background(),
		bson.M{
			"username": user.Username,
		},
		bson.M{"$set": bson.M{"clients": user.Clients}},
	)

	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "could not login, please try again"})
		return
	}



	_ = json.NewEncoder(w).Encode(registeringResult{
		User:  user,
		Token: token,
	})
}

func verifyRequest(r *http.Request, w http.ResponseWriter) (User, jwt.Claims, bool) {
	token := r.Header.Get("Authorization")
	splitToken := strings.Split(token, "Bearer ")
	if len(splitToken) == 1 {
		json.NewEncoder(w).Encode(map[string]string{"error": "bad authorisation header"})
		return User{}, jwt.Claims{}, true
	}
	token = splitToken[1]

	if token == "" || len(token) > maxTokenLength {
		json.NewEncoder(w).Encode(map[string]string{"error": "bad_token"})
	}
	user, claims, err := getUserWithToken(token)

	if err != nil {
		fmt.Println(err.Error())
		json.NewEncoder(w).Encode(map[string]string{"error": "could not verify"})
		return User{}, jwt.Claims{}, true
	}
	return user, claims, false
}

func getUserWithToken(token string) (User, jwt.Claims, error) {
	claims, err := auth.VerifyJwt(token)
	if err != nil {
		return User{}, claims, err
	}
	var user User
	err = Users.FindOne(context.Background(), bson.M{
		"username": claims.Subject,
	}).Decode(&user)

	if err == nil {
		for _, val := range user.Clients {
			if val == claims.Set["client"] {
				return user, claims, nil
			}
		}
		return User{}, claims, errors.New("claimed client not found")
	}
	return User{}, claims, err
}

func LogoutUserWithToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, claims, unauthorised := verifyRequest(r, w)
	if unauthorised {
		return
	}

	var clients []string
	for i, val := range user.Clients {
		if val == claims.Set["client"] {
			clients = append(user.Clients[:i], user.Clients[i+1:]...)
			break
		}
	}

	_, err := Users.UpdateOne(
		context.Background(),
		bson.M{
			"username": user.Username,
		},
		bson.M{"$set": bson.M{"clients": clients}},
	)

	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "could not logout, please try again"})
		return
	}

	json.NewEncoder(w).Encode(user)
}