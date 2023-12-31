package service

import (
	"context"
	"distributed_gin/registry"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Start( // 入口函数不应承担太多功能
	reg registry.Registration,
	ctx context.Context,
	registerHandlers func(*gin.Engine),
) (context.Context, error) {

	r := gin.Default()
	registerHandlers(r)
	r, err := registry.RegisterUpdateService(r, reg)
	if err != nil {
		log.Println(err)
		return ctx, err 
	}
	ctx = startService(reg, ctx, r)
	
	return ctx, nil
}

func startService(reg registry.Registration, ctx context.Context,
	r *gin.Engine) context.Context {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		server := http.Server{
			Addr: reg.ServiceURL,
			Handler: r,
		}
		go func() {
			log.Println(server.ListenAndServe())
			if err := registry.Shutdown(reg.ServiceURL); err != nil {
				log.Print(err)
			}
			cancel()
		}()

		go func() {
			fmt.Printf("Service %v started, press any key to shutdown\n", reg.ServiceName)
			var s string
			fmt.Scanf("%v", &s)
			if err := registry.Shutdown(reg.ServiceURL); err != nil {
				log.Print(err)
			}
			server.Shutdown(ctx)
			cancel()
		}()
	return ctx
}
