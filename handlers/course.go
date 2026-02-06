package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"AP_Final/db"
	"AP_Final/models"
)

// Описания структур для входящих данных (теперь ошибки Unresolved type исчезнут)
type courseModuleInput struct {
	ID    string `json:"id,omitempty"`
	Title string `json:"title"`
	Order int    `json:"order"`
}

type courseCreateInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	TeacherID   string `json:"teacherId"`
}

// GetCourses получает список курсов с фильтрацией и пагинацией.
func GetCourses(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	category := r.URL.Query().Get("category")
	teacherId := r.URL.Query().Get("teacherId")
	sortParam := r.URL.Query().Get("sort")

	filter := bson.M{}
	if search != "" {
		filter["title"] = bson.M{"$regex": search, "$options": "i"}
	}
	if category != "" {
		filter["category"] = category
	}
	if teacherId != "" {
		if tOID, err := primitive.ObjectIDFromHex(teacherId); err == nil {
			filter["teacherId"] = tOID
		}
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 9
	skip := (page - 1) * limit

	findOptions := options.Find().SetLimit(int64(limit)).SetSkip(int64(skip))
	if sortParam != "" {
		if sortDoc, err := parseSort(sortParam); err == nil {
			findOptions.SetSort(sortDoc)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := db.GetCollection("courses").Find(ctx, filter, findOptions)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка БД")
		return
	}
	defer cursor.Close(ctx)

	var courses []models.Course
	if err := cursor.All(ctx, &courses); err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка декодирования")
		return
	}

	total, _ := db.GetCollection("courses").CountDocuments(ctx, filter)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": courses,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// AddModule добавляет новый модуль в массив курса.
func AddModule(w http.ResponseWriter, r *http.Request) {
	courseID, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid course ID")
		return
	}

	var input courseModuleInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	newModule := models.CourseModule{
		ID:    primitive.NewObjectID(),
		Title: input.Title,
		Order: input.Order,
		Items: []models.CourseItem{},
	}

	filter := bson.M{"_id": courseID}
	update := bson.M{"$push": bson.M{"modules": newModule}}

	_, err = db.GetCollection("courses").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Update failed")
		return
	}
	writeJSON(w, http.StatusCreated, newModule)
}

// PatchModule обновляет заголовок конкретного модуля используя arrayFilters.
func PatchModule(w http.ResponseWriter, r *http.Request) {
	courseID, _ := primitive.ObjectIDFromHex(r.PathValue("id"))
	moduleID, _ := primitive.ObjectIDFromHex(r.PathValue("moduleId"))

	var input courseModuleInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	filter := bson.M{"_id": courseID}
	update := bson.M{"$set": bson.M{"modules.$[mod].title": input.Title}}

	opts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{bson.M{"mod._id": moduleID}},
	})

	_, err := db.GetCollection("courses").UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Patch failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteModule удаляет модуль из массива курса.
func DeleteModule(w http.ResponseWriter, r *http.Request) {
	courseID, _ := primitive.ObjectIDFromHex(r.PathValue("id"))
	moduleID, _ := primitive.ObjectIDFromHex(r.PathValue("moduleId"))

	filter := bson.M{"_id": courseID}
	update := bson.M{"$pull": bson.M{"modules": bson.M{"_id": moduleID}}}

	_, err := db.GetCollection("courses").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Delete failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// CreateCourse создает новый курс.
func CreateCourse(w http.ResponseWriter, r *http.Request) {
	var in courseCreateInput
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	tID, _ := primitive.ObjectIDFromHex(in.TeacherID)
	c := models.Course{
		ID:          primitive.NewObjectID(),
		Title:       in.Title,
		Description: in.Description,
		Category:    in.Category,
		TeacherID:   tID,
		CreatedAt:   time.Now(),
	}
	if _, err := db.GetCollection("courses").InsertOne(context.TODO(), c); err != nil {
		writeError(w, http.StatusInternalServerError, "Insert failed")
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func parseSort(s string) (bson.D, error) {
	switch s {
	case "createdAt_desc":
		return bson.D{{Key: "createdAt", Value: -1}}, nil
	case "createdAt_asc":
		return bson.D{{Key: "createdAt", Value: 1}}, nil
	case "title_asc":
		return bson.D{{Key: "title", Value: 1}}, nil
	case "title_desc":
		return bson.D{{Key: "title", Value: -1}}, nil
	default:
		return bson.D{{Key: "createdAt", Value: -1}}, nil
	}
}

func GetCourse(w http.ResponseWriter, r *http.Request) {
	id, _ := primitive.ObjectIDFromHex(r.PathValue("id"))
	var c models.Course
	err := db.GetCollection("courses").FindOne(context.TODO(), bson.M{"_id": id}).Decode(&c)
	if err != nil {
		writeError(w, 404, "Not found")
		return
	}
	writeJSON(w, 200, c)
}

func PatchCourse(w http.ResponseWriter, _ *http.Request)  { w.WriteHeader(204) }
func DeleteCourse(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) }
