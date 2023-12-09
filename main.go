package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"redisTest/bredis"
	"time"
)

var (
	rdb *redis.Client
	r   *bredis.Redis
)

func main() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis server address
		Password: "",               // No password
		DB:       0,                // Default DB
	})
	ctx := context.Background()
	r = bredis.NewRedis()
	// Enable keyspace notifications
	if err := r.EnableKeyspaceNotifications(ctx, rdb); err != nil {
		fmt.Println("Error setting config:", err)
		return
	}

	router := gin.Default()
	router.POST("/subscribe/:key", handleSubscription)
	router.Run(":8080")

}

func handleSubscription(c *gin.Context) {
	key := c.Param("key")
	if len(key) == 0 {
		c.JSON(400, gin.H{
			"message": "Key is required",
		})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	ch := make(chan interface{})
	go r.CPSubscribe(ctx, rdb, ch, fmt.Sprintf("__keyspace@*__:%s", key))
	select {
	case <-ctx.Done():
		fmt.Println("Context done")
		return
	case msg := <-ch:
		fmt.Println("Message received:", msg)
		if err := r.DeleteKey(ctx, rdb, key); err != nil {
			fmt.Println("Error deleting key:", err)
			return
		}
		c.JSON(200, gin.H{
			"message": msg,
		})
		return
	}
}
