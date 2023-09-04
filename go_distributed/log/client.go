package log

import (
	"bytes"
	"distributed/registry"
	"fmt"
	stlog "log"
	"net/http"
)

func SetClientLogger(url string, name registry.ServiceName) {
	stlog.SetPrefix(fmt.Sprintf("[%v] - ", name))
	stlog.SetFlags(0)
	stlog.SetOutput(&clientLogger{url: url})
} 

type clientLogger struct {
	url string
}

func(c *clientLogger) Write(data []byte) (int, error) {
	b := bytes.NewBuffer([]byte(data))
	res, err := http.Post(c.url, "text/plain", b)
	if err != nil {
		return 0, err 
	}
	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to send log message, service responded with code: %v", res.StatusCode)
	}
	return len(data), nil
}