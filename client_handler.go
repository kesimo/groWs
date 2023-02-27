package groWs

import (
	"errors"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type ClientHandler struct {
	// map of string, function(client)
	onConnect    func(*Client) error
	onDisconnect func(*Client) error
	on           map[string]func(client *Client, data []byte) error
	onEvent      map[string]func(client *Client, data []byte) error
	middlewares  []HandlerFunc
}

func NewClientHandler() ClientHandler {
	return ClientHandler{
		onDisconnect: nil,
		onConnect:    nil,
		on:           make(map[string]func(*Client, []byte) error),
		onEvent:      make(map[string]func(*Client, []byte) error),
	}
}

// OnConnect sets the onConnect function
func (ch *ClientHandler) OnConnect(f func(client *Client) error) {
	ch.onConnect = f
}

// OnDisconnect sets the onDisconnect function
func (ch *ClientHandler) OnDisconnect(f func(client *Client) error) {
	ch.onDisconnect = f
}

// On sets the on function
func (ch *ClientHandler) On(event string, f func(client *Client, data []byte) error) {
	ch.on[event] = f
}

// OnEvent sets the onEvent function
func (ch *ClientHandler) OnEvent(event string, f func(client *Client, data []byte) error) {
	ch.onEvent[event] = f
}

// handle handles an incoming on
func (ch *ClientHandler) handle(data []byte, op ws.OpCode, c *Client) error {
	if ch == nil {
		return errors.New("client handler is nil")
	}
	// switch handle by opcode
	switch op {
	case ws.OpClose:
		if ch.onDisconnect == nil {
			return nil
		}
		return ch.onDisconnect(c)
	case ws.OpText:
		if isJSON(data) && isEvent(data) {
			return ch.handleOnEvent(data, c)
		}
		return ch.handleOn(data, c)
	case ws.OpPing:
		return wsutil.WriteServerMessage(c.conn, ws.OpPong, data)
	case ws.OpPong:
		return nil

	default:
		return nil
	}
}

// handleEvent handles an incoming event
func (ch *ClientHandler) handleOnEvent(data []byte, c *Client) error {
	event, err := eventFromJSON(data)
	if err != nil {
		return err
	}
	if ch.onEvent[event.Identifier] == nil {
		if ch.onEvent["*"] == nil {
			return nil
		}
		return ch.onEvent["*"](c, event.Data)
	}
	return ch.onEvent[event.Identifier](c, event.Data)
}

// handleOn handles an incoming on
func (ch *ClientHandler) handleOn(data []byte, c *Client) error {
	if ch.on[string(data)] == nil {
		if ch.on["*"] == nil {
			return nil
		}
		return ch.on["*"](c, data)
	}
	return ch.on[string(data)](c, data)
}
