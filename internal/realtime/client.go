package realtime

import (
	"sync"

	"golangchatapp/internal/models"

	"github.com/coder/websocket"
)

// a client is a user connected to websocket
type Client struct {
	User *models.User    `json:"user"`
	Conn *websocket.Conn `json:"-"`
	Send chan Event      `json:"-"`
	once sync.Once       `json:"-"`
}

func NewClient(user *models.User, conn *websocket.Conn) *Client {
	return &Client{
		User: user,
		Conn: conn,
		Send: make(chan Event, 512), // 512 items can be send to a user
	}
}

func (c *Client) SendEvent(event Event) {
	select {
	case c.Send <- event:
	default:
	}
}

func (c *Client) Close() {
	c.once.Do(func() {
		if c.Conn != nil {
			_ = c.Conn.Close(websocket.StatusNormalClosure, "Closing connection")
		}
		close(c.Send)
	})
}
