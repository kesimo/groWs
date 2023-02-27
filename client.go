package groWs

import (
	"encoding/json"
	"errors"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"net"
)

var ErrMetaNotFount = errors.New("metadata not found")

type Client struct {
	Meta map[string]interface{}
	// websocket connection
	conn            net.Conn
	sendMiddlewares []SendMiddleware
}

func NewClient(conn net.Conn, middlewares []SendMiddleware) *Client {
	return &Client{
		conn:            conn,
		Meta:            make(map[string]interface{}),
		sendMiddlewares: middlewares,
	}
}

// SetMeta adds metadata to a key value map
// Can be used to store data about the client (e.g. username, password, etc.)
// To get the metadata, use the GetMeta function
func (c *Client) SetMeta(key string, value interface{}) {
	c.Meta[key] = value
}

// GetMeta returns the metadata of the client by key
func (c *Client) GetMeta(key string) (interface{}, error) {
	if c.Meta[key] == nil {
		return nil, ErrMetaNotFount
	}
	return c.Meta[key], nil
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

// Read reads data from the client
func (c *Client) Read() ([]byte, ws.OpCode, error) {
	return wsutil.ReadClientData(c.conn)
}
