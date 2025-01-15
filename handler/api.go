package handler

import (
	"alparslanahmed/qrGo/database"
	"alparslanahmed/qrGo/helper"
	"alparslanahmed/qrGo/model"
	"bytes"
	"fmt"
	"github.com/chai2010/webp"
	"github.com/gofiber/fiber/v2"
	"github.com/nfnt/resize"
	"image"
	"os"
	"strconv"
	"time"
)

func Hello(c *fiber.Ctx) error {
	return c.SendString("Email sent successfully")
}

func Categories(c *fiber.Ctx) error {
	db := database.DB
	user := GetUser(c.Locals("user"))

	var categories []model.Category
	db.Where("user_id", user.ID).Find(&categories)
	return c.JSON(categories)
}

type CreateCategoryInput struct {
	Name  string `json:"name"`
	Image []byte `json:"image"`
}

func CreateCategory(c *fiber.Ctx) error {
	db := database.DB
	user := GetUser(c.Locals("user"))

	// Parse the image from the request body
	input := new(CreateCategoryInput)
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error parsing input", "data": err})
	}

	// Check the image format
	_, format, err := image.DecodeConfig(bytes.NewReader(input.Image))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Unknown image format", "data": err.Error(), "format": format})
	}

	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(input.Image))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error decoding image", "data": err.Error()})
	}

	// Resize the image to a max height of 200 pixels
	resizedImg := resize.Resize(0, 200, img, resize.Lanczos3)

	// Encode the resized image to a buffer
	var buf bytes.Buffer
	if err := webp.Encode(&buf, resizedImg, &webp.Options{Quality: 100}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error encoding image", "data": err.Error()})
	}

	// Generate a unique file name
	fileName := fmt.Sprintf("category_user_%d_%d.webp", user.ID, time.Now().Unix())

	// Save the image bytes to the public folder
	filePath := fmt.Sprintf("./public/%s", fileName)
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error saving file", "data": err.Error()})
	}

	var category model.Category
	category.UserID = user.ID
	category.Name = input.Name
	category.Slug = helper.GenerateSlug(input.Name)
	category.Image = fmt.Sprintf("/public/%s", fileName)
	db.Create(&category)
	return c.JSON(category)
}

func UpdateCategory(c *fiber.Ctx) error {
	db := database.DB
	user := GetUser(c.Locals("user"))

	// Parse the image from the request body
	input := new(CreateCategoryInput)
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error parsing input", "data": err})
	}

	var categoryId = c.Params("id")
	categoryIdInt, err := strconv.Atoi(categoryId)
	if err != nil {
		fmt.Println("Error:", err)
	}

	var category model.Category
	db.Where("user_id", user.ID).Where("id", categoryIdInt).First(&category)
	if len(input.Image) > 0 {
		// Check the image format
		_, format, err := image.DecodeConfig(bytes.NewReader(input.Image))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Unknown image format", "data": err.Error(), "format": format})
		}

		// Decode the image
		img, _, err := image.Decode(bytes.NewReader(input.Image))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error decoding image", "data": err.Error()})
		}

		// Resize the image to a max height of 200 pixels
		resizedImg := resize.Resize(0, 200, img, resize.Lanczos3)

		// Encode the resized image to a buffer
		var buf bytes.Buffer
		if err := webp.Encode(&buf, resizedImg, &webp.Options{Quality: 100}); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error encoding image", "data": err.Error()})
		}

		// Generate a unique file name
		fileName := fmt.Sprintf("category_user_%d_%d.webp", user.ID, time.Now().Unix())

		// Save the image bytes to the public folder
		filePath := fmt.Sprintf("./public/%s", fileName)
		if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error saving file", "data": err.Error()})
		}

		category.Image = fmt.Sprintf("/public/%s", fileName)

	}

	if input.Name != "" {
		category.Name = input.Name
		category.Slug = helper.GenerateSlug(input.Name)
	}

	db.Save(&category)
	return c.JSON(category)
}

func DeleteCategory(c *fiber.Ctx) error {
	db := database.DB
	user := GetUser(c.Locals("user"))

	var category model.Category
	db.Where("user_id = ? AND id = ?", user.ID, c.Params("id")).Delete(&category)
	return c.JSON(fiber.Map{"message": "Category deleted successfully"})
}

func GetCategory(c *fiber.Ctx) error {
	db := database.DB
	user := GetUser(c.Locals("user"))

	var category model.Category
	db.Where("user_id = ? AND id = ?", user.ID, c.Params("id")).First(&category)
	return c.JSON(category)
}

