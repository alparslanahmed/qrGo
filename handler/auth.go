package handler

import (
	"alparslanahmed/qrGo/config"
	"alparslanahmed/qrGo/database"
	"alparslanahmed/qrGo/email"
	"alparslanahmed/qrGo/helper"
	"alparslanahmed/qrGo/model"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/nfnt/resize"
	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"net/mail"
	"os"
	"strings"
	"time"

	"google.golang.org/api/idtoken"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

// Register new user
func RegisterUser(c *fiber.Ctx) error {
	type RegisterInput struct {
		Email        string `json:"email"`
		Name         string `json:"name"`
		Password     string `json:"password"`
		BusinessName string `json:"business_name"`
		TaxOffice    string `json:"tax_office"`
		TaxNumber    string `json:"tax_number"`
		Address      string `json:"address"`
		Phone        string `json:"phone"`
	}
	input := new(RegisterInput)

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error on parsing", "data": err})
	}

	email := input.Email
	pass := input.Password

	if email == "" || pass == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error on input", "data": nil})
	}

	var user_with_email model.User
	db := database.DB
	db.Where("email = ?", email).First(&user_with_email)

	if user_with_email.Email != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "This email already in use", "data": nil})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error on register request", "data": err})
	}

	user := &model.User{
		Email:        email,
		Name:         input.Name,
		Password:     string(hash),
		BusinessName: input.BusinessName,
		BusinessSlug: fmt.Sprintf("%s-%d", helper.GenerateSlug(input.BusinessName), time.Now().Unix()),
		TaxOffice:    input.TaxOffice,
		TaxNumber:    input.TaxNumber,
		Address:      input.Address,
		Phone:        input.Phone,
	}

	if err := db.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error on register request", "data": err})
	}

	go SendVerificationEmail(user)

	return c.JSON(fiber.Map{"status": "success", "message": "Success register", "data": user.UserPublic(db)})
}

func RequestVerificationCode(c *fiber.Ctx) error {
	redis := database.RedisClient

	user := GetUser(c.Locals("user"))

	if user.EmailVerified {
		return c.JSON(fiber.Map{"status": "success", "message": "Success verify", "data": nil})
	}

	redisKey := fmt.Sprintf("verification_code.%d", user.ID)

	if redis.Get(context.Background(), redisKey).Val() != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "You have to wait before requesting code again", "data": nil})
	}

	go SendVerificationEmail(&user)

	redis.Set(context.Background(), redisKey, user.ID, time.Minute*5)

	return c.JSON(fiber.Map{"status": "success", "message": "Success verification code sent", "data": nil})
}

func SendVerificationEmail(user *model.User) {
	code, err := helper.GenerateRandomString(16)
	if err != nil {
		fmt.Println("Error generating random string:", err)
		return
	}

	var verify model.VerifyCode = model.VerifyCode{
		Email: user.Email,
		Code:  code,
	}

	db := database.DB
	db.Create(&verify)

	// Send email
	var message = `
<p>Sayın [User],</p>

<p>Enfes Menü'ye kaydolduğunuz için teşekkür ederiz. Kaydınızı tamamlamak ve hesabınızın güvenliğini sağlamak için lütfen aşağıdaki butona tıklayarak e-posta adresinizi doğrulayın:</p>

<p style="text-align: center;">
    <a href="[verification_link]" style="background-color: #4CAF50; color: white; padding: 14px 20px; text-align: center; text-decoration: none; display: inline-block; font-size: 16px; margin: 4px 2px; cursor: pointer; border-radius: 4px;">E-posta Adresini Doğrula</a>
</p>

<p>Yukarıdaki buton çalışmazsa, aşağıdaki bağlantıyı tarayıcınıza kopyalayıp yapıştırabilirsiniz:</p>

<p style="word-break: break-all;">
    <a href="[verification_link]" style="color: #1a73e8;">[verification_link]</a>
</p>

<p>Bu bağlantı güvenlik nedeniyle 24 saat içinde sona erecektir. Bu süre içinde e-postanızı doğrulamazsanız, yeni bir doğrulama bağlantısı talep etmeniz gerekebilir.</p>

<p>Eğer Enfes Menü ile bir hesap oluşturmadıysanız, lütfen bu e-postayı dikkate almayın.</p>

<p>Herhangi bir sorunuz veya endişeniz varsa, lütfen destek ekibimizle <a href="mailto:iletisim@enfesmenu.com" style="color: #1a73e8;">iletisim@enfesmenu.com</a> adresinden iletişime geçin.</p>

<p>Enfes Menü'yi tercih ettiğiniz için teşekkür ederiz!</p>

<p>Saygılarımızla,<br>
Enfes Menü Ekibi</p>
`

	message = strings.Replace(message, "[User]", user.Name, -1)
	message = strings.Replace(message, "[verification_link]", fmt.Sprintf("%s/verification?token=%s", config.Config("FRONTEND_URL"), code), -1)
	email.SendHTMLEmail(user.Email, "Email Hesabınızı Onaylayın", message, "/email.html")
}

