package realtime

import (
	"sync"

	"golangchatapp/internal/models"

	"github.com/coder/websocket"
)

// think of Client as a user session. It contains the user information, the
// websocket connection, and a channel to send events to the user. The once field
// is used to ensure that the Close method is only called once, even if it is
// called multiple times.
type Client struct {
	User *models.User `json:"user"` // we can put other object like Admin,
	// Moderator, etc. here if we want to send them to the client
	Conn *websocket.Conn `json:"-"` // - here means that this field will not be
	// serialized to JSON when the struct is converted to JSON. This is useful for
	// fields that are not relevant to the client or should not be exposed.
	Send chan Event `json:"-"`
	once sync.Once  `json:"-"`
}

// NewClient creates a new Client instance with the provided user and websocket
// connection. It initializes the Send channel with a buffer size of 512, allowing
// for up to 512 events to be sent to the user without blocking. This is useful for
// handling bursts of events without overwhelming the client or causing delays in
// event delivery.
// the standard events can range between 10-100 events per second, so a buffer
// size of 512 is a reasonable choice

// the standard practice in Go is to use a constructor function named
// NewTypeName to create a new instance of a struct. In this case, NewClient is the
// constructor function for the Client struct.
func NewClient(user *models.User, conn *websocket.Conn) *Client {
	return &Client{
		User: user,
		Conn: conn,
		Send: make(chan Event, 512), // 512 items can be send to a user
	}
}

func (c *Client) SendEvent(event Event) {
	select { // we use select here to avoid blocking the sending of events to
	// the client. If the Send channel is full, the event will be dropped and
	// not sent to the client. This is a common pattern in Go to prevent
	// blocking when sending to channels. It allows the program to continue
	// executing other code instead of waiting for the channel to be ready to
	// receive the event.

	// so channel and select are used together to implement non-blocking
	// communication between goroutines. In this case, we are using a select
	// statement to send an event to the client's Send channel. If the channel is
	// full, the default case will be executed, which means the event will be
	// dropped and not sent to the client.
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
