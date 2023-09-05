package main

import (
	"context"
	"distributed/registry"
	"fmt"
	"log"
	"net/http"
)

func main() {
	registry.SetupRegistryService()
	http.Handle("/services", &registry.RegistryService{})
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	srv := http.Server{
		Addr: registry.ServerPort,
	}

	go func ()  {
		log.Println(srv.ListenAndServe())
		cancel()
	}()

	go func() {
		fmt.Println("Registry starting, press any key to shutdown")
		var s string
		fmt.Scanln(&s)
		cancel()
	}()

	<- ctx.Done()
	fmt.Println("Shutting down registry service")
}