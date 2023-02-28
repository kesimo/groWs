package groWs

import (
	"context"
	"github.com/redis/go-redis/v9"
)

// implementation of redis client for pub/sub

type PubSubClient struct {
	redis *redis.Client
}

func NewPubSubClient(host string, port int) *PubSubClient {

	client := redis.NewClient(&redis.Options{
		Addr:     host + ":" + string(rune(port)),
		Password: "",
	})
	// ping redis
	_, err := client.Ping(context.TODO()).Result()
	if err != nil {
		panic(err)
	}
	return &PubSubClient{
		redis: client,
	}
}

func (c *PubSubClient) Close() error {
	return c.redis.Close()
}

func (c *PubSubClient) Ping() error {
	return c.redis.Ping(context.TODO()).Err()
}

func (c *PubSubClient) Publish(channel string, message string) error {
	return c.redis.Publish(context.TODO(), channel, message).Err()
}

func (c *PubSubClient) Subscribe(channel string, handler func(message string)) error {
	pubsub := c.redis.Subscribe(context.TODO(), channel)
	_, err := pubsub.Receive(context.TODO())
	if err != nil {
		return err
	}

	ch := pubsub.Channel()
	for msg := range ch {
		handler(msg.Payload)
	}

	return nil
}
