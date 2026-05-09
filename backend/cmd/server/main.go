// main.go — Backend entry point
// Reads config from environment variables and starts the HTTP server.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Read config from environment variables set by the EC2 user_data script
	mongoURI := mustGetEnv("MONGO_URI")
	redisAddr := mustGetEnv("REDIS_ADDR")
	port := getEnvOrDefault("PORT", "8080")

	// ---- Connect to MongoDB ----
	mongoClient, err := mongo.Connect(
		context.Background(),
		options.Client().ApplyURI(mongoURI),
	)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())
	log.Println("Connected to MongoDB")

	// ---- Connect to Redis ----
	// Redis is used as a cache. If it is temporarily unavailable,
	// keep the API online and serve directly from MongoDB.
	redisClient := connectRedisWithRetry(redisAddr, 6, 5*time.Second)

	// ---- Set up HTTP routes ----
	r := gin.Default()

	// Health check endpoint — the ALB calls this to decide if the instance is healthy
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Friendly root endpoint for manual browser checks on the ALB URL
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "StartTech backend is running"})
	})

	// To-do routes (the actual app logic lives in the handlers package)
	db := mongoClient.Database("muchToDo")
	registerTodoRoutes(r, db, redisClient)

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func connectRedisWithRetry(addr string, attempts int, wait time.Duration) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	for i := 1; i <= attempts; i++ {
		if err := client.Ping(context.Background()).Err(); err == nil {
			log.Println("Connected to Redis")
			return client
		}
		log.Printf("Redis not ready (attempt %d/%d). Retrying in %s...", i, attempts, wait)
		time.Sleep(wait)
	}

	log.Println("Redis unavailable after retries. Continuing without cache.")
	_ = client.Close()
	return nil
}

// mustGetEnv exits the process if the variable is not set.
// This catches misconfiguration at startup rather than silently failing later.
func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return val
}

func getEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
