package database

import (
    "context"
    "encoding/json"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "log"
    "net/http"
    "time"
)

type Project struct {
    ID            *primitive.ObjectID   `bson:"_id,omitempty" json:"_id,omitempty"`
    Name          string                `bson:"name" json:"name"`
    DateCreated   time.Time             `bson:"date_created" json:"date_created"`
    Tags          []string              `bson:"tags" json:"tags"`
    NotesLocation string                `bson:"notes_location" json:"notes_location"`
    StudySessions []*primitive.ObjectID `bson:"study_sessions" json:"study_sessions"`
}

func GetAllProject(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    payload := getAllProject()
    json.NewEncoder(w).Encode(payload)
}

func getAllProject() []primitive.M {
    cur, err := Projects.Find(context.Background(), bson.D{{}})
    return queryForResult(err, cur)
}

func updateProjectWithSession(sessionId *primitive.ObjectID, projectName string) Project {
    // get the project
    opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

    filter := bson.M{"name": projectName}
    update := bson.D{{"$push", bson.D{{"study_sessions", sessionId}}}}

    var project = Project{}

    err := Projects.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&project)
    if err != nil {
        // ErrNoDocuments means that the filter did not match any documents in the collection
        if err == mongo.ErrNoDocuments {
            // TODO: handle
            panic(err)
        }
    }

    return project
}

func decoder(err error, cur *mongo.Cursor) []primitive.M {

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
