package groWs

import (
	"context"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"log"
	"net"
	"net/http"
	"regexp"
	"strconv"
)

type Config struct {
	// Server
	Host string `json:"host"`
	Port int    `json:"port"`
	// tls
	UseTLS bool   `json:"use_tls"`
	Cert   string `json:"cert"`
	Key    string `json:"key"`
	// PubSub
	EnablePubSub bool   `json:"enable_pub_sub"`
	RedisHost    string `json:"pub_sub_host"`
	RedisPort    int    `json:"pub_sub_port"`
}

type App struct {
	config               Config
	server               *Server
	router               *Router
	handshakeMiddlewares map[string]HandshakeMiddleware
	receiveMiddlewares   map[string][]ReceiveMiddleware
	sendMiddlewares      map[string][]SendMiddleware
	ctx                  context.Context
}

func NewApp(config Config) *App {
	if config.Port == 0 {
		config.Port = 8080
	}
	if config.EnablePubSub {
		log.Println("PubSub enabled")
		initPubSubClient(context.Background(), config.RedisHost, config.RedisPort)
		log.Println("Redis connection established")
	}
	return &App{
		config:               config,
		server:               NewServer(config.Host + ":" + strconv.Itoa(config.Port)),
		router:               nil,
		handshakeMiddlewares: make(map[string]HandshakeMiddleware, 0),
		receiveMiddlewares:   make(map[string][]ReceiveMiddleware, 0),
		sendMiddlewares:      make(map[string][]SendMiddleware, 0),
		ctx:                  context.Background(),
	}
}

func (a *App) AddRouter(router *Router) {
	a.router = router
}

// AddHandshakeMiddleware adds a middleware that is called before the websocket handshake
// only one per route allowed if multiple regex match the same route the first one will be used
func (a *App) AddHandshakeMiddleware(route string, middleware HandshakeMiddleware) {
	a.handshakeMiddlewares[route] = middleware
}

// AddReceiveMiddleware adds a middleware to the route regex (e.g. "/test" or "/test/:Id")
// multiple middlewares can be added to the same route
// (Caution: not ordered by adding order if multiple regex match the same route)
// Example:
// - "/test" will match "/test"
// - * will match everything
func (a *App) AddReceiveMiddleware(route string, middleware ReceiveMiddleware) {
	a.receiveMiddlewares[route] = append(a.receiveMiddlewares[route], middleware)
}

// AddSendMiddleware adds a middleware to the route regex (e.g. "/test" or ".*")
func (a *App) AddSendMiddleware(route string, middleware SendMiddleware) {
	a.sendMiddlewares[route] = append(a.sendMiddlewares[route], middleware)
}

// ListenAndServe starts the server and listens for incoming connections
// It will use TLS if the config.UseTLS is set to true and a cert and key are provided
// It will panic if no router is added (or for TLS no cert or key is provided)
func (a *App) ListenAndServe() error {
	defer func() {
		if pubSubEnabled {
			_ = getPubSubClient().Close()
		}
	}()
	if a.router == nil {
		panic("No router added")
	}
	for _, route := range a.router.routes {
		// todo add middleware (global and route specific)
		log.Println("Registering route: " + route.Path)
		a.server.AddHandleFunc(route.Path, a.buildHandlerFunc(route.Path, route.Handler))
	}
	if a.config.UseTLS {
		if a.config.Cert == "" || a.config.Key == "" {
			panic("No cert or key provided")
		}
		return a.server.ListenAndServeTLS(a.config.Cert, a.config.Key)
	}
	return a.server.ListenAndServe()
}

// getMiddlewaresForRoute returns all middlewares the matches the given route
func (a *App) getMiddlewaresForRoute(route string) (HandshakeMiddleware, []ReceiveMiddleware, []SendMiddleware) {
	hMiddlewares := func(r *http.Request, client *Client) bool { return true }
	rMiddlewares := make([]ReceiveMiddleware, 0)
	sMiddlewares := make([]SendMiddleware, 0)
	// generate route specific middleware chain by regex matching
	for regex, receiveMiddlewares := range a.receiveMiddlewares {
		// check if regex from middleware matches the route
		if match, _ := regexp.MatchString(regex, route); match {
			rMiddlewares = append(rMiddlewares, receiveMiddlewares...)
		}
	}
	for regex, sendMiddlewares := range a.sendMiddlewares {
		// check if regex from middleware matches the route
		if match, _ := regexp.MatchString(regex, route); match {
			sMiddlewares = append(sMiddlewares, sendMiddlewares...)
		}
	}
	for regex, handshake := range a.handshakeMiddlewares {
		// check if regex from middleware matches the route
		if match, _ := regexp.MatchString(regex, route); match {
			hMiddlewares = handshake
			// skip after first match because only one handshake middleware per route is allowed
			break
		}
	}
	return hMiddlewares, rMiddlewares, sMiddlewares
}

// buildHandlerFunc builds a http.HandlerFunc that handles the websocket connection
// it applies the middlewares for the given route
// HandshakeMiddleware is only applied once per connection and called before loop -> false if client should not connect
// ReceiveMiddleware is applied for every Message received (in loop)
// SendMiddleware is applied to the Client and is called on Client.WriteJSON or Client.Write
func (a *App) buildHandlerFunc(route string, handler ClientHandler) HandlerFunc {
	handshakeMiddleware, receiveMiddlewares, sendMiddlewares := a.getMiddlewaresForRoute(route)
	if handshakeMiddleware != nil {
		log.Printf("apply HandshakeMiddleware for route %s", route)
	}
	if len(receiveMiddlewares) > 0 {
		log.Printf("apply %d ReceiveMiddleware for route %s", len(receiveMiddlewares), route)
	}
	if len(sendMiddlewares) > 0 {
		log.Printf("apply %d SendMiddleware for route %s", len(sendMiddlewares), route)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		// Upgrade connection
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			panic(err)
		}
		// Create client
		client := NewClient(conn, sendMiddlewares)

		// run handshake and check if client is authorized
		handshakeResult := handshakeMiddleware(r, client)
		if !handshakeResult {
			// todo error handling
			return
		}

		go webSocketHandler(client, handler, receiveMiddlewares)

	}
}

// webSocketHandler handles the websocket connection in a loop on a separate goroutine
func webSocketHandler(client *Client, handler ClientHandler, receiveMiddlewares []ReceiveMiddleware) {

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
		err = handler.onDisconnect(client)
		if err != nil {
			log.Println(err)
		}
		GetClientPool().RemoveClient(client)
		GetClientPool().RemoveClientFromAllRooms(client, client.GetRooms())
	}(client.getConn())
	err := handler.onConnect(client)
	if err != nil {
		// todo error handling
		return
	}

	// add client to pool
	GetClientPool().AddClient(client)

	// authorized, continue with WebSocket connection
	for {
		msg, opCode, err := wsutil.ReadClientData(client.getConn())
		if err != nil {
			break
		}
		go func(msg []byte, opCode ws.OpCode) {
			// handle Message
			var middlewareError error
			for _, middleware := range receiveMiddlewares {
				msg, middlewareError = middleware(client, msg)
				if middlewareError != nil {
					log.Println("middleware error: ", middlewareError) // todo error handling
				}
			}
			handlerErr := handler.handle(msg, opCode, client)
			if handlerErr != nil {
				log.Println(handlerErr)
			}
		}(msg, opCode)

	}
}
