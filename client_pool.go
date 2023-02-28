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
	cp.clients[c.getId()] = c
}

// RemoveClient removes a client from the pool
func (cp *ClientPool) RemoveClient(c *Client) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	delete(cp.clients, c.getId())
}

// GetClient returns a client by id
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
		if client.Meta[key] == value {
			clients = append(clients, client)
		}
	}
	return clients
}

// AddClientToRoom adds a client to a room
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
	cp.rooms[roomId].clients[c.getId()] = c
}

// RemoveClientFromRoom removes a client from a room
func (cp *ClientPool) RemoveClientFromRoom(c *Client, roomId string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.rooms[roomId].mu.Lock()
	defer cp.rooms[roomId].mu.Unlock()
	delete(cp.rooms[roomId].clients, c.getId())
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
		delete(cp.rooms[roomId].clients, c.getId())
		if len(cp.rooms[roomId].clients) == 0 {
			delete(cp.rooms, roomId)
		}
		cp.rooms[roomId].mu.Unlock()
	}
}

// GetRoom returns a room by id
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

// SendToRoom sends a message to all clients in a room
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

// SendToAll sends a message to all clients
func (cp *ClientPool) SendToAll(message []byte) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	for _, client := range cp.clients {
		client.Write(message)
	}
}

// SendToAllExcept sends a message to all clients except the client with the given id
func (cp *ClientPool) SendToAllExcept(id string, message []byte) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	for _, client := range cp.clients {
		if client.getId() != id {
			client.Write(message)
		}
	}
}

// SendToAllByMeta sends a message to all clients with a specific metadata
func (cp *ClientPool) SendToAllByMeta(key string, value interface{}, message []byte) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	for _, client := range cp.clients {
		if client.Meta[key] == value {
			client.Write(message)
		}
	}
}

// SendToClient sends a message to a client with the given id
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
