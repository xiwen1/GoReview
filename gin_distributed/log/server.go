package log

import (
	stlog "log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var log stlog.Logger

type filelog string

func (f filelog) Write(data []byte) (int, error) {
	file, err := os.OpenFile("./distributed.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return file.Write(data)
}

func Run(destination string) {
	log = *stlog.New(filelog(destination), "[go] - ", stlog.LstdFlags)
}

type logMessage struct {
	sender string
	msg string
}

func RegisterHandlers(r *gin.Engine) {// engine is a super of group
	 r.POST("/log", func (c *gin.Context) {
		var msg logMessage
		if err := c.ShouldBindJSON(&msg); err != nil {
			stlog.Println("Failed receiving message")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		write(msg)
	 }) 
}

func write(msg logMessage) {
	log.Printf("[%v] - %v", msg.sender, msg.msg)
}
