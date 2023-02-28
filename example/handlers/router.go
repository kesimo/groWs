package handlers

import groWs "github.com/kesimo/grows"

func ExampleRouter() *groWs.Router {
	router := groWs.NewRouter()
	router.AddRoute("/example", NewBroadcastHandler().GetHandler())
	return router
}
