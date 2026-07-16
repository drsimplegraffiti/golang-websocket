package realtime

type EventType string

// Events are used to communicate between the server and the client. They are
// sent over WebSocket connections and can be used to notify the client of various
// events, such as new messages, user status changes, etc.
const (
	// -------------------- WebSocket --------------------
	EventCurrentUsers EventType = "current_users"
	EventUserOnline   EventType = "online"
	EventUserOffline  EventType = "offline"
	EventNewPrivate   EventType = "new_private"
	EventMessage      EventType = "message"
	EventDelivered    EventType = "delivered"
	EventRead         EventType = "read"
	EventTyping       EventType = "typing"
	EventError        EventType = "error"
	EventHeartbeat    EventType = "heartbeat"
	// -------------------- Shutdown --------------------
	EventServerShutdown EventType = "shutdown"
)

type Event struct {
	EventType EventType `json:"event_type"`
	Payload   any       `json:"payload"`
}
