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

	opts := &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	}

	conn, err := websocket.Accept(w, r, opts)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, false, "failed to upgrade to websocket", nil)
		return
	}

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

func heartbeat(ctx context.Context, client *realtime.Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := client.Conn.Ping(pingCtx)
			if err != nil {
				log.Println("Ping failed, disconnecting client.")
				cancel()
				return
			}
			cancel()

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

		case event, ok := <-client.Send:
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

		msgBytes, _ := json.Marshal(payload)
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
