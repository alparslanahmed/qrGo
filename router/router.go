package router

import (
	"alparslanahmed/qrGo/handler"
	"alparslanahmed/qrGo/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// SetupRoutes setup router api
func SetupRoutes(app *fiber.App) {
	app.Static("/public", "./public")

	// Middleware
	api := app.Group("/api", logger.New())
	api.Get("/", handler.Hello)
	api.Post("/billing-info", middleware.Protected(), handler.UpdateBillingInfo)
	api.Post("/categories", middleware.Protected(), handler.CreateCategory)
	api.Put("/categories/:id", middleware.Protected(), handler.UpdateCategory)
	api.Delete("/categories/:id", middleware.Protected(), handler.DeleteCategory)
	api.Get("/categories", middleware.Protected(), handler.Categories)
	api.Get("/categories/:id", middleware.Protected(), handler.GetCategory)
	api.Post("/products", middleware.Protected(), handler.CreateProduct)
	api.Get("/products/:category_id", middleware.Protected(), handler.Products)
	api.Get("/product/:id", middleware.Protected(), handler.GetProduct)
	api.Put("/products/:id", middleware.Protected(), handler.UpdateProduct)
	api.Delete("/products/:id", middleware.Protected(), handler.DeleteProduct)

	customer := api.Group("/customer")
	customer.Get("/:business_slug/categories", handler.CustomerCategories)
	customer.Get("/:business_slug/products/:category_id", handler.CustomerProducts)
	customer.Get("/:business_slug/business", handler.CustomerBusiness)
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
	auth.Post("/avatar", middleware.Protected(), handler.UpdateAvatar)
	auth.Get("/user", middleware.Protected(), handler.User)

}
