package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/KashyretsIvanna/voice-balance/config"
	"github.com/KashyretsIvanna/voice-balance/internals/model"
	"github.com/KashyretsIvanna/voice-balance/internals/repositories"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleUser struct {
	Email string `json:"email"`
}

// Define the LoginRequest struct globally so it's recognized by Swagger
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

var (
	googleOauthConfig = oauth2.Config{
		RedirectURL:  config.Config("GOOGLE_LOGIN_CALLBACK_URL"), // Adjust for your environment
		ClientID:     config.Config("GOOGLE_CLIENT_ID"),          // Replace with your Client ID
		ClientSecret: config.Config("GOOGLE_CLIENT_SECRET"),      // Replace with your Client Secret
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	oauthStateString = "random" // Use a more secure state in production
	jwtAccessKey     = []byte(config.Config("ACCESS_SECRET_KEY"))
	jwtRefreshKey    = []byte(config.Config("REFRESH_SECRET_KEY"))
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func generateToken(email string, secret []byte, duration time.Duration) (string, error) {
	claims := &jwt.StandardClaims{
		Subject:   email,
		ExpiresAt: time.Now().Add(duration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func parseToken(tokenStr string, secret []byte) (*jwt.StandardClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return token.Claims.(*jwt.StandardClaims), nil
}

// isValidEmail checks if the email format is valid
func isValidEmail(email string) bool {
	// Simple email validation regex
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// RegisterRequest defines the body structure for the registration request
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register handles user registration
// @Summary      Register a new user
// @Description  Registers a new user with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body RegisterRequest true "User Registration"
// @Success      201  {string} string "User created successfully"
// @Failure      400  {string} string "Invalid request"
// @Failure      409  {string} string "User already exists"
// @Failure      500  {string} string "Internal Server Error"
// @Router       /api/auth/register [post]
func Register(c *fiber.Ctx) error {
	var regReq RegisterRequest

	// Parse the body to get the user email and password
	if err := c.BodyParser(&regReq); err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid request")
	}

	// Validate email format
	if !isValidEmail(regReq.Email) {
		return c.Status(http.StatusBadRequest).SendString("Invalid email format")
	}

	// Check if the user already exists
	existingUser, err := repositories.GetUserByEmail(regReq.Email)
	fmt.Print(existingUser)
	if err == nil && existingUser.Email != "" {
		return c.Status(http.StatusConflict).SendString("User already exists")
	}

	// Hash the password before saving it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(regReq.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error hashing password")
	}

	// Create a new user
	user := &model.User{
		Email:    regReq.Email,
		Password: string(hashedPassword),
	}

	// Save the new user to the database
	if err := repositories.AddUser(user); err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Could not create user")
	}

	// Return success response
	return c.Status(http.StatusCreated).SendString("User created successfully")
}

// EmailPasswordLogin handles login with email/password
// @Summary      Login with Email and Password
// @Description  Authenticates a user with email and password, returns access and refresh tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body LoginRequest true "Login credentials"
// @Success      200  {object} TokenPair "Successful login"
// @Failure      400  {string} string "Invalid request"
// @Failure      401  {string} string "Invalid credentials"
// @Failure      500  {string} string "Internal Server Error"
// @Router       /api/auth/email-login [post]
func EmailPasswordLogin(c *fiber.Ctx) error {
	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var loginReq LoginRequest
	if err := c.BodyParser(&loginReq); err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid request")
	}

	user, err := repositories.GetUserByEmail(loginReq.Email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password)) != nil {
		return c.Status(http.StatusUnauthorized).SendString("Invalid credentials")
	}

	accessToken, err := generateToken(user.Email, jwtAccessKey, 15*time.Hour)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Could not generate access token")
	}

	refreshToken, err := generateToken(user.Email, jwtRefreshKey, 7*24*time.Hour)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Could not generate refresh token")
	}

	// Save refresh token in database
	repositories.SaveRefreshToken(user.Email, refreshToken)

	return c.JSON(TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// RefreshToken generates a new access token using a refresh token
// @Summary      Refresh Access Token
// @Description  Uses a valid refresh token to generate a new access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body RefreshRequest true "Refresh token request"
// @Success      200  {object} TokenPair "Access token successfully refreshed"
// @Failure      400  {string} string "Invalid request"
// @Failure      401  {string} string "Invalid or expired refresh token"
// @Failure      500  {string} string "Internal Server Error"
// @Router       /api/auth/refresh [post]
func RefreshToken(c *fiber.Ctx) error {
	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token"`
	}

	var refreshReq RefreshRequest
	if err := c.BodyParser(&refreshReq); err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid request")
	}

	claims, err := parseToken(refreshReq.RefreshToken, jwtRefreshKey)
	if err != nil {
		return c.Status(http.StatusUnauthorized).SendString("Invalid refresh token")
	}

	// Verify the refresh token exists in the database
	if !repositories.IsValidRefreshToken(claims.Subject, refreshReq.RefreshToken) {
		return c.Status(http.StatusUnauthorized).SendString("Refresh token not recognized")
	}

	accessToken, err := generateToken(claims.Subject, jwtAccessKey, 15*time.Minute)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Could not generate access token")
	}

	return c.JSON(fiber.Map{
		"access_token": accessToken,
	})
}

