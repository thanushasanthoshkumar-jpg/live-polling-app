package db

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"

	"live-polling-backend/config"
)

var RedisClient *redis.Client

func ConnectRedis(cfg *config.Config) {
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("invalid REDIS_URL: %v", err)
	}

	client := redis.NewClient(opts)

	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}

	RedisClient = client
	log.Println("connected to Redis")
}

// PollVotesKey returns the Redis hash key holding live vote counts for a poll.
func PollVotesKey(pollID string) string {
	return "poll_votes:" + pollID
}

// PollUpdatesChannel returns the Redis Pub/Sub channel name for a poll.
func PollUpdatesChannel(pollID string) string {
	return "poll_updates:" + pollID
}
