package users

import (
	"database/sql"
	"take-home-test/internal/auth"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/slog"
)

func RegisterAdmin(db *sql.DB, jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&req); err != nil {
			return errorResponse(c, "Invalid request body", 400)
		}

		if len(req.Username) < 3 {
			return errorResponse(c, "Username must be at least 3 characters", 400)
		}
		if len(req.Password) < 6 {
			return errorResponse(c, "Password must be at least 6 characters", 400)
		}

		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", req.Email).Scan(&count)
		if err != nil {
			slog.Error("Database error", "error", err)
			return errorResponse(c, "Internal server error", 500)
		}
		if count > 0 {
			return errorResponse(c, "Email already registered", 400)
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return errorResponse(c, "Failed to process password", 500)
		}

		var userID int
		err = db.QueryRow(
			"INSERT INTO users (username, email, password, role) VALUES ($1, $2, $3, $4) RETURNING user_id",
			req.Username, req.Email, string(hashedPassword), "admin",
		).Scan(&userID)

		if err != nil {
			slog.Error("Failed to create user", "error", err)
			return errorResponse(c, "Failed to create user", 500)
		}

		token, err := auth.GenerateJWT(jwtSecret, userID, req.Email, "admin")
		if err != nil {
			return errorResponse(c, "Failed to generate token", 500)
		}

		return c.Status(201).JSON(fiber.Map{
			"message": "User registered successfully",
			"user": fiber.Map{
				"email": req.Email,
				"token": token,
			},
		})
	}
}

func RegisterUser(db *sql.DB, jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&req); err != nil {
			return errorResponse(c, "Invalid request body", 400)
		}

		if len(req.Username) < 3 {
			return errorResponse(c, "Username must be at least 3 characters", 400)
		}
		if len(req.Password) < 6 {
			return errorResponse(c, "Password must be at least 6 characters", 400)
		}

		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", req.Email).Scan(&count)
		if err != nil {
			slog.Error("Database error", "error", err)
			return errorResponse(c, "Internal server error", 500)
		}
		if count > 0 {
			return errorResponse(c, "Email already registered", 400)
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return errorResponse(c, "Failed to process password", 500)
		}

		var userID int
		err = db.QueryRow(
			"INSERT INTO users (username, email, password, role) VALUES ($1, $2, $3, $4) RETURNING user_id",
			req.Username, req.Email, string(hashedPassword), "user",
		).Scan(&userID)

		if err != nil {
			slog.Error("Failed to create user", "error", err)
			return errorResponse(c, "Failed to create user", 500)
		}

		token, err := auth.GenerateJWT(jwtSecret, userID, req.Email, "user")
		if err != nil {
			return errorResponse(c, "Failed to generate token", 500)
		}

		return c.Status(201).JSON(fiber.Map{
			"message": "User registered successfully",
			"user": fiber.Map{
				"email": req.Email,
				"token": token,
			},
		})
	}
}

func Login(db *sql.DB, jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&req); err != nil {
			return errorResponse(c, "Invalid request body", 400)
		}

		var user struct {
			UserID   int
			Username string
			Email    string
			Password string
			Role     string
		}

		err := db.QueryRow(
			"SELECT user_id, username, email, password, role FROM users WHERE email = $1",
			req.Email,
		).Scan(&user.UserID, &user.Username, &user.Email, &user.Password, &user.Role)

		if err != nil {
			return errorResponse(c, "Invalid email or password", 401)
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
		if err != nil {
			return errorResponse(c, "Invalid email or password", 401)
		}

		token, err := auth.GenerateJWT(jwtSecret, user.UserID, user.Email, user.Role)
		if err != nil {
			return errorResponse(c, "Failed to generate token", 500)
		}

		return c.JSON(fiber.Map{
			"message": "Login successful",
			"user": fiber.Map{
				"email": user.Email,
				"token": token,
			},
		})
	}
}

func errorResponse(c *fiber.Ctx, message string, status int) error {
	return c.Status(status).JSON(fiber.Map{"error": message})
}
