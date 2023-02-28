package services

import groWs "github.com/kesimo/grows"

type BroadcastService struct {
}

func NewBroadcastService() *BroadcastService {
	return &BroadcastService{}
}

func (s *BroadcastService) BroadcastToAllClients(data any) error {
	event := groWs.Event{
		Identifier: "broadcast",
		Data:       data,
	}
	return groWs.BroadcastEventToAll(event)
}
