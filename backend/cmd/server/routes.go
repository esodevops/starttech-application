// routes.go — Registers all /api/todos HTTP routes
package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Todo struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"  json:"id"`
	Title     string             `bson:"title"          json:"title"`
	Completed bool               `bson:"completed"      json:"completed"`
	CreatedAt time.Time          `bson:"created_at"     json:"created_at"`
}

func registerTodoRoutes(r *gin.Engine, db *mongo.Database, rdb *redis.Client) {
	col := db.Collection("todos")

	r.GET("/api/todos", func(c *gin.Context) {
		ctx := context.Background()

		// Try to serve from Redis cache first
		cached, err := rdb.Get(ctx, "todos:all").Bytes()
		if err == nil {
			c.Data(http.StatusOK, "application/json", cached)
			return
		}

		// Cache miss — fetch from MongoDB
		cursor, err := col.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var todos []Todo
		if err := cursor.All(ctx, &todos); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, todos)
	})

	r.POST("/api/todos", func(c *gin.Context) {
		var input struct{ Title string `json:"title"` }
		if err := c.ShouldBindJSON(&input); err != nil || input.Title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
			return
		}
		todo := Todo{
			ID:        primitive.NewObjectID(),
			Title:     input.Title,
			Completed: false,
			CreatedAt: time.Now(),
		}
		if _, err := col.InsertOne(context.Background(), todo); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Bust the cache so the next GET returns fresh data
		rdb.Del(context.Background(), "todos:all")
		c.JSON(http.StatusCreated, todo)
	})
}
// trigger
