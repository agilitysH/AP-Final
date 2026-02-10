package handlers

import (
	"context"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"AP_Final/db"
)

// Структура для получения данных из JSON (Backend B)
type progressUpdateInput struct {
	Status string  `json:"status"`
	Score  float64 `json:"score"`
}

// UpdateProgress — ТВОЯ ЧАСТЬ: обновляет прогресс с использованием Upsert и $inc.
func UpdateProgress(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// r.PathValue берет данные из роута: /courses/{courseId}/items/{itemId}/progress
	courseOID, err := primitive.ObjectIDFromHex(r.PathValue("courseId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid course id")
		return
	}
	itemOID, err := primitive.ObjectIDFromHex(r.PathValue("itemId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	var input progressUpdateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{"$match": bson.M{"_id": courseOID}},
		{"$unwind": "$modules"},
		{"$unwind": "$modules.items"},
		{"$match": bson.M{"modules.items._id": itemOID}},
		{"$project": bson.M{"maxScore": "$modules.items.maxScore"}},
	}

	cursor, err := db.GetCollection("courses").Aggregate(ctx, pipeline)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load course item")
		return
	}
	defer cursor.Close(ctx)

	type itemScore struct {
		MaxScore float64 `bson:"maxScore"`
	}
	if !cursor.Next(ctx) {
		writeError(w, http.StatusNotFound, "course item not found")
		return
	}

	var item itemScore
	if err := cursor.Decode(&item); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to decode course item")
		return
	}

	if input.Score > item.MaxScore {
		writeError(w, http.StatusBadRequest, "score exceeds maxScore")
		return
	}

	filter := bson.M{
		"userId":   userID,
		"courseId": courseOID,
		"itemId":   itemOID,
	}

	// Операция обновления: $set меняет данные, $inc увеличивает счетчик попыток
	update := bson.M{
		"$set": bson.M{
			"status":    input.Status,
			"score":     input.Score,
			"updatedAt": time.Now(),
		},
		"$inc": bson.M{"attempts": 1},
	}

	// Upsert: true создаст запись, если её нет в базе
	opts := options.Update().SetUpsert(true)

	if _, err := db.GetCollection("progress").UpdateOne(ctx, filter, update, opts); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update progress")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// GetMyProgress — ЧАСТЬ ДРУГА: сложная агрегация для личного кабинета.
func GetMyProgress(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Pipeline агрегации (оставляем без изменений, как в оригинале)
	pipeline := []bson.M{
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
				bson.M{"$multiply": []interface{}{bson.M{"$divide": []interface{}{"$doneCount", "$itemsCount"}}, 100}},
				0,
			}},
		}},
		{"$project": bson.M{
			"courseId":       1,
			"courseTitle":    "$course.title",
			"itemsCount":     1,
			"doneCount":      1,
			"avgScore":       1,
			"completionRate": 1,
		}},
	}

	cursor, err := db.GetCollection("enrollments").Aggregate(ctx, pipeline)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "aggregation failed")
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		writeError(w, http.StatusInternalServerError, "decoding failed")
		return
	}

	writeJSON(w, http.StatusOK, results)
}
