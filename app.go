package groWs

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"log"
	"net"
	"net/http"
	"regexp"
)

type Config struct {
	// Server
	Host string `json:"host"`
	Port string `json:"port"`
	// tls
	UseTLS bool   `json:"use_tls"`
	Cert   string `json:"cert"`
	Key    string `json:"key"`
}

type App struct {
	config               Config
	server               *Server
	router               *Router
	handshakeMiddlewares map[string]HandshakeMiddleware
	receiveMiddlewares   map[string][]ReceiveMiddleware
	sendMiddlewares      map[string][]SendMiddleware
}

func NewApp(config Config) *App {
	return &App{
		config:               config,
		server:               NewServer(config.Host + ":" + config.Port),
		router:               nil,
		handshakeMiddlewares: make(map[string]HandshakeMiddleware, 0),
		receiveMiddlewares:   make(map[string][]ReceiveMiddleware, 0),
		sendMiddlewares:      make(map[string][]SendMiddleware, 0),
	}
}

func (a *App) AddRouter(router *Router) {
	a.router = router
}

// only one per route allowed if multiple regex match the same route the first one will be used
func (a *App) AddHandshakeMiddleware(route string, middleware HandshakeMiddleware) {
	a.handshakeMiddlewares[route] = middleware
}

// adds a middleware to the route regex (e.g. "/test" or "/test/:id")
func (a *App) AddReceiveMiddleware(route string, middleware ReceiveMiddleware) {
	a.receiveMiddlewares[route] = append(a.receiveMiddlewares[route], middleware)
}

func (a *App) AddSendMiddleware(route string, middleware SendMiddleware) {
	a.sendMiddlewares[route] = append(a.sendMiddlewares[route], middleware)
}

func (a *App) ListenAndServe() error {
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

func (a *App) getMiddlewaresForRoute(route string) (HandshakeMiddleware, []ReceiveMiddleware, []SendMiddleware) {
	hMiddlewares := func(client *Client) bool { return true }
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

		go webSocketHandler(client, handler, handshakeMiddleware, receiveMiddlewares)

	}
}

func webSocketHandler(client *Client, handler ClientHandler, handshakeMiddleware HandshakeMiddleware, receiveMiddlewares []ReceiveMiddleware) {

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
	// run handshake and check if client is authorized
	handshakeResult := handshakeMiddleware(client)
	if !handshakeResult {
		// todo error handling
		return
	}
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

		// handle message
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
	}
}
