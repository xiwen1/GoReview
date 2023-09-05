package registry

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type registry struct {
	registrations []Registration
	mutex         *sync.RWMutex
}

var reg registry

var pattern = "/services"

func (r registry) add(regis Registration) error {
	r.mutex.RLock()
	r.registrations = append(r.registrations, regis)
	r.mutex.RUnlock()

	return nil
}

func (r registry) remove(regisURL string) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for k, v := range r.registrations {
		if regisURL == v.ServiceURL {
			r.registrations = append(r.registrations[:k], r.registrations[k+1:]...)
			return nil
		}
	}
	return fmt.Errorf("failed to find the target service: %v", regisURL)
}

func RegistryRouter() *gin.Engine {
	r := gin.Default()
	r.POST(pattern, func(c *gin.Context) {
		var regis Registration
		if err := c.ShouldBindJSON(&regis); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if err := reg.add(regis); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	})

	r.DELETE(pattern, func(c *gin.Context) {
		var url string
		if err := c.ShouldBindJSON(&url); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if err := reg.remove(url); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	})
	return r
}
