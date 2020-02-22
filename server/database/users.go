package database

import (
    "go.mongodb.org/mongo-driver/bson/primitive"
    "time"
)


type User struct {
    ID          *primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
    Username    string              `bson:"username" json:"username"`
    DateCreated time.Time           `bson:"date_created" json:"date_created"`
    Hash        []byte              `bson:"hash" json:"-"`
    Sessions    []string            `bson:"sessions" json:"sessions"`
}

