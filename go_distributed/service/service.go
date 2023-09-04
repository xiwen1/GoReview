package service

import (
	"context"
	"distributed/registry"
	"fmt"
	"log"
	"net/http"
)

func Start(ctx context.Context, host, port string, r registry.Registration,
	RegisterHandlers func()) (context.Context, error) {
	RegisterHandlers()
	ctx = startService(ctx, host, port, r)
	if err := registry.RegisterService(r); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func startService(ctx context.Context, host, port string, 
	r registry.Registration) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	//http.ServeMux is to create a mux for server's handler
	srv := http.Server{
		Addr: ":" + port,
	}
	
	go func() {
		log.Println(srv.ListenAndServe())
		if err := registry.ShutdownService(r.ServiceURL); err != nil {
			log.Println(err)
		}
		cancel()
	}()

	go func() {
		fmt.Printf("%v started, press any key to stop", r.ServiceName)
		var s string
		fmt.Scanln(&s)
		if err := registry.ShutdownService(r.ServiceURL); err != nil {
			log.Println(err)
		}
		srv.Shutdown(ctx)
		cancel()
	}()

	return ctx

}