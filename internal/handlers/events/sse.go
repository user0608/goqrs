package events

import (
	"encoding/json"
	"goqrs/security"
	"log"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/r3labs/sse/v2"
)

var server *sse.Server
var one sync.Once

type EventData struct {
	EventName string
	Data      interface{}
}

func init() {
	one.Do(func() {
		server = sse.New()
		server.AutoStream = true
		server.AutoReplay = false
	})
}

func Handler(c echo.Context) error {
	qrs := c.Request().URL.Query()
	qrs.Set("stream", security.UserName(c.Request().Context())) //adapter
	c.Request().URL.RawQuery = qrs.Encode()
	server.ServeHTTP(c.Response().Writer, c.Request())
	return nil
}

type EventPublicher interface {
	Publish(username string, event EventData)
}
type publicher struct{}

func NewEventPublicher() EventPublicher {
	return &publicher{}
}
func (*publicher) Publish(username string, event EventData) {
	var data []byte
	if event.Data != nil {
		var err error
		data, err = json.Marshal(event.Data)
		if err != nil {
			log.Println("Err: events.Publish.JsonMarshal:", err)
		}
	}
	server.Publish(username, &sse.Event{
		Event: []byte(event.EventName),
		Data:  data,
	})
}