func VerifyUser(c *fiber.Ctx) error {
	type VerifyInput struct {
		Code string `json:"code"`
	}
	input := new(VerifyInput)

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error on parsing", "data": err})
	}

	code := input.Code

	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error on input", "data": nil})
	}

	var verify model.VerifyCode
	db := database.DB

	db.Where("code = ?", code).First(&verify)

	if verify.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Invalid code", "data": nil})
	}

	var user model.User
	db.Where("email = ?", verify.Email).First(&user)

	// Update user
	user.EmailVerified = true
	db.Save(&user)

	db.Delete(&verify)

	return c.JSON(fiber.Map{"status": "success", "message": "Success verify", "data": user.UserPublic(db)})
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login get user and password
func Login(c *fiber.Ctx) error {

	input := new(LoginInput)

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error on login request", "data": err})
	}

	identity := input.Email
	pass := input.Password

	var user *model.User
	var err error

	if valid(identity) {
		user, err = getUserByEmail(identity)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error on fetching user", "data": err})
	}

	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "User not found", "data": nil})
	}

	if !CheckPasswordHash(pass, user.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Invalid password", "data": nil})
	}

	t, err := IssueJWT(*user)

	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Success login", "token": t, "user": user.UserPublic(database.DB)})
}

func IssueJWT(user model.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["email"] = user.Email
	claims["user_id"] = user.ID
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString([]byte(config.Config("SECRET")))

	return t, err
}

// Get user
func User(c *fiber.Ctx) error {
	user := GetUser(c.Locals("user"))

	if user.Name == "" {
		return c.Status(401).JSON(fiber.Map{"status": "error", "message": "Error!", "data": nil})
	}

	return c.JSON(fiber.Map{"status": "success", "data": user.UserPublic(database.DB)})
}

func GetUser(userToken interface{}) model.User {
	token := userToken.(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	user_id := claims["user_id"]

	var user model.User
	db := database.DB
	db.Find(&user, user_id)

	return user
}

// CheckPasswordHash compare password with hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func getUserByEmail(e string) (*model.User, error) {
	db := database.DB
	var user model.User
	if err := db.Where(&model.User{Email: e}).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func valid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func GoogleLogin(c *fiber.Ctx) error {
	type TokenRequest struct {
		Token string `json:"token"`
	}

	var req TokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request",
		})
	}

	token := req.Token

	// Replace with your Google Client ID
	clientID := config.Config("GOOGLE_CLIENT_ID")

	payload, err := verifyGoogleIDToken(token, clientID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid token",
		})
	}

	// Extract user information from the payload
	user := model.User{
		Email:    payload.Claims["email"].(string),
		Name:     payload.Claims["name"].(string),
		Password: "google",
	}

	// Check if the user exists in the database
	if err := database.DB.Where("email = ?", user.Email).First(&user).Error; err != nil {
		// Create a new user if not exists
		if err := database.DB.Create(&user).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	// Generate JWT token
	t, err := IssueJWT(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not issue token",
		})
	}

	if !user.EmailVerified {
		go SendVerificationEmail(&user)
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Success login", "token": t, "user": user.UserPublic(database.DB)})
}

