package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"AP_Final/db"
	"AP_Final/routes"
)

func main() {
	_ = godotenv.Load()

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017" // дефолт, если .env пуст
	}

	db.Connect(mongoURI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.EnsureIndexes(ctx); err != nil {
		log.Fatal(err)
	}

	// Инициализируем маршруты
	routes.RegisterRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Сервер запущен на http://localhost:%s/home", port)
	// nil означает использование стандартного ServeMux, в который мы пишем через http.HandleFunc
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
