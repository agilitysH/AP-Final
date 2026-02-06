package routes

import (
	"net/http"
	"strings"

	"AP_Final/handlers"
)

func RegisterRoutes() {
	// 1. АВТОРИЗАЦИЯ И ГЛАВНАЯ
	http.HandleFunc("GET /home", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "views/home.html")
	})
	http.HandleFunc("GET /auth", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "views/auth.html")
	})
	http.HandleFunc("POST /register", handlers.Register)
	http.HandleFunc("POST /login", handlers.Login)

	// 2. КУРСЫ (HTML + API)
	http.HandleFunc("GET /courses", func(w http.ResponseWriter, r *http.Request) {
		// Если запрос НЕ содержит application/json в заголовке Accept, отдаем HTML
		if !strings.Contains(r.Header.Get("Accept"), "application/json") {
			http.ServeFile(w, r, "views/courses.html")
			return
		}
		// Иначе отдаем JSON данные через хендлер
		handlers.GetCourses(w, r)
	})

	http.HandleFunc("GET /courses/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept"), "application/json") {
			http.ServeFile(w, r, "views/course.html")
			return
		}
		handlers.GetCourse(w, r)
	})

	// 3. УПРАВЛЕНИЕ КУРСАМИ (Protected)
	http.HandleFunc("POST /courses", handlers.AuthMiddleware(handlers.CreateCourse))
	http.HandleFunc("PATCH /courses/{id}", handlers.AuthMiddleware(handlers.PatchCourse))
	http.HandleFunc("DELETE /courses/{id}", handlers.AuthMiddleware(handlers.DeleteCourse))

	// 4. МОДУЛИ КУРСА
	http.HandleFunc("POST /courses/{id}/modules", handlers.AuthMiddleware(handlers.AddModule))
	http.HandleFunc("PATCH /courses/{id}/modules/{moduleId}", handlers.AuthMiddleware(handlers.PatchModule))
	http.HandleFunc("DELETE /courses/{id}/modules/{moduleId}", handlers.AuthMiddleware(handlers.DeleteModule))

	// 5. ПРОГРЕСС И СТАТИСТИКА
	http.HandleFunc("PUT /courses/{courseId}/items/{itemId}/progress", handlers.AuthMiddleware(handlers.UpdateProgress))
	http.HandleFunc("GET /me/progress", handlers.AuthMiddleware(handlers.GetMyProgress))

	// 6. ЗАПИСЬ НА КУРС (Enrollments)
	http.HandleFunc("POST /enrollments", handlers.AuthMiddleware(handlers.CreateEnrollment))
	http.HandleFunc("GET /enrollments/my", handlers.AuthMiddleware(handlers.GetMyEnrollments))

	// 7. СТАТИЧЕСКИЕ ФАЙЛЫ (JS, CSS)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
}
