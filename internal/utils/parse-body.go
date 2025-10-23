package utils

import (
	"fmt"
	"nexa/internal/model"

	"github.com/gofiber/fiber/v2"
)

func ParseBody(c *fiber.Ctx, user *model.User) (string, error) {
	err := c.BodyParser(user)

	if err != nil {
		return "INVALID_BODY_FORMAT", fmt.Errorf("user decoding failed")
	}

	return "", nil
}
