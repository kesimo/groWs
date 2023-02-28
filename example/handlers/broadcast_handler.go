package handlers

import (
	"example/services"
	groWs "github.com/kesimo/grows"
	"log"
	"math/rand"
	"strconv"
)

type BroadcastHandler struct {
	handler groWs.ClientHandler
}

func NewBroadcastHandler() *BroadcastHandler {
	handler := groWs.NewClientHandler()
	broadcastService := services.NewBroadcastService()
	infoService := services.NewInfoService()

	handler.OnConnect(func(client *groWs.Client) error {
		client.SetMeta("Identifier", rand.Int()) // Set Metadata to the client
		client.SetMeta("Name", "TestUser")
		groWs.AddClientToRoom(client, "broadcastingRoom")
		return nil
	})
	handler.OnDisconnect(func(client *groWs.Client) error {
		identifier, _ := client.GetMeta("Identifier") // Get Metadata from the client
		log.Printf("Client %s disconnected", strconv.Itoa(identifier.(int)))
		return nil
	})
	handler.On("TEST REQUEST", func(client *groWs.Client, data []byte) error {
		return client.Write([]byte("TEST RESPONSE"))
	})
	handler.OnEvent("server-info", func(client *groWs.Client, data any) error {
		info := infoService.GetServerInfo()
		return client.WriteEvent(groWs.Event{
			Identifier: "server-info",
			Data:       info,
		})
	})
	handler.OnEvent("broadcast", func(client *groWs.Client, data any) error {
		return broadcastService.BroadcastToAllClients(data)
	})
	return &BroadcastHandler{handler: handler}
}

func (h *BroadcastHandler) GetHandler() groWs.ClientHandler {
	return h.handler
}
