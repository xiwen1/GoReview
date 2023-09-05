package log

import (
	"bytes"
	"encoding/json"
	stlog "log"
	"net/http"
)

func SetClientLogger(logURL string, serviceName string) {
	stlog.SetOutput(&clientLogger{url:logURL, name:serviceName})
	stlog.SetFlags(0)
}

type clientLogger struct {
	name string
	url string
}

func (c clientLogger) Write(data []byte) (int, error) {
	msg := logMessage{
		sender: c.name,
		msg: string(data),
	}
	msgJson, err := json.Marshal(msg)
	if err != nil {
		return 0, err
	}
	_, err = http.Post(c.url, "application/json", bytes.NewBuffer(msgJson))
	if err != nil {
		return 0, nil
	}
	return len(data), nil 
}

