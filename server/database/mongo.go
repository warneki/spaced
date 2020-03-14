package database

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/warneki/spaced/server/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
)

var Projects, Sessions, Repeats, Users *mongo.Collection

func init() {
	clientOptions := options.Client().ApplyURI(config.MONGO_URL())
	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	Projects = client.Database(config.DbName).Collection("projects")
	Sessions = client.Database(config.DbName).Collection("sessions")
	Repeats = client.Database(config.DbName).Collection("repeats")
	Users = client.Database(config.DbName).Collection("users")

}

type dataForToday struct {
	Sessions []Session     `json:"sessions"`
	Projects []primitive.M `json:"projects"`
	Repeats  []Repeat      `json:"repeats"`
	User     User          `json:"user"`
}

func GetDataForToday(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	sessions := make(chan []Session, 1)
	projects := make(chan []primitive.M, 1)
	repeats := make(chan []Repeat, 1)

	user, _, unauthorised := verifyRequest(r, w)
	if unauthorised {
		return
	}

	go getAllSession(sessions, user)
	go getAllProject(projects, user)
	go getAllRepeat(repeats, user)

	payload := dataForToday{}
	payload.User = user

	// TODO: bettwer way to merge results?
	payload.Sessions = <-sessions
	payload.Projects = <-projects
	payload.Repeats = <-repeats

	_ = json.NewEncoder(w).Encode(payload)
}

func queryForResult(err error, cur *mongo.Cursor) []primitive.M {
	if err != nil {
		log.Fatal(err)
	}

	var results []primitive.M
	for cur.Next(context.Background()) {
		var result bson.M
		e := cur.Decode(&result)
		if e != nil {
			log.Fatal(e)
		}
		results = append(results, result)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	cur.Close(context.Background())
	return results
}
