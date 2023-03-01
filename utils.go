package groWs

// Broadcast sends a Message to all clients in a room
func Broadcast(roomId string, message []byte) {
	GetClientPool().SendToRoom(roomId, message)
}

// BroadcastToAll sends a Message to all clients connected
func BroadcastToAll(message []byte) {
	GetClientPool().SendToAll(message)
}

// BroadcastEvent sends an event to all clients in a room
func BroadcastEvent(roomId string, event Event) error {
	json, err := event.ToJSON()
	if err != nil {
		return err
	}

	if pubSubEnabled {
		return getPubSubClient().PublishEventToRoom(roomId, event)
	} else {
		GetClientPool().SendToRoom(roomId, json)
	}
	return nil
}

// BroadcastEventToAll sends an event to all clients
func BroadcastEventToAll(event Event) error {
	json, err := event.ToJSON()
	if err != nil {
		return err
	}
	if pubSubEnabled {
		return getPubSubClient().PublishEventToAll(event)
	} else {
		GetClientPool().SendToAll(json)
	}
	return nil
}

// BroadcastExcept sends a Message to all clients except the client with the given id
func BroadcastExcept(id string, message []byte) error {
	if pubSubEnabled {
		return getPubSubClient().PublishToAllExcept(id, message)
	} else {
		GetClientPool().SendToAllExcept(id, message)
	}
	return nil
}

// BroadcastEventExcept sends an event to all clients except the client with the given Id
func BroadcastEventExcept(id string, event Event) error {
	json, err := event.ToJSON()
	if err != nil {
		return err
	}
	if pubSubEnabled {
		return getPubSubClient().PublishEventToAllExcept(id, event)
	} else {
		GetClientPool().SendToAllExcept(id, json)
	}
	return nil
}

// BroadcastByMeta sends a Message to all clients with a specific metadata
// TODO: implement pub/sub
func BroadcastByMeta(key string, value interface{}, message []byte) {
	GetClientPool().SendToAllByMeta(key, value, message)
}

// BroadcastEventByMeta sends an event to all clients with a specific metadata
// TODO: implement pub/sub
func BroadcastEventByMeta(key string, value interface{}, event Event) {
	//Convert event to json
	json, err := event.ToJSON()
	if err != nil {
		return
	}
	GetClientPool().SendToAllByMeta(key, value, json)
}

// BroadcastToClient sends a Message to a client with the given Id
func BroadcastToClient(id string, message []byte) error {
	if pubSubEnabled {
		return getPubSubClient().PublishToClient(id, message)
	} else {
		return GetClientPool().SendToClient(id, message)
	}
}

// BroadcastEventToClient sends an event to a client with the given Id
func BroadcastEventToClient(id string, event Event) error {
	json, err := event.ToJSON()
	if err != nil {
		return err
	}
	if pubSubEnabled {
		return getPubSubClient().PublishEventToClient(id, event)
	} else {
		return GetClientPool().SendToClient(id, json)
	}
}

// GetConnectedClientIds returns a list of all connected client ids
func GetConnectedClientIds() []string {
	clientIds := make([]string, 0)
	for id := range GetClientPool().clients {
		clientIds = append(clientIds, id)
	}
	return clientIds
}

// GetConnectedClientIdsByMeta returns a list of all connected client ids with a specific metadata
func GetConnectedClientIdsByMeta(key string, value interface{}) []string {
	clientIds := make([]string, 0)
	for id, client := range GetClientPool().clients {
		if client.meta[key] == value {
			clientIds = append(clientIds, id)
		}
	}
	return clientIds
}

func GetConnectedClientIdsByRoom(roomId string) []string {
	clientIds := make([]string, 0)
	for id := range GetClientPool().rooms[roomId].clients {
		clientIds = append(clientIds, id)
	}
	return clientIds
}

// GetClient returns a client with the given ID
func GetClient(id string) *Client {
	return GetClientPool().GetClient(id)
}

// AddClientToRoom adds a client to a room
func AddClientToRoom(client *Client, roomId string) {
	GetClientPool().AddClientToRoom(client, roomId)
	client.joinRoom(roomId)
}

// RemoveClientFromRoom removes a client from a room
func RemoveClientFromRoom(client *Client, roomId string) {
	GetClientPool().RemoveClientFromRoom(client, roomId)
	client.leaveRoom(roomId)
}

// GetClientRooms returns a list of all rooms the client is in
func GetClientRooms(client *Client) []string {
	return client.GetRooms()
}

// RemoveClientFromAllRooms removes a client from all rooms
func RemoveClientFromAllRooms(client *Client) {
	for _, roomId := range client.GetRooms() {
		RemoveClientFromRoom(client, roomId)
	}
}
