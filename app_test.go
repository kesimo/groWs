package groWs

import (
	"log"
	"testing"
)

func TestNewApp(t *testing.T) {
	config := Config{
		Host: "localhost",
		Port: "8080",
	}
	handler := NewClientHandler()
	handler.OnConnect(func(client *Client) error {
		log.Println("Client connected")
		log.Println(client.GetMeta("user"))
		return nil
	})
	handler.OnDisconnect(func(client *Client) error {
		log.Println("Client disconnected")
		return nil
	})
	handler.On("test", func(client *Client, data []byte) error {
		log.Println("test")
		err := client.Write([]byte("test-back"))
		if err != nil {
			return err
		}
		return nil
	})

	router := NewRouter()
	router.AddRoute("/test", handler)

	app := NewApp(config)
	app.AddRouter(router)
	app.AddHandshakeMiddleware("/test", func(client *Client) bool {
		log.Println("Handshake")
		client.SetMeta("user", "testUser122312")
		return true
	})

	log.Fatalln(app.ListenAndServe())
}
