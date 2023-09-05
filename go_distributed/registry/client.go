package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
)

// 注意：客户端是跑在每个使用该服务的客户的进程上的
func RegisterService(r Registration) error {
	serviceUpdateURL, err := url.Parse(r.ServiceUpdateURL)
	if err != nil {
		return err
	}
	http.Handle(serviceUpdateURL.Path, &serviceUpdateHandler{})

	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(r); err != nil {
		return err
	}

	res, err := http.Post(ServicesURL, "application/json", buf)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register service. Registry service "+
			"response with code: %v", res.StatusCode)
	}

	return nil
}

type serviceUpdateHandler struct{}

func (s *serviceUpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	dec := json.NewDecoder(r.Body)
	var p patch
	if err := dec.Decode(&p); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println("Request received")
	prov.Update(p)
}

func ShutdownService(url string) error {
	req, err := http.NewRequest(http.MethodDelete, ServicesURL,
		bytes.NewBuffer([]byte(url)))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "text/plain")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to deregister service. Registry service"+
			"service responded with code: %v", res.StatusCode)
	}
	return nil
}

type providers struct {
	services map[ServiceName][]string
	mutex    *sync.RWMutex
}

var prov = providers{
	services: make(map[ServiceName][]string),
	mutex:    &sync.RWMutex{},
}

func (p *providers) Update(pat patch) error {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	for _, patchEntry := range pat.Added {
		if _, ok := p.services[patchEntry.Name]; !ok {
			p.services[patchEntry.Name] = make([]string, 0)
		}
		var urls = p.services[patchEntry.Name] // slice is passsed by reference
		checked := true
		for _, url := range urls {
			if url == patchEntry.URL {
				checked = false
				break
			}
		}
		if checked {
			urls = append(urls, patchEntry.URL)
		}
	}

	for _, patchEntry := range pat.Removed {
		if urls, ok := p.services[patchEntry.Name]; ok {
			for k, url := range urls {
				if url == patchEntry.URL {
					urls = append(urls[:k], urls[k+1:]...)
				}
			}
		}
	}
	return nil
}

func  (p *providers) get(name ServiceName) (string, error) {
	if _, ok := p.services[name]; !ok {
		return "", fmt.Errorf("failed to get service from providers")
	}
	idx := int(rand.Float32() * float32(len(p.services[name])))
	return p.services[name][idx], nil
}

func GetProvider(name ServiceName) (string, error) {
	return prov.get(name)
}
