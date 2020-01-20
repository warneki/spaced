package database

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

func GetAllSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	payload := getAllSession()
	json.NewEncoder(w).Encode(payload)
}

func getAllSession() []primitive.M {
	cur, err := Sessions.Find(context.Background(), bson.D{{}})
	return queryForResult(err, cur)
}
