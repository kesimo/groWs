package groWs

import "net/http"

type ReceiveMiddleware func(*Client, []byte) ([]byte, error)

type SendMiddleware func(*Client, []byte) ([]byte, error)

type HandshakeMiddleware = func(r *http.Request, client *Client) bool
