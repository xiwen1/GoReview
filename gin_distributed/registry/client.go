package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/gin-gonic/gin"
)

// 注意加锁
type providers struct {
	services map[ServiceName][]string
	mutex    *sync.RWMutex
}

var prov providers

func RegisterUpdateService(r *gin.Engine, reg Registration) (*gin.Engine, error) {
	serviceURL, err := url.Parse(reg.ServiceUpdateURL)
	if err != nil {
		return r, err
	}
	r.POST(serviceURL.Path, UpdateService)
	regJson, err := json.Marshal(reg)
	if err != nil {
		return r, err
	}

	res, err := http.Post(Addr+Pattern, "application/json", bytes.NewBuffer(regJson))
	if err != nil {
		return r, err
	}
	if res.StatusCode != http.StatusOK {
		return r, fmt.Errorf("failed to register service to registry, with code: %v", res.StatusCode)
	}
	return r, nil
}

func UpdateService(c *gin.Context) {
	var patch patch
	if err := c.ShouldBindJSON(&patch); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if err := prov.Update(patch); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	c.Status(http.StatusOK)
}

func (p *providers) Update(pat patch) error {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	for _, added := range pat.Added {
		if _, ok := p.services[added.Name]; !ok {
			p.services[added.Name] = make([]string, 0)
			p.services[added.Name] = append(p.services[added.Name], added.URL)
			continue
		}
		serv := p.services[added.Name]
		flag := true
		for _, url := range serv {
			if url == added.URL {
				flag = false 
				break
			}
		}
		if flag {
			serv = append(serv, added.URL)
		}
	}

	for _, removed := range pat.Removed {
		if _, ok := p.services[removed.Name]; !ok {
			return fmt.Errorf("failed to find service in providers: %v", removed.Name)
		}
		serv := p.services[removed.Name]
		for k, v := range serv {
			if v == removed.URL {
				serv = append(serv[:k], serv[k+1:]...)
				return nil
			}
		}
		return fmt.Errorf("failed ot find service in providers: %v", removed.URL)
	}
	return nil
}


func Shutdown(url string) error {
	urlJson, err := json.Marshal(gin.H{"url": url})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodDelete, ServicesURL, bytes.NewBuffer(urlJson))
	if err != nil {
		return err 
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err 
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to remove registration in registry, status code: %v", res.StatusCode)
	}
	return nil 
}


