package groWs

import (
	"log"
	"sync"
)

var (
	// ClientPool is the global client pool
	clientPool *ClientPool
)

type Room struct {
	// mu Mutex
	mu sync.RWMutex
	// Room participants
	clients map[string]*Client
}

type ClientPool struct {
	clients map[string]*Client
	mu      sync.RWMutex
	rooms   map[string]*Room
}

func newClientPool() *ClientPool {
	return &ClientPool{
		clients: make(map[string]*Client),
		rooms:   make(map[string]*Room),
	}
}

func GetClientPool() *ClientPool {
	if clientPool == nil {
		clientPool = newClientPool()
	}
	return clientPool
}

// AddClient adds a client to the pool
func (cp *ClientPool) AddClient(c *Client) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.clients[c.GetID()] = c
}

// RemoveClient removes a client from the pool
func (cp *ClientPool) RemoveClient(c *Client) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	delete(cp.clients, c.GetID())
}

// GetClient returns a client by Id
func (cp *ClientPool) GetClient(id string) *Client {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.clients[id]
}

// GetClients returns all clients
func (cp *ClientPool) GetClients() map[string]*Client {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.clients
}

// GetClientsByMeta returns all clients with a specific metadata
func (cp *ClientPool) GetClientsByMeta(key string, value interface{}) []*Client {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	var clients []*Client
	for _, client := range cp.clients {
		if client.meta[key] == value {
			clients = append(clients, client)
		}
	}
	return clients
}

// AddClientToRoom adds a client to a Id
func (cp *ClientPool) AddClientToRoom(c *Client, roomId string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	if cp.rooms[roomId] == nil {
		cp.rooms[roomId] = &Room{
			clients: make(map[string]*Client),
		}
	}
	cp.rooms[roomId].mu.Lock()
	defer cp.rooms[roomId].mu.Unlock()
	cp.rooms[roomId].clients[c.GetID()] = c
}

// RemoveClientFromRoom removes a client from a Id
func (cp *ClientPool) RemoveClientFromRoom(c *Client, roomId string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.rooms[roomId].mu.Lock()
	defer cp.rooms[roomId].mu.Unlock()
	delete(cp.rooms[roomId].clients, c.GetID())
	if len(cp.rooms[roomId].clients) == 0 {
		delete(cp.rooms, roomId)
	}
}

// RemoveClientFromAllRooms removes a client from all rooms
func (cp *ClientPool) RemoveClientFromAllRooms(c *Client, rooms []string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	for _, roomId := range rooms {
		cp.rooms[roomId].mu.Lock()
		delete(cp.rooms[roomId].clients, c.GetID())
		if len(cp.rooms[roomId].clients) == 0 {
			delete(cp.rooms, roomId)
		} else {
			cp.rooms[roomId].mu.Unlock()
		}
	}
}

// GetRoom returns a room by identifier
func (cp *ClientPool) GetRoom(roomId string) *Room {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.rooms[roomId]
}

// GetRooms returns all rooms
func (cp *ClientPool) GetRooms() map[string]*Room {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.rooms
}

// SendToRoom sends a Message to all clients in a Id
func (cp *ClientPool) SendToRoom(roomId string, message []byte) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	if cp.rooms[roomId] != nil {
		cp.rooms[roomId].mu.RLock()
		defer cp.rooms[roomId].mu.RUnlock()
		for _, client := range cp.rooms[roomId].clients {
			err := client.Write(message)
			if err != nil {
				log.Println(err) // TODO: handle error
			}
		}
	}
}

// SendToAll sends a Message to all clients
func (cp *ClientPool) SendToAll(message []byte) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	for _, client := range cp.clients {
		client.Write(message)
	}
}

// SendToAllExcept sends a Message to all clients except the client with the given identifier
func (cp *ClientPool) SendToAllExcept(id string, message []byte) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	for _, client := range cp.clients {
		if client.GetID() != id {
			client.Write(message)
		}
	}
}

// SendToAllByMeta sends a Message to all clients with a specific metadata
func (cp *ClientPool) SendToAllByMeta(key string, value interface{}, message []byte) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	for _, client := range cp.clients {
		if client.meta[key] == value {
			client.Write(message)
		}
	}
}

// SendToClient sends a Message to a client with the given Id
func (cp *ClientPool) SendToClient(id string, message []byte) error {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	if client, ok := cp.clients[id]; ok {
		err := client.Write(message)
		if err != nil {

			return err
		}
	}
	return nil
}
