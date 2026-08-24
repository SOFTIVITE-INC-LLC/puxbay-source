package websocket

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// ServeWs handles websocket requests from the peer.
func ServeWs(hub *Hub, c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// TenantID is populated by the auth middleware
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		log.Println("Unauthorized WS connection attempt")
		conn.Close()
		return
	}

	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), tenantID: tenantID.(uuid.UUID).String()}
	
	role, exists := c.Get("role")
	if exists && (role == "superadmin" || role == "admin") {
		client.isAdmin = true
	}

	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
