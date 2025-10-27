package payments

import (
	"database/sql"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func UpdatePayment(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			BookingID int `json:"booking_id"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body: " + err.Error(),
			})
		}

		if req.BookingID <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid booking ID",
			})
		}

		var currentStatus string
		var totalPrice int
		err := db.QueryRow(`
			SELECT status, total_price FROM bookings WHERE booking_id = $1
		`, req.BookingID).Scan(&currentStatus, &totalPrice)

		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Booking not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check booking: " + err.Error(),
			})
		}

		if currentStatus != "confirmed" && currentStatus != "pending" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Cannot update payment for booking with status: %s. Only 'confirmed' or 'pending' bookings can be paid.", currentStatus),
			})
		}

		result, err := db.Exec(`
			UPDATE bookings SET status = $1 WHERE booking_id = $2
		`, "paid", req.BookingID)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update payment: " + err.Error(),
			})
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Booking not found",
			})
		}

		var booking struct {
			BookingID   int
			UserID      int
			FieldID     int
			FieldName   string
			BookingDate string
			StartTime   string
			EndTime     string
			TotalPrice  int
			Status      string
			CreatedAt   string
		}

		err = db.QueryRow(`
			SELECT 
				b.booking_id, b.user_id, b.field_id, f.name as field_name,
				b.booking_date, b.start_time, b.end_time, 
				b.total_price, b.status, b.created_at
			FROM bookings b
			JOIN fields f ON b.field_id = f.field_id
			WHERE b.booking_id = $1
		`, req.BookingID).Scan(
			&booking.BookingID,
			&booking.UserID,
			&booking.FieldID,
			&booking.FieldName,
			&booking.BookingDate,
			&booking.StartTime,
			&booking.EndTime,
			&booking.TotalPrice,
			&booking.Status,
			&booking.CreatedAt,
		)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch updated booking: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Payment completed successfully",
			"payment": fiber.Map{
				"booking_id":   booking.BookingID,
				"user_id":      booking.UserID,
				"field_id":     booking.FieldID,
				"field_name":   booking.FieldName,
				"booking_date": booking.BookingDate,
				"start_time":   booking.StartTime,
				"end_time":     booking.EndTime,
				"total_price":  booking.TotalPrice,
				"status":       booking.Status,
			},
		})
	}
}
