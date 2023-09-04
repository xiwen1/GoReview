package log

import (
	"io"
	stlog "log"
	"net/http"
	"os"
)

var log *stlog.Logger

type filelog string

//业务逻辑以及http请求处理逻辑

func(f filelog) Write(data []byte) (int, error) {
	file, err := os.OpenFile(string(f), os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0600)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return file.Write(data)
}

func Run(destination string) {
	log = stlog.New(filelog(destination), "[go]", stlog.LstdFlags)
}

func RegisterHandlers() {
	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			msg, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			write(string(msg))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func write(data string) {
	log.Printf("%v\n", data)
}

