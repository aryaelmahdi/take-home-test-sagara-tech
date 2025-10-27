package main

import (
	"fmt"
	"log"
	"take-home-test/internal/bookings"
	"take-home-test/internal/configs"
	"take-home-test/internal/fields"
	"take-home-test/internal/middleware"
	"take-home-test/internal/payments"
	"take-home-test/internal/postgres"
	"take-home-test/internal/users"

	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg, err := configs.InitConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := postgres.Open(cfg)
	if err != nil {
		log.Fatalf("sql connection error: %v", err)
	}
	defer db.Close()

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Take Home Test Sagara")
	})

	//Auth
	app.Post("/auth/register", users.RegisterUser(db, cfg.AppConfig.JWTSecret))
	app.Get("/auth/login", users.Login(db, cfg.AppConfig.JWTSecret))
	app.Post("/admin/auth/register", users.RegisterAdmin(db, cfg.AppConfig.JWTSecret))

	//Fields
	app.Get("/fields", fields.GetFieldsHandler(db))
	app.Get("/fields/:id", fields.GetFieldHandler(db))
	app.Post("/fields", middleware.AdminMiddleware(cfg.AppConfig.JWTSecret), fields.CreateFieldHandler(db))
	app.Put("/fields/:id", middleware.AdminMiddleware(cfg.AppConfig.JWTSecret), fields.UpdateFieldHandler(db))
	app.Delete("/fields/:id", middleware.AdminMiddleware(cfg.AppConfig.JWTSecret), fields.DeleteFieldHandler(db))

	//Booking
	app.Post("/bookings", middleware.UserMiddleware(cfg.AppConfig.JWTSecret), bookings.CreateBookingHandler(db))

	//Payment
	app.Post("/payments", middleware.UserMiddleware(cfg.AppConfig.JWTSecret), payments.UpdatePayment(db))

	port := fmt.Sprintf(":%d", cfg.AppConfig.Port)
	log.Printf("Server running on port %s", port)
	log.Fatal(app.Listen(port))
}
