package groWs

import (
	"encoding/json"
	"errors"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
	"net"
	"sync"
)

var ErrMetaNotFound = errors.New("metadata not found")

type Client struct {
	metaMu sync.RWMutex
	meta   map[string]interface{}
	// websocket connection
	conn            net.Conn
	sendMiddlewares []SendMiddleware
	id              string
	roomsMu         sync.RWMutex
	rooms           []string
}

func NewClient(conn net.Conn, middlewares []SendMiddleware) *Client {
	id, _ := uuid.NewUUID()
	return &Client{
		conn:            conn,
		meta:            make(map[string]interface{}),
		sendMiddlewares: middlewares,
		id:              id.String(),
		rooms:           make([]string, 0),
	}
}

// joinRoom adds a room to the client's room list
func (c *Client) joinRoom(room string) {
	c.roomsMu.Lock()
	defer c.roomsMu.Unlock()
	c.rooms = append(c.rooms, room)
}

// leaveRoom is used to remove a room from the client's room list
func (c *Client) leaveRoom(room string) {
	c.roomsMu.Lock()
	defer c.roomsMu.Unlock()
	for i, r := range c.rooms {
		if r == room {
			c.rooms = append(c.rooms[:i], c.rooms[i+1:]...)
		}
	}
}

// GetRooms returns all rooms the client is in
func (c *Client) GetRooms() []string {
	c.roomsMu.RLock()
	defer c.roomsMu.RUnlock()
	return c.rooms
}

// SetMeta adds metadata to a key value map
// Can be used to store data about the client (e.g. username, password, etc.)
// To get the metadata, use the GetMeta function
func (c *Client) SetMeta(key string, value interface{}) {
	c.metaMu.Lock()
	defer c.metaMu.Unlock()
	c.meta[key] = value
}

// GetMeta returns the metadata of the client by key
func (c *Client) GetMeta(key string) (interface{}, error) {
	c.metaMu.RLock()
	defer c.metaMu.RUnlock()
	if c.meta[key] == nil {
		return nil, ErrMetaNotFound
	}
	return c.meta[key], nil
}

// GetID returns the ID of the client
func (c *Client) GetID() string {
	return c.id
}

func (c *Client) setID(id string) {
	c.id = id
}

// getConn returns the connection of the client
func (c *Client) getConn() net.Conn {
	return c.conn
}

// Close closes the connection of the client
func (c *Client) Close() error {
	return c.conn.Close()
}

// Write writes data to the client
func (c *Client) Write(data []byte) error {
	//call send middlewares
	for _, middleware := range c.sendMiddlewares {
		data, _ = middleware(c, data)
	}
	return wsutil.WriteServerMessage(c.conn, ws.OpText, data)
}

// WriteJSON writes JSON data to the client
func (c *Client) WriteJSON(data interface{}) error {
	//convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	//call send middlewares
	for _, middleware := range c.sendMiddlewares {
		jsonData, _ = middleware(c, jsonData)
	}
	//write data to client
	return wsutil.WriteServerMessage(c.conn, ws.OpText, jsonData)
}

// WriteEvent writes an event to the client as JSON
func (c *Client) WriteEvent(event Event) error {
	return c.WriteJSON(event)
}

// read reads data from the client
func (c *Client) read() ([]byte, ws.OpCode, error) {
	return wsutil.ReadClientData(c.conn)
}
