package routes

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"golangchatapp/internal/middlewares"
	"golangchatapp/internal/models"
	"golangchatapp/internal/realtime"
	"golangchatapp/internal/utils"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func handleWebSocket(hub *realtime.Hub, w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get(middlewares.CtxAuthorization)
	if authHeader == "" || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		utils.JSON(w, http.StatusUnauthorized, false, "unauthorised", nil)
		return
	}

	accessToken := strings.TrimSpace(authHeader[7:])

	userId, _, _, err := utils.VerifyJWT(accessToken)
	if err != nil {
		utils.JSON(w, http.StatusUnauthorized, false, "unauthorised", nil)
		return
	}

	user, err := models.GetUserById(userId)
	if err != nil {
		utils.JSON(w, http.StatusUnauthorized, false, "unauthorised", nil)
		return
	}

	// this is where we upgrade the HTTP connection to a WebSocket connection.
	// The AcceptOptions struct allows us to specify options for the WebSocket
	// connection, such as allowed origin patterns. In this case, we are allowing
	// connections from any origin by using the wildcard "*".
	opts := &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	}

	conn, err := websocket.Accept(w, r, opts)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, false, "failed to upgrade to websocket", nil)
		return
	}

	// this is where we create a new client instance for the connected user. The
	// NewClient function initializes a new Client struct with the user's
	// information and the WebSocket connection. It also creates a buffered channel
	// for sending events to the client.
	client := realtime.NewClient(user, conn)

	hub.RegisterClientConnection(client)
	hub.SendCurrentClients(client)

	defer func() {
		hub.UnregisterClientConnection(client)
		client.Close()
	}()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// sends requestes periodically to front end
	go heartbeat(ctx, client)
	go writePump(ctx, client)
	readPump(ctx, cancel, hub, client)
}

// Heartbeat function is used to send a heartbeat event to the client every 30
// seconds. This is done to keep the connection alive and to detect if the client
// has disconnected. If the ping fails, the client is disconnected.
func heartbeat(ctx context.Context, client *realtime.Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C: // C is a channel that receives the current time every
			// 30 seconds. When the ticker ticks, we send a heartbeat event to the
			//    client.
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second) //
			// ensures that the ping does not block indefinitely. If the ping takes
			// longer than 5 seconds, it will be cancelled.
			err := client.Conn.Ping(pingCtx) // pingCtx is used to set a timeout
			// for the ping operation. If the ping takes longer than 5 seconds, it
			// will be cancelled and an error will be returned.client
			if err != nil {
				log.Println("Ping failed, disconnecting client.")
				cancel()
				return
			}
			cancel()

			// this is where we send the heartbeat event to the client. The
			// Event struct contains the event type and payload. In this case, the
			// event type is realtime.EventHeartbeat and the payload is nil. The
			// SendEvent method sends the event to the client's Send channel, which
			// is then picked up by the writePump goroutine and sent to the client over the WebSocket connection.
			client.Send <- realtime.Event{
				EventType: realtime.EventHeartbeat,
				Payload:   nil,
			}
		}
	}
}

func writePump(ctx context.Context, client *realtime.Client) {
	for {
		select {
		case <-ctx.Done():
			return

		case event, ok := <-client.Send: // we use a select statement to listen
			// for events on the client's Send channel. When an event is received, we
			// write it to the WebSocket connection using the wsjson.Write function. If
			// the context is done (e.g., the client has disconnected), we exit the
			// loop and return.
			if !ok {
				return
			}

			writeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			_ = wsjson.Write(writeCtx, client.Conn, event)
			cancel()
		}
	}
}

func readPump(ctx context.Context, cancel context.CancelFunc, hub *realtime.Hub, client *realtime.Client) {
	defer cancel()
	defer func() {
		r := recover()
		if r != nil {
			log.Printf("Recovered from panic in readPump for client %d: %v", client.User.ID, r)
		}
	}()

	// // this is like a while loop that continuously reads events from the
	// WebSocket connection. The select statement is used to listen for the
	// context being done (e.g., the client has disconnected) and to read events
	// from the WebSocket connection. When an event is received, it is passed to
	// the handleIncomingEvent function for processing.
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var event realtime.Event
		err := wsjson.Read(ctx, client.Conn, &event)
		if err != nil {
			return
		}

		handleIncomingEvent(hub, client, event)
	}
}

