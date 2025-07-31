// controllers/auth_controller.go
package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"notedteam.backend/config"
	"notedteam.backend/models"
	"notedteam.backend/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// RegisterInput mendefinisikan data yang dibutuhkan untuk registrasi
type RegisterInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginInput mendefinisikan data yang dibutuhkan untuk login
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// user := models.User{Name: input.Name, Email: input.Email, Password: string(hashedPassword)}

	token, err := generateSecureToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate verification token"})
		return
	}

	user := models.User{
		Name:              input.Name,
		Email:             input.Email,
		Password:          string(hashedPassword),
		IsVerified:        false,
		VerificationToken: token,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Kirim email verifikasi dalam sebuah goroutine agar tidak memblokir respons
	go utils.SendVerificationEmail(user.Email, user.VerificationToken)

	c.JSON(http.StatusCreated, gin.H{"message": "Registration successful. Please check your email to verify your account."})
}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	// Cari user berdasarkan email
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Bandingkan password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate token JWT
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}
	if !user.IsVerified {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account not verified. Please check your email."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}
func VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verification token is required"})
		return
	}

	var user models.User
	if err := config.DB.Where("verification_token = ?", token).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid or expired verification token"})
		return
	}

	// Update status user
	user.IsVerified = true
	user.VerificationToken = "" // Hapus token setelah digunakan
	config.DB.Save(&user)

	// Tampilkan halaman sukses sederhana
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<h1>Email Verified!</h1><p>Your account has been successfully verified. You can now close this window and log in to the application.</p>"))
}
func ForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		// Kirim respons sukses bahkan jika email tidak ada, untuk mencegah user enumeration
		c.JSON(http.StatusOK, gin.H{"message": "If an account with that email exists, a password reset link has been sent."})
		return
	}

	token, err := generateSecureToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Set token dan waktu kedaluwarsa (1 jam dari sekarang)
	expTime := time.Now().Add(time.Hour * 1)
	user.PasswordResetToken = token
	user.PasswordResetTokenExp = &expTime
	config.DB.Save(&user)

	// Kirim email
	go utils.SendPasswordResetEmail(user.Email, user.PasswordResetToken)

	c.JSON(http.StatusOK, gin.H{"message": "If an account with that email exists, a password reset link has been sent."})
}
func ShowResetPasswordPage(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("<h1>Invalid Token</h1>"))
		return
	}

	// (Opsional) Verifikasi token ada sebelum menampilkan halaman
	var user models.User
	if err := config.DB.Where("password_reset_token = ? AND password_reset_token_exp > ?", token, time.Now()).First(&user).Error; err != nil {
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("<h1>Invalid or Expired Token</h1><p>Please request a new password reset link.</p>"))
		return
	}

	// Sajikan halaman HTML dengan form
	html := `
    <!DOCTYPE html><html><head><title>Reset Password</title></head><body>
    <h2>Set New Password</h2>
    <form action="/auth/reset-password" method="POST">
        <input type="hidden" name="token" value="` + token + `">
        <label for="password">New Password:</label><br>
        <input type="password" id="password" name="password" required><br><br>
        <input type="submit" value="Reset Password">
    </form>
    </body></html>
    `
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
func ResetPassword(c *gin.Context) {
	token := c.PostForm("token")
	password := c.PostForm("password")

	if token == "" || password == "" {
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("<h1>Error</h1><p>Token and password are required.</p>"))
		return
	}

	var user models.User
	if err := config.DB.Where("password_reset_token = ? AND password_reset_token_exp > ?", token, time.Now()).First(&user).Error; err != nil {
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("<h1>Invalid or Expired Token</h1>"))
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)
	user.PasswordResetToken = ""
	user.PasswordResetTokenExp = nil // Set ke null
	config.DB.Save(&user)

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<h1>Success!</h1><p>Your password has been reset. You can now close this window and log in with your new password.</p>"))
}
