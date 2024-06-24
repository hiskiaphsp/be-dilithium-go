package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Document struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Filename string             `json:"filename" bson:"filename"`
	Path     string             `json:"path" bson:"path"`
}
