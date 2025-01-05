package handler

import (
	"alparslanahmed/qrGo/config"
	"alparslanahmed/qrGo/database"
	"alparslanahmed/qrGo/email"
	"alparslanahmed/qrGo/model"
	"fmt"

	"github.com/gofiber/fiber/v2"
	parasut "github.com/ozgur-yalcin/parasut.go/src"
)

var cfg = parasut.Config{CompanyID: config.Config("PARASUT_COMPANYID"), ClientID: config.Config("PARASUT_CLIENTID"),
	ClientSecret: config.Config("PARASUT_CLIENTSECRET"), Username: config.Config("PARASUT_USERNAME"), Password: config.Config("PARASUT_PASSWORD")}
var api = &parasut.API{Config: cfg}
var auth = api.Authorize()

func CreatePayment(c *fiber.Ctx) error {
	db := database.DB
	user := GetUser(c.Locals("user"))

	// Parse the request body
	var input struct {
		CardNumber string `json:"cardNumber"`
		ExpiryDate string `json:"expiryDate"`
		CVV        string `json:"cvv"`
		Name       string `json:"name"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Error on payment parse",
			"data":    err,
		})
	}

	// TODO: Implement actual payment processing logic here
	// This could involve calling a payment gateway API

	// For now, we'll just create a payment record
	payment := model.Payment{
		UserID: user.ID,
		Amount: 15,
		Status: "completed", // In a real scenario, this would depend on the payment gateway response
	}

	if err := db.Create(&payment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Couldn't create payment",
			"data":    err,
		})
	}

	card := model.Card{
		UserID:         user.ID,
		MaskedNumber:   "maskednumberfrompaymentprocessor", // Replace with actual masked number
		Brand:          "Unknown",                          // You should determine this from the card number
		ExpiryDate:     input.ExpiryDate,
		CardholderName: input.Name,
		IsDefault:      true,                           // Make this card the default
		Token:          "token_from_payment_processor", // Replace with actual token
	}

	// Set all other cards to non-default
	db.Model(&model.Card{}).Where("user_id = ?", user.ID).Update("is_default", false)

	if err := db.Create(&card).Error; err != nil {
		// Log the error, but don't fail the payment
		fmt.Printf("Error saving card: %v\n", err)
	}

	// Send confirmation email
	emailSubject := "Subscription Payment Confirmation"
	emailBody := fmt.Sprintf(`
		Dear %s,

		Thank you for your subscription payment of $%.2f.
		Payment Details:
		- Amount: $%.2f
		- Card: %s
		- Date: %s

		If you have any questions, please don't hesitate to contact us.

		Best regards,
		Team
	`, user.Name, payment.Amount, payment.Amount, card.MaskedNumber, payment.CreatedAt.Format("2006-01-02 15:04:05"))

	err := email.SendHTMLEmail(user.Email, emailSubject, emailBody, "/email.html")
	if err != nil {
		// Log the error, but don't return it to the user
		fmt.Printf("Error sending confirmation email: %v\n", err)
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Payment processed successfully",
		"data":    payment,
	})
}
