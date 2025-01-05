// main.go
package main

import (
	"alparslanahmed/qrGo/database"
	"alparslanahmed/qrGo/handler"
	"alparslanahmed/qrGo/router"
	"fmt"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	go handler.WebsocketRunner()
	app := fiber.New(fiber.Config{
		Prefork: true,
	})

	app.Use(limiter.New(limiter.Config{
		Max:        200,
		Expiration: 30 * time.Second,
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))
	app.Use(helmet.New())

	// Initialize database
	database.ConnectDB()
	database.ConnectRedis()

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	router.SetupRoutes(app)
	router.SetupWebsocket(app)
	err := app.Listen(":5001")

	if err != nil {
		fmt.Println("Error: ", err.Error())
		panic(err)
	}
}
