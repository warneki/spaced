package database

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"time"
)

var DaysOffsets = [7]int{0, 1, 9, 25, 55, 131, 241}

type Repeat struct {
	ID        *primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Done      bool                `bson:"done" json:"done"`
	Days      int                 `bson:"days" json:"days"`
	SessionID *primitive.ObjectID `bson:"session_id" json:"session_id"`
	RepeatOn  time.Time           `bson:"repeat_on" json:"repeat_on"`
	Username  string              `bson:"username" json:"-"`
}

func getAllRepeat(c chan []Repeat, user User) {
	cur, err := Repeats.Find(context.Background(), bson.M{
		"username": user.Username,
	})
	if err != nil {
		log.Fatal(err)
	}
	var repeats []Repeat
	for cur.Next(context.Background()) {
		var repeat Repeat
		e := cur.Decode(&repeat)
		if e != nil {
			log.Fatal(e)
		}
		repeats = append(repeats, repeat)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	_ = cur.Close(context.Background())

	c <- repeats
	close(c)
}

func generateRepeatsForSession(session Session) [7]Repeat {
	// Create repeats
	var repeats = [7]Repeat{}

	for i, days := range DaysOffsets {
		repeats[i] = Repeat{
			Done:      false,
			Days:      days,
			SessionID: session.ID,
			RepeatOn:  session.Date.AddDate(0, 0, days),
			Username:  session.Username,
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
