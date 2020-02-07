package database

import (
    "context"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "log"
    "time"
)

var DAYS_OFFSETS = [7]int{0, 1, 9, 25, 55, 131, 241}

type Repeat struct {
    ID        *primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
    Done      bool                `bson:"done" json:"done"`
    Days      int                 `bson:"days" json:"days"`
    SessionID *primitive.ObjectID `bson:"session_id" json:"session_id"`
    RepeatOn  time.Time           `bson:"repeat_on" json:"repeat_on"`
}

func getAllRepeat() []primitive.M {
    cur, err := Repeats.Find(context.Background(), bson.D{{}})
    return queryForResult(err, cur)
}

func generateRepeatsForSession(sessionId *primitive.ObjectID, sessionCreated time.Time) [7]Repeat {
    // Create repeats
    var repeats = [7]Repeat{}

    for i, days := range DAYS_OFFSETS {
        repeats[i] = Repeat{
            Done:      false,
            Days:      days,
            SessionID: sessionId,
            RepeatOn:  sessionCreated.AddDate(0, 0, days),
        }
    }

    // make an interface to insert TODO: investigate
    var interfaceSlice = make([]interface{}, len(repeats))
    for i, r := range repeats {
        interfaceSlice[i] = r
    }

    res, err := Repeats.InsertMany(context.Background(), interfaceSlice)
    if err != nil {
        log.Fatal("Got error: ", err)
    }

    for i, _ := range repeats {
        if oid, ok := res.InsertedIDs[i].(primitive.ObjectID); ok {
            repeats[i].ID = &oid
        }
    }
    return repeats
}
