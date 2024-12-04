package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func newRedisClient() *RedisClient {
	return &RedisClient{
		client: redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
			Protocol: 2,
		}),
	}
}

func (rc *RedisClient) init(key string) {
	ctx := context.Background()
	_, err1 := rc.client.Get(ctx, key).Result()
	if err1 == nil {
		fmt.Printf("Key: %s already set \n", key)
		return
	}
	err := rc.client.SetBit(ctx, key, 1_000_000-1, 0).Err()
	if err != nil {
		panic(err)
	}

	// Confirm the string is set correctly by retrieving the first few bits
	val, err := rc.client.Get(ctx, key).Result()
	if err != nil {
		fmt.Printf("could not get key: %v\n", err)
	}
	fmt.Printf("Len of the stored value: %d\n", len(val))
}

func (rc *RedisClient) setBit(key string, index int64, value int) {
	ctx := context.Background()

	err := rc.client.SetBit(ctx, key, index, value).Err()
	if err != nil {
		panic(err)
	}
}

func (rc *RedisClient) get(key string) (string, error) {
	ctx := context.Background()

	val, err := rc.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (rc *RedisClient) incr(key string) error {
	ctx := context.Background()

	_, err := rc.client.Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	return nil
}

func (rc *RedisClient) publish(channel string, message string) error {
	ctx := context.Background()

	return rc.client.Publish(ctx, channel, message).Err()
}

func (rc *RedisClient) subscribe(channel string) *redis.PubSub {
	ctx := context.Background()

	return rc.client.Subscribe(ctx, channel)

}
