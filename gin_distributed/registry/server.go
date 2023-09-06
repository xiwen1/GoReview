package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type registry struct {
	registrations []Registration
	mutex         *sync.RWMutex
}

var reg registry

var Addr = "http://localhost:5000"
var Pattern = "/services"

var ServicesURL = Addr + Pattern

func (r registry) add(regis Registration) error {
	r.mutex.RLock()
	r.registrations = append(r.registrations, regis)
	r.notify(patch{
		Added: []patchEntry {
			{
				Name: regis.ServiceName,
				URL: regis.ServiceURL,
			},
		},
	})
	r.mutex.RUnlock()

	return nil
}

func (r registry) remove(regisURL string) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for k, v := range r.registrations {
		if regisURL == v.ServiceURL {
			r.registrations = append(r.registrations[:k], r.registrations[k+1:]...)
			r.notify(patch{
				Removed: []patchEntry {
					{
						Name: v.ServiceName,
						URL: regisURL,
					},
				},
			})
			return nil
		}
	}
	return fmt.Errorf("failed to find the target service: %v", regisURL)
}

func (r *registry) notify(fullpatch patch) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, regis := range r.registrations {
		go func(regis Registration) {
			p := patch {}
			sendUpdate := false
			for _, req := range regis.RequiredServices {
				for _, added := range fullpatch.Added {
					if added.Name == req {
						p.Added = append(p.Added, added)
						sendUpdate = true
					}
				}
				for _, removed := range fullpatch.Removed {
					if removed.Name == req {
						p.Removed = append(p.Removed, removed)
						sendUpdate = true
					}
				}
			}
			if sendUpdate {
				if err := r.sendPatch(p, regis.ServiceUpdateURL); err != nil {
					log.Println(err)
					return
				}
			}
		}(regis)
	}
}

func (r *registry) sendPatch(p patch, url string) error {
	patchJson, err := json.Marshal(p)
	if err != nil {
		return err 
	}
	res, err := http.Post(url, "application/json", bytes.NewBuffer(patchJson))
	if err != nil {
		return err 
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send patch to service: %v", url)
	}
	return nil 
}

var once sync.Once

func SetupHeartbeat() {
	once.Do(func ()  {
		reg.heartbeat(60 * time.Second)
	})
}

func (r *registry) heartbeat(freq time.Duration) {
	for {
		wg := &sync.WaitGroup{}

		for _, regis := range r.registrations {
			wg.Add(1)
			go func(regis Registration) {
				removed := false
				defer wg.Done()
				for attempt := 0; attempt < 3; attempt ++ {
					res, err := http.Get(regis.HeartbeatURL)
					if err != nil {
						log.Println(err) 
					} else if res.StatusCode == http.StatusOK {
						log.Printf("HeartBeat check passed for %v", regis.ServiceName) 
						if removed {
							r.add(regis)
						}
						break
					}
					if !removed {
						removed = true
						r.remove(regis.ServiceURL)
					}
				}
			}(regis) // 将参数传入确保不受外界变量的影响
		}

		wg.Wait()
		time.Sleep(freq)
	}
}

func RegistryRouter() *gin.Engine {
	r := gin.Default()
	r.POST(Pattern, func(c *gin.Context) {
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

	r.DELETE(Pattern, func(c *gin.Context) {
		var url struct {
			url string
		}
		if err := c.ShouldBind(&url); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if err := reg.remove(url.url); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	})
	return r
}
