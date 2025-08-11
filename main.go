package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// upgrader is used to upgrade an HTTP connection to a WebSocket connection.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections
	},
}

// hub manages all active WebSocket connections.
type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.Mutex
}

// newHub creates a new Hub.
func newHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// run starts the hub's event loop.
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
					// Client is likely disconnected, remove them.
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

func main() {
	hub := newHub()
	go hub.run()

	// 1. Serve the static web files (HTML, CSS, JS)
	http.Handle("/", http.FileServer(http.Dir("./web")))

	// 2. Handle WebSocket connections
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}
		hub.register <- conn
		// We don't read from the client in this simple case, but a real app might.
		// When the client disconnects, the WriteMessage in the hub will fail,
		// and the client will be unregistered.
	})

	// 3. Handle file uploads
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		file, _, err := r.FormFile("logfile")
		if err != nil {
			http.Error(w, "Error retrieving the file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		log.Println("File uploaded successfully. Starting to process...")

		// Process the file in a separate goroutine so the HTTP request can return immediately.
		go func() {
			scanner := bufio.NewScanner(file)
			linesProcessed := 0
			for scanner.Scan() {
				line := scanner.Bytes()
				var entry map[string]interface{}
				// Check if it's valid JSON before broadcasting
				if err := json.Unmarshal(line, &entry); err == nil {
					hub.broadcast <- line
				}
				linesProcessed++
			}
			if err := scanner.Err(); err != nil {
				log.Println("Error reading uploaded file:", err)
			}
			log.Printf("Finished processing %d lines.\n", linesProcessed)
		}()

		// Respond to the client immediately
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "File upload received. Processing...",
		})
	})

	fmt.Println("Starting LogLens server on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}