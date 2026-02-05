package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CourseItem struct {
	ID       primitive.ObjectID `bson:"_id" json:"id"`
	Type     string             `bson:"type" json:"type"`
	Title    string             `bson:"title" json:"title"`
	MaxScore float64            `bson:"maxScore" json:"maxScore"`
	Order    int                `bson:"order" json:"order"`
}

type CourseModule struct {
	ID    primitive.ObjectID `bson:"_id" json:"id"`
	Title string             `bson:"title" json:"title"`
	Order int                `bson:"order" json:"order"`
	Items []CourseItem       `bson:"items,omitempty" json:"items,omitempty"`
}

type Course struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	Category    string             `bson:"category" json:"category"`
	TeacherID   primitive.ObjectID `bson:"teacherId" json:"teacherId"`
	Modules     []CourseModule     `bson:"modules,omitempty" json:"modules,omitempty"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}
