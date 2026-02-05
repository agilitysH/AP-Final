package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Enrollment struct {
	ID           primitive.ObjectID `bson:"_id" json:"id"`
	UserID       primitive.ObjectID `bson:"userId" json:"userId"`
	CourseID     primitive.ObjectID `bson:"courseId" json:"courseId"`
	Status       string             `bson:"status" json:"status"`
	EnrolledAt   time.Time          `bson:"enrolledAt" json:"enrolledAt"`
	LastAccessAt *time.Time         `bson:"lastAccessAt,omitempty" json:"lastAccessAt,omitempty"`
}
