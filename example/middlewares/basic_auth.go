package middlewares

import (
	groWs "github.com/kesimo/grows"
	"net/http"
)

type BasicAuthMiddleware struct {
	users map[string]string
}

func NewBasicAuthMiddleware() *BasicAuthMiddleware {
	users := make(map[string]string)
	users["admin"] = "admin"
	users["user"] = "user"
	return &BasicAuthMiddleware{users: users}
}

func (m *BasicAuthMiddleware) HandleHandshake() groWs.HandshakeMiddleware {
	return func(r *http.Request, client *groWs.Client) bool {
		user, pass, ok := r.BasicAuth()
		if !ok {
			return false
		}
		if m.users[user] != pass {
			return false
		}
		client.SetMeta("Role", user)
		return true
	}
}
