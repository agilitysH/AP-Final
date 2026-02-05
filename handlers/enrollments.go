package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"AP_Final/db"
	"AP_Final/models"
)

type enrollmentCreateInput struct {
	CourseID string `json:"courseId"`
}

func CreateEnrollment(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input enrollmentCreateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if strings.TrimSpace(input.CourseID) == "" {
		writeError(w, http.StatusBadRequest, "courseId is required")
		return
	}

	courseOID, err := primitive.ObjectIDFromHex(input.CourseID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid courseId")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ensure course exists
	var course models.Course
	if err := db.GetCollection("courses").FindOne(ctx, bson.M{"_id": courseOID}).Decode(&course); err != nil {
		if err == mongo.ErrNoDocuments {
			writeError(w, http.StatusNotFound, "course not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to check course")
		return
	}

	now := time.Now()
	doc := models.Enrollment{
		ID:         primitive.NewObjectID(),
		UserID:     userID,
		CourseID:   courseOID,
		Status:     "active",
		EnrolledAt: now,
	}

	_, err = db.GetCollection("enrollments").InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			writeError(w, http.StatusConflict, "enrollment already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create enrollment")
		return
	}

	writeJSON(w, http.StatusCreated, doc)
}

func GetMyEnrollments(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := db.GetCollection("enrollments").Find(ctx, bson.M{"userId": userID})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch enrollments")
		return
	}
	defer cursor.Close(ctx)

	results := []models.Enrollment{}
	if err := cursor.All(ctx, &results); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to decode enrollments")
		return
	}

	writeJSON(w, http.StatusOK, results)
}

func DeleteEnrollment(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := r.PathValue("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid enrollment id")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := db.GetCollection("enrollments").DeleteOne(ctx, bson.M{"_id": oid, "userId": userID})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete enrollment")
		return
	}
	if res.DeletedCount == 0 {
		writeError(w, http.StatusNotFound, "enrollment not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "enrollment deleted"})
}

func DeleteEnrollmentsByCourse(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	courseID := strings.TrimSpace(r.URL.Query().Get("courseId"))
	if courseID == "" {
		writeError(w, http.StatusBadRequest, "courseId is required")
		return
	}

	courseOID, err := primitive.ObjectIDFromHex(courseID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid courseId")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var course models.Course
	err = db.GetCollection("courses").FindOne(ctx, bson.M{"_id": courseOID}).Decode(&course)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			writeError(w, http.StatusNotFound, "course not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to check course")
		return
	}

	if course.TeacherID != userID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	res, err := db.GetCollection("enrollments").DeleteMany(ctx, bson.M{"courseId": courseOID})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete enrollments")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"deletedCount": res.DeletedCount})
}
