package groWs

type ReceiveMiddleware func(*Client, []byte) ([]byte, error)

type SendMiddleware func(*Client, []byte) ([]byte, error)

type HandshakeMiddleware = func(client *Client) bool
