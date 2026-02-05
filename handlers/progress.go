package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"AP_Final/db"
)

type progressInput struct {
	Status string  `json:"status"`
	Score  float64 `json:"score"`
}

func UpdateProgress(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	courseID := r.PathValue("courseId")
	itemID := r.PathValue("itemId")

	courseOID, err := primitive.ObjectIDFromHex(courseID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid course id")
		return
	}

	itemOID, err := primitive.ObjectIDFromHex(itemID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	var input progressInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	status := strings.TrimSpace(input.Status)
	if status == "" {
		writeError(w, http.StatusBadRequest, "status is required")
		return
	}

	if status != "not_started" && status != "in_progress" && status != "done" {
		writeError(w, http.StatusBadRequest, "invalid status")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"userId":    userID,
			"courseId":  courseOID,
			"itemId":    itemOID,
			"status":    status,
			"score":     input.Score,
			"updatedAt": time.Now(),
		},
		"$inc": bson.M{"attempts": 1},
	}

	opts := options.Update().SetUpsert(true)

	_, err = db.GetCollection("progress").UpdateOne(
		ctx,
		bson.M{"userId": userID, "courseId": courseOID, "itemId": itemOID},
		update,
		opts,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update progress")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "progress updated"})
}

func GetMyProgress(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := mongoProgressPipeline(userID)

	cursor, err := db.GetCollection("enrollments").Aggregate(ctx, pipeline)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to aggregate progress")
		return
	}
	defer cursor.Close(ctx)

	results := []bson.M{}
	if err := cursor.All(ctx, &results); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to decode progress")
		return
	}

	writeJSON(w, http.StatusOK, results)
}

func mongoProgressPipeline(userID primitive.ObjectID) []bson.M {
	return []bson.M{
		{"$match": bson.M{"userId": userID}},
		{"$lookup": bson.M{
			"from":         "courses",
			"localField":   "courseId",
			"foreignField": "_id",
			"as":           "course",
		}},
		{"$unwind": "$course"},
		{"$lookup": bson.M{
			"from": "progress",
			"let":  bson.M{"courseId": "$courseId", "userId": "$userId"},
			"pipeline": []bson.M{
				{"$match": bson.M{"$expr": bson.M{"$and": []bson.M{
					{"$eq": []interface{}{"$courseId", "$$courseId"}},
					{"$eq": []interface{}{"$userId", "$$userId"}},
				}}}},
			},
			"as": "progress",
		}},
		{"$addFields": bson.M{
			"itemsCount": bson.M{"$sum": bson.M{"$map": bson.M{
				"input": "$course.modules",
				"as":    "m",
				"in":    bson.M{"$size": bson.M{"$ifNull": []interface{}{"$$m.items", []interface{}{}}}},
			}}},
			"doneCount": bson.M{"$size": bson.M{"$filter": bson.M{
				"input": "$progress",
				"as":    "p",
				"cond":  bson.M{"$eq": []interface{}{"$$p.status", "done"}},
			}}},
			"avgScore": bson.M{"$ifNull": []interface{}{bson.M{"$avg": "$progress.score"}, 0}},
		}},
		{"$addFields": bson.M{
			"completionRate": bson.M{"$cond": []interface{}{
				bson.M{"$gt": []interface{}{"$itemsCount", 0}},
				bson.M{"$divide": []interface{}{"$doneCount", "$itemsCount"}},
				0,
			}},
		}},
		{"$project": bson.M{
			"_id":              0,
			"courseId":         "$course._id",
			"courseTitle":      "$course.title",
			"itemsCount":       1,
			"doneCount":        1,
			"completionRate":   1,
			"avgScore":         1,
			"enrollmentStatus": "$status",
			"enrolledAt":       "$enrolledAt",
		}},
		{"$sort": bson.M{"completionRate": -1, "courseTitle": 1}},
	}
}
