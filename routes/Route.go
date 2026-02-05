package routes

import (
	"net/http"

	"AP_Final/handlers"
)

func RegisterRoutes() {
	// Public routes
	http.HandleFunc("POST /register", handlers.Register)
	http.HandleFunc("POST /login", handlers.Login)
	http.HandleFunc("GET /home", handlers.Home)

	// Courses (HTML + API via content negotiation)
	http.HandleFunc("GET /courses", handlers.GetCourses)
	http.HandleFunc("GET /courses/{id}", handlers.GetCourse)

	// Protected course mutations
	http.HandleFunc("POST /courses", handlers.AuthMiddleware(handlers.CreateCourse))
	http.HandleFunc("PATCH /courses/{id}", handlers.AuthMiddleware(handlers.PatchCourse))
	http.HandleFunc("DELETE /courses/{id}", handlers.AuthMiddleware(handlers.DeleteCourse))

	// Modules
	http.HandleFunc("POST /courses/{id}/modules", handlers.AuthMiddleware(handlers.AddModule))
	http.HandleFunc("PATCH /courses/{id}/modules/{moduleId}", handlers.AuthMiddleware(handlers.PatchModule))
	http.HandleFunc("DELETE /courses/{id}/modules/{moduleId}", handlers.AuthMiddleware(handlers.DeleteModule))

	// Progress
	http.HandleFunc("PUT /courses/{courseId}/items/{itemId}/progress", handlers.AuthMiddleware(handlers.UpdateProgress))
	http.HandleFunc("GET /me/progress", handlers.AuthMiddleware(handlers.GetMyProgress))

	// Enrollments
	http.HandleFunc("POST /enrollments", handlers.AuthMiddleware(handlers.CreateEnrollment))
	http.HandleFunc("GET /enrollments/my", handlers.AuthMiddleware(handlers.GetMyEnrollments))
	http.HandleFunc("DELETE /enrollments", handlers.AuthMiddleware(handlers.DeleteEnrollmentsByCourse))
	http.HandleFunc("DELETE /enrollments/{id}", handlers.AuthMiddleware(handlers.DeleteEnrollment))

	// Static
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
}
