package handlers

import (
	"designmypdf/api/handlers/presenter"
	"designmypdf/pkg/entities"
	"designmypdf/pkg/logs"
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func CreateLog(logService logs.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var log entities.Log
		if err := c.BodyParser(&log); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		if err := logService.CreateLog(&log); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.Status(fiber.StatusCreated).JSON(presenter.LogSuccessResponse(&log))
	}
}

func GetLogStats(logService logs.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDValue := c.Locals("userID")
		userIDFloat, ok := userIDValue.(float64)
		if !ok {
			err := errors.New("invalid user ID type")
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.LogErrorResponse(err))
		}
		userID := uint(userIDFloat)

		period := c.Query("period")
		if period == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "period is required"})
		}

		stats, err := logService.GetLogStats(userID, period)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.LogErrorResponse(err))
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"stats": stats})
	}
}

func GetLogsByUserID(logService logs.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDValue := c.Locals("userID")
		userIDFloat, ok := userIDValue.(float64)
		if !ok {
			err := errors.New("invalid user ID type")
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.LogErrorResponse(err))
		}
		userID := uint(userIDFloat)

		logs, err := logService.GetLogsByUserID(userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.LogErrorResponse(err))
		}

		return c.Status(fiber.StatusOK).JSON(presenter.LogsSuccessResponse(logs))
	}
}
