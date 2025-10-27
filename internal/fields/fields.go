package fields

import (
	"database/sql"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func CreateFieldHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Name         string `json:"name"`
			PricePerHour int    `json:"price_per_hour"`
			Location     string `json:"location"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if req.Name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Field name is required",
			})
		}
		if req.PricePerHour <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Price per hour must be greater than 0",
			})
		}
		if req.Location == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Location is required",
			})
		}

		var fieldID int
		err := db.QueryRow(
			"INSERT INTO fields (name, price_per_hour, location) VALUES ($1, $2, $3) RETURNING field_id",
			req.Name, req.PricePerHour, req.Location,
		).Scan(&fieldID)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create field",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "Field created successfully",
			"field": fiber.Map{
				"field_id":       fieldID,
				"name":           req.Name,
				"price_per_hour": req.PricePerHour,
				"location":       req.Location,
			},
		})
	}
}

func GetFieldsHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rows, err := db.Query(`
			SELECT field_id, name, price_per_hour, location 
			FROM fields
		`)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch fields",
			})
		}
		defer rows.Close()

		var fields []fiber.Map
		for rows.Next() {
			var field struct {
				FieldID      int
				Name         string
				PricePerHour int
				Location     string
			}
			err := rows.Scan(
				&field.FieldID,
				&field.Name,
				&field.PricePerHour,
				&field.Location,
			)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to read field data",
				})
			}
			fields = append(fields, fiber.Map{
				"field_id":       field.FieldID,
				"name":           field.Name,
				"price_per_hour": field.PricePerHour,
				"location":       field.Location,
			})
		}

		return c.JSON(fiber.Map{
			"message": "Fields retrieved successfully",
			"fields":  fields,
			"count":   len(fields),
		})
	}
}

func GetFieldHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid field ID",
			})
		}

		var field struct {
			FieldID      int
			Name         string
			PricePerHour int
			Location     string
		}

		err = db.QueryRow(`
			SELECT field_id, name, price_per_hour, location 
			FROM fields 
			WHERE field_id = $1
		`, id).Scan(
			&field.FieldID,
			&field.Name,
			&field.PricePerHour,
			&field.Location,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Field not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch field",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Field retrieved successfully",
			"field": fiber.Map{
				"field_id":       field.FieldID,
				"name":           field.Name,
				"price_per_hour": field.PricePerHour,
				"location":       field.Location,
			},
		})
	}
}

func UpdateFieldHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid field ID",
			})
		}

		var req struct {
			Name         string `json:"name"`
			PricePerHour int    `json:"price_per_hour"`
			Location     string `json:"location"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if req.Name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Field name is required",
			})
		}
		if req.PricePerHour <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Price per hour must be greater than 0",
			})
		}
		if req.Location == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Location is required",
			})
		}

		result, err := db.Exec(`
			UPDATE fields 
			SET name = $1, price_per_hour = $2, location = $3 
			WHERE field_id = $4
		`, req.Name, req.PricePerHour, req.Location, id)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update field",
			})
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Field not found",
			})
		}

		var field struct {
			FieldID      int
			Name         string
			PricePerHour int
			Location     string
		}

		err = db.QueryRow(`
			SELECT field_id, name, price_per_hour, location 
			FROM fields 
			WHERE field_id = $1
		`, id).Scan(
			&field.FieldID,
			&field.Name,
			&field.PricePerHour,
			&field.Location,
		)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch updated field",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Field updated successfully",
			"field": fiber.Map{
				"field_id":       field.FieldID,
				"name":           field.Name,
				"price_per_hour": field.PricePerHour,
				"location":       field.Location,
			},
		})
	}
}

func DeleteFieldHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid field ID",
			})
		}

		result, err := db.Exec("DELETE FROM fields WHERE field_id = $1", id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete field",
			})
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Field not found",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Field deleted successfully",
		})
	}
}
