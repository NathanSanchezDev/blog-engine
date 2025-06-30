package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/nathansanchezdev/blog-engine/handlers"
	"github.com/nathansanchezdev/blog-engine/pkg/insight"
)

func main() {
	loadConfig()
	initServices()
	setupRoutes()
	startServer()
}

func loadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Check if observability logging is enabled
	handlers.LoggingEnabled = os.Getenv("ENABLE_OBSERVABILITY") == "true"

	fmt.Println("🚀 Blog Engine starting...")
	if handlers.LoggingEnabled {
		fmt.Println("📊 Observability enabled")
	}
}

func initServices() {
	initDatabase()
	initInsight()
}

func initDatabase() {
	dbURL := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	var err error
	handlers.DB, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	if err := handlers.DB.Ping(); err != nil {
		log.Fatal("Database ping failed:", err)
	}

	fmt.Println("✅ Database connected")
}

func initInsight() {
	if !handlers.LoggingEnabled {
		return
	}

	insightURL := os.Getenv("INSIGHT_URL")
	insightKey := os.Getenv("INSIGHT_API_KEY")

	if insightURL == "" || insightKey == "" {
		log.Println("⚠️  Observability config missing, skipping Go-Insight")
		handlers.LoggingEnabled = false
		return
	}

	handlers.InsightClient = insight.NewClient(insightURL, insightKey)

	if err := handlers.InsightClient.Health(); err != nil {
		log.Printf("⚠️  Go-Insight not available: %v", err)
		handlers.LoggingEnabled = false
	} else {
		fmt.Println("✅ Go-Insight connected")
	}
}

func setupRoutes() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/health", handlers.HealthHandler)
}

func startServer() {
	port := getEnvOrDefault("PORT", "8080")
	fmt.Printf("Server running on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Blog Engine v0.1.0"))
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
