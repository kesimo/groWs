package main

import (
	"example/handlers"
	"example/middlewares"
	"github.com/kesimo/grows"
	"log"
)

func main() {
	config := groWs.Config{
		Host:         "localhost",
		Port:         8080,
		EnablePubSub: true,
		RedisHost:    "localhost",
		RedisPort:    6379,
	}

	app := groWs.NewApp(config)
	app.AddRouter(handlers.ExampleRouter())
	app.AddHandshakeMiddleware("/example", middlewares.NewBasicAuthMiddleware().HandleHandshake())
	app.AddReceiveMiddleware("/example", middlewares.LoggerMiddlewareInstance.HandleReceive())
	app.AddSendMiddleware("/example", middlewares.LoggerMiddlewareInstance.HandleSend())
	app.AddSendMiddleware("/example", func(client *groWs.Client, data []byte) ([]byte, error) {
		if groWs.IsEvent(data) {
			log.Println("Event sent: " + string(data))
		}
		return data, nil
	})

	log.Fatalln(app.ListenAndServe())
}
