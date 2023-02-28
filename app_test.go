package groWs

import (
	"log"
	"testing"
)

func TestNewApp(t *testing.T) {
	config := Config{
		Host:         "localhost",
		Port:         "8080",
		EnablePubSub: true,
		RedisHost:    "localhost",
		RedisPort:    6379,
	}
	handler := NewClientHandler()
	handler.OnConnect(func(client *Client) error {
		log.Println("Client connected")
		log.Println(client.GetMeta("Id"))
		AddClientToRoom(client, "testRoom")
		return nil
	})
	handler.OnDisconnect(func(client *Client) error {
		log.Println("Client disconnected")
		return nil
	})
	handler.On("test", func(client *Client, data []byte) error {
		log.Println(string(data))
		err := client.Write([]byte("test-back"))
		if err != nil {
			return err
		}
		return nil
	})
	handler.OnEvent("testRoom", func(client *Client, data any) error {
		log.Println("event: " + data.(string))
		err := BroadcastEvent("testRoom", Event{
			Identifier: "test-back-event",
			Data:       "test-back-event-data",
		})
		if err != nil {
			t.Errorf("Error broadcasting event: %s", err.Error())
		}
		return nil
	})

	router := NewRouter()
	router.AddRoute("/test", handler)

	app := NewApp(config)
	app.AddRouter(router)
	app.AddHandshakeMiddleware("/test", func(client *Client) bool {
		log.Println("Handshake")
		client.SetMeta("Id", "testUser122312")
		return true
	})
	app.AddReceiveMiddleware("/test", func(client *Client, data []byte) ([]byte, error) {
		log.Println("Receive: " + string(data))
		return data, nil
	})
	app.AddSendMiddleware("/test", func(client *Client, data []byte) ([]byte, error) {
		log.Println("Send: " + string(data))
		return data, nil
	})

	log.Fatalln(app.ListenAndServe())
}