// GoogleLogin handles Google OAuth login
// @Summary      Google OAuth Login
// @Description  Redirects to Google OAuth for login
// @Tags         auth
// @Success      302 {string} string "Redirecting to Google OAuth"
// @Router       /api/auth/google [get]
func GoogleLogin(c *fiber.Ctx) error {
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	return c.Redirect(url)
}

// GoogleCallback handles the Google OAuth callback
// @Summary      Google OAuth Callback
// @Description  Handles the callback from Google OAuth, registers the user if necessary, and returns tokens
// @Tags         auth
// @Success      200 {object} TokenPair "Successful login"
// @Failure      500 {string} string "Internal Server Error"
// @Router       /api/auth/callback [get]
func GoogleCallback(c *fiber.Ctx) error {
	state := c.Query("state")
	if state != oauthStateString {
		return c.Status(http.StatusBadRequest).SendString("Invalid OAuth state")
	}

	code := c.Query("code")
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Could not exchange token")
	}

	client := googleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v1/userinfo?alt=json")
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Could not get user info")
	}
	defer resp.Body.Close()

	var googleUser GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Could not parse user info")
	}

	if _, err := repositories.GetUserByEmail(googleUser.Email); err != nil {
		// Register new user
		repositories.AddUser(&model.User{Email: googleUser.Email})
	}

	accessToken, err := generateToken(googleUser.Email, jwtAccessKey, 15*time.Hour)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Could not generate access token")
	}

	refreshToken, err := generateToken(googleUser.Email, jwtRefreshKey, 7*24*time.Hour)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Could not generate refresh token")
	}

	repositories.SaveRefreshToken(googleUser.Email, refreshToken)

	return c.JSON(TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// Logout clears the user's session tokens
// @Summary      Logout
// @Description  Clears the authentication tokens, effectively logging out the user
// @Tags         auth
// @Success      200 {string} string "Successfully logged out!"
// @Failure      500 {string} string "Internal Server Error"
// @Router       /api/auth/logout [get]
func Logout(c *fiber.Ctx) error {
	// Retrieve the access token from the cookies (if stored there)
	tokenStr := c.Get("Authorization")
	if !strings.HasPrefix(tokenStr, "Bearer ") {
		return fmt.Errorf("invalid Bearer token format")
	}

	// Extract the token by removing the "Bearer " prefix
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
	if tokenStr == "" {
		return c.Status(http.StatusBadRequest).SendString("No active session found")
	}

	// Parse and validate the token
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtAccessKey, nil
	})
	if err != nil || !token.Valid {
		return c.Status(http.StatusUnauthorized).SendString("Invalid session")
	}

	// Extract the email (subject) from the token claims
	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok || claims.Subject == "" {
		return c.Status(http.StatusUnauthorized).SendString("Invalid token claims")
	}

	fmt.Print(claims.Subject)
	// Invalidate the refresh token in the database
	err = repositories.SaveRefreshToken(claims.Subject, "")
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Could not clear session tokens")
	}

	// Clear the access token cookie
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour), // Expire immediately
		HTTPOnly: true,
	})

	return c.SendString("Successfully logged out!")
}

// AuthMiddleware verifies JWT and authorizes the user
// @Summary      Auth Middleware
// @Description  Verifies the user's JWT and allows access to protected routes
func AuthMiddleware(c *fiber.Ctx) error {
	// Retrieve the token from the "Authorization" header or cookies
	tokenStr := c.Get("Authorization")
	// Check if the bearer string starts with "Bearer "
	if !strings.HasPrefix(tokenStr, "Bearer ") {
		return fmt.Errorf("invalid Bearer token format")
	}

	// Extract the token by removing the "Bearer " prefix
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
	// If no token is provided, return unauthorized
	if tokenStr == "" {
		fmt.Print("empty")
		return c.Status(http.StatusUnauthorized).SendString("Unauthorized")
	}


	// Parse and validate the token
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtAccessKey, nil
	})

	if err != nil || !token.Valid {
		return c.Status(http.StatusUnauthorized).SendString("Invalid or expired token")
	}

	// Extract claims and validate
	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok || claims.Subject == "" {
		return c.Status(http.StatusUnauthorized).SendString("Invalid token claims")
	}

	// Check if user exists in the database
	user, err := repositories.GetUserByEmail(claims.Subject)
	if err != nil {
		return c.Status(http.StatusNotFound).SendString("User not found")
	}

	if user.RefreshToken == "" {
		return c.Status(http.StatusUnauthorized).SendString("Unauthorized")

	}

	// Set user information in context for later use
	c.Locals("ID", user.ID)

	// Proceed to the next handler
	return c.Next()
}
