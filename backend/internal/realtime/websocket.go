package realtime

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// EventType represents the type of real-time event
type EventType string

const (
	EventDownloadStarted   EventType = "download.started"
	EventDownloadProgress  EventType = "download.progress"
	EventDownloadCompleted EventType = "download.completed"
	EventDownloadFailed    EventType = "download.failed"
	EventImportStarted     EventType = "import.started"
	EventImportCompleted   EventType = "import.completed"
	EventImportFailed      EventType = "import.failed"
	EventBookAdded         EventType = "book.added"
	EventBookUpdated       EventType = "book.updated"
	EventBookDeleted       EventType = "book.deleted"
	EventScanStarted       EventType = "scan.started"
	EventScanCompleted     EventType = "scan.completed"
	EventSystemStatus      EventType = "system.status"
)

// Event represents a real-time event
type Event struct {
	Type      EventType   `json:"type"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// Client represents a WebSocket client
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	userID   uint
	isAdmin  bool
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("WebSocket client connected, total: %d", len(h.clients))

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mutex.Unlock()
			log.Printf("WebSocket client disconnected, total: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// Broadcast sends an event to all connected clients
func (h *Hub) Broadcast(event Event) {
	event.Timestamp = time.Now().Unix()
	
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return
	}

	h.broadcast <- data
}

// BroadcastToUser sends an event to a specific user
func (h *Hub) BroadcastToUser(userID uint, event Event) {
	event.Timestamp = time.Now().Unix()
	
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for client := range h.clients {
		if client.userID == userID {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}

// ClientCount returns the number of connected clients
func (h *Hub) ClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// WebSocketHandler handles WebSocket connections
func (h *Hub) WebSocketHandler(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	// Get user info from context (if authenticated)
	var userID uint
	var isAdmin bool
	if id, ok := c.Get("userId").(uint); ok {
		userID = id
	}
	if admin, ok := c.Get("isAdmin").(bool); ok {
		isAdmin = admin
	}

	client := &Client{
		hub:     h,
		conn:    conn,
		send:    make(chan []byte, 256),
		userID:  userID,
		isAdmin: isAdmin,
	}

	h.register <- client

	// Start client goroutines
	go client.writePump()
	go client.readPump()

	return nil
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming messages (e.g., subscription requests)
		c.handleMessage(message)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Batch any queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *Client) handleMessage(message []byte) {
	// Parse message
	var msg struct {
		Type string `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(message, &msg); err != nil {
		return
	}

	// Handle different message types
	switch msg.Type {
	case "ping":
		// Respond with pong
		response, _ := json.Marshal(map[string]string{"type": "pong"})
		c.send <- response

	case "subscribe":
		// Handle subscription to specific events
		// (Could implement filtering here)

	case "unsubscribe":
		// Handle unsubscription
	}
}

// EventEmitter provides a convenient way to emit events
type EventEmitter struct {
	hub *Hub
}

// NewEventEmitter creates a new event emitter
func NewEventEmitter(hub *Hub) *EventEmitter {
	return &EventEmitter{hub: hub}
}

// DownloadStarted emits a download started event
func (e *EventEmitter) DownloadStarted(bookID uint, title string) {
	e.hub.Broadcast(Event{
		Type: EventDownloadStarted,
		Data: map[string]interface{}{
			"bookId": bookID,
			"title":  title,
		},
	})
}

// DownloadProgress emits a download progress event
func (e *EventEmitter) DownloadProgress(bookID uint, progress float64, speed int64) {
	e.hub.Broadcast(Event{
		Type: EventDownloadProgress,
		Data: map[string]interface{}{
			"bookId":   bookID,
			"progress": progress,
			"speed":    speed,
		},
	})
}

// DownloadCompleted emits a download completed event
func (e *EventEmitter) DownloadCompleted(bookID uint, title string, path string) {
	e.hub.Broadcast(Event{
		Type: EventDownloadCompleted,
		Data: map[string]interface{}{
			"bookId": bookID,
			"title":  title,
			"path":   path,
		},
	})
}

// DownloadFailed emits a download failed event
func (e *EventEmitter) DownloadFailed(bookID uint, title string, error string) {
	e.hub.Broadcast(Event{
		Type: EventDownloadFailed,
		Data: map[string]interface{}{
			"bookId": bookID,
			"title":  title,
			"error":  error,
		},
	})
}

// ImportCompleted emits an import completed event
func (e *EventEmitter) ImportCompleted(bookID uint, mediaFileID uint, path string) {
	e.hub.Broadcast(Event{
		Type: EventImportCompleted,
		Data: map[string]interface{}{
			"bookId":      bookID,
			"mediaFileId": mediaFileID,
			"path":        path,
		},
	})
}

// BookAdded emits a book added event
func (e *EventEmitter) BookAdded(bookID uint, title string, authorName string) {
	e.hub.Broadcast(Event{
		Type: EventBookAdded,
		Data: map[string]interface{}{
			"bookId":     bookID,
			"title":      title,
			"authorName": authorName,
		},
	})
}

// BookUpdated emits a book updated event
func (e *EventEmitter) BookUpdated(bookID uint, changes map[string]interface{}) {
	data := map[string]interface{}{
		"bookId":  bookID,
		"changes": changes,
	}
	e.hub.Broadcast(Event{
		Type: EventBookUpdated,
		Data: data,
	})
}

// ScanCompleted emits a scan completed event
func (e *EventEmitter) ScanCompleted(filesFound int, booksImported int) {
	e.hub.Broadcast(Event{
		Type: EventScanCompleted,
		Data: map[string]interface{}{
			"filesFound":    filesFound,
			"booksImported": booksImported,
		},
	})
}

// SystemStatus emits a system status event
func (e *EventEmitter) SystemStatus(status map[string]interface{}) {
	e.hub.Broadcast(Event{
		Type: EventSystemStatus,
		Data: status,
	})
}

