package groWs

import (
	"context"
	json2 "encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"log"
	"strconv"
)

// implementation of redis client for pub/sub

var (
	pubSubClientInternal   *pubSubClient
	pubSubEnabled          = false
	ErrPubSubIsNil         = errors.New("pub/sub client is nil")
	defaultChannel         = "grows:default"
	clientChannel          = "grows:client"
	clientEventChannel     = "grows:client:event"
	roomChannel            = "grows:Id"
	roomEventChannel       = "grows:Id:event"
	allClientsChannel      = "grows:all:clients"
	allClientsEventChannel = "grows:all:clients:event"
)

type Payload struct {
	Id      string `json:"Id"`
	Message []byte `json:"Message"`
	Event   Event  `json:"event"`
}

func (p *Payload) toJsonString() string {
	json, _ := json2.Marshal(p)
	return string(json)
}

func (p *Payload) fromJsonString(json string) error {
	return json2.Unmarshal([]byte(json), p)
}

type pubSubClient struct {
	redis *redis.Client
	ctx   context.Context
}

func getPubSubClient() *pubSubClient {
	if pubSubClientInternal == nil {
		panic(ErrPubSubIsNil)
	}
	return pubSubClientInternal
}

func initPubSubClient(ctx context.Context, host string, port int) {
	client := redis.NewClient(&redis.Options{
		Addr:     host + ":" + strconv.Itoa(port),
		Password: "",
	})
	// ping redis
	_, err := client.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	pubSubClientInternal = &pubSubClient{
		redis: client,
		ctx:   ctx,
	}
	pubSubEnabled = true
	pubSubClientInternal.StartSubscribing()
}

func (c *pubSubClient) StartSubscribing() {
	go c.subscribeToAllChannels(c.handleIncomingMessages())
}

func (c *pubSubClient) Close() error {
	return c.redis.Close()
}

func (c *pubSubClient) Ping() error {
	return c.redis.Ping(c.ctx).Err()
}

func (c *pubSubClient) PublishDefault(message string) error {
	payload := Payload{Message: []byte(message)}
	return c.redis.Publish(c.ctx, defaultChannel, payload.toJsonString()).Err()
}

func (c *pubSubClient) PublishToRoom(room string, message []byte) error {
	payload := Payload{
		Id:      room,
		Message: message,
	}
	return c.redis.Publish(c.ctx, roomChannel, payload.toJsonString()).Err()
}

// PublishEventToRoom sends event to a specific room
func (c *pubSubClient) PublishEventToRoom(room string, event Event) error {
	payload := Payload{
		Id:    room,
		Event: event,
	}
	return c.redis.Publish(c.ctx, roomChannel, payload.toJsonString()).Err()
}

// PublishToClient sends message to a specific client
func (c *pubSubClient) PublishToClient(userId string, message []byte) error {
	payload := Payload{
		Id:      userId,
		Message: message,
	}
	return c.redis.Publish(c.ctx, clientChannel, payload.toJsonString()).Err()
}

// PublishToAll sends message to all clients
func (c *pubSubClient) PublishToAll(message []byte) error {
	payload := Payload{Message: message}
	return c.redis.Publish(c.ctx, allClientsChannel, payload.toJsonString()).Err()
}

// PublishEventToAll sends event to all clients
func (c *pubSubClient) PublishEventToAll(event Event) error {
	payload := Payload{
		Event: event,
	}
	return c.redis.Publish(c.ctx, allClientsEventChannel, payload.toJsonString()).Err()
}

// PublishToAllExcept sends message to all clients except the one with the given id
func (c *pubSubClient) PublishToAllExcept(userId string, message []byte) error {
	payload := Payload{
		Id:      userId,
		Message: message,
	}
	return c.redis.Publish(c.ctx, allClientsChannel, payload.toJsonString()).Err()
}

// PublishEventToAllExcept sends event to all clients except the one with the given id
func (c *pubSubClient) PublishEventToAllExcept(userId string, event Event) error {
	payload := Payload{
		Id:    userId,
		Event: event,
	}
	return c.redis.Publish(c.ctx, allClientsEventChannel, payload.toJsonString()).Err()
}

// PublishEventToClient sends event to a specific client
func (c *pubSubClient) PublishEventToClient(userId string, event Event) error {
	payload := Payload{
		Id:    userId,
		Event: event,
	}
	return c.redis.Publish(c.ctx, clientEventChannel, payload).Err()
}

// subscribeToAllChannels subscribes to all channels and calls the handler function
// for each incoming message in a goroutine
func (c *pubSubClient) subscribeToAllChannels(handler func(channel string, message string)) {
	subs := c.redis.Subscribe(c.ctx, defaultChannel, clientChannel, clientEventChannel, roomChannel,
		roomEventChannel, allClientsChannel, allClientsEventChannel)
	for {
		msg, err := subs.ReceiveMessage(c.ctx)
		if err != nil {
			panic(err)
		}
		go handler(msg.Channel, msg.Payload)
	}
}

// handleIncomingMessages handles one messages from pub/sub and sends it to the client pool
// based on the channel identifier and the payload
func (c *pubSubClient) handleIncomingMessages() func(channel string, message string) {
	return func(channel string, message string) {
		//switch case by channel name
		payload := &Payload{}
		err := payload.fromJsonString(message)
		if err != nil {
			return
		}
		switch channel {
		case defaultChannel:
			log.Println("received Message from default channel: " + string(payload.Message))
		case clientChannel:
			_ = GetClientPool().SendToClient(payload.Id, payload.Message)
		case clientEventChannel:
			json, err := payload.Event.ToJSON()
			if err != nil {
				return
			}
			_ = GetClientPool().SendToClient(payload.Id, json)
		case roomChannel:
			GetClientPool().SendToRoom(payload.Id, payload.Message)
		case roomEventChannel:
			json, err := payload.Event.ToJSON()
			if err != nil {
				return
			}
			GetClientPool().SendToRoom(payload.Id, json)
		case allClientsChannel:
			if payload.Id != "" {
				GetClientPool().SendToAllExcept(payload.Id, payload.Message)
				return
			}
			GetClientPool().SendToAll(payload.Message)
		case allClientsEventChannel:
			json, err := payload.Event.ToJSON()
			if err != nil {
				return
			}
			if payload.Id != "" {
				GetClientPool().SendToAllExcept(payload.Id, json)
				return
			}
			GetClientPool().SendToAll(json)
		default:
			return
		}
	}
}
