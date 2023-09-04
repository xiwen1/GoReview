package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func RegisterService(r Registration) error {
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
		return fmt.Errorf("failed to register service. Registry service " + 
		"response with code: %v", res.StatusCode)
	}

	return nil
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
		return fmt.Errorf("failed to deregister service. Registry service" + 
			"service responded with code: %v", res.StatusCode)
	}
	return nil
}