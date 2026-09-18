package handlers

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"live-polling-backend/db"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Polling page is meant to be embedded/shared publicly; allow all origins.
	// Tighten this with an allow-list in production if needed.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// hub tracks connected clients per poll so a single Redis subscription can
// fan out to many browser tabs watching the same poll.
type pollHub struct {
	mu      sync.Mutex
	clients map[string]map[*websocket.Conn]bool
}

var hub = &pollHub{clients: make(map[string]map[*websocket.Conn]bool)}

func (h *pollHub) add(pollID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[pollID] == nil {
		h.clients[pollID] = make(map[*websocket.Conn]bool)
	}
	h.clients[pollID][conn] = true
}

func (h *pollHub) remove(pollID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.clients[pollID]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.clients, pollID)
		}
	}
}

func (h *pollHub) broadcast(pollID string, message []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.clients[pollID] {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			conn.Close()
			delete(h.clients[pollID], conn)
		}
	}
}

// subscribedPolls tracks which polls already have an active Redis
// subscriber goroutine, so we don't spawn duplicates when multiple
// browsers watch the same poll.
var (
	subMu           sync.Mutex
	subscribedPolls = make(map[string]bool)
)

// PollWebSocket handles GET /ws/polls/:id. Each new connection registers
// itself with the hub; the first connection for a given poll also starts a
// Redis Pub/Sub subscriber goroutine that relays messages to the hub.
func PollWebSocket(c *gin.Context) {
	pollID := c.Param("id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	hub.add(pollID, conn)
	defer hub.remove(pollID, conn)

	ensureSubscriber(pollID)

	// Keep the connection open; we don't expect inbound messages from the
	// client, but we must read to detect disconnects and respond to pings.
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

// ensureSubscriber starts exactly one goroutine per poll that listens on
// the poll's Redis Pub/Sub channel and forwards messages to all WebSocket
// clients currently watching that poll via the hub.
func ensureSubscriber(pollID string) {
	subMu.Lock()
	defer subMu.Unlock()

	if subscribedPolls[pollID] {
		return
	}
	subscribedPolls[pollID] = true

	go func() {
		ctx := context.Background()
		pubsub := db.RedisClient.Subscribe(ctx, db.PollUpdatesChannel(pollID))
		defer pubsub.Close()

		ch := pubsub.Channel()
		for msg := range ch {
			hub.broadcast(pollID, []byte(msg.Payload))
		}

		subMu.Lock()
		delete(subscribedPolls, pollID)
		subMu.Unlock()
	}()
}