func Products(c *fiber.Ctx) error {
	db := database.DB
	user := GetUser(c.Locals("user"))

	var categoryId = c.Params("category_id")
	categoryIdInt, err := strconv.Atoi(categoryId)
	if err != nil {
		fmt.Println("Products Error:", err)
	}

	var products []model.Product
	db.Where("user_id", user.ID).Where("category_id", categoryIdInt).Find(&products)
	return c.JSON(products)
}

type CreateProductInput struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Image       []byte  `json:"image"`
	CategoryID  uint    `json:"category_id"`
}

func CreateProduct(c *fiber.Ctx) error {
	db := database.DB
	user := GetUser(c.Locals("user"))

	// Parse the image from the request body
	input := new(CreateProductInput)
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error parsing input", "data": err})
	}

	// Check the image format
	_, format, err := image.DecodeConfig(bytes.NewReader(input.Image))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Unknown image format", "data": err.Error(), "format": format})
	}

	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(input.Image))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error decoding image", "data": err.Error()})
	}

	// Resize the image to a max height of 200 pixels
	resizedImg := resize.Resize(0, 200, img, resize.Lanczos3)

	// Encode the resized image to a buffer
	var buf bytes.Buffer
	if err := webp.Encode(&buf, resizedImg, &webp.Options{Quality: 100}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error encoding image", "data": err.Error()})
	}

	// Generate a unique file name
	fileName := fmt.Sprintf("product_user_%d_%d.webp", user.ID, time.Now().Unix())

	// Save the image bytes to the public folder
	filePath := fmt.Sprintf("./public/%s", fileName)
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error saving file", "data": err.Error()})
	}

	var product model.Product
	product.UserID = user.ID
	product.CategoryID = input.CategoryID
	product.Name = input.Name
	product.Price = input.Price
	product.Description = input.Description
	product.Slug = helper.GenerateSlug(input.Name)
	product.Image = fmt.Sprintf("/public/%s", fileName)
	db.Create(&product)
	return c.JSON(product)
}

func UpdateProduct(c *fiber.Ctx) error {
	db := database.DB
	user := GetUser(c.Locals("user"))

	// Parse the image from the request body
	input := new(CreateProductInput)
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error parsing input", "data": err})
	}

	var productId = c.Params("id")
	productIdInt, err := strconv.Atoi(productId)
	if err != nil {
		fmt.Println("Error:", err)
	}

	var product model.Product
	db.Where("user_id", user.ID).Where("id", productIdInt).First(&product)
	if len(input.Image) > 0 {
		// Check the image format
		_, format, err := image.DecodeConfig(bytes.NewReader(input.Image))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Unknown image format", "data": err.Error(), "format": format})
		}

		// Decode the image
		img, _, err := image.Decode(bytes.NewReader(input.Image))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error decoding image", "data": err.Error()})
		}

		// Resize the image to a max height of 200 pixels
		resizedImg := resize.Resize(0, 200, img, resize.Lanczos3)

		// Encode the resized image to a buffer
		var buf bytes.Buffer
		if err := webp.Encode(&buf, resizedImg, &webp.Options{Quality: 100}); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error encoding image", "data": err.Error()})
		}

		// Generate a unique file name
		fileName := fmt.Sprintf("product_user_%d_%d.webp", user.ID, time.Now().Unix())

		// Save the image bytes to the public folder
		filePath := fmt.Sprintf("./public/%s", fileName)
		if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error saving file", "data": err.Error()})
		}

		product.Image = fmt.Sprintf("/public/%s", fileName)
	}

	if input.Name != "" {
		product.Name = input.Name
		product.Slug = helper.GenerateSlug(input.Name)
	}

	if input.Description != "" {
		product.Description = input.Description
	}

	if input.Price != 0 {
		product.Price = input.Price
	}

	db.Save(&product)

	return c.JSON(product)
}

func DeleteProduct(c *fiber.Ctx) error {
	db := database.DB
	user := GetUser(c.Locals("user"))

	var product model.Product
	db.Where("user_id", user.ID).Where("id", c.Params("id")).Delete(&product)
	return c.JSON(fiber.Map{"message": "Product deleted successfully"})
}

func GetProduct(c *fiber.Ctx) error {
	db := database.DB
	user := GetUser(c.Locals("user"))

	var product model.Product
	db.Where("user_id", user.ID).Where("id", c.Params("id")).First(&product)
	return c.JSON(product)
}
