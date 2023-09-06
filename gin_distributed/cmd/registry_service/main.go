package main

import (
	"context"
	"distributed_gin/registry"
	"fmt"
	"log"
	"net/http"
)

func main() {
	serviceURL := registry.Addr + registry.Pattern

	server := http.Server{
		Addr: serviceURL,
		Handler: registry.RegistryRouter(),
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		log.Println(server.ListenAndServe())
		cancel()
	}()

	go func() {
		fmt.Println("Registry started, press any key to shutdown")
		var s string 
		fmt.Scanf("%v", &s)
		cancel()
	}()

	<-ctx.Done()
	fmt.Println("Registry shutting down")
}