package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	clients   = make(map[string]*websocket.Conn)
	broadcast = make(chan string)
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	mu sync.RWMutex
)

type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

/******************************************************************************
* FUNCTION:        WebSocketHandler
*
* DESCRIPTION:     This function is used to create web socket handler
* INPUT:					 gin context
* RETURNS:         void
******************************************************************************/
func WebSocketHandler(c *gin.Context) {
	var (
		userId string
	)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}

	val, _ := c.Get("user_id")
	if val != nil {
		userId = val.(string)
	}

	mu.Lock()
	clients[userId] = conn
	mu.Unlock()

	conn.SetPongHandler(func(appData string) error {
		fmt.Printf("Received Pong from user %s\n", userId)
		return nil
	})

	go func() {
		for {
			err := conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				mu.Lock()
				delete(clients, userId)
				mu.Unlock()
				conn.Close()
				break
			}
			time.Sleep(10 * time.Second)
		}
	}()
}

/******************************************************************************
* FUNCTION:        BroadcastMessage
*
* DESCRIPTION:     Helper function to be called to brodcast
* INPUT:					 gin context
* RETURNS:         void
******************************************************************************/
func BroadcastMessage(message interface{}, mssgType, userId string) {
	mu.RLock()
	conn, exists := clients[userId]
	mu.RUnlock()

	if !exists {
		fmt.Printf("User %s not connected\n", userId)
		return
	}

	jsonMssg, _ := json.Marshal(WebSocketMessage{
		Type: mssgType,
		Data: message,
	})

	err := conn.WriteMessage(websocket.TextMessage, []byte(jsonMssg))
	if err != nil {
		fmt.Printf("Failed to send message to user %s: %v\n", userId, err)
		conn.Close()

		mu.RLock()
		delete(clients, userId)
		mu.RUnlock()
	}

}