func handleIncomingEvent(hub *realtime.Hub, client *realtime.Client, event realtime.Event) {
	payload, ok := event.Payload.(map[string]any)
	if !ok {
		hub.SendError(client.User.ID, "invalid event payload format")
		return
	}

	switch event.EventType {
	case realtime.EventMessage:
		privateIdAny, ok := payload["private_id"]
		if !ok {
			hub.SendError(client.User.ID, "private id is mising")
			return
		}

		privateIdFloat, ok := privateIdAny.(float64)
		if !ok {
			hub.SendError(client.User.ID, "private id must be a number")
			return
		}

		privateId := int64(privateIdFloat) // here we convert the float64 to
		// int64, which is the expected type for privateId.

		receiverIdAny, ok := payload["receiver_id"]
		if !ok {
			hub.SendError(client.User.ID, "receiver id is mising")
			return
		}

		receiverIdFloat, ok := receiverIdAny.(float64)
		if !ok {
			hub.SendError(client.User.ID, "receiver id must be a number")
			return
		}

		receiverId := int64(receiverIdFloat)

		messageTypeAny, ok := payload["message_type"]
		if !ok {
			hub.SendError(client.User.ID, "message_type is missing")
			return
		}

		messageType, ok := messageTypeAny.(string)
		if !ok {
			hub.SendError(client.User.ID, "message_type must be a string")
			return
		}

		msgBytes, _ := json.Marshal(payload) // json.Marshal is used to convert
		// the payload map into a JSON byte slice. This is necessary because the
		// next step involves unmarshalling this JSON into a Message struct.
		var msg models.Message
		err := json.Unmarshal(msgBytes, &msg)
		if err != nil {
			hub.SendError(client.User.ID, "invalid message format.")
			return
		}

		msg.FromID = client.User.ID
		msg.PrivateID = privateId
		msg.MessageType = messageType
		msg.CreatedAt = time.Now()

		err = models.CreateMessage(&msg)
		if err != nil {
			hub.SendError(client.User.ID, "failed to save message")
			return
		}

		hub.SendEventToUserIds([]int64{msg.FromID, receiverId}, msg.FromID, realtime.EventMessage, map[string]any{
			"message": msg,
		})

	case realtime.EventDelivered:
		msgIdAny, ok := payload["message_id"]
		if !ok {
			hub.SendError(client.User.ID, "message id is mising")
			return
		}

		msgIdFloat, ok := msgIdAny.(float64)
		if !ok {
			hub.SendError(client.User.ID, "message id must be a number")
			return
		}

		msgId := int64(msgIdFloat)

		msg, err := models.GetMessageByID(msgId)
		if err != nil {
			hub.SendError(client.User.ID, "message not found")
			return
		}

		if msg.FromID == client.User.ID {
			hub.SendError(client.User.ID, "cannot mark own message as delivered")
			return
		}

		err = models.MarkMessageDelivered(msgId)
		if err != nil {
			hub.SendError(client.User.ID, "failed to mark message as delivered")
			return
		}

		hub.SendEventToUserIds([]int64{msg.FromID}, client.User.ID, realtime.EventDelivered, map[string]any{
			"message_id": msgId,
			"to_id":      client.User.ID,
		})

	case realtime.EventRead:
		msgIdAny, ok := payload["message_id"]
		if !ok {
			hub.SendError(client.User.ID, "message id is mising")
			return
		}

		msgIdFloat, ok := msgIdAny.(float64)
		if !ok {
			hub.SendError(client.User.ID, "message id must be a number")
			return
		}

		msgId := int64(msgIdFloat)

		msg, err := models.GetMessageByID(msgId)
		if err != nil {
			hub.SendError(client.User.ID, "message not found")
			return
		}

		if msg.FromID == client.User.ID {
			hub.SendError(client.User.ID, "cannot mark own message as read")
			return
		}

		err = models.MarkMessageRead(msgId)
		if err != nil {
			hub.SendError(client.User.ID, "failed to mark message as read")
			return
		}

		hub.SendEventToUserIds([]int64{msg.FromID}, client.User.ID, realtime.EventRead, map[string]any{
			"message_id": msgId,
		})

	case realtime.EventTyping:
		privateIdAny, ok := payload["private_id"]
		if !ok {
			hub.SendError(client.User.ID, "private id is mising")
			return
		}

		privateIdFloat, ok := privateIdAny.(float64)
		if !ok {
			hub.SendError(client.User.ID, "private id must be a number")
			return
		}

		privateId := int64(privateIdFloat)

		receiverIdAny, ok := payload["receiver_id"]
		if !ok {
			hub.SendError(client.User.ID, "receiver id is mising")
			return
		}

		receiverIdFloat, ok := receiverIdAny.(float64)
		if !ok {
			hub.SendError(client.User.ID, "receiver id must be a number")
			return
		}

		receiverId := int64(receiverIdFloat)

		isTypingAny, ok := payload["is_typing"]
		if !ok {
			hub.SendError(client.User.ID, "is typing is missing")
			return
		}

		isTyping := isTypingAny.(bool)

		hub.SendEventToUserIds([]int64{receiverId}, client.User.ID, realtime.EventTyping, map[string]any{
			"private_id": privateId,
			"user_id":    client.User.ID,
			"is_typing":  isTyping,
		})

	default:
		hub.SendError(client.User.ID, "unknown event type: "+string(event.EventType))
	}
}
