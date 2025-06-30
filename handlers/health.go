package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/nathansanchezdev/blog-engine/pkg/insight"
)

var (
	DB             *sql.DB
	InsightClient  *insight.Client
	LoggingEnabled bool
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Get database counts
	authorCount, postCount, err := getDatabaseCounts()
	if err != nil {
		SendErrorLog("Database query failed", err, "/health")
		http.Error(w, "Database error", 500)
		return
	}

	// Send observability data
	duration := time.Since(start).Milliseconds()
	metadata := map[string]interface{}{
		"authors": authorCount,
		"posts":   postCount,
	}
	SendSuccessObservability("/health", "GET", 200, duration, "Health check completed", metadata, r)

	// Return response
	response := map[string]interface{}{
		"status":  "ok",
		"authors": authorCount,
		"posts":   postCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getDatabaseCounts() (int, int, error) {
	var authorCount, postCount int

	err := DB.QueryRow("SELECT COUNT(*) FROM authors").Scan(&authorCount)
	if err != nil {
		return 0, 0, err
	}

	err = DB.QueryRow("SELECT COUNT(*) FROM posts").Scan(&postCount)
	if err != nil {
		return 0, 0, err
	}

	return authorCount, postCount, nil
}

func SendSuccessObservability(path, method string, statusCode int, duration int64, message string, metadata map[string]any, r *http.Request) {
	if !LoggingEnabled || InsightClient == nil {
		return
	}

	go func() {
		// Add request info to metadata
		if metadata == nil {
			metadata = make(map[string]interface{})
		}
		metadata["user_ip"] = r.RemoteAddr
		metadata["user_agent"] = r.UserAgent()

		// Send log
		InsightClient.SendLog("blog-service", "INFO", message, metadata)

		// Send metric
		InsightClient.SendMetric("blog-service", path, method, statusCode, float64(duration))
	}()
}

// SendErrorLog sends error log for failed requests
func SendErrorLog(message string, err error, endpoint string) {
	if !LoggingEnabled || InsightClient == nil {
		return
	}

	go func() {
		InsightClient.SendLog("blog-service", "ERROR", message, map[string]interface{}{
			"error":    err.Error(),
			"endpoint": endpoint,
		})
	}()
}
