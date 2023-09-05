package main

import (
	stlog "log"
	"context"
	"distributed_gin/log"
	"distributed_gin/registry"
	"distributed_gin/service"
	"fmt"
)

func main() {
	host, port := "localhost", "4000"
	ServiceAddr := fmt.Sprintf("http://%v:%v", host, port)
	reg := registry.Registration {
		ServiceName: "LogService",
		ServiceURL: ServiceAddr + "/log",
		RequiredServices: make([]registry.ServiceName, 0),
		ServiceUpdateURL: ServiceAddr + "/services",
		HeartbeatURL: ServiceAddr + "/heartbeat",
	}

	ctx, err := service.Start(reg, context.Background(), log.RegisterHandlers)
	if err != nil {
		stlog.Fatal(err)
	}
	<- ctx.Done()
	fmt.Println("Shutting down log service")
}