package services

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

type PubSubService struct {
	RedisClient *redis.Client
}

func NewPubSubService(redisClient *redis.Client) *PubSubService {
	return &PubSubService{RedisClient: redisClient}
}

func (s *PubSubService) StartDriverUpdateSubscriber() {
	go func() {
		ctx := context.Background()
		subscriber := s.RedisClient.Subscribe(ctx, "driver:updates")
		defer subscriber.Close()

		log.Println("PubSub: subscribed to driver:updates channel")

		ch := subscriber.Channel()

		for msg := range ch {
			log.Printf("pubSub received: %s", msg.Payload)
		}
	}()
}
