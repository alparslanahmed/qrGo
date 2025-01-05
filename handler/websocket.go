package handler

import (
	"alparslanahmed/qrGo/database"
	"alparslanahmed/qrGo/model"
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

type WsClient struct {
	isClosing bool
	mu        sync.Mutex
}

type SocketEvent struct {
	Event string
	Data  interface{}
}

type AuthorizedConnection struct {
	Connection *websocket.Conn
	User       model.User
	Client     *WsClient
	UserID     string
}

var Clients = make(map[*websocket.Conn]*WsClient) // Note: although large maps with pointer-like types (e.g. strings) as keys are slow, using pointers themselves as keys is acceptable and fast
var Authorized = []*AuthorizedConnection{}

var Authenticate = make(chan *AuthorizedConnection)
var Register = make(chan *websocket.Conn)
var Broadcast = make(chan SocketEvent)
var Unregister = make(chan *websocket.Conn)

func WebsocketRunner() {
	for {
		select {
		case connection := <-Register:
			Clients[connection] = &WsClient{}
			log.Println("connection registered")
		case authenticate := <-Authenticate:
			Authorized = append(Authorized, authenticate)
			log.Println("connection authenticated")
		case event := <-Broadcast:
			log.Println("message received:", event.Event)
			// Send the message to all clients
			for _, authorized := range Authorized {
				go func(a *AuthorizedConnection) { // send to each client in parallel so we don't block on a slow client
					a.Client.mu.Lock()
					defer a.Client.mu.Unlock()
					if a.Client.isClosing {
						return
					}

					sendData, _ := json.Marshal(event)

					if err := a.Connection.WriteMessage(websocket.TextMessage, []byte(sendData)); err != nil {
						a.Client.isClosing = true
						log.Println("write error:", err)

						a.Connection.WriteMessage(websocket.CloseMessage, []byte{})
						a.Connection.Close()
						Unregister <- a.Connection
					}
				}(authorized)
			}

		case connection := <-Unregister:
			// Remove the client from the hub
			delete(Clients, connection)
			log.Println("connection unregistered")
		}
	}

}

// Create a function to send an event to a specific user
func SendEventToUser(userID string, event SocketEvent) {
	// Publish the event to the Redis channel for the specific user
	eventData, _ := json.Marshal(event)
	database.RedisClient.Publish(context.Background(), "user:"+userID, eventData)
}

// Subscribe to a Redis channel for each user upon authentication
func SubscribeToUserChannel(userID string, conn *websocket.Conn) {
	pubsub := database.RedisClient.Subscribe(context.Background(), "user:"+userID)
	go func() {
		for msg := range pubsub.Channel() {
			var event SocketEvent
			json.Unmarshal([]byte(msg.Payload), &event)
			sendEventToConnection(conn, event)
		}
	}()
}

// Handle incoming messages from Redis and send them to the appropriate WebSocket connection
func sendEventToConnection(conn *websocket.Conn, event SocketEvent) {
	sendData, _ := json.Marshal(event)
	conn.WriteMessage(websocket.TextMessage, sendData)
}
