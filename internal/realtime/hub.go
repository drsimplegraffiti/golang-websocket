package realtime

import (
	"log"
	"sync"

	"golangchatapp/internal/models"
)

type Hub struct {
	Clients map[int64]map[*Client]struct{}
	mu      sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Clients: make(map[int64]map[*Client]struct{}),
	}
}

func (h *Hub) broadcastToAll(event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, conns := range h.Clients {
		for c := range conns {
			select {
			case c.Send <- event:
			default:
				log.Printf("warning: dropped event for client %d, channel full", c.User.ID)
			}
		}
	}
}

func (h *Hub) GetClients(userId int64) ([]*Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns, ok := h.Clients[userId]
	if !ok || len(conns) == 0 {
		return nil, false
	}

	clients := make([]*Client, 0, len(conns))
	for c := range conns {
		clients = append(clients, c)
	}

	return clients, true
}

func (h *Hub) SendEventToUserIds(userIds []int64, sendId int64, eventType EventType, payload map[string]any) {
	for _, id := range userIds {
		h.mu.RLock()
		conns, ok := h.Clients[id]
		h.mu.RUnlock()

		if !ok {
			continue
		}

		for c := range conns {
			c.SendEvent(Event{
				EventType: eventType,
				Payload:   payload,
			})
		}
	}
}

func (h *Hub) RegisterClientConnection(client *Client) {
	h.mu.Lock()
	conns, ok := h.Clients[client.User.ID]
	if !ok {
		conns = make(map[*Client]struct{})
		h.Clients[client.User.ID] = conns
	}
	conns[client] = struct{}{}
	firstConnection := len(conns) == 1
	h.mu.Unlock()

	if firstConnection {
		h.broadcastToAll(Event{
			EventType: EventUserOnline,
			Payload:   client.User.ToMap(),
		})

		go func() {
			privates, err := models.GetPrivatesForUser(client.User.ID)
			if err != nil {
				log.Println("failed to get privates:", err)
				return
			}

			for _, p := range privates {
				msgs, err := models.GetUndeliveredMessagesByPrivateID(p.ID)
				if err != nil {
					log.Println("failed to get undelivered messages for private:", p.ID, err)
					continue
				}

				for _, msg := range msgs {
					if msg.FromID == client.User.ID {
						continue
					}

					h.SendEventToUserIds([]int64{msg.FromID}, client.User.ID, EventDelivered, map[string]any{
						"message_id": msg.ID,
						"to_id":      client.User.ID,
					})
				}
			}
		}()
	}
}

func (h *Hub) UnregisterClientConnection(client *Client) {
	h.mu.Lock()
	conns, ok := h.Clients[client.User.ID]
	if !ok {
		h.mu.Unlock()
		return
	}

	delete(conns, client)
	noConnectionsLeft := len(conns) == 0
	if noConnectionsLeft {
		delete(h.Clients, client.User.ID)
	}
	h.mu.Unlock()

	if noConnectionsLeft {
		h.broadcastToAll(Event{
			EventType: EventUserOffline,
			Payload:   client.User.ToMap(),
		})
	}
}

func (h *Hub) SendCurrentClients(toClient *Client) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := []map[string]any{}
	seen := make(map[int64]struct{})

	for userId, conns := range h.Clients {
		if userId == toClient.User.ID {
			continue
		}

		_, ok := seen[userId]
		if ok {
			continue
		}

		for c := range conns {
			users = append(users, c.User.ToMap())
			seen[userId] = struct{}{}
			break
		}
	}

	toClient.Send <- Event{
		EventType: EventCurrentUsers,
		Payload:   users,
	}
}

func (h *Hub) SendError(clientId int64, message string) {
	clients, ok := h.GetClients(clientId)
	if !ok || len(clients) == 0 {
		return
	}

	for _, c := range clients {
		c.SendEvent(Event{
			EventType: EventError,
			Payload:   message,
		})
	}
}

func (h *Hub) Shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	log.Println("Shutting down Hub, notifying all clients...")

	for _, conns := range h.Clients {
		for c := range conns {
			c.SendEvent(Event{
				EventType: EventServerShutdown,
				Payload:   "Server is shutting down",
			})
			c.Close()
		}
	}

	h.Clients = make(map[int64]map[*Client]struct{})

	log.Println("Hub shutdown complete.")
}
