package services

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan string)
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	mu sync.Mutex
)

/******************************************************************************
* FUNCTION:        WebSocketHandler
*
* DESCRIPTION:     This function is used to create web socket handler
* INPUT:					 gin context
* RETURNS:         void
******************************************************************************/
func WebSocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	mu.Lock()
	clients[conn] = true
	mu.Unlock()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			mu.Lock()
			delete(clients, conn)
			mu.Unlock()
			conn.Close()
			break
		}
	}
}

/******************************************************************************
* FUNCTION:        BroadcastMessage
*
* DESCRIPTION:     Helper function to be called to brodcast
* INPUT:					 gin context
* RETURNS:         void
******************************************************************************/
func BroadcastMessage(message string) {
	mu.Lock()
	defer mu.Unlock()

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}
