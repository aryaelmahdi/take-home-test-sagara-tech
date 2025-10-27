package bookings

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

func CreateBookingHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(int)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User not authenticated",
			})
		}

		var req struct {
			FieldID     int    `json:"field_id"`
			BookingDate string `json:"booking_date"`
			StartTime   string `json:"start_time"`
			EndTime     string `json:"end_time"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body: " + err.Error(),
			})
		}

		if req.FieldID <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid field ID",
			})
		}

		bookingDate, err := time.Parse("2006-01-02", req.BookingDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid booking date format. Use YYYY-MM-DD",
			})
		}

		startTime, err := time.Parse("15:04", req.StartTime)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid start time format. Use HH:MM",
			})
		}

		endTime, err := time.Parse("15:04", req.EndTime)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid end time format. Use HH:MM",
			})
		}

		if endTime.Before(startTime) || endTime.Equal(startTime) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "End time must be after start time",
			})
		}

		now := time.Now()
		bookingDateTime := time.Date(
			bookingDate.Year(), bookingDate.Month(), bookingDate.Day(),
			startTime.Hour(), startTime.Minute(), 0, 0, time.Local,
		)
		if bookingDateTime.Before(now) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot book in the past",
			})
		}

		var pricePerHour int
		err = db.QueryRow("SELECT price_per_hour FROM fields WHERE field_id = $1", req.FieldID).Scan(&pricePerHour)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Field not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check field: " + err.Error(),
			})
		}

		isAvailable, err := checkTimeAvailability(db, req.FieldID, req.BookingDate, req.StartTime, req.EndTime)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check availability: " + err.Error(),
			})
		}
		if !isAvailable {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Field is already booked at the selected time",
			})
		}

		duration := endTime.Sub(startTime).Hours()
		totalPrice := int(duration * float64(pricePerHour))

		var bookingID int
		err = db.QueryRow(`
			INSERT INTO bookings (user_id, field_id, booking_date, start_time, end_time, total_price, status)
			VALUES ($1, $2, $3, $4, $5, $6, 'pending')
			RETURNING booking_id
		`, userID, req.FieldID, req.BookingDate, req.StartTime, req.EndTime, totalPrice).Scan(&bookingID)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create booking: " + err.Error(),
			})
		}

		var fieldName, fieldLocation string
		db.QueryRow("SELECT name, location FROM fields WHERE field_id = $1", req.FieldID).Scan(&fieldName, &fieldLocation)

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "Booking created successfully",
			"booking": fiber.Map{
				"booking_id":   bookingID,
				"field_id":     req.FieldID,
				"field_name":   fieldName,
				"location":     fieldLocation,
				"booking_date": req.BookingDate,
				"start_time":   req.StartTime,
				"end_time":     req.EndTime,
				"duration":     fmt.Sprintf("%.1f hours", duration),
				"total_price":  totalPrice,
				"status":       "pending",
			},
		})
	}
}

func checkTimeAvailability(db *sql.DB, fieldID int, bookingDate, startTime, endTime string) (bool, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) 
		FROM bookings 
		WHERE field_id = $1 
		AND booking_date = $2 
		AND (status = 'pending' or status = 'paid')
		AND (start_time, end_time) OVERLAPS ($3::time, $4::time)
	`, fieldID, bookingDate, startTime, endTime).Scan(&count)

	if err != nil {
		return false, err
	}

	return count == 0, nil
}
