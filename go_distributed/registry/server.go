package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

const ServerPort = ":3000"
const ServicesURL = "http://localhost" + ServerPort + "/services"

type registry struct {
	registrations []Registration
	mutex         *sync.RWMutex
}

var reg = registry{ //create variable in package level
	registrations: make([]Registration, 0),
	mutex:         &sync.RWMutex{}, //new(sync.Mutex)
}

func (r *registry) add(reg Registration) error {
	r.mutex.RLock()
	r.registrations = append(r.registrations, reg)
	r.mutex.RUnlock()
	if err := r.sendRequiredServices(reg); err != nil {
		return err
	}
	r.notify(patch{
		Added: []patchEntry {
			{
				Name: reg.ServiceName,
				URL: reg.ServiceURL,
			},
		},
	})
	return nil
}

func (r *registry) notify(fullpatch patch) {
	r.mutex.RLock()
	defer r.mutex.RUnlock() //大胆随便加，一个rw锁可以被多个goroutine持有

	for _, reg := range r.registrations { // 善用并发优化循环
		go func (reg Registration) {
			p := patch{Added: []patchEntry{}, Removed: []patchEntry{}}
			sendUpdate := false
			for _, reqService := range reg.RequiredServices {
				for _, added :=   range fullpatch.Added {
					if added.Name == reqService {
						p.Added = append(p.Added, added)
						sendUpdate = true
					}
				}
				for _, removed := range fullpatch.Removed {
					if removed.Name == reqService {
						p.Removed = append(p.Removed, removed)
						sendUpdate = true
					}
				}
			}
			if sendUpdate {
				if err := r.sendPatch(p, reg.ServiceUpdateURL); err != nil {
					log.Println(err)
					return
				}
			}
		}(reg)
	}
}

func (r *registry) sendRequiredServices(reg Registration) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	var p patch
	for _, serviceReg := range r.registrations {
		for _, reqService := range reg.RequiredServices {
			if serviceReg.ServiceName == reqService {
				p.Added = append(p.Added, patchEntry{
					Name: serviceReg.ServiceName,
					URL:  serviceReg.ServiceURL,
				})
			}
		}
	}
	err := r.sendPatch(p, reg.ServiceUpdateURL)
	if err != nil {
		return err
	}
	return nil
}

func (r *registry) sendPatch(p patch, url string) error {
	d, err := json.Marshal(p)
	if err != nil {
		return err
	}
	res, err := http.Post(url, "application/json", bytes.NewBuffer(d))
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed at send patch request to target service with response"+
			" status code: %v", res.StatusCode)
	}
	return nil
}

func (r *registry) remove(url string) error {
	for k, v := range r.registrations {
		if v.ServiceURL == url {
			r.notify(patch{
				Removed: []patchEntry{
					{
						Name: v.ServiceName,
						URL: v.ServiceURL,
					},
				},
			})
			r.mutex.Lock()
			r.registrations = append(r.registrations[:k], r.registrations[k+1:]...)
			r.mutex.Unlock()
			return nil
		}
	}
	return fmt.Errorf("service at URL %s not found", url)
}

func (r *registry) heartBeat(freq time.Duration) {
	for {
		var wg sync.WaitGroup 
		for _, reg := range r.registrations {
			wg.Add(1)
			go func (reg Registration) {
				defer wg.Done()
				success := true
				for attempts := 0; attempts < 3; attempts ++ {
					res, err := http.Get(reg.HeartBeatURL)
					if err != nil {
						log.Println(err)
					} else if res.StatusCode == http.StatusOK {
						log.Printf("HeartBeat check passed for %v", reg.ServiceName)
						if !success {
							r.add(reg)
						}
						break;
					}
					log.Printf("HeartBeat check failed for %v", reg.ServiceName)
					if success {
						success = false
						r.remove(reg.ServiceURL)
					}
					time.Sleep(time.Second)
				}
			}(reg)
			wg.Wait()
			time.Sleep(freq)
		}
	}
}

// only run once
var once sync.Once 

func SetupRegistryService() {
	once.Do(func() {
		go reg.heartBeat(3 * time.Second)
	})
}


type RegistryService struct{}

func (s RegistryService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received")
	switch r.Method {
	case http.MethodPost:
		dec := json.NewDecoder(r.Body)
		var r Registration
		if err := dec.Decode(&r); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Adding service: %v, at URL: %v\n", r.ServiceName, r.ServiceURL)
		if err := reg.add(r); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

	case http.MethodDelete:
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := reg.remove(string(payload)); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Removing service at URL: %v\n", string(payload))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
