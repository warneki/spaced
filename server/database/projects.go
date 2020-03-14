package database

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Project struct {
	ID            *primitive.ObjectID   `bson:"_id,omitempty" json:"_id,omitempty"`
	Name          string                `bson:"name" json:"name"`
	DateCreated   time.Time             `bson:"date_created" json:"date_created"`
	Tags          []string              `bson:"tags" json:"tags"`
	NotesLocation string                `bson:"notes_location" json:"notes_location"`
	StudySessions []*primitive.ObjectID `bson:"study_sessions" json:"study_sessions"`
	Username      string                `bson:"username" json:"-"`
}

func getAllProject(c chan []primitive.M, user User) {
	cur, err := Projects.Find(context.Background(), bson.M{
		"username": user.Username,
	})
	c <- queryForResult(err, cur)
	close(c)
}

func updateProjectWithSession(session Session) Project {
	// get the project
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	filter := bson.M{
		"name":     session.Project,
		"username": session.Username,
	}
	update := bson.D{{"$push", bson.D{{"study_sessions", session.ID}}}}

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
