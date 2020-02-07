package database

import (
    "context"
    "encoding/json"
    "fmt"
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
}

type SessionInsertResult struct {
    Session Session   `json:"session"`
    Project Project   `json:"project"`
    Repeats [7]Repeat `json:"repeats"`
}

func getAllSession() []primitive.M {
    cur, err := Sessions.Find(context.Background(), bson.D{{}})
    return queryForResult(err, cur)
}

func ReturnOptions(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
    w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func AddNewSession(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Server", "A Go Web Server")
    w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")

    decoder := json.NewDecoder(r.Body)

    var session Session
    err := decoder.Decode(&session)
    if err != nil {
        panic(err)
    }

    sessionCreated := time.Now()
    session.Date = sessionCreated

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
    repeats := generateRepeatsForSession(session.ID, session.Date)
    project := updateProjectWithSession(session.ID, session.Project)

    return SessionInsertResult{
        Session: session,
        Project: project,
        Repeats: repeats,
    }
}
