package bredis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strings"
	"time"
)

type Redis struct{}

func NewRedis() *Redis {
	return &Redis{}
}

func (r *Redis) EnableKeyspaceNotifications(ctx context.Context, rdb *redis.Client) error {
	return rdb.ConfigSet(ctx, "notify-keyspace-events", "AKE").Err()
}

func (r *Redis) setKeyIfNotExists(rdb *redis.Client, ctx context.Context, key, value string) error {
	// Set the key with an expiration of 24 hours (86400 seconds), only if it does not exist
	statusCmd := rdb.SetNX(ctx, key, value, 24*time.Hour)
	if err := statusCmd.Err(); err != nil {
		return err
	}

	if statusCmd.Val() {
		fmt.Println("Key set successfully")
	} else {
		fmt.Println("Key already exists")
	}

	return nil
}

func (r *Redis) CPSubscribe(ctx context.Context, client *redis.Client, isreceive chan interface{}, value ...string) {
	pubsub := client.PSubscribe(ctx, value...)
	defer func() {
		if err := pubsub.Close(); err != nil {
			fmt.Println("Error closing pubsub:", err)
		}
		close(isreceive)
	}()
	for msg := range pubsub.Channel() {
		fmt.Println(msg.Payload, strings.Split(msg.Pattern, ":"))
		val, err := client.Get(ctx, strings.Split(msg.Pattern, ":")[1]).Result()
		if err != nil {
			fmt.Println("Error fetching key:", err)
		} else {
			fmt.Println("New value of key:", val)
			isreceive <- val
			return
		}
	}
}

func (r *Redis) DeleteKey(ctx context.Context, client *redis.Client, key string) error {
	return client.Del(ctx, key).Err()
}
