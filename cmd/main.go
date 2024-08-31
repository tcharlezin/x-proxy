package main

import (
	"fmt"
	"log"
	"net/http"
	"x-proxy/app"
	"x-proxy/cmd/proxy"
)

func main() {

	mutex := http.NewServeMux()
	mutex.HandleFunc("/", proxy.Proxy)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", app.Application.WebPort),
		Handler: mutex,
	}

	app.Application.Log.Info(fmt.Sprintf("Starting server on :%s", app.Application.WebPort))
	err := server.ListenAndServe()

	if err != nil {
		log.Fatalln(err)
	}
}
