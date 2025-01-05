package router

import (
	"alparslanahmed/qrGo/config"
	"alparslanahmed/qrGo/database"
	"alparslanahmed/qrGo/handler"
	"alparslanahmed/qrGo/model"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
	"github.com/golang-jwt/jwt/v4"
	"log"
	"strconv"
)

func SetupWebsocket(app *fiber.App) {

	ws := app.Group("/ws", logger.New())

	ws.Use("/", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", false)
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	ws.Get("/", websocket.New(func(c *websocket.Conn) {
		var user model.User
		db := database.DB
		allowed := false
		// When the function returns, unregister the client and close the connection
		defer func() {
			handler.Unregister <- c
			c.Close()
		}()

		// Register the client
		handler.Register <- c

		// c.Locals is added to the *websocket.Conn
		// log.Println(c.Locals("allowed"))  // true
		// log.Println(c.Params("id"))       // 123
		// log.Println(c.Query("v"))         // 1.0
		// log.Println(c.Cookies("session")) // ""

		// websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index
		var (
			mt  int
			msg []byte
			err error
		)

		for {
			if mt, msg, err = c.ReadMessage(); err != nil {
				// log.Println("read:", err)
				break
			}

			var jsonEvent handler.SocketEvent

			jsonString := string(msg[:])

			json.Unmarshal([]byte(jsonString), &jsonEvent)

			if !allowed {

				if jsonEvent.Event == "auth" {
					println("socket auth started")
					tokenString := jsonEvent.Data.(string)
					claims := jwt.MapClaims{}
					_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
						return []byte(config.Config("SECRET")), nil
					})

					if err != nil {
						println("socket auth failed token invalid")
						break
					}

					user_id := claims["user_id"]

					db.Find(&user, user_id)

					if user.Name == "" {
						println("socket auth user error", user.Name, c.Params("id"), fmt.Sprintf("%d", user.ID))
						break
					}

					allowed = true
					authConn := &handler.AuthorizedConnection{Connection: c, User: user, Client: &handler.WsClient{}, UserID: strconv.FormatUint(uint64(user.ID), 10)}
					handler.Authenticate <- authConn
					handler.SubscribeToUserChannel(strconv.FormatUint(uint64(user.ID), 10), c)
					authResponse, _ := json.Marshal(handler.SocketEvent{Event: "auth", Data: "Authorized"})
					c.WriteMessage(mt, authResponse)
				}

				continue
			}

			switch jsonEvent.Event {
			//case "generate":
			//	fmt.Println("GENERATE RECEIVED")
			//	handler.GenerateTattoo(jsonEvent.Data.(map[string]interface{}), c, user)
			//case "leave_table":
			}

			if mt == websocket.TextMessage {
				// Broadcast the received message
				//handler.Broadcast <- jsonEvent
			} else {
				log.Println("websocket message received of type", mt)
			}
		}

	}))

}
