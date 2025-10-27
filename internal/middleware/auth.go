package middleware

import (
	"strings"
	"take-home-test/internal/auth"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(jwtSecret string, condition string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authorization := c.Get("Authorization")

		if authorization == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		userToken := strings.Split(authorization, " ")

		if len(userToken) <= 1 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid token",
			})
		}

		if userToken[0] != "Bearer" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid token type",
			})
		}

		claims, err := auth.ExtractToken(jwtSecret, userToken[1])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		var userID int
		var email, role string

		if uid, ok := claims["user_id"].(float64); ok {
			userID = int(uid)
		} else if uid, ok := claims["id"].(float64); ok {
			userID = int(uid)
		} else {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "user_id not found in token",
			})
		}

		if em, ok := claims["email"].(string); ok {
			email = em
		} else if un, ok := claims["username"].(string); ok {
			email = un
		}

		if r, ok := claims["role"].(string); ok {
			role = r
		}

		if condition == "user" {
			if role != "admin" && role != "user" {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "access forbidden - user role required",
				})
			}
		}

		if condition == "admin" {
			if role != "admin" {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "access forbidden - admin role required",
				})
			}
		}

		c.Locals("user_id", userID)
		c.Locals("role", role)
		c.Locals("email", email)

		return c.Next()
	}
}

func AdminMiddleware(jwtSecret string) fiber.Handler {
	return AuthMiddleware(jwtSecret, "admin")
}

func UserMiddleware(jwtSecret string) fiber.Handler {
	return AuthMiddleware(jwtSecret, "user")
}
