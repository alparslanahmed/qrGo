package handler

import (
	"alparslanahmed/qrGo/database"
	"alparslanahmed/qrGo/model"
	"github.com/gofiber/fiber/v2"
)

func UpdateBillingInfo(c *fiber.Ctx) error {
	// Parse the request body into a BillingInfo struct
	var billingInfo model.BillingInfo
	if err := c.BodyParser(&billingInfo); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse billing info",
		})
	}

	// Get the user ID from the context (assuming it's set by your auth middleware)
	user := GetUser(c.Locals("user"))

	// Check if billing info already exists for this user
	var existingBillingInfo model.BillingInfo
	result := database.DB.Where("user_id = ?", user.ID).First(&existingBillingInfo)

	if result.Error == nil {
		// Update existing billing info
		existingBillingInfo.Address = billingInfo.Address
		existingBillingInfo.City = billingInfo.City
		existingBillingInfo.State = billingInfo.State
		existingBillingInfo.ZipCode = billingInfo.ZipCode
		existingBillingInfo.Country = billingInfo.Country

		if err := database.DB.Save(&existingBillingInfo).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not update billing info",
			})
		}
	} else {
		// Create new billing info
		billingInfo.UserID = user.ID
		if err := database.DB.Create(&billingInfo).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not create billing info",
				"data":  err.Error(),
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Billing info updated successfully",
	})
}
