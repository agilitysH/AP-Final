package handlers

import (
	"context"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"AP_Final/db"
	"AP_Final/models"
)

type courseItemInput struct {
	ID       string  `json:"id,omitempty"`
	Type     string  `json:"type"`
	Title    string  `json:"title"`
	MaxScore float64 `json:"maxScore"`
	Order    int     `json:"order"`
}

type courseModuleInput struct {
	ID    string            `json:"id,omitempty"`
	Title string            `json:"title"`
	Order int               `json:"order"`
	Items []courseItemInput `json:"items,omitempty"`
}

type courseCreateInput struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Category    string              `json:"category"`
	TeacherID   string              `json:"teacherId"`
	Modules     []courseModuleInput `json:"modules,omitempty"`
}

type coursePatchInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	TeacherID   *string `json:"teacherId"`
}

type moduleCreateInput struct {
	Title string            `json:"title"`
	Order int               `json:"order"`
	Items []courseItemInput `json:"items,omitempty"`
}

type modulePatchInput struct {
	Title *string `json:"title"`
	Order *int    `json:"order"`
}

func wantsHTML(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "text/html")
}

func GetCourses(w http.ResponseWriter, r *http.Request) {
	if wantsHTML(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl, err := template.ParseFiles("views/courses.html")
		if err != nil {
			http.Error(w, "Template parse error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tmpl.Execute(w, map[string]interface{}{"Title": "Courses"}); err != nil {
			http.Error(w, "Template execute error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	page, limit, err := parsePagination(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	filter := bson.M{}

	search := strings.TrimSpace(r.URL.Query().Get("search"))
	if search != "" {
		filter["title"] = bson.M{"$regex": search, "$options": "i"}
	}

	category := strings.TrimSpace(r.URL.Query().Get("category"))
	if category != "" {
		filter["category"] = category
	}

	teacherID := strings.TrimSpace(r.URL.Query().Get("teacherId"))
	if teacherID != "" {
		oid, err := primitive.ObjectIDFromHex(teacherID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid teacherId")
			return
		}
		filter["teacherId"] = oid
	}

	sortSpec, err := parseSort(r.URL.Query().Get("sort"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := db.GetCollection("courses")

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to count courses")
		return
	}

	opts := options.Find().SetSkip(int64((page - 1) * limit)).SetLimit(int64(limit)).SetSort(sortSpec)
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch courses")
		return
	}
	defer cursor.Close(ctx)

	courses := []models.Course{}
	if err := cursor.All(ctx, &courses); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to decode courses")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": courses,
		"page":  page,
		"limit": limit,
		"total": total,
	})
}

func CreateCourse(w http.ResponseWriter, r *http.Request) {
	var input courseCreateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if strings.TrimSpace(input.Title) == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if strings.TrimSpace(input.Category) == "" {
		writeError(w, http.StatusBadRequest, "category is required")
		return
	}
	if strings.TrimSpace(input.TeacherID) == "" {
		writeError(w, http.StatusBadRequest, "teacherId is required")
		return
	}

	teacherOID, err := primitive.ObjectIDFromHex(input.TeacherID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid teacherId")
		return
	}

	now := time.Now()
	course := models.Course{
		ID:          primitive.NewObjectID(),
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		Category:    strings.TrimSpace(input.Category),
		TeacherID:   teacherOID,
		Modules:     mapModulesInput(input.Modules),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = db.GetCollection("courses").InsertOne(ctx, course)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create course")
		return
	}

	writeJSON(w, http.StatusCreated, course)
}

func GetCourse(w http.ResponseWriter, r *http.Request) {
	if wantsHTML(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl, err := template.ParseFiles("views/course.html")
		if err != nil {
			http.Error(w, "Template parse error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tmpl.Execute(w, map[string]interface{}{"Title": "Course"}); err != nil {
			http.Error(w, "Template execute error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	id := r.PathValue("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid course id")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var course models.Course
	err = db.GetCollection("courses").FindOne(ctx, bson.M{"_id": oid}).Decode(&course)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			writeError(w, http.StatusNotFound, "course not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch course")
		return
	}

	writeJSON(w, http.StatusOK, course)
}

func PatchCourse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid course id")
		return
	}

	var input coursePatchInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	setFields := bson.M{}

	if input.Title != nil {
		if strings.TrimSpace(*input.Title) == "" {
			writeError(w, http.StatusBadRequest, "title cannot be empty")
			return
		}
		setFields["title"] = strings.TrimSpace(*input.Title)
	}

	if input.Description != nil {
		setFields["description"] = strings.TrimSpace(*input.Description)
	}

	if input.Category != nil {
		if strings.TrimSpace(*input.Category) == "" {
			writeError(w, http.StatusBadRequest, "category cannot be empty")
			return
		}
		setFields["category"] = strings.TrimSpace(*input.Category)
	}

	if input.TeacherID != nil {
		if strings.TrimSpace(*input.TeacherID) == "" {
			writeError(w, http.StatusBadRequest, "teacherId cannot be empty")
			return
		}
		teacherOID, err := primitive.ObjectIDFromHex(*input.TeacherID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid teacherId")
			return
		}
		setFields["teacherId"] = teacherOID
	}

	if len(setFields) == 0 {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	setFields["updatedAt"] = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := db.GetCollection("courses").UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": setFields})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update course")
		return
	}
	if res.MatchedCount == 0 {
		writeError(w, http.StatusNotFound, "course not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "course updated"})
}

func DeleteCourse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid course id")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := db.GetCollection("courses").DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete course")
		return
	}
	if res.DeletedCount == 0 {
		writeError(w, http.StatusNotFound, "course not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func AddModule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	courseOID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid course id")
		return
	}

	var input moduleCreateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if strings.TrimSpace(input.Title) == "" {
		writeError(w, http.StatusBadRequest, "module title is required")
		return
	}

	module := models.CourseModule{
		ID:    primitive.NewObjectID(),
		Title: strings.TrimSpace(input.Title),
		Order: input.Order,
		Items: mapItemsInput(input.Items),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := db.GetCollection("courses").UpdateOne(
		ctx,
		bson.M{"_id": courseOID},
		bson.M{
			"$push": bson.M{"modules": module},
			"$set":  bson.M{"updatedAt": time.Now()},
		},
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add module")
		return
	}
	if res.MatchedCount == 0 {
		writeError(w, http.StatusNotFound, "course not found")
		return
	}

	writeJSON(w, http.StatusCreated, module)
}

func PatchModule(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("id")
	moduleID := r.PathValue("moduleId")

	courseOID, err := primitive.ObjectIDFromHex(courseID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid course id")
		return
	}

	moduleOID, err := primitive.ObjectIDFromHex(moduleID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid module id")
		return
	}

	var input modulePatchInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	setFields := bson.M{}
	if input.Title != nil {
		if strings.TrimSpace(*input.Title) == "" {
			writeError(w, http.StatusBadRequest, "module title cannot be empty")
			return
		}
		setFields["modules.$[mod].title"] = strings.TrimSpace(*input.Title)
	}
	if input.Order != nil {
		setFields["modules.$[mod].order"] = *input.Order
	}

	if len(setFields) == 0 {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	setFields["updatedAt"] = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{bson.M{"mod._id": moduleOID}},
	})

	res, err := db.GetCollection("courses").UpdateOne(
		ctx,
		bson.M{"_id": courseOID},
		bson.M{"$set": setFields},
		opts,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update module")
		return
	}
	if res.MatchedCount == 0 {
		writeError(w, http.StatusNotFound, "course not found")
		return
	}
	if res.ModifiedCount == 0 {
		writeError(w, http.StatusNotFound, "module not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "module updated"})
}

func DeleteModule(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("id")
	moduleID := r.PathValue("moduleId")

	courseOID, err := primitive.ObjectIDFromHex(courseID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid course id")
		return
	}

	moduleOID, err := primitive.ObjectIDFromHex(moduleID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid module id")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := db.GetCollection("courses").UpdateOne(
		ctx,
		bson.M{"_id": courseOID},
		bson.M{
			"$pull": bson.M{"modules": bson.M{"_id": moduleOID}},
			"$set":  bson.M{"updatedAt": time.Now()},
		},
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete module")
		return
	}
	if res.MatchedCount == 0 {
		writeError(w, http.StatusNotFound, "course not found")
		return
	}
	if res.ModifiedCount == 0 {
		writeError(w, http.StatusNotFound, "module not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func parsePagination(r *http.Request) (int, int, error) {
	page := 1
	limit := 10

	if v := r.URL.Query().Get("page"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil || p < 1 {
			return 0, 0, errorf("invalid page")
		}
		page = p
	}

	if v := r.URL.Query().Get("limit"); v != "" {
		l, err := strconv.Atoi(v)
		if err != nil || l < 1 {
			return 0, 0, errorf("invalid limit")
		}
		if l > 100 {
			l = 100
		}
		limit = l
	}

	return page, limit, nil
}

func parseSort(sortParam string) (bson.D, error) {
	switch strings.TrimSpace(sortParam) {
	case "", "createdAt_desc":
		return bson.D{{Key: "createdAt", Value: -1}}, nil
	case "createdAt_asc":
		return bson.D{{Key: "createdAt", Value: 1}}, nil
	case "title_asc":
		return bson.D{{Key: "title", Value: 1}}, nil
	case "title_desc":
		return bson.D{{Key: "title", Value: -1}}, nil
	default:
		return nil, errorf("invalid sort")
	}
}

func mapModulesInput(inputs []courseModuleInput) []models.CourseModule {
	modules := []models.CourseModule{}
	for _, m := range inputs {
		modID := primitive.NewObjectID()
		if strings.TrimSpace(m.ID) != "" {
			if parsed, err := primitive.ObjectIDFromHex(m.ID); err == nil {
				modID = parsed
			}
		}

		module := models.CourseModule{
			ID:    modID,
			Title: strings.TrimSpace(m.Title),
			Order: m.Order,
			Items: mapItemsInput(m.Items),
		}
		modules = append(modules, module)
	}
	return modules
}

func mapItemsInput(inputs []courseItemInput) []models.CourseItem {
	items := []models.CourseItem{}
	for _, i := range inputs {
		itemID := primitive.NewObjectID()
		if strings.TrimSpace(i.ID) != "" {
			if parsed, err := primitive.ObjectIDFromHex(i.ID); err == nil {
				itemID = parsed
			}
		}

		item := models.CourseItem{
			ID:       itemID,
			Type:     strings.TrimSpace(i.Type),
			Title:    strings.TrimSpace(i.Title),
			MaxScore: i.MaxScore,
			Order:    i.Order,
		}
		items = append(items, item)
	}
	return items
}

func errorf(message string) error {
	return &simpleError{message: message}
}

type simpleError struct {
	message string
}

func (e *simpleError) Error() string {
	return e.message
}

var _ error = (*simpleError)(nil)