func verifyGoogleIDToken(idToken, clientID string) (*idtoken.Payload, error) {
	ctx := context.Background()
	payload, err := idtoken.Validate(ctx, idToken, clientID)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func ForgotPassword(c *fiber.Ctx) error {
	redis := database.RedisClient

	type ForgotPasswordInput struct {
		Email string `json:"email"`
	}
	input := new(ForgotPasswordInput)

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error on parsing", "data": err})
	}

	mail := input.Email

	if mail == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error on input", "data": nil})
	}

	var user model.User
	db := database.DB
	db.Where("email = ?", mail).First(&user)

	if user.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "User not found", "data": nil})
	}

	redisKey := fmt.Sprintf("password_reset.%d", user.ID)

	if redis.Get(context.Background(), redisKey).Val() != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "You have to wait before requesting code again", "data": nil})
	}

	code, err := helper.GenerateRandomString(16)
	if err != nil {
		fmt.Println("Error generating random string:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error on generating code", "data": nil})
	}

	var reset = model.PasswordCode{
		UserId: user.ID,
		Code:   code,
	}

	db.Create(&reset)

	// Send email
	var message = `
    <p>Sevgili [User],</p>

<p>Enfes Menü hesabınızın şifresini sıfırlamanız için bir talep aldık. Bu talebi siz yapmadıysanız, lütfen bu e-postayı göz ardı edin. Aksi takdirde, aşağıdaki bağlantıyı kullanarak şifrenizi sıfırlayabilirsiniz:</p>

<p style="text-align: center;">
    <a href="[reset_link]" style="background-color: #4CAF50; color: white; padding: 14px 20px; text-align: center; text-decoration: none; display: inline-block; font-size: 16px; margin: 4px 2px; cursor: pointer; border-radius: 4px;">Şifreyi Sıfırla</a>
</p>

<p>Yukarıdaki buton çalışmıyorsa, aşağıdaki bağlantıyı tarayıcınıza kopyalayıp yapıştırabilirsiniz:</p>

<p style="word-break: break-all;">
    <a href="[reset_link]" style="color: #1a73e8;">[reset_link]</a>
</p>

<p>Güvenlik nedeniyle bu bağlantı 24 saat sonra geçersiz olacaktır. Bu süre içinde şifrenizi sıfırlamazsanız, yeni bir şifre sıfırlama bağlantısı talep etmeniz gerekebilir.</p>

<p>Herhangi bir sorunuz varsa veya daha fazla yardıma ihtiyacınız olursa, lütfen <a href="mailto:iletisim@enfesmenu.com" style="color: #1a73e8;">iletisim@enfesmenu.com</a> adresinden destek ekibimizle iletişime geçin.</p>

<p>Enfes Menü'yü kullandığınız için teşekkür ederiz!</p>

<p>Saygılarımızla,<br>
Enfes Menü Ekibi</p>
    `

	message = strings.Replace(message, "[User]", user.Name, -1)
	message = strings.Replace(message, "[reset_link]", fmt.Sprintf("%s/password-reset?token=%s", config.Config("FRONTEND_URL"), code), -1)
	go email.SendHTMLEmail(user.Email, "Şifre Sıfırlama Talebi", message, "/email.html")
	redis.Set(context.Background(), redisKey, user.ID, time.Minute*5)
	return c.JSON(fiber.Map{"status": "success", "message": "Password reset email sent", "data": nil})
}

func ResetPassword(c *fiber.Ctx) error {
	type ResetPasswordInput struct {
		Code     string `json:"code"`
		Password string `json:"password"`
	}
	input := new(ResetPasswordInput)

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error on parsing", "data": err})
	}

	code := input.Code
	newPassword := input.Password

	if code == "" || newPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error on input", "data": nil})
	}

	var reset model.PasswordCode
	db := database.DB
	db.Where("code = ?", code).First(&reset)

	if reset.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Invalid or expired reset code", "data": nil})
	}

	var user model.User
	db.Where("id = ?", reset.UserId).First(&user)

	if user.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "User not found", "data": nil})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error on password reset", "data": err})
	}

	user.Password = string(hash)
	db.Save(&user)
	db.Delete(&reset)

	// Send password reset successful email
	var message = `
    <p>Sevgili [User],</p>

<p>Enfes Menü hesabınızın şifresi başarıyla sıfırlandı. Bu değişikliği siz yapmadıysanız, lütfen hemen destek ekibimizle iletişime geçin.</p>

<p>Herhangi bir sorunuz varsa veya daha fazla yardıma ihtiyacınız olursa, lütfen <a href="mailto:iletisim@enfesmenu.com" style="color: #1a73e8;">iletisim@enfesmenu.com</a> adresinden destek ekibimizle iletişime geçin.</p>

<p>Enfes Menü’yü kullandığınız için teşekkür ederiz!</p>

<p>Saygılarımızla,<br>
Enfes Menü Ekibi</p>
    `

	message = strings.Replace(message, "[User]", user.Name, -1)
	go email.SendHTMLEmail(user.Email, "Şifre Sıfırlama Başarılı", message, "/email.html")

	return c.JSON(fiber.Map{"status": "success", "message": "Password reset successful", "data": nil})
}

