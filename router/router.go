package router

import (
	"alparslanahmed/qrGo/handler"
	"alparslanahmed/qrGo/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// SetupRoutes setup router api
func SetupRoutes(app *fiber.App) {
	// Middleware
	api := app.Group("/api", logger.New())
	api.Get("/", handler.Hello)
	api.Post("/billing-info", middleware.Protected(), handler.UpdateBillingInfo)

	// Payment
	payment := api.Group("/payment")
	payment.Post("/create", middleware.Protected(), handler.CreatePayment)

	// Auth
	auth := api.Group("/auth")
	auth.Post("/token", handler.Login)
	auth.Post("/register", handler.RegisterUser)
	auth.Post("/verify_email", handler.VerifyUser)
	auth.Get("/send_verification_code", middleware.Protected(), handler.RequestVerificationCode)
	auth.Post("/forgot-password", handler.ForgotPassword)
	auth.Post("/reset-password", handler.ResetPassword)
	auth.Post("/change-password", middleware.Protected(), handler.ChangePassword)
	auth.Post("/google", handler.GoogleLogin)
	auth.Post("/profile", middleware.Protected(), handler.UpdateProfile)
	auth.Get("/user", middleware.Protected(), handler.User)
}
