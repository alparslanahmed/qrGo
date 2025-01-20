package handler

import (
	"alparslanahmed/qrGo/database"
	"alparslanahmed/qrGo/model"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

func CustomerCategories(c *fiber.Ctx) error {
	db := database.DB
	businessSlug := c.Params("business_slug")

	var user model.User
	db.Where("business_slug", businessSlug).First(&user)

	var categories []model.Category
	db.Where("user_id", user.ID).Find(&categories)
	return c.JSON(categories)
}

func CustomerProducts(c *fiber.Ctx) error {
	db := database.DB
	businessSlug := c.Params("business_slug")

	var user model.User
	db.Where("business_slug", businessSlug).First(&user)

	var categoryId = c.Params("category_id")
	categoryIdInt, err := strconv.Atoi(categoryId)
	if err != nil {
		fmt.Println("Products Error:", err)
	}

	var products []model.Product
	db.Where("user_id", user.ID).Where("category_id", categoryIdInt).Find(&products)
	return c.JSON(products)
}

func CustomerBusiness(c *fiber.Ctx) error {
	db := database.DB
	businessSlug := c.Params("business_slug")

	var user model.User
	db.Where("business_slug", businessSlug).First(&user)

	return c.JSON(user.UserBusiness())
}
