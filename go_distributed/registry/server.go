package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

const ServerPort = ":3000"
const ServicesURL = "http://localhost" + ServerPort + "/services"

type registry struct {
	registrations []Registration
	mutex         *sync.Mutex
}

var reg = registry{ //create variable in package level
	registrations: make([]Registration, 0),
	mutex: &sync.Mutex{}, //new(sync.Mutex)
}

func(r *registry) add(reg Registration) error {
	r.mutex.Lock()
	r.registrations = append(r.registrations, reg)
	r.mutex.Unlock()
	return nil
}

func(r *registry) remove(url string) error {
	for k, v := range r.registrations {
		if v.ServiceURL == url {
			r.mutex.Lock()
			r.registrations = append(r.registrations[:k], r.registrations[k+1:]...)
			r.mutex.Unlock()
			return nil
		}
	}
	return fmt.Errorf("service at URL %s not found", url)
}

type RegistryService struct {}

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