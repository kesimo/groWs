package groWs

import (
	"log"
	"net/http"
	"sync"
)

type Server struct {
	Server *http.Server
	mu     sync.Mutex
	sMux   *http.ServeMux
}

func NewServer(addr string) *Server {
	// Create mux
	sMux := http.NewServeMux()

	// create server
	s := &http.Server{
		Addr:    addr,
		Handler: sMux,
	}

	// return server
	return &Server{
		Server: s,
		sMux:   sMux,
	}
}

func (s *Server) AddHandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	if s.Server == nil || s.sMux == nil {
		panic("Server not initialized")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	// Add handler to server mux
	s.sMux.HandleFunc(pattern, handler)
}

func (s *Server) ListenAndServe() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Listen and serve
	log.Println("Listening on " + s.Server.Addr)
	return s.Server.ListenAndServe()
}

func (s *Server) ListenAndServeTLS(certFile, keyFile string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Listen and serve
	log.Println("Listening on " + s.Server.Addr)
	return s.Server.ListenAndServeTLS(certFile, keyFile)
}
