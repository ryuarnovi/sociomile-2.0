package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	amqp "github.com/rabbitmq/amqp091-go"

	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

type WebsocketHandler struct {
	rabbitConn *amqp.Connection
	jwtSecret  string
	hub        *wsHub
}

func NewWebsocketHandler(rabbitConn *amqp.Connection, jwtSecret string) *WebsocketHandler {
	h := &WebsocketHandler{
		rabbitConn: rabbitConn,
		jwtSecret:  jwtSecret,
		hub:        newHub(),
	}
	// start rabbit consumer broadcaster if connection present
	if rabbitConn != nil {
		go h.startRabbitConsumer(context.Background())
	}
	return h
}

// wsHub holds clients and broadcasts messages
type wsHub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]string // conn -> tenantID
}

func newHub() *wsHub {
	return &wsHub{clients: make(map[*websocket.Conn]string)}
}

func (h *wsHub) add(conn *websocket.Conn, tenant string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[conn] = tenant
}

func (h *wsHub) remove(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, conn)
	conn.Close()
}

func (h *wsHub) broadcast(msg []byte, tenant string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c, t := range h.clients {
		if tenant != "" && t != tenant {
			continue
		}
		// write in goroutine to avoid blocking
		go func(conn *websocket.Conn) {
			conn.SetWriteDeadline(timeNowAdd())
			_ = conn.WriteMessage(websocket.TextMessage, msg)
		}(c)
	}
}

// helper to provide a short write deadline duration
func timeNowAdd() (t time.Time) {
	return time.Now().Add(10 * time.Second)
}

// startRabbitConsumer consumes from conversation.events and broadcasts to clients
func (w *WebsocketHandler) startRabbitConsumer(ctx context.Context) {
	ch, err := w.rabbitConn.Channel()
	if err != nil {
		log.Printf("ws: failed to open rabbit channel: %v", err)
		return
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		log.Printf("ws: failed to declare queue: %v", err)
		return
	}

	// bind to conversation.events
	if err := ch.QueueBind(q.Name, "#", "conversation.events", false, nil); err != nil {
		log.Printf("ws: failed to bind queue: %v", err)
		return
	}

	msgs, err := ch.Consume(q.Name, "", true, true, false, false, nil)
	if err != nil {
		log.Printf("ws: failed to consume: %v", err)
		return
	}

	for d := range msgs {
		// try to extract tenant_id from message body
		var payload map[string]interface{}
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			// broadcast raw body
			w.hub.broadcast(d.Body, "")
			continue
		}
		tenant := ""
		if v, ok := payload["tenant_id"]; ok {
			if s, ok := v.(string); ok {
				tenant = s
			}
		}
		w.hub.broadcast(d.Body, tenant)
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Handle upgrades and register clients
func (w *WebsocketHandler) Handle(c *gin.Context) {
	// Allow token via query param or Authorization header
	tokenStr := c.Query("token")
	if tokenStr == "" {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	claims := &middleware.Claims{}
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "token required"})
		return
	}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(w.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "invalid token"})
		return
	}

	// upgrade
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("ws: upgrade failed: %v", err)
		return
	}

	tenant := claims.TenantID
	w.hub.add(conn, tenant)

	// Listen for client close/read messages
	go func() {
		defer w.hub.remove(conn)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	}()

	// keep connection open
}