func ChangePassword(c *fiber.Ctx) error {
	type ChangePasswordInput struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	input := new(ChangePasswordInput)

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error on parsing", "data": err})
	}

	oldPassword := input.OldPassword
	newPassword := input.NewPassword

	if oldPassword == "" || newPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error on input", "data": nil})
	}

	user := GetUser(c.Locals("user"))

	if !CheckPasswordHash(oldPassword, user.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Invalid old password", "data": nil})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error on password change", "data": err})
	}

	user.Password = string(hash)
	db := database.DB
	if err := db.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error on password change", "data": err})
	}

	// Send password change notification email
	var message = `
    <p>Sevgili [User],</p> <p>Enfes Menü hesabınızın şifresi başarıyla değiştirildi. Bu değişikliği siz yapmadıysanız, lütfen hemen destek ekibimizle iletişime geçin.</p> <p>Herhangi bir sorunuz veya ek yardıma ihtiyacınız olursa, lütfen şu adresten bize ulaşın: <a href="mailto:iletisim@enfesmenu.com" style="color: #1a73e8;">iletisim@enfesmenu.com</a>.</p> <p>Enfes Menü’yü kullandığınız için teşekkür ederiz!</p> <p>Saygılarımızla,<br> Enfes Menü Ekibi</p>
    `

	message = strings.Replace(message, "[User]", user.Name, -1)
	go email.SendHTMLEmail(user.Email, "Password Change Notification", message, "/email.html")

	return c.JSON(fiber.Map{"status": "success", "message": "Password changed successfully", "data": nil})
}

// Add this function to the existing auth.go file

func UpdateProfile(c *fiber.Ctx) error {
	type UpdateProfileInput struct {
		Name         string `json:"name"`
		BusinessName string `json:"business_name"`
		TaxOffice    string `json:"tax_office"`
		TaxNumber    string `json:"tax_number"`
		Address      string `json:"address"`
		Phone        string `json:"phone"`
	}
	input := new(UpdateProfileInput)

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error on parsing", "data": err})
	}

	user := GetUser(c.Locals("user"))

	if input.Name != "" {
		user.Name = input.Name
	}
	if input.BusinessName != "" {
		user.BusinessName = input.BusinessName
		user.BusinessSlug = fmt.Sprintf("%s-%d", helper.GenerateSlug(input.BusinessName), time.Now().Unix())
	}
	if input.TaxOffice != "" {
		user.TaxOffice = input.TaxOffice
	}
	if input.TaxNumber != "" {
		user.TaxNumber = input.TaxNumber
	}
	if input.Address != "" {
		user.Address = input.Address
	}
	if input.Phone != "" {
		user.Phone = input.Phone
	}

	db := database.DB
	if err := db.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error updating profile", "data": err})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Profile updated successfully", "data": user.UserPublic(db)})
}

type UpdateAvatarInput struct {
	Image []byte `json:"image"`
}

func UpdateAvatar(c *fiber.Ctx) error {

	// Parse the image from the request body
	input := new(UpdateAvatarInput)
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Error parsing image", "data": err})
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
	if err := png.Encode(&buf, resizedImg); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error encoding image", "data": err.Error()})
	}

	// Generate a unique file name
	fileName := fmt.Sprintf("avatar_%d.png", time.Now().Unix())

	// Save the image bytes to the public folder
	filePath := fmt.Sprintf("./public/%s", fileName)
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error saving file", "data": err.Error()})
	}

	// Get the user from the context
	user := GetUser(c.Locals("user"))

	// Remove the old image if it exists
	if user.LogoURL != "" {
		oldFilePath := fmt.Sprintf(".%s", user.LogoURL)
		if err := os.Remove(oldFilePath); err != nil && !os.IsNotExist(err) {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error removing old file", "data": err.Error()})
		}
	}

	// Update the user's avatar URL
	user.LogoURL = fmt.Sprintf("/public/%s", fileName)

	// Save the user to the database
	db := database.DB
	if err := db.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error updating user", "data": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Avatar updated successfully", "data": user.UserPublic(db)})
}
