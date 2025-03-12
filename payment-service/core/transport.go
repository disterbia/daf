package core

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/gofiber/fiber/v2"
)

func PaymentCallbackHandler(endpoint endpoint.Endpoint) fiber.Handler {
	return func(c *fiber.Ctx) error {

		var req PaymentCallbackResponse

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		response, err := endpoint(c.Context(), req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		resp := response.(BasicResponse)
		return c.Status(fiber.StatusOK).JSON(resp)
	}
}
func RefundHandler(endpoint endpoint.Endpoint) fiber.Handler {
	return func(c *fiber.Ctx) error {

		response, err := endpoint(c.Context(), nil)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		resp := response.(BasicResponse)
		return c.Status(fiber.StatusOK).JSON(resp)
	}
}
