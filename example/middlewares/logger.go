package middlewares

import (
	groWs "github.com/kesimo/grows"
	"log"
)

var LoggerMiddlewareInstance = &LoggerMiddleware{}

type LoggerMiddleware struct {
}

func (m *LoggerMiddleware) HandleReceive() groWs.ReceiveMiddleware {
	return func(client *groWs.Client, data []byte) ([]byte, error) {
		role, _ := client.GetMeta("Role")
		log.Printf("Receive: %s - %s", role, string(data))
		return data, nil
	}
}

func (m *LoggerMiddleware) HandleSend() groWs.SendMiddleware {
	return func(client *groWs.Client, data []byte) ([]byte, error) {
		log.Println("Send: " + string(data))
		return data, nil
	}
}
