package handlers

import (
    "time"
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "gorm.io/gorm"
    "github.com/gopro/internal/models"
)

type CastVoteRequest struct {
    OptionID string `json:"option_id"`
    UserID   string `json:"user_id"`
}

func CastVote(db *gorm.DB) fiber.Handler {
    return func(c *fiber.Ctx) error {
       userID := c.Locals("user_id")
		userIDStr := "00000000-0000-0000-0000-000000000000"
		if userID != nil {
			if s, ok := userID.(string); ok {
				userIDStr = s
			}
		}
        pollIDStr := c.Params("poll_id")
        if pollIDStr == "" {
            return fiber.NewError(fiber.StatusBadRequest, "Missing poll_id")
        }
        var req CastVoteRequest
        if err := c.BodyParser(&req); err != nil {
            return fiber.NewError(fiber.StatusBadRequest, "Invalid request")
        }
        pollID, err := uuid.Parse(pollIDStr)
        if err != nil {
            return fiber.NewError(fiber.StatusBadRequest, "Invalid poll_id")
        }
        optionID, err := uuid.Parse(req.OptionID)
        if err != nil {
            return fiber.NewError(fiber.StatusBadRequest, "Invalid option_id")
        }
        // userUUID, err := uuid.Parse(userID.(string))
        // if err != nil {
        //     return fiber.ErrUnauthorized
        // }

        // Ensure the option belongs to the poll
        var option models.PollOption
        if err := db.Where("id = ? AND poll_id = ?", optionID, pollID).First(&option).Error; err != nil {
            return fiber.NewError(fiber.StatusBadRequest, "Option does not belong to poll")
        }

        // Insert vote, handle unique constraint violation for (poll_id, user_id)
        var userUUID uuid.UUID
        if userIDStr == "anonymous" {
            userUUID = uuid.Nil
        } else {
            userUUID, err = uuid.Parse(userIDStr)
            if err != nil {
                return fiber.NewError(fiber.StatusBadRequest, "Invalid user_id")
            }
        }
        vote := models.Vote{
            ID:       uuid.New(),
            PollID:   pollID,
            OptionID: optionID,
            UserID:   userUUID,
            VotedAt:  time.Now(),
        }
        if err := db.Create(&vote).Error; err != nil {
            if db.Error != nil && db.Error.Error() == "UNIQUE constraint failed: votes.poll_id, votes.user_id" {
                return fiber.NewError(fiber.StatusConflict, "User has already voted for this poll")
            }
            return fiber.NewError(fiber.StatusInternalServerError, "Failed to cast vote")
        }

        return c.JSON(fiber.Map{"message": "Vote cast successfully", "poll_id": pollIDStr, "option_id": req.OptionID, "user_id": userIDStr,})
    }
}