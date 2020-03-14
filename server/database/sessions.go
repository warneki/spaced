package database

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/warneki/spaced/server/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"time"
)

type Session struct {
	ID          *primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Description string              `bson:"description" json:"description"`
	Project     string              `bson:"project" json:"project"`
	Date        time.Time           `bson:"date" json:"date,omitempty"`
	Username    string              `bson:"username" json:"-"`
}

type SessionInsertResult struct {
	Session Session   `json:"session"`
	Project Project   `json:"project"`
	Repeats [7]Repeat `json:"repeats"`
}

func getAllSession(c chan []Session, user User) {
	cur, err := Sessions.Find(context.Background(), bson.M{
		"username": user.Username,
	})
	if err != nil {
		log.Fatal(err)
	}
	var sessions []Session
	for cur.Next(context.Background()) {
		var session Session
		e := cur.Decode(&session)
		if e != nil {
			log.Fatal(e)
		}
		sessions = append(sessions, session)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	_ = cur.Close(context.Background())

	c <- sessions
	close(c)
}

func ReturnOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", config.OriginUrl)
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func AddNewSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", config.OriginUrl)
	w.Header().Set("Server", "A Go Web Server")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")

	user, _, unauthorised := verifyRequest(r, w)
	if unauthorised {
		return
	}

	decoder := json.NewDecoder(r.Body)

	var session Session
	err := decoder.Decode(&session)
	if err != nil {
		panic(err)
	}

	sessionCreated := time.Now()
	session.Date = sessionCreated
	session.Username = user.Username

	result := insertNewSession(session)
	json.NewEncoder(w).Encode(result)
}

func insertNewSession(session Session) SessionInsertResult {
	// Todo: do in transaction
	// Create new session
	res, err := Sessions.InsertOne(context.Background(), session)
	if err != nil {
		log.Fatal("Got error: ", err)
	}

	fmt.Println("Inserted a single document: ", res.InsertedID)

	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		session.ID = &oid
	}

	// Create new repeats and update the project
	repeats := generateRepeatsForSession(session)
	project := updateProjectWithSession(session)

	return SessionInsertResult{
		Session: session,
		Project: project,
		Repeats: repeats,
	}
}
