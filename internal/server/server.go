package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"sir-john-shell/internal/acp"
	"sir-john-shell/web"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Server holds the ACP client and WebSocket hub.
type Server struct {
	addr    string
	cwd     string
	acp     *acp.Client
	session string
	hub     *Hub
}

// Hub keeps track of connected WebSocket clients.
// Only one client (active) is allowed at a time to avoid duplicate UIs.
type Hub struct {
	mu       sync.Mutex
	clients  map[*Client]bool
	sessions map[string]map[*Client]bool
	active   *Client
}

// Client is a WebSocket connection.
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	sessionID string
	send      chan []byte
	closeOnce sync.Once
}

// New creates a server.
func New(addr, cwd string, acpClient *acp.Client, sessionID string) *Server {
	h := &Hub{
		clients:  make(map[*Client]bool),
		sessions: make(map[string]map[*Client]bool),
	}
	return &Server{
		addr:    addr,
		cwd:     cwd,
		acp:     acpClient,
		session: sessionID,
		hub:     h,
	}
}

// Run starts the HTTP server and the ACP notification forwarder.
func (s *Server) Run() error {
	fs, err := staticFS()
	if err != nil {
		return fmt.Errorf("static fs: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(fs))
	mux.HandleFunc("/ws", s.handleWebSocket)

	go s.forwardACP()

	srv := &http.Server{
		Addr:              s.addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("sir-john-shell listening on %s", s.addr)
	return srv.ListenAndServe()
}

// staticFS returns the embedded front-end filesystem.
func staticFS() (http.FileSystem, error) {
	if path := os.Getenv("SIR_JOHN_SHELL_STATIC"); path != "" {
		return http.Dir(path), nil
	}
	return web.StaticFS()
}

// forwardACP reads ACP notifications and broadcasts them to WebSocket clients.
func (s *Server) forwardACP() {
	for msg := range s.acp.Notifications() {
		data, err := json.Marshal(msg)
		if err != nil {
			log.Printf("marshal notification: %v", err)
			continue
		}
		s.hub.broadcast(s.session, data)
	}
}

// handleWebSocket upgrades the HTTP connection and starts read/write pumps.
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade: %v", err)
		return
	}

	client := &Client{
		hub:       s.hub,
		conn:      conn,
		sessionID: s.session,
		send:      make(chan []byte, 256),
	}
	if !s.hub.register(client) {
		return
	}

	// Send welcome message.
	welcome, _ := json.Marshal(map[string]any{
		"type":      "welcome",
		"sessionId": s.session,
		"cwd":       s.cwd,
	})
	client.send <- welcome

	go client.writePump()
	client.readPump(s)
}

// readPump reads messages from the WebSocket client.
func (c *Client) readPump(s *Server) {
	defer func() {
		s.hub.unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(64 * 1024)
	c.conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(15 * time.Second))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket read error: %v", err)
			}
			break
		}

		var msg map[string]any
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("websocket json: %v", err)
			continue
		}

		msgType, _ := msg["type"].(string)
		switch msgType {
		case "prompt":
			text, _ := msg["text"].(string)
			if text != "" {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				if err := s.acp.Prompt(ctx, c.sessionID, text); err != nil {
					log.Printf("prompt: %v", err)
				}
				cancel()
			}
		case "cancel":
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := s.acp.Cancel(ctx, c.sessionID); err != nil {
				log.Printf("cancel: %v", err)
			}
			cancel()
		default:
			log.Printf("unknown ws message type: %s", msgType)
		}
	}
}

// writePump sends messages to the WebSocket client.
func (c *Client) writePump() {
	ticker := time.NewTicker(10 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("websocket write: %v", err)
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

func (c *Client) closeSend() { c.closeOnce.Do(func() { close(c.send) }) }

// register adds a client to the hub. Only the first client becomes active;
// any later client is rejected with "session in use".
func (h *Hub) register(c *Client) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.active != nil {
		closeMsg := websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "session in use")
		_ = c.conn.WriteMessage(websocket.CloseMessage, closeMsg)
		c.closeSend()
		_ = c.conn.Close()
		return false
	}
	h.clients[c] = true
	if h.sessions[c.sessionID] == nil {
		h.sessions[c.sessionID] = make(map[*Client]bool)
	}
	h.sessions[c.sessionID][c] = true
	h.active = c
	return true
}

// unregister removes a client from the hub.
func (h *Hub) unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, c)
	if set := h.sessions[c.sessionID]; set != nil {
		delete(set, c)
		if len(set) == 0 {
			delete(h.sessions, c.sessionID)
		}
	}
	if h.active == c {
		h.active = nil
	}
	c.closeSend()
}

// broadcast sends a message to the active client for the given session.
func (h *Hub) broadcast(sessionID string, msg []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.active == nil || h.active.sessionID != sessionID {
		return
	}
	select {
	case h.active.send <- msg:
	default:
		c := h.active
		h.active = nil
		delete(h.clients, c)
		delete(h.sessions[sessionID], c)
		c.closeSend()
	}
}
