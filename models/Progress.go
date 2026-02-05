package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Progress struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	CourseID  primitive.ObjectID `bson:"courseId" json:"courseId"`
	ItemID    primitive.ObjectID `bson:"itemId" json:"itemId"`
	Status    string             `bson:"status" json:"status"`
	Score     float64            `bson:"score" json:"score"`
	Attempts  int                `bson:"attempts" json:"attempts"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}
