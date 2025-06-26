package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gopro/internal/models"
)

type Poll struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Options     []string  `json:"options"`
	CreatedAt   time.Time `json:"created_at"`
	PublicURL   string    `json:"public_url"`
	CreatedBy   string    `json:"created_by"`
}

func CreatePoll(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		userIDStr := "00000000-0000-0000-0000-000000000000"
		if userID != nil {
			if s, ok := userID.(string); ok {
				userIDStr = s
			}
		}

		var req struct {
			PollName   string   `json:"poll_name"`
			Title       string   `json:"title"`
			Description string   `json:"description"`
			Options     []string `json:"options"`
		}
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request")
		}
		if len(req.Options) < 2 {
			return fiber.NewError(fiber.StatusBadRequest, "At least 2 options required")
		}

		pollID := uuid.New()
		poll := models.Poll{
			ID:          pollID,
			WebsiteID:   req.PollName,
			Title:       req.Title,
			Description: req.Description,
			CreatedBy:   userIDStr,
			CreatedAt:   time.Now(),
		}
		if err := db.Create(&poll).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to create poll")
		}

		for _, opt := range req.Options {
			option := models.PollOption{
				ID:         uuid.New(),
				PollID:     pollID,
				OptionText: opt,
			}
			db.Create(&option)
		}

		publicURL := c.BaseURL() + "/poll/" + pollID.String()
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"poll_id":     pollID,
			"public_url":  publicURL,
			"title":       poll.Title,
			"created_at":  poll.CreatedAt,
			"description": poll.Description,
			"options":     req.Options,
			"created_by":  userIDStr,
			"poll_name":   req.PollName,
		})
	}
}

type UserVoteInfo struct {
	PhoneNumber string `json:"phone_number"`
	OptionText  string `json:"option_text"`
}

func GetPollData(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		pollID := c.Params("poll_id")
		if pollID == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Missing poll_id")
		}

		var results []UserVoteInfo
		err := db.Table("votes").
			Select("users.identifier as phone_number, poll_options.option_text").
			Joins("JOIN users ON users.id = votes.user_id").
			Joins("JOIN poll_options ON poll_options.id = votes.option_id").
			Where("votes.poll_id = ?", pollID).
			Scan(&results).Error
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch votes")
		}

		return c.JSON(results)
	}
}
