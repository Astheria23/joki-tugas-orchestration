package pipeline

import (
	"net/http"
	"sync"

	"github.com/Astheria23/jokiOrchestrator/shared/logging"
	"github.com/Astheria23/jokiOrchestrator/shared/models"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// Hub manages client subscriptions to task progress updates over WebSocket.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*websocket.Conn]bool
}

// GlobalHub is the singleton instance of the connection hub for tasks.
var GlobalHub = &Hub{
	clients: make(map[string]map[*websocket.Conn]bool),
}

// ConversationHub broadcasts new chat messages for a conversation.
var ConversationHub = &Hub{
	clients: make(map[string]map[*websocket.Conn]bool),
}

// ChatWSPayload is pushed to conversation subscribers.
type ChatWSPayload struct {
	Type    string         `json:"type"`
	Message models.Message `json:"message"`
}

// Register adds a client connection to the list of subscribers for a task ID.
func (h *Hub) Register(taskID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[taskID]; !ok {
		h.clients[taskID] = make(map[*websocket.Conn]bool)
	}
	h.clients[taskID][conn] = true
	logging.Log.Info().Str("taskId", taskID).Msg("Client registered to task WebSocket updates")
}

// Unregister removes a client connection and closes it.
func (h *Hub) Unregister(taskID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.clients[taskID]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.clients, taskID)
		}
	}
	_ = conn.Close()
	logging.Log.Info().Str("taskId", taskID).Msg("Client unregistered from task WebSocket updates")
}

// Broadcast sends a message to all active WebSocket connections for a given task.
func (h *Hub) Broadcast(taskID string, msg interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conns, ok := h.clients[taskID]
	if !ok {
		return
	}
	for conn := range conns {
		// Run writes in Goroutines to prevent blocking other clients
		go func(c *websocket.Conn) {
			err := c.WriteJSON(msg)
			if err != nil {
				h.Unregister(taskID, c)
			}
		}(conn)
	}
}

// HandleWS upgrades HTTP connection and registers the client to the task WebSocket updates channel.
func HandleWS(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id is required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logging.Log.Error().Err(err).Msg("Failed to upgrade connection to WebSocket")
		return
	}

	GlobalHub.Register(taskID, conn)

	// Hold connection open until client disconnects
	go func() {
		defer GlobalHub.Unregister(taskID, conn)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}

// HandleConversationWS streams new messages for a conversation.
func HandleConversationWS(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation id is required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logging.Log.Error().Err(err).Msg("Failed to upgrade conversation WebSocket")
		return
	}

	ConversationHub.Register(id, conn)
	go func() {
		defer ConversationHub.Unregister(id, conn)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}
