package database

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "github.com/warneki/spaced/server/config"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "golang.org/x/crypto/bcrypt"
    "log"
    "net/http"
    "time"
)


type User struct {
    ID          *primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
    Username    string              `bson:"username" json:"username"`
    DateCreated time.Time           `bson:"date_created" json:"date_created"`
    Hash        []byte              `bson:"hash" json:"-"`
    Sessions    []string            `bson:"sessions" json:"sessions"`
}

type registeringUser struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", config.OriginUrl)
    w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")

    decoder := json.NewDecoder(r.Body)

    var userData registeringUser
    err := decoder.Decode(&userData)
    if err != nil {
        panic(err)
    }

    for _, reserved := range config.ListOfReservedUsernames  {
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
    user.Username = userData.Username

    hash, err := bcrypt.GenerateFromPassword([]byte(userData.Password), config.HashCost)
    if err != nil {
        http.Error(w, "Could not register user, please try again", 500)
        return
    }
    user.Hash = hash

    // TODO: generate signed session
    user.Sessions = []string{"sample_session_id"}

    result, err := insertNewUser(user)
    if err != nil {
        http.Error(w, "Could not register user, please try again", 500)
        return
    }

    json.NewEncoder(w).Encode(result)
}

func insertNewUser(user User) (User, error) {
    res, err := Users.InsertOne(context.Background(), user)
    if err != nil {
        log.Fatal("Got error when inserting new user: ", err)
        return User{}, errors.New("Failed to insert user")
    }

    if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
        user.ID = &oid
    }

    fmt.Printf("Inserted user: %+v \n", user)

    return user, nil
}
