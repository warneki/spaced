package database

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

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


